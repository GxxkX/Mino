package service

import (
	"context"
)

// StructuredResult is what the LLM extracts from a conversation transcript.
type StructuredResult struct {
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	ActionItems []string `json:"action_items"`
	Memories    []Memory `json:"memories"`
}

// Memory represents a single memory point extracted from a conversation.
type Memory struct {
	Content    string `json:"content"`
	Category   string `json:"category"`   // insight, fact, preference, event
	Importance int    `json:"importance"` // 1-10
}

// ToolCall represents a single tool invocation requested by the LLM.
type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// LLMProvider defines the distinct agent capabilities for the Mino AI service:
//
//  1. ChatAgent — used by the web/mobile/desktop frontend for interactive chat (RAG-based Q&A).
//     It receives the user's message along with retrieved context and returns a conversational reply.
//
//  2. ChatAgentStream — streaming variant of ChatAgent; calls chunkFn for each token chunk.
//
//  3. ChatAgentWithTools — ChatAgent with function/tool calling support. The LLM can invoke
//     tools (e.g. memory_create, task_list) during the conversation. The toolExecutor callback
//     is called for each tool invocation and its result is fed back to the LLM.
//
//  4. ChatAgentWithToolsStream — streaming variant of ChatAgentWithTools.
//
//  5. ExtractAgent — used after STT produces a transcript. It extracts structured data
//     (title, summary, action_items, memories) from the raw transcript text.
//
//  6. SummarizeTitle — generates a short title for a chat session based on the first exchange.
type LLMProvider interface {
	// ChatAgent sends a user message with optional retrieved context to the chat agent
	// and returns the assistant's reply. The context parameter contains relevant conversation
	// history retrieved via RAG (semantic search, full-text search, etc.).
	ChatAgent(ctx context.Context, userMessage string, retrievedContext string) (string, error)

	// ChatAgentStream is the streaming variant of ChatAgent. It calls chunkFn for each
	// token chunk as it arrives and returns the full accumulated reply when done.
	ChatAgentStream(ctx context.Context, userMessage string, retrievedContext string, chunkFn func(chunk string)) (string, error)

	// ChatAgentWithTools sends a user message with tools available for the LLM to call.
	// toolExecutor is invoked when the LLM requests a tool call; its result is fed back
	// to the LLM for the next turn. Returns the final text reply.
	ChatAgentWithTools(ctx context.Context, userMessage string, retrievedContext string, tools []ToolDefinition, toolExecutor func(name string, args map[string]interface{}) (*ToolResult, error)) (string, error)

	// ChatAgentWithToolsStream is the streaming variant of ChatAgentWithTools.
	ChatAgentWithToolsStream(ctx context.Context, userMessage string, retrievedContext string, tools []ToolDefinition, toolExecutor func(name string, args map[string]interface{}) (*ToolResult, error), chunkFn func(chunk string)) (string, error)

	// ExtractAgent takes a raw STT transcript and extracts structured information
	// (title, summary, action items, memories) using a dedicated extraction prompt.
	ExtractAgent(ctx context.Context, transcript string) (*StructuredResult, error)

	// SummarizeTitle generates a concise session title from a user message and assistant reply.
	SummarizeTitle(ctx context.Context, userMessage, assistantReply string) (string, error)
}

// EmbeddingProvider generates vector embeddings from text.
// Implementations may use OpenAI, Zhipu, or other embedding APIs.
type EmbeddingProvider interface {
	// EmbedQuery generates a single embedding vector for a query text.
	EmbedQuery(ctx context.Context, text string) ([]float32, error)

	// EmbedDocuments generates embedding vectors for multiple texts.
	EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error)
}
