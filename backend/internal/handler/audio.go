package handler

import (
	"context"
	"fmt"
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
	resp.OK(c, conv)
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
