package service

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/mino/backend/internal/model"
	"github.com/mino/backend/internal/repository"
)

// MemoryTaskService provides unified business logic for memory and task operations.
// It is shared by ChatAgent tools, ExtractAgent post-processing, and REST handlers.
type MemoryTaskService struct {
	memRepo     *repository.MemoryRepository
	taskRepo    *repository.TaskRepository
	vectorStore *VectorStoreService
	logger      *logrus.Logger
}

func NewMemoryTaskService(
	memRepo *repository.MemoryRepository,
	taskRepo *repository.TaskRepository,
	vectorStore *VectorStoreService,
	logger *logrus.Logger,
) *MemoryTaskService {
	return &MemoryTaskService{
		memRepo:     memRepo,
		taskRepo:    taskRepo,
		vectorStore: vectorStore,
		logger:      logger,
	}
}

// ── Memory operations ──

func (s *MemoryTaskService) CreateMemory(ctx context.Context, userID string, content, category string, importance int, conversationID *string) (*model.Memory, error) {
	if content == "" {
		return nil, fmt.Errorf("memory content is required")
	}
	if importance < 1 {
		importance = 5
	}
	if importance > 10 {
		importance = 10
	}

	mem := &model.Memory{
		UserID:         userID,
		ConversationID: conversationID,
		Content:        content,
		Importance:     importance,
	}
	if category != "" {
		mem.Category = &category
	}

	if err := s.memRepo.Create(mem); err != nil {
		return nil, fmt.Errorf("failed to create memory: %w", err)
	}

	// Index in vector store (async, non-blocking)
	if s.vectorStore != nil {
		go func() {
			if err := s.vectorStore.IndexMemories(context.Background(), userID, []string{mem.ID}, []string{mem.Content}); err != nil {
				s.logger.Warnf("failed to index memory in vector store: %v", err)
			}
		}()
	}

	return mem, nil
}

// CreateMemories batch-creates memories and indexes them in the vector store.
func (s *MemoryTaskService) CreateMemories(ctx context.Context, userID string, memories []MemoryInput, conversationID *string) ([]*model.Memory, error) {
	if len(memories) == 0 {
		return nil, nil
	}

	var mems []*model.Memory
	for _, m := range memories {
		importance := m.Importance
		if importance < 1 {
			importance = 5
		}
		if importance > 10 {
			importance = 10
		}
		cat := m.Category
		mem := &model.Memory{
			UserID:         userID,
			ConversationID: conversationID,
			Content:        m.Content,
			Importance:     importance,
		}
		if cat != "" {
			mem.Category = &cat
		}
		mems = append(mems, mem)
	}

	if err := s.memRepo.CreateBatch(mems); err != nil {
		return nil, fmt.Errorf("failed to batch create memories: %w", err)
	}

	// Index in vector store (async, non-blocking)
	if s.vectorStore != nil {
		go func() {
			var ids, contents []string
			for _, m := range mems {
				ids = append(ids, m.ID)
				contents = append(contents, m.Content)
			}
			if err := s.vectorStore.IndexMemories(context.Background(), userID, ids, contents); err != nil {
				s.logger.Warnf("failed to index memories in vector store: %v", err)
			}
		}()
	}

	return mems, nil
}

func (s *MemoryTaskService) SearchMemories(ctx context.Context, userID, query string, limit int) ([]*model.Memory, error) {
	if limit <= 0 {
		limit = 10
	}
	mems, _, err := s.memRepo.List(userID, limit, 0)
	return mems, err
}

func (s *MemoryTaskService) DeleteMemory(ctx context.Context, userID, memoryID string) error {
	return s.memRepo.Delete(memoryID, userID)
}

// ── Task operations ──

func (s *MemoryTaskService) CreateTask(ctx context.Context, userID, title string, description, priority string, conversationID *string) (*model.Task, error) {
	if title == "" {
		return nil, fmt.Errorf("task title is required")
	}
	if priority == "" {
		priority = "medium"
	}

	task := &model.Task{
		UserID:         userID,
		ConversationID: conversationID,
		Title:          title,
		Status:         "pending",
		Priority:       priority,
	}
	if description != "" {
		task.Description = &description
	}

	if err := s.taskRepo.Create(task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	return task, nil
}

// CreateTasks batch-creates tasks.
func (s *MemoryTaskService) CreateTasks(ctx context.Context, userID string, tasks []TaskInput, conversationID *string) ([]*model.Task, error) {
	if len(tasks) == 0 {
		return nil, nil
	}

	var models []*model.Task
	for _, t := range tasks {
		priority := t.Priority
		if priority == "" {
			priority = "medium"
		}
		task := &model.Task{
			UserID:         userID,
			ConversationID: conversationID,
			Title:          t.Title,
			Status:         "pending",
			Priority:       priority,
		}
		if t.Description != "" {
			task.Description = &t.Description
		}
		models = append(models, task)
	}

	if err := s.taskRepo.CreateBatch(models); err != nil {
		return nil, fmt.Errorf("failed to batch create tasks: %w", err)
	}
	return models, nil
}

func (s *MemoryTaskService) ListTasks(ctx context.Context, userID string, limit int) ([]*model.Task, int, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.taskRepo.List(userID, limit, 0)
}

func (s *MemoryTaskService) UpdateTaskStatus(ctx context.Context, userID, taskID, status string) (*model.Task, error) {
	task, err := s.taskRepo.FindByID(taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}
	task.Status = status
	if err := s.taskRepo.Update(task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}
	return task, nil
}

func (s *MemoryTaskService) DeleteTask(ctx context.Context, userID, taskID string) error {
	return s.taskRepo.Delete(taskID, userID)
}

// ── Input types ──

// MemoryInput is the input for creating a memory (used by both tools and ExtractAgent).
type MemoryInput struct {
	Content    string `json:"content"`
	Category   string `json:"category"`
	Importance int    `json:"importance"`
}

// TaskInput is the input for creating a task (used by both tools and ExtractAgent).
type TaskInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
}
