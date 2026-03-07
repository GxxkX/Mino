package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mino/backend/internal/middleware"
	"github.com/mino/backend/internal/model"
	resp "github.com/mino/backend/internal/pkg/response"
	"github.com/mino/backend/internal/repository"
)

type MemoryHandler struct {
	memRepo *repository.MemoryRepository
}

func NewMemoryHandler(memRepo *repository.MemoryRepository) *MemoryHandler {
	return &MemoryHandler{memRepo: memRepo}
}

type updateMemoryRequest struct {
	Content    string `json:"content"`
	Category   string `json:"category"`
	Importance int    `json:"importance"`
}

func (h *MemoryHandler) List(c *gin.Context) {
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

	mems, total, err := h.memRepo.List(userID, limit, offset)
	if err != nil {
		resp.InternalError(c, "failed to fetch memories, reason: "+err.Error())
		return
	}
	if mems == nil {
		mems = []*model.Memory{}
	}
	resp.Paginated(c, mems, total, page, limit)
}

func (h *MemoryHandler) Get(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	mem, err := h.memRepo.FindByID(id, userID)
	if err != nil {
		resp.InternalError(c, "failed to fetch memory, reason: "+err.Error())
		return
	}
	if mem == nil {
		resp.NotFound(c, "memory not found, reason: no record with given id")
		return
	}
	resp.OK(c, mem)
}

func (h *MemoryHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	var req updateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	mem, err := h.memRepo.FindByID(id, userID)
	if err != nil {
		resp.NotFound(c, "memory not found, reason: "+err.Error())
		return
	}
	if mem == nil {
		resp.NotFound(c, "memory not found, reason: no record with given id")
		return
	}

	if req.Content != "" {
		mem.Content = req.Content
	}
	if req.Category != "" {
		mem.Category = &req.Category
	}
	if req.Importance > 0 {
		mem.Importance = req.Importance
	}

	if err := h.memRepo.Update(mem); err != nil {
		resp.InternalError(c, "failed to update memory, reason: "+err.Error())
		return
	}
	resp.OK(c, mem)
}

func (h *MemoryHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	if err := h.memRepo.Delete(id, userID); err != nil {
		resp.NotFound(c, "memory not found, reason: "+err.Error())
		return
	}
	resp.NoContent(c)
}
