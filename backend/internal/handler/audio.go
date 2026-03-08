package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mino/backend/internal/middleware"
	"github.com/mino/backend/internal/model"
	resp "github.com/mino/backend/internal/pkg/response"
	"github.com/mino/backend/internal/pkg/storage"
	"github.com/mino/backend/internal/repository"
)

type ConversationHandler struct {
	convRepo *repository.ConversationRepository
	storage  *storage.Client // may be nil when MinIO is not configured
}

func NewConversationHandler(convRepo *repository.ConversationRepository, storageClient *storage.Client) *ConversationHandler {
	return &ConversationHandler{convRepo: convRepo, storage: storageClient}
}

func (h *ConversationHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	convs, total, err := h.convRepo.List(userID, limit, offset)
	if err != nil {
		resp.InternalError(c, "failed to fetch conversations, reason: "+err.Error())
		return
	}
	if convs == nil {
		convs = []*model.Conversation{}
	}
	// Rewrite audioUrl to the backend proxy endpoint so clients never see
	// raw MinIO URLs (works for both old public URLs and new object keys).
	for _, conv := range convs {
		rewriteAudioURL(conv)
	}
	resp.Paginated(c, convs, total, page, limit)
}

func (h *ConversationHandler) Get(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	conv, err := h.convRepo.FindByID(id, userID)
	if err != nil {
		resp.InternalError(c, "failed to fetch conversation, reason: "+err.Error())
		return
	}
	if conv == nil {
		resp.NotFound(c, "conversation not found, reason: no record with given id")
		return
	}
	rewriteAudioURL(conv)
	resp.OK(c, conv)
}

// StreamAudio proxies the audio file from MinIO to the client.
// GET /v1/conversations/:id/audio
func (h *ConversationHandler) StreamAudio(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	if h.storage == nil {
		resp.InternalError(c, "audio storage not configured")
		return
	}

	conv, err := h.convRepo.FindByID(id, userID)
	if err != nil {
		resp.InternalError(c, "failed to fetch conversation, reason: "+err.Error())
		return
	}
	if conv == nil {
		resp.NotFound(c, "conversation not found")
		return
	}
	if conv.AudioURL == nil || *conv.AudioURL == "" {
		resp.NotFound(c, "no audio available for this conversation")
		return
	}

	objectKey := extractObjectKey(*conv.AudioURL, userID, id)
	reader, contentType, size, err := h.storage.GetAudio(c.Request.Context(), objectKey)
	if err != nil {
		resp.InternalError(c, "failed to retrieve audio: "+err.Error())
		return
	}
	defer reader.Close()

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Length", fmt.Sprintf("%d", size))
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "private, max-age=3600")
	c.Status(http.StatusOK)
	io.Copy(c.Writer, reader)
}

// rewriteAudioURL replaces the stored audio URL (MinIO internal URL, public
// URL, or object key) with the backend proxy path so clients always use
// GET /v1/conversations/:id/audio.
func rewriteAudioURL(conv *model.Conversation) {
	if conv == nil || conv.AudioURL == nil || *conv.AudioURL == "" {
		return
	}
	proxyURL := fmt.Sprintf("/v1/conversations/%s/audio", conv.ID)
	conv.AudioURL = &proxyURL
}

func (h *ConversationHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")
	deleteAudio := c.Query("delete_audio") == "true"

	// When the caller wants to delete the audio file we need the conversation
	// record first so we can derive the MinIO object key from the stored URL.
	if deleteAudio && h.storage != nil {
		conv, err := h.convRepo.FindByID(id, userID)
		if err != nil {
			resp.InternalError(c, "failed to fetch conversation, reason: "+err.Error())
			return
		}
		if conv == nil {
			resp.NotFound(c, "conversation not found, reason: no record with given id")
			return
		}
		if conv.AudioURL != nil && *conv.AudioURL != "" {
			objectKey := extractObjectKey(*conv.AudioURL, userID, id)
			if err := h.storage.DeleteAudio(context.Background(), objectKey); err != nil {
				fmt.Printf("warning: failed to delete audio from MinIO (%s): %v\n", objectKey, err)
			}
		}
	}

	if err := h.convRepo.Delete(id, userID); err != nil {
		resp.NotFound(c, "conversation not found, reason: "+err.Error())
		return
	}
	resp.NoContent(c)
}

// extractObjectKey derives the MinIO object key from the stored audio URL.
//
// The URL is typically one of:
//   - "https://localhost/mino-audio/{userID}/{convID}.webm"  (public URL)
//   - "https://localhost/mino-audio/{userID}/{convID}.wav"   (public URL, PCM→WAV)
//   - "/mino-audio/{userID}/{convID}.webm"                     (relative)
//
// We strip everything up to and including the bucket name prefix "/mino-audio/"
// to get the raw object key. If parsing fails, fall back to trying both
// "{userID}/{convID}.webm" and ".wav" extensions.
func extractObjectKey(audioURL, userID, convID string) string {
	const bucketPrefix = "/mino-audio/"
	if idx := strings.Index(audioURL, bucketPrefix); idx >= 0 {
		return audioURL[idx+len(bucketPrefix):]
	}
	// Fallback: derive from IDs — check URL hint for extension
	if strings.HasSuffix(audioURL, ".wav") {
		return userID + "/" + convID + ".wav"
	}
	return userID + "/" + convID + ".webm"
}
