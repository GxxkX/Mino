package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mino/backend/internal/middleware"
	resp "github.com/mino/backend/internal/pkg/response"
	"github.com/mino/backend/internal/service"
)

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

// --------------- Session endpoints ---------------

func (h *ChatHandler) CreateSession(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		Title string `json:"title"`
	}
	_ = c.ShouldBindJSON(&req)

	sess, err := h.chatService.CreateSession(userID, req.Title)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Created(c, sess)
}

func (h *ChatHandler) ListSessions(c *gin.Context) {
	userID := middleware.GetUserID(c)

	sessions, err := h.chatService.ListSessions(userID)
	if err != nil {
		resp.InternalError(c, "failed to list sessions: "+err.Error())
		return
	}
	resp.OK(c, sessions)
}

func (h *ChatHandler) UpdateSession(c *gin.Context) {
	userID := middleware.GetUserID(c)
	sessionID := c.Param("id")

	var req struct {
		Title string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	if err := h.chatService.UpdateSessionTitle(sessionID, userID, req.Title); err != nil {
		resp.NotFound(c, "session not found")
		return
	}
	resp.OK(c, gin.H{"id": sessionID, "title": req.Title})
}

func (h *ChatHandler) DeleteSession(c *gin.Context) {
	userID := middleware.GetUserID(c)
	sessionID := c.Param("id")

	if err := h.chatService.DeleteSession(sessionID, userID); err != nil {
		resp.NotFound(c, "session not found")
		return
	}
	resp.NoContent(c)
}

// --------------- Message endpoints ---------------

type sendMessageRequest struct {
	Message string `json:"message" binding:"required"`
}

func (h *ChatHandler) Send(c *gin.Context) {
	userID := middleware.GetUserID(c)
	sessionID := c.Param("id")

	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	result, err := h.chatService.SendMessage(c.Request.Context(), sessionID, userID, req.Message)
	if err != nil {
		resp.InternalError(c, "failed to process message: "+err.Error())
		return
	}
	resp.OK(c, result)
}

// SendStream handles POST /chat/sessions/:id/messages/stream
// It streams the LLM response as Server-Sent Events (SSE).
//
// SSE event types:
//
//	data: {"type":"chunk","content":"..."}   — token chunk
//	data: {"type":"sources","sources":[...]} — RAG sources (sent once before first chunk)
//	data: {"type":"done","id":"...","createdAt":"..."} — stream complete, includes persisted message ID
//	data: {"type":"error","message":"..."}   — error
func (h *ChatHandler) SendStream(c *gin.Context) {
	userID := middleware.GetUserID(c)
	sessionID := c.Param("id")

	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ctx := c.Request.Context()

	// Step 1: RAG context retrieval + save user message
	retrievedContext, sources, err := h.chatService.PrepareContext(ctx, sessionID, userID, req.Message)
	if err != nil {
		writeSSEEvent(c.Writer, map[string]interface{}{
			"type":    "error",
			"message": "failed to prepare context: " + err.Error(),
		})
		return
	}

	// Step 2: Send sources event so the client can display citations immediately
	if len(sources) > 0 {
		writeSSEEvent(c.Writer, map[string]interface{}{
			"type":    "sources",
			"sources": sources,
		})
	}

	// Step 3: Stream LLM tokens (with tool calling if tools are registered)
	var reply string
	var llmErr error

	toolRegistry := h.chatService.GetToolRegistry()
	if toolRegistry != nil {
		toolDefs := toolRegistry.DefinitionsForScope("chat")
		if len(toolDefs) > 0 {
			reply, llmErr = h.chatService.GetLLMService().ChatAgentWithToolsStream(
				ctx, req.Message, retrievedContext, toolDefs,
				func(name string, args map[string]interface{}) (*service.ToolResult, error) {
					return toolRegistry.Execute(ctx, userID, name, args)
				},
				func(chunk string) {
					writeSSEEvent(c.Writer, map[string]interface{}{
						"type":    "chunk",
						"content": chunk,
					})
				},
			)
		} else {
			reply, llmErr = h.chatService.GetLLMService().ChatAgentStream(ctx, req.Message, retrievedContext, func(chunk string) {
				writeSSEEvent(c.Writer, map[string]interface{}{
					"type":    "chunk",
					"content": chunk,
				})
			})
		}
	} else {
		reply, llmErr = h.chatService.GetLLMService().ChatAgentStream(ctx, req.Message, retrievedContext, func(chunk string) {
			writeSSEEvent(c.Writer, map[string]interface{}{
				"type":    "chunk",
				"content": chunk,
			})
		})
	}

	if llmErr != nil && reply == "" {
		writeSSEEvent(c.Writer, map[string]interface{}{
			"type":    "error",
			"message": "LLM error: " + llmErr.Error(),
		})
		return
	}

	// Step 4: Persist assistant message
	saved, saveErr := h.chatService.SaveAssistantMessage(ctx, sessionID, userID, req.Message, reply, sources)
	if saveErr != nil {
		writeSSEEvent(c.Writer, map[string]interface{}{
			"type":    "error",
			"message": "failed to save message: " + saveErr.Error(),
		})
		return
	}

	// Step 5: Done event
	writeSSEEvent(c.Writer, map[string]interface{}{
		"type":      "done",
		"id":        saved.ID,
		"createdAt": saved.CreatedAt,
	})
}

func (h *ChatHandler) Messages(c *gin.Context) {
	userID := middleware.GetUserID(c)
	sessionID := c.Param("id")

	msgs, err := h.chatService.GetMessages(sessionID, userID)
	if err != nil {
		resp.InternalError(c, "failed to fetch messages: "+err.Error())
		return
	}
	resp.OK(c, msgs)
}

// writeSSEEvent serialises v as JSON and writes it as an SSE data line, then flushes.
func writeSSEEvent(w http.ResponseWriter, v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		fmt.Fprintf(w, "data: {\"type\":\"error\",\"message\":\"marshal error\"}\n\n")
	} else {
		fmt.Fprintf(w, "data: %s\n\n", b)
	}
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}
