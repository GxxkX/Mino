package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mino/backend/internal/middleware"
	"github.com/mino/backend/internal/model"
	resp "github.com/mino/backend/internal/pkg/response"
	"github.com/mino/backend/internal/repository"
)

type TaskHandler struct {
	taskRepo *repository.TaskRepository
}

func NewTaskHandler(taskRepo *repository.TaskRepository) *TaskHandler {
	return &TaskHandler{taskRepo: taskRepo}
}

type createTaskRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description"`
	Priority    string  `json:"priority"`
	DueDate     *string `json:"dueDate"`
}

type updateTaskRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	DueDate     *string `json:"dueDate"`
}

func (h *TaskHandler) List(c *gin.Context) {
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

	tasks, total, err := h.taskRepo.List(userID, limit, offset)
	if err != nil {
		resp.InternalError(c, "failed to fetch tasks, reason: "+err.Error())
		return
	}
	if tasks == nil {
		tasks = []*model.Task{}
	}
	resp.Paginated(c, tasks, total, page, limit)
}

func (h *TaskHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req createTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}

	task := &model.Task{
		UserID:   userID,
		Title:    req.Title,
		Status:   "pending",
		Priority: priority,
	}
	if req.Description != "" {
		task.Description = &req.Description
	}
	if req.DueDate != nil {
		t, err := time.Parse(time.RFC3339, *req.DueDate)
		if err == nil {
			task.DueDate = &t
		}
	}

	if err := h.taskRepo.Create(task); err != nil {
		resp.InternalError(c, "failed to create task, reason: "+err.Error())
		return
	}
	resp.Created(c, task)
}

func (h *TaskHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	var req updateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	task, err := h.taskRepo.FindByID(id, userID)
	if err != nil {
		resp.NotFound(c, "task not found, reason: "+err.Error())
		return
	}
	if task == nil {
		resp.NotFound(c, "task not found, reason: no record with given id")
		return
	}

	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = &req.Description
	}
	if req.Status != "" {
		task.Status = req.Status
		if req.Status == "completed" && task.CompletedAt == nil {
			now := time.Now()
			task.CompletedAt = &now
		}
	}
	if req.Priority != "" {
		task.Priority = req.Priority
	}
	if req.DueDate != nil {
		t, err := time.Parse(time.RFC3339, *req.DueDate)
		if err == nil {
			task.DueDate = &t
		}
	}

	if err := h.taskRepo.Update(task); err != nil {
		resp.InternalError(c, "failed to update task, reason: "+err.Error())
		return
	}
	resp.OK(c, task)
}

func (h *TaskHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	if err := h.taskRepo.Delete(id, userID); err != nil {
		resp.NotFound(c, "task not found, reason: "+err.Error())
		return
	}
	resp.NoContent(c)
}
