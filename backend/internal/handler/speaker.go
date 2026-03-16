package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mino/backend/internal/middleware"
	"github.com/mino/backend/internal/model"
	resp "github.com/mino/backend/internal/pkg/response"
	"github.com/mino/backend/internal/service"
)

// ensure imports are used
var _ = (*model.SpeakerProfile)(nil)

type SpeakerHandler struct {
	speakerSvc *service.SpeakerService
}

func NewSpeakerHandler(speakerSvc *service.SpeakerService) *SpeakerHandler {
	return &SpeakerHandler{speakerSvc: speakerSvc}
}

// List returns all speaker profiles for the authenticated user.
// GET /v1/speakers
func (h *SpeakerHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	speakers, err := h.speakerSvc.ListSpeakers(c.Request.Context(), userID)
	if err != nil {
		resp.InternalError(c, "failed to list speakers: "+err.Error())
		return
	}
	if speakers == nil {
		speakers = []*model.SpeakerProfile{}
	}
	resp.OK(c, speakers)
}

// Register creates a new speaker profile with a name and embedding.
// POST /v1/speakers
func (h *SpeakerHandler) Register(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		Name      string    `json:"name" binding:"required"`
		Embedding []float32 `json:"embedding" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "invalid request: name and embedding are required")
		return
	}

	sp, err := h.speakerSvc.RegisterSpeaker(c.Request.Context(), userID, req.Name, req.Embedding)
	if err != nil {
		resp.InternalError(c, "failed to register speaker: "+err.Error())
		return
	}
	resp.Created(c, sp)
}

// UpdateName updates the display name of a speaker profile.
// PUT /v1/speakers/:id
func (h *SpeakerHandler) UpdateName(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "invalid request: name is required")
		return
	}

	if err := h.speakerSvc.UpdateSpeakerName(c.Request.Context(), userID, id, req.Name); err != nil {
		resp.NotFound(c, "speaker not found")
		return
	}
	resp.OK(c, gin.H{"id": id, "name": req.Name})
}

// Delete removes a speaker profile.
// DELETE /v1/speakers/:id
func (h *SpeakerHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	if err := h.speakerSvc.DeleteSpeaker(c.Request.Context(), userID, id); err != nil {
		resp.NotFound(c, "speaker not found")
		return
	}
	resp.NoContent(c)
}
