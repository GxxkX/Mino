package tools

import (
	"context"
	"fmt"

	"github.com/mino/backend/internal/service"
)

// ── TaskCreateTool ──

type TaskCreateTool struct {
	svc *service.MemoryTaskService
}

func NewTaskCreateTool(svc *service.MemoryTaskService) *TaskCreateTool {
	return &TaskCreateTool{svc: svc}
}

func (t *TaskCreateTool) Definition() service.ToolDefinition {
	return service.ToolDefinition{
		Name:        "task_create",
		Description: "创建一个待办任务。适用场景：1) 用户在对话中要求创建任务或待办事项；2) 从语音转写文本中提取到需要执行的行动项。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type":        "string",
					"description": "任务标题",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "任务描述（可选）",
				},
				"priority": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"low", "medium", "high", "urgent"},
					"description": "优先级，默认 medium",
				},
			},
			"required": []string{"title"},
		},
	}
}

func (t *TaskCreateTool) Execute(ctx context.Context, userID string, args map[string]interface{}) (*service.ToolResult, error) {
	title, _ := args["title"].(string)
	if title == "" {
		return &service.ToolResult{Success: false, Error: "title is required"}, nil
	}

	description, _ := args["description"].(string)
	priority, _ := args["priority"].(string)

	task, err := t.svc.CreateTask(ctx, userID, title, description, priority, nil)
	if err != nil {
		return &service.ToolResult{Success: false, Error: err.Error()}, nil
	}
	return &service.ToolResult{Success: true, Data: task}, nil
}

// ── TaskCreateBatchTool ──

type TaskCreateBatchTool struct {
	svc *service.MemoryTaskService
}

func NewTaskCreateBatchTool(svc *service.MemoryTaskService) *TaskCreateBatchTool {
	return &TaskCreateBatchTool{svc: svc}
}

func (t *TaskCreateBatchTool) Definition() service.ToolDefinition {
	return service.ToolDefinition{
		Name:        "task_create_batch",
		Description: "批量创建多个待办任务。适用于从语音转写文本中一次性提取多个行动项，或用户要求同时创建多个任务的场景。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tasks": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"title": map[string]interface{}{
								"type":        "string",
								"description": "任务标题",
							},
							"description": map[string]interface{}{
								"type":        "string",
								"description": "任务描述",
							},
							"priority": map[string]interface{}{
								"type":        "string",
								"enum":        []string{"low", "medium", "high", "urgent"},
								"description": "优先级",
							},
						},
						"required": []string{"title"},
					},
					"description": "任务列表",
				},
				"conversation_id": map[string]interface{}{
					"type":        "string",
					"description": "关联的对话ID（可选）",
				},
			},
			"required": []string{"tasks"},
		},
	}
}

func (t *TaskCreateBatchTool) Execute(ctx context.Context, userID string, args map[string]interface{}) (*service.ToolResult, error) {
	rawTasks, ok := args["tasks"].([]interface{})
	if !ok || len(rawTasks) == 0 {
		return &service.ToolResult{Success: false, Error: "tasks array is required"}, nil
	}

	var convID *string
	if cid, ok := args["conversation_id"].(string); ok && cid != "" {
		convID = &cid
	}

	var inputs []service.TaskInput
	for _, raw := range rawTasks {
		m, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		title, _ := m["title"].(string)
		if title == "" {
			continue
		}
		description, _ := m["description"].(string)
		priority, _ := m["priority"].(string)
		inputs = append(inputs, service.TaskInput{
			Title:       title,
			Description: description,
			Priority:    priority,
		})
	}

	if len(inputs) == 0 {
		return &service.ToolResult{Success: false, Error: "no valid tasks provided"}, nil
	}

	tasks, err := t.svc.CreateTasks(ctx, userID, inputs, convID)
	if err != nil {
		return &service.ToolResult{Success: false, Error: err.Error()}, nil
	}
	return &service.ToolResult{
		Success: true,
		Data:    map[string]interface{}{"count": len(tasks), "tasks": tasks},
	}, nil
}

// ── TaskListTool ──

type TaskListTool struct {
	svc *service.MemoryTaskService
}

func NewTaskListTool(svc *service.MemoryTaskService) *TaskListTool {
	return &TaskListTool{svc: svc}
}

func (t *TaskListTool) Definition() service.ToolDefinition {
	return service.ToolDefinition{
		Name:        "task_list",
		Description: "查询用户的待办任务列表。当用户询问自己有哪些任务、待办事项时使用。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "返回数量上限，默认20",
				},
			},
		},
	}
}

func (t *TaskListTool) Execute(ctx context.Context, userID string, args map[string]interface{}) (*service.ToolResult, error) {
	limit := intFromArgs(args, "limit", 20)

	tasks, total, err := t.svc.ListTasks(ctx, userID, limit)
	if err != nil {
		return &service.ToolResult{Success: false, Error: err.Error()}, nil
	}
	return &service.ToolResult{
		Success: true,
		Data:    map[string]interface{}{"total": total, "tasks": tasks},
	}, nil
}

// ── TaskUpdateTool ──

type TaskUpdateTool struct {
	svc *service.MemoryTaskService
}

func NewTaskUpdateTool(svc *service.MemoryTaskService) *TaskUpdateTool {
	return &TaskUpdateTool{svc: svc}
}

func (t *TaskUpdateTool) Definition() service.ToolDefinition {
	return service.ToolDefinition{
		Name:        "task_update",
		Description: "更新任务状态或信息。当用户要求标记任务完成、修改任务优先级等时使用。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"task_id": map[string]interface{}{
					"type":        "string",
					"description": "任务ID",
				},
				"status": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"pending", "in_progress", "completed", "cancelled"},
					"description": "新状态",
				},
			},
			"required": []string{"task_id", "status"},
		},
	}
}

func (t *TaskUpdateTool) Execute(ctx context.Context, userID string, args map[string]interface{}) (*service.ToolResult, error) {
	taskID, _ := args["task_id"].(string)
	status, _ := args["status"].(string)
	if taskID == "" || status == "" {
		return &service.ToolResult{Success: false, Error: "task_id and status are required"}, nil
	}

	task, err := t.svc.UpdateTaskStatus(ctx, userID, taskID, status)
	if err != nil {
		return &service.ToolResult{Success: false, Error: err.Error()}, nil
	}
	return &service.ToolResult{Success: true, Data: task}, nil
}

// ── TaskDeleteTool ──

type TaskDeleteTool struct {
	svc *service.MemoryTaskService
}

func NewTaskDeleteTool(svc *service.MemoryTaskService) *TaskDeleteTool {
	return &TaskDeleteTool{svc: svc}
}

func (t *TaskDeleteTool) Definition() service.ToolDefinition {
	return service.ToolDefinition{
		Name:        "task_delete",
		Description: "删除一个任务。当用户要求移除某个任务时使用。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"task_id": map[string]interface{}{
					"type":        "string",
					"description": "要删除的任务ID",
				},
			},
			"required": []string{"task_id"},
		},
	}
}

func (t *TaskDeleteTool) Execute(ctx context.Context, userID string, args map[string]interface{}) (*service.ToolResult, error) {
	taskID, _ := args["task_id"].(string)
	if taskID == "" {
		return &service.ToolResult{Success: false, Error: "task_id is required"}, nil
	}

	if err := t.svc.DeleteTask(ctx, userID, taskID); err != nil {
		return &service.ToolResult{Success: false, Error: err.Error()}, nil
	}
	return &service.ToolResult{Success: true, Data: fmt.Sprintf("task %s deleted", taskID)}, nil
}
