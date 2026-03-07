package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	jwtpkg "github.com/mino/backend/internal/pkg/jwt"
	"github.com/mino/backend/internal/service"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

type WSHandler struct {
	jwtMgr       *jwtpkg.Manager
	audioService *service.AudioService
	sttService   *service.STTService
}

func NewWSHandler(jwtMgr *jwtpkg.Manager, audioService *service.AudioService, sttService *service.STTService) *WSHandler {
	return &WSHandler{
		jwtMgr:       jwtMgr,
		audioService: audioService,
		sttService:   sttService,
	}
}

// Incoming control message from client (JSON text frame)
type wsClientMessage struct {
	Type      string `json:"type"`
	Action    string `json:"action"` // "start" | "stop" | "pause" | "resume"
	Timestamp int64  `json:"timestamp"`
}

// Outgoing message types to client
type wsServerMessage struct {
	Type           string      `json:"type"`
	Text           string      `json:"text,omitempty"`
	IsFinal        bool        `json:"is_final,omitempty"`
	Timestamp      int64       `json:"timestamp,omitempty"`
	ConversationID string      `json:"conversation_id,omitempty"`
	Title          string      `json:"title,omitempty"`
	Summary        string      `json:"summary,omitempty"`
	ActionItems    []string    `json:"action_items,omitempty"`
	Memories       interface{} `json:"memories,omitempty"`
	Error          string      `json:"error,omitempty"`
}

// AudioWS handles the WebSocket endpoint for real-time audio streaming.
// The client sends PCM 16-bit 16kHz mono binary frames; the backend
// forwards them to the STT provider for real-time transcription.
func (h *WSHandler) AudioWS(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}
	claims, err := h.jwtMgr.Validate(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}
	userID := claims.UserID

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	var (
		recording   bool
		startTime   time.Time
		// transcript accumulates final sentences for LLM processing
		transcriptMu sync.Mutex
		transcript   strings.Builder
		// realtimeSession is the STT real-time session (nil when STT not configured)
		realtimeSession service.RealtimeSession
		// audioBuffer accumulates PCM audio data for storage
		audioBuffer []byte
	)

	// writeMu guards concurrent writes to the client WebSocket
	var writeMu sync.Mutex
	sendMsg := func(msg wsServerMessage) {
		data, _ := json.Marshal(msg)
		writeMu.Lock()
		conn.WriteMessage(websocket.TextMessage, data)
		writeMu.Unlock()
	}

	for {
		msgType, rawMsg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			// Client disconnected mid-recording — close STT session and process
			if recording {
				if realtimeSession != nil {
					realtimeSession.Close()
					realtimeSession = nil
				}
				transcriptMu.Lock()
				tx := transcript.String()
				transcriptMu.Unlock()
				if tx != "" {
					duration := int(time.Since(startTime).Seconds())
					audioCopy := make([]byte, len(audioBuffer))
					copy(audioCopy, audioBuffer)
					go func(uid, t string, dur int, audio []byte) {
						bgCtx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
						defer cancel()
						if _, err := h.audioService.ProcessTranscript(bgCtx, uid, t, dur, audio); err != nil {
							log.Printf("failed to process transcript after disconnect (userID=%s): %v", uid, err)
						}
					}(userID, tx, duration, audioCopy)
				}
			}
			break
		}

		// Binary frames are PCM audio chunks — forward to STT provider and accumulate
		if msgType == websocket.BinaryMessage {
			if recording {
				// Accumulate audio data for later storage
				audioBuffer = append(audioBuffer, rawMsg...)
				
				// Forward to STT for real-time transcription
				if realtimeSession != nil {
					if err := realtimeSession.SendAudio(rawMsg); err != nil {
						log.Printf("failed to send audio to STT: %v", err)
					}
				}
			}
			continue
		}

		// Text frames are JSON control messages
		var msg wsClientMessage
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			sendMsg(wsServerMessage{Type: "error", Error: "invalid message format"})
			continue
		}

		switch msg.Action {
		case "start":
			recording = true
			startTime = time.Now()
			transcript.Reset()
			audioBuffer = nil // Clear audio buffer

			// Start STT real-time session if STT is configured
			if h.sttService != nil {
				session, err := h.sttService.NewRealtimeSession(
					func(text string, isFinal bool) {
						if isFinal {
							transcriptMu.Lock()
							if transcript.Len() > 0 {
								transcript.WriteString(" ")
							}
							transcript.WriteString(text)
							transcriptMu.Unlock()
						}
						sendMsg(wsServerMessage{
							Type:      "transcript",
							Text:      text,
							IsFinal:   isFinal,
							Timestamp: time.Now().UnixMilli(),
						})
					},
				)
				if err != nil {
					log.Printf("failed to start STT realtime session: %v", err)
					sendMsg(wsServerMessage{Type: "error", Error: "failed to start transcription: " + err.Error()})
				} else {
					realtimeSession = session
				}
			}

			sendMsg(wsServerMessage{Type: "status", Text: "recording started"})

		case "stop":
			if !recording {
				continue
			}
			recording = false
			duration := int(time.Since(startTime).Seconds())

			// Close STT session and wait for final results
			if realtimeSession != nil {
				realtimeSession.Close()
				realtimeSession = nil
			}

			// Give STT provider a moment to flush the last result
			time.Sleep(500 * time.Millisecond)

			transcriptMu.Lock()
			finalTranscript := transcript.String()
			transcriptMu.Unlock()

			// Send final transcript to client
			sendMsg(wsServerMessage{
				Type:      "transcript",
				Text:      finalTranscript,
				IsFinal:   true,
				Timestamp: time.Now().UnixMilli(),
			})

			// Copy audio buffer before passing to goroutine to avoid data race
			audioCopy := make([]byte, len(audioBuffer))
			copy(audioCopy, audioBuffer)
			audioBuffer = nil

			// Run LLM extraction + storage asynchronously
			go func(uid, tx string, dur int, audio []byte) {
				bgCtx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
				defer cancel()

				conv, err := h.audioService.ProcessTranscript(bgCtx, uid, tx, dur, audio)
				if err != nil {
					sendMsg(wsServerMessage{Type: "error", Error: "processing failed"})
					return
				}

				title := ""
				summary := ""
				if conv.Title != nil {
					title = *conv.Title
				}
				if conv.Summary != nil {
					summary = *conv.Summary
				}

				sendMsg(wsServerMessage{
					Type:           "completed",
					ConversationID: conv.ID,
					Title:          title,
					Summary:        summary,
				})
			}(userID, finalTranscript, duration, audioCopy)

		case "pause":
			recording = false
		case "resume":
			recording = true
		}
	}
}
