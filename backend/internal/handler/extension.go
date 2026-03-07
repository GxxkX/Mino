package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mino/backend/internal/middleware"
	"github.com/mino/backend/internal/model"
	resp "github.com/mino/backend/internal/pkg/response"
	"github.com/mino/backend/internal/repository"
)

type ExtensionHandler struct {
	extRepo *repository.ExtensionRepository
}

func NewExtensionHandler(extRepo *repository.ExtensionRepository) *ExtensionHandler {
	return &ExtensionHandler{extRepo: extRepo}
}

type createExtensionRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
	Config      *string `json:"config"`
}

type updateExtensionRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
	Enabled     *bool   `json:"enabled"`
	Config      *string `json:"config"`
}

func (h *ExtensionHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	exts, err := h.extRepo.List(userID)
	if err != nil {
		resp.InternalError(c, "failed to fetch extensions, reason: "+err.Error())
		return
	}
	if exts == nil {
		exts = []*model.Extension{}
	}
	resp.OK(c, exts)
}

func (h *ExtensionHandler) Get(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")
	ext, err := h.extRepo.FindByID(id, userID)
	if err != nil {
		resp.NotFound(c, "extension not found, reason: "+err.Error())
		return
	}
	if ext == nil {
		resp.NotFound(c, "extension not found, reason: no record with given id")
		return
	}
	resp.OK(c, ext)
}

func (h *ExtensionHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req createExtensionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	icon := req.Icon
	if icon == "" {
		icon = "zap"
	}

	ext := &model.Extension{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Icon:        icon,
		Enabled:     false,
		Config:      req.Config,
	}

	if err := h.extRepo.Create(ext); err != nil {
		resp.InternalError(c, "failed to create extension, reason: "+err.Error())
		return
	}
	resp.Created(c, ext)
}

func (h *ExtensionHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	var req updateExtensionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	ext, err := h.extRepo.FindByID(id, userID)
	if err != nil {
		resp.NotFound(c, "extension not found, reason: "+err.Error())
		return
	}
	if ext == nil {
		resp.NotFound(c, "extension not found, reason: no record with given id")
		return
	}

	if req.Name != "" {
		ext.Name = req.Name
	}
	if req.Description != "" {
		ext.Description = req.Description
	}
	if req.Icon != "" {
		ext.Icon = req.Icon
	}
	if req.Enabled != nil {
		ext.Enabled = *req.Enabled
	}
	if req.Config != nil {
		ext.Config = req.Config
	}

	if err := h.extRepo.Update(ext); err != nil {
		resp.InternalError(c, "failed to update extension, reason: "+err.Error())
		return
	}
	resp.OK(c, ext)
}

func (h *ExtensionHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")
	if err := h.extRepo.Delete(id, userID); err != nil {
		resp.NotFound(c, "extension not found, reason: "+err.Error())
		return
	}
	resp.NoContent(c)
}
