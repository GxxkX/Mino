package service

import (
	"context"
	"fmt"
	"time"

	"github.com/mino/backend/internal/model"
	"github.com/mino/backend/internal/pkg/search"
	"github.com/mino/backend/internal/repository"
	"github.com/typesense/typesense-go/v2/typesense/api"
)

// SearchResult represents a unified search result item.
type SearchResult struct {
	Type       string      `json:"type"` // "conversation" or "memory"
	ID         string      `json:"id"`
	Title      string      `json:"title"`
	Snippet    string      `json:"snippet"`
	Category   string      `json:"category,omitempty"`
	Importance int         `json:"importance,omitempty"`
	CreatedAt  string      `json:"createdAt"`
	Highlights interface{} `json:"highlights,omitempty"`
}

// SearchResponse wraps the full search response.
type SearchResponse struct {
	Conversations []SearchResult `json:"conversations"`
	Memories      []SearchResult `json:"memories"`
	TotalFound    int            `json:"totalFound"`
}

// SearchService handles Typesense indexing and searching.
type SearchService struct {
	client   *search.Client
	convRepo *repository.ConversationRepository
	memRepo  *repository.MemoryRepository
}

func NewSearchService(client *search.Client, convRepo *repository.ConversationRepository, memRepo *repository.MemoryRepository) *SearchService {
	return &SearchService{client: client, convRepo: convRepo, memRepo: memRepo}
}

// --- Index sync methods ---

// SyncConversation upserts a conversation document into Typesense.
func (s *SearchService) SyncConversation(ctx context.Context, conv *model.Conversation) error {
	title := ""
	if conv.Title != nil {
		title = *conv.Title
	}
	summary := ""
	if conv.Summary != nil {
		summary = *conv.Summary
	}

	doc := map[string]interface{}{
		"id":         conv.ID,
		"user_id":    conv.UserID,
		"title":      title,
		"summary":    summary,
		"transcript": conv.Transcript,
		"language":   conv.Language,
		"status":     conv.Status,
		"created_at": conv.CreatedAt.Unix(),
	}

	_, err := s.client.Typesense().Collection(search.CollectionConversations).Documents().Upsert(ctx, doc)
	return err
}

// SyncMemory upserts a memory document into Typesense.
func (s *SearchService) SyncMemory(ctx context.Context, mem *model.Memory) error {
	convID := ""
	if mem.ConversationID != nil {
		convID = *mem.ConversationID
	}
	category := ""
	if mem.Category != nil {
		category = *mem.Category
	}

	doc := map[string]interface{}{
		"id":              mem.ID,
		"user_id":         mem.UserID,
		"conversation_id": convID,
		"content":         mem.Content,
		"category":        category,
		"importance":      int32(mem.Importance),
		"created_at":      mem.CreatedAt.Unix(),
	}

	_, err := s.client.Typesense().Collection(search.CollectionMemories).Documents().Upsert(ctx, doc)
	return err
}

// DeleteConversation removes a conversation from the index.
func (s *SearchService) DeleteConversation(ctx context.Context, id string) error {
	_, err := s.client.Typesense().Collection(search.CollectionConversations).Document(id).Delete(ctx)
	return err
}

// DeleteMemory removes a memory from the index.
func (s *SearchService) DeleteMemory(ctx context.Context, id string) error {
	_, err := s.client.Typesense().Collection(search.CollectionMemories).Document(id).Delete(ctx)
	return err
}

// --- Search ---

func ptr[T any](v T) *T { return &v }

// Search performs a multi-collection search filtered by user ID.
func (s *SearchService) Search(ctx context.Context, userID, query string, limit int) (*SearchResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	filterBy := fmt.Sprintf("user_id:=%s", userID)

	searchParams := api.MultiSearchSearchesParameter{
		Searches: []api.MultiSearchCollectionParameters{
			{
				Collection:     search.CollectionConversations,
				Q:              &query,
				QueryBy:        ptr("title,summary,transcript"),
				QueryByWeights: ptr("3,2,1"),
				FilterBy:       &filterBy,
				PerPage:        &limit,
				SortBy:         ptr("created_at:desc"),
			},
			{
				Collection: search.CollectionMemories,
				Q:          &query,
				QueryBy:    ptr("content"),
				FilterBy:   &filterBy,
				PerPage:    &limit,
				SortBy:     ptr("importance:desc,created_at:desc"),
			},
		},
	}

	result, err := s.client.Typesense().MultiSearch.Perform(ctx, &api.MultiSearchParams{}, searchParams)
	if err != nil {
		return nil, fmt.Errorf("typesense multi-search: %w", err)
	}

	resp := &SearchResponse{}

	// Parse conversation results
	if len(result.Results) > 0 {
		convResult := result.Results[0]
		if convResult.Found != nil {
			resp.TotalFound += *convResult.Found
		}
		if convResult.Hits != nil {
			for _, hit := range *convResult.Hits {
				if hit.Document == nil {
					continue
				}
				doc := *hit.Document
				resp.Conversations = append(resp.Conversations, SearchResult{
					Type:       "conversation",
					ID:         strVal(doc["id"]),
					Title:      strVal(doc["title"]),
					Snippet:    searchTruncate(strVal(doc["summary"]), 120),
					CreatedAt:  timeFromUnix(doc["created_at"]),
					Highlights: hit.Highlight,
				})
			}
		}
	}

	// Parse memory results
	if len(result.Results) > 1 {
		memResult := result.Results[1]
		if memResult.Found != nil {
			resp.TotalFound += *memResult.Found
		}
		if memResult.Hits != nil {
			for _, hit := range *memResult.Hits {
				if hit.Document == nil {
					continue
				}
				doc := *hit.Document
				resp.Memories = append(resp.Memories, SearchResult{
					Type:       "memory",
					ID:         strVal(doc["id"]),
					Title:      searchTruncate(strVal(doc["content"]), 60),
					Snippet:    searchTruncate(strVal(doc["content"]), 120),
					Category:   strVal(doc["category"]),
					Importance: intVal(doc["importance"]),
					CreatedAt:  timeFromUnix(doc["created_at"]),
					Highlights: hit.Highlight,
				})
			}
		}
	}

	return resp, nil
}

// SyncAllFromDB re-indexes all conversations and memories for a given user.
func (s *SearchService) SyncAllFromDB(ctx context.Context, userID string) (int, error) {
	count := 0

	convs, _, err := s.convRepo.List(userID, 10000, 0)
	if err != nil {
		return 0, fmt.Errorf("list conversations: %w", err)
	}
	for _, conv := range convs {
		if err := s.SyncConversation(ctx, conv); err != nil {
			return count, fmt.Errorf("sync conversation %s: %w", conv.ID, err)
		}
		count++
	}

	mems, _, err := s.memRepo.List(userID, 10000, 0)
	if err != nil {
		return count, fmt.Errorf("list memories: %w", err)
	}
	for _, mem := range mems {
		if err := s.SyncMemory(ctx, mem); err != nil {
			return count, fmt.Errorf("sync memory %s: %w", mem.ID, err)
		}
		count++
	}

	return count, nil
}

// --- helpers ---

func strVal(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func intVal(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int32:
		return int(n)
	}
	return 0
}

func timeFromUnix(v interface{}) string {
	var ts int64
	switch n := v.(type) {
	case float64:
		ts = int64(n)
	case int64:
		ts = n
	default:
		return ""
	}
	return time.Unix(ts, 0).UTC().Format(time.RFC3339)
}

func searchTruncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
