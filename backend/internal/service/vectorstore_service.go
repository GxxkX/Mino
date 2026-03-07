package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/mino/backend/internal/pkg/vectordb"
	"github.com/sirupsen/logrus"
)

// VectorStoreService orchestrates embedding generation and Milvus vector storage/retrieval.
// It bridges the EmbeddingProvider (LangchainGo) with the Milvus client.
type VectorStoreService struct {
	milvus   *vectordb.Client
	embedder EmbeddingProvider
	logger   *logrus.Logger
}

// NewVectorStoreService creates a new VectorStoreService.
// Returns nil (not an error) if either milvus or embedder is nil — this makes
// the service optional so the app can still run without vector search.
func NewVectorStoreService(milvus *vectordb.Client, embedder EmbeddingProvider, logger *logrus.Logger) *VectorStoreService {
	if milvus == nil || embedder == nil {
		if logger != nil {
			logger.Warn("VectorStoreService disabled: milvus or embedder not available")
		}
		return nil
	}
	return &VectorStoreService{
		milvus:   milvus,
		embedder: embedder,
		logger:   logger,
	}
}

// IndexConversation generates an embedding for a conversation's text and stores it in Milvus.
// The text is typically the transcript or summary of the conversation.
func (s *VectorStoreService) IndexConversation(ctx context.Context, userID, conversationID, text string) error {
	if text == "" {
		return nil
	}

	// Truncate very long texts to avoid embedding API limits.
	text = truncateForEmbedding(text, 6000)

	vec, err := s.embedder.EmbedQuery(ctx, text)
	if err != nil {
		return fmt.Errorf("embed conversation: %w", err)
	}

	return s.milvus.Insert(ctx, s.milvus.ConversationsCollection(), userID, conversationID, text, vec)
}

// IndexMemories generates embeddings for a batch of memories and stores them in Milvus.
func (s *VectorStoreService) IndexMemories(ctx context.Context, userID string, memoryIDs, contents []string) error {
	if len(memoryIDs) == 0 {
		return nil
	}

	vectors, err := s.embedder.EmbedDocuments(ctx, contents)
	if err != nil {
		return fmt.Errorf("embed memories: %w", err)
	}

	userIDs := make([]string, len(memoryIDs))
	for i := range userIDs {
		userIDs[i] = userID
	}

	return s.milvus.InsertBatch(ctx, s.milvus.MemoriesCollection(), userIDs, memoryIDs, contents, vectors)
}

// IndexMemory generates an embedding for a single memory and stores it in Milvus.
func (s *VectorStoreService) IndexMemory(ctx context.Context, userID, memoryID, content string) error {
	if content == "" {
		return nil
	}

	vec, err := s.embedder.EmbedQuery(ctx, content)
	if err != nil {
		return fmt.Errorf("embed memory: %w", err)
	}

	return s.milvus.Insert(ctx, s.milvus.MemoriesCollection(), userID, memoryID, content, vec)
}

// VectorSearchResult holds a semantic search result with its source metadata.
type VectorSearchResult struct {
	SourceID string  // conversation_id or memory_id
	Text     string  // the original text stored alongside the vector
	Score    float32 // cosine similarity score
}

// SearchConversations performs semantic search over conversation vectors.
func (s *VectorStoreService) SearchConversations(ctx context.Context, userID, query string, topK int) ([]VectorSearchResult, error) {
	return s.search(ctx, s.milvus.ConversationsCollection(), userID, query, topK)
}

// SearchMemories performs semantic search over memory vectors.
func (s *VectorStoreService) SearchMemories(ctx context.Context, userID, query string, topK int) ([]VectorSearchResult, error) {
	return s.search(ctx, s.milvus.MemoriesCollection(), userID, query, topK)
}

// SearchAll performs semantic search across both conversations and memories,
// merging results by score.
func (s *VectorStoreService) SearchAll(ctx context.Context, userID, query string, topK int) ([]VectorSearchResult, error) {
	convResults, err := s.SearchConversations(ctx, userID, query, topK)
	if err != nil {
		s.logger.Warnf("vector search conversations: %v", err)
		convResults = nil
	}

	memResults, err := s.SearchMemories(ctx, userID, query, topK)
	if err != nil {
		s.logger.Warnf("vector search memories: %v", err)
		memResults = nil
	}

	// Merge and sort by score (descending)
	all := append(convResults, memResults...)
	// Simple insertion sort — topK is small
	for i := 1; i < len(all); i++ {
		for j := i; j > 0 && all[j].Score > all[j-1].Score; j-- {
			all[j], all[j-1] = all[j-1], all[j]
		}
	}

	if len(all) > topK {
		all = all[:topK]
	}

	return all, nil
}

// DeleteConversation removes a conversation's vectors from Milvus.
func (s *VectorStoreService) DeleteConversation(ctx context.Context, conversationID string) error {
	return s.milvus.DeleteBySourceID(ctx, s.milvus.ConversationsCollection(), conversationID)
}

// DeleteMemory removes a memory's vectors from Milvus.
func (s *VectorStoreService) DeleteMemory(ctx context.Context, memoryID string) error {
	return s.milvus.DeleteBySourceID(ctx, s.milvus.MemoriesCollection(), memoryID)
}

// search is the internal helper that embeds the query and searches a collection.
func (s *VectorStoreService) search(ctx context.Context, collection, userID, query string, topK int) ([]VectorSearchResult, error) {
	queryVec, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	hits, err := s.milvus.Search(ctx, collection, userID, queryVec, topK)
	if err != nil {
		return nil, err
	}

	results := make([]VectorSearchResult, len(hits))
	for i, h := range hits {
		results[i] = VectorSearchResult{
			SourceID: h.SourceID,
			Text:     h.Text,
			Score:    h.Score,
		}
	}
	return results, nil
}

// truncateForEmbedding truncates text to roughly maxChars characters,
// breaking at the last space before the limit.
func truncateForEmbedding(text string, maxChars int) string {
	runes := []rune(text)
	if len(runes) <= maxChars {
		return text
	}
	truncated := string(runes[:maxChars])
	if idx := strings.LastIndex(truncated, " "); idx > maxChars/2 {
		truncated = truncated[:idx]
	}
	return truncated
}
