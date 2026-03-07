package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/mino/backend/internal/repository"
)

type ChatSource struct {
	ConversationID string `json:"conversationId"`
	Title          string `json:"title"`
	Excerpt        string `json:"excerpt"`
}

type ChatService struct {
	chatRepo    *repository.ChatRepository
	convRepo    *repository.ConversationRepository
	llmService  LLMProvider
	vectorStore *VectorStoreService // nil if vector search is unavailable
	logger      *logrus.Logger
}

func NewChatService(
	chatRepo *repository.ChatRepository,
	convRepo *repository.ConversationRepository,
	llmService LLMProvider,
	vectorStore *VectorStoreService,
	logger *logrus.Logger,
) *ChatService {
	return &ChatService{
		chatRepo:    chatRepo,
		convRepo:    convRepo,
		llmService:  llmService,
		vectorStore: vectorStore,
		logger:      logger,
	}
}

type ChatResponse struct {
	ID        string       `json:"id"`
	SessionID string       `json:"sessionId"`
	Role      string       `json:"role"`
	Content   string       `json:"content"`
	Sources   []ChatSource `json:"sources,omitempty"`
	CreatedAt interface{}  `json:"createdAt"`
}

// --------------- Session operations ---------------

func (s *ChatService) CreateSession(userID, title string) (*repository.ChatSession, error) {
	if title == "" {
		title = "新对话"
	}
	sess := &repository.ChatSession{UserID: userID, Title: title}
	if err := s.chatRepo.CreateSession(sess); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	return sess, nil
}

func (s *ChatService) ListSessions(userID string) ([]*repository.ChatSession, error) {
	return s.chatRepo.ListSessions(userID)
}

func (s *ChatService) UpdateSessionTitle(sessionID, userID, title string) error {
	return s.chatRepo.UpdateSessionTitle(sessionID, userID, title)
}

func (s *ChatService) DeleteSession(sessionID, userID string) error {
	return s.chatRepo.DeleteSession(sessionID, userID)
}

// --------------- Message operations ---------------

func (s *ChatService) GetMessages(sessionID, userID string) ([]*repository.ChatMessage, error) {
	return s.chatRepo.ListMessages(sessionID, userID)
}

// GetLLMService exposes the underlying LLMProvider for streaming use in handlers.
func (s *ChatService) GetLLMService() LLMProvider {
	return s.llmService
}

func (s *ChatService) SendMessage(ctx context.Context, sessionID, userID, userMessage string) (*ChatResponse, error) {
	// Verify session ownership
	sess, err := s.chatRepo.GetSession(sessionID, userID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Save user message
	userMsg := &repository.ChatMessage{
		SessionID: sessionID,
		UserID:    userID,
		Role:      "user",
		Content:   userMessage,
	}
	if err := s.chatRepo.CreateMessage(userMsg); err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// Retrieve relevant context via semantic vector search (preferred) or keyword fallback.
	var contextParts []string
	var sources []ChatSource

	if s.vectorStore != nil {
		// Semantic search via Milvus
		results, err := s.vectorStore.SearchAll(ctx, userID, userMessage, 5)
		if err != nil {
			// Non-fatal: fall through to keyword search
			s.logger.Warnf("vector search failed, falling back to keyword: %v", err)
			results = nil
		}
		for _, r := range results {
			excerpt := r.Text
			if len(excerpt) > 300 {
				excerpt = excerpt[:300] + "..."
			}
			contextParts = append(contextParts, fmt.Sprintf("[相关内容 (相似度 %.2f)]: %s", r.Score, excerpt))
			sources = append(sources, ChatSource{
				ConversationID: r.SourceID,
				Title:          "",
				Excerpt:        excerpt[:min(100, len(excerpt))],
			})
		}
	}

	// Fallback to keyword search if vector search returned nothing
	if len(contextParts) == 0 {
		convs, err := s.convRepo.Search(userID, userMessage, 5)
		if err != nil {
			convs = nil // non-fatal
		}
		for _, c := range convs {
			title := "Untitled"
			if c.Title != nil {
				title = *c.Title
			}
			excerpt := c.Transcript
			if len(excerpt) > 300 {
				excerpt = excerpt[:300] + "..."
			}
			contextParts = append(contextParts, fmt.Sprintf("[%s]: %s", title, excerpt))
			sources = append(sources, ChatSource{
				ConversationID: c.ID,
				Title:          title,
				Excerpt:        excerpt[:min(100, len(excerpt))],
			})
		}
	}

	retrievedContext := ""
	if len(contextParts) > 0 {
		retrievedContext = strings.Join(contextParts, "\n\n")
	}

	reply, err := s.llmService.ChatAgent(ctx, userMessage, retrievedContext)
	if err != nil {
		reply = "I'm sorry, I'm having trouble connecting to the AI service right now. Please try again later. reason: " + err.Error()
		sources = nil
	}

	// Serialize sources
	var sourcesJSON json.RawMessage
	if len(sources) > 0 {
		sourcesJSON, _ = json.Marshal(sources)
	}

	// Save assistant message
	assistantMsg := &repository.ChatMessage{
		SessionID: sessionID,
		UserID:    userID,
		Role:      "assistant",
		Content:   reply,
		Sources:   sourcesJSON,
	}
	if err := s.chatRepo.CreateMessage(assistantMsg); err != nil {
		return nil, fmt.Errorf("failed to save assistant message: %w", err)
	}

	// Bump session updated_at
	s.chatRepo.TouchSession(sessionID)

	// Auto-summarize session title on first exchange
	if sess.Title == "新对话" {
		go func() {
			title, err := s.llmService.SummarizeTitle(context.Background(), userMessage, reply)
			if err != nil || title == "" {
				return
			}
			_ = s.chatRepo.UpdateSessionTitle(sessionID, userID, title)
		}()
	}

	return &ChatResponse{
		ID:        assistantMsg.ID,
		SessionID: sessionID,
		Role:      "assistant",
		Content:   reply,
		Sources:   sources,
		CreatedAt: assistantMsg.CreatedAt,
	}, nil
}

// PrepareContext retrieves RAG context for a user message without calling the LLM.
// Returns retrieved context string and sources slice.
func (s *ChatService) PrepareContext(ctx context.Context, sessionID, userID, userMessage string) (string, []ChatSource, error) {
	// Verify session ownership and save user message
	if _, err := s.chatRepo.GetSession(sessionID, userID); err != nil {
		return "", nil, fmt.Errorf("session not found: %w", err)
	}

	userMsg := &repository.ChatMessage{
		SessionID: sessionID,
		UserID:    userID,
		Role:      "user",
		Content:   userMessage,
	}
	if err := s.chatRepo.CreateMessage(userMsg); err != nil {
		return "", nil, fmt.Errorf("failed to save user message: %w", err)
	}

	var contextParts []string
	var sources []ChatSource

	if s.vectorStore != nil {
		results, err := s.vectorStore.SearchAll(ctx, userID, userMessage, 5)
		if err != nil {
			results = nil
		}
		for _, r := range results {
			excerpt := r.Text
			if len(excerpt) > 300 {
				excerpt = excerpt[:300] + "..."
			}
			contextParts = append(contextParts, fmt.Sprintf("[相关内容 (相似度 %.2f)]: %s", r.Score, excerpt))
			sources = append(sources, ChatSource{
				ConversationID: r.SourceID,
				Title:          "",
				Excerpt:        excerpt[:min(100, len(excerpt))],
			})
		}
	}

	if len(contextParts) == 0 {
		convs, err := s.convRepo.Search(userID, userMessage, 5)
		if err != nil {
			convs = nil
		}
		for _, c := range convs {
			title := "Untitled"
			if c.Title != nil {
				title = *c.Title
			}
			excerpt := c.Transcript
			if len(excerpt) > 300 {
				excerpt = excerpt[:300] + "..."
			}
			contextParts = append(contextParts, fmt.Sprintf("[%s]: %s", title, excerpt))
			sources = append(sources, ChatSource{
				ConversationID: c.ID,
				Title:          title,
				Excerpt:        excerpt[:min(100, len(excerpt))],
			})
		}
	}

	retrievedContext := strings.Join(contextParts, "\n\n")
	return retrievedContext, sources, nil
}

// SaveAssistantMessage persists the completed assistant reply and triggers side effects.
func (s *ChatService) SaveAssistantMessage(ctx context.Context, sessionID, userID, userMessage, reply string, sources []ChatSource) (*ChatResponse, error) {
	sess, err := s.chatRepo.GetSession(sessionID, userID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	var sourcesJSON json.RawMessage
	if len(sources) > 0 {
		sourcesJSON, _ = json.Marshal(sources)
	}

	assistantMsg := &repository.ChatMessage{
		SessionID: sessionID,
		UserID:    userID,
		Role:      "assistant",
		Content:   reply,
		Sources:   sourcesJSON,
	}
	if err := s.chatRepo.CreateMessage(assistantMsg); err != nil {
		return nil, fmt.Errorf("failed to save assistant message: %w", err)
	}

	s.chatRepo.TouchSession(sessionID)

	if sess.Title == "新对话" {
		go func() {
			title, err := s.llmService.SummarizeTitle(context.Background(), userMessage, reply)
			if err != nil || title == "" {
				return
			}
			_ = s.chatRepo.UpdateSessionTitle(sessionID, userID, title)
		}()
	}

	return &ChatResponse{
		ID:        assistantMsg.ID,
		SessionID: sessionID,
		Role:      "assistant",
		Content:   reply,
		Sources:   sources,
		CreatedAt: assistantMsg.CreatedAt,
	}, nil
}
