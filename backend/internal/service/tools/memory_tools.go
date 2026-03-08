package tools

import (
	"context"
	"fmt"

	"github.com/mino/backend/internal/service"
)

// ── MemoryCreateTool ──

type MemoryCreateTool struct {
	svc *service.MemoryTaskService
}

func NewMemoryCreateTool(svc *service.MemoryTaskService) *MemoryCreateTool {
	return &MemoryCreateTool{svc: svc}
}

func (t *MemoryCreateTool) Definition() service.ToolDefinition {
	return service.ToolDefinition{
		Name:        "memory_create",
		Description: "保存一条记忆信息。适用场景：1) 用户在对话中要求记住某些信息；2) 从语音转写文本中提取到值得记住的关键信息（见解、事实、偏好、事件等）。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"content": map[string]interface{}{
					"type":        "string",
					"description": "记忆内容",
				},
				"category": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"insight", "fact", "preference", "event"},
					"description": "记忆分类：insight(见解)、fact(事实)、preference(偏好)、event(事件)",
				},
				"importance": map[string]interface{}{
					"type":        "integer",
					"minimum":     1,
					"maximum":     10,
					"description": "重要程度，1-10，默认5",
				},
			},
			"required": []string{"content"},
		},
	}
}

func (t *MemoryCreateTool) Execute(ctx context.Context, userID string, args map[string]interface{}) (*service.ToolResult, error) {
	content, _ := args["content"].(string)
	if content == "" {
		return &service.ToolResult{Success: false, Error: "content is required"}, nil
	}

	category, _ := args["category"].(string)
	importance := intFromArgs(args, "importance", 5)

	mem, err := t.svc.CreateMemory(ctx, userID, content, category, importance, nil)
	if err != nil {
		return &service.ToolResult{Success: false, Error: err.Error()}, nil
	}
	return &service.ToolResult{Success: true, Data: mem}, nil
}

// ── MemoryCreateBatchTool ──

type MemoryCreateBatchTool struct {
	svc *service.MemoryTaskService
}

func NewMemoryCreateBatchTool(svc *service.MemoryTaskService) *MemoryCreateBatchTool {
	return &MemoryCreateBatchTool{svc: svc}
}

func (t *MemoryCreateBatchTool) Definition() service.ToolDefinition {
	return service.ToolDefinition{
		Name:        "memory_create_batch",
		Description: "批量保存多条记忆信息。适用于从语音转写文本中一次性提取多条记忆点，或用户要求同时记住多条信息的场景。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"memories": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"content": map[string]interface{}{
								"type":        "string",
								"description": "记忆内容",
							},
							"category": map[string]interface{}{
								"type":        "string",
								"enum":        []string{"insight", "fact", "preference", "event"},
								"description": "记忆分类",
							},
							"importance": map[string]interface{}{
								"type":        "integer",
								"minimum":     1,
								"maximum":     10,
								"description": "重要程度",
							},
						},
						"required": []string{"content"},
					},
					"description": "记忆列表",
				},
				"conversation_id": map[string]interface{}{
					"type":        "string",
					"description": "关联的对话ID（可选）",
				},
			},
			"required": []string{"memories"},
		},
	}
}

func (t *MemoryCreateBatchTool) Execute(ctx context.Context, userID string, args map[string]interface{}) (*service.ToolResult, error) {
	rawMemories, ok := args["memories"].([]interface{})
	if !ok || len(rawMemories) == 0 {
		return &service.ToolResult{Success: false, Error: "memories array is required"}, nil
	}

	var convID *string
	if cid, ok := args["conversation_id"].(string); ok && cid != "" {
		convID = &cid
	}

	var inputs []service.MemoryInput
	for _, raw := range rawMemories {
		m, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		content, _ := m["content"].(string)
		if content == "" {
			continue
		}
		category, _ := m["category"].(string)
		importance := intFromArgs(m, "importance", 5)
		inputs = append(inputs, service.MemoryInput{
			Content:    content,
			Category:   category,
			Importance: importance,
		})
	}

	if len(inputs) == 0 {
		return &service.ToolResult{Success: false, Error: "no valid memories provided"}, nil
	}

	mems, err := t.svc.CreateMemories(ctx, userID, inputs, convID)
	if err != nil {
		return &service.ToolResult{Success: false, Error: err.Error()}, nil
	}
	return &service.ToolResult{
		Success: true,
		Data:    map[string]interface{}{"count": len(mems), "memories": mems},
	}, nil
}

// ── MemorySearchTool ──

type MemorySearchTool struct {
	svc *service.MemoryTaskService
}

func NewMemorySearchTool(svc *service.MemoryTaskService) *MemorySearchTool {
	return &MemorySearchTool{svc: svc}
}

func (t *MemorySearchTool) Definition() service.ToolDefinition {
	return service.ToolDefinition{
		Name:        "memory_search",
		Description: "搜索用户的历史记忆。当用户询问过去记住的信息，或需要查找相关记忆时使用。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "搜索关键词",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "返回数量上限，默认10",
				},
			},
			"required": []string{"query"},
		},
	}
}

func (t *MemorySearchTool) Execute(ctx context.Context, userID string, args map[string]interface{}) (*service.ToolResult, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return &service.ToolResult{Success: false, Error: "query is required"}, nil
	}
	limit := intFromArgs(args, "limit", 10)

	mems, err := t.svc.SearchMemories(ctx, userID, query, limit)
	if err != nil {
		return &service.ToolResult{Success: false, Error: err.Error()}, nil
	}
	return &service.ToolResult{Success: true, Data: mems}, nil
}

// ── MemoryDeleteTool ──

type MemoryDeleteTool struct {
	svc *service.MemoryTaskService
}

func NewMemoryDeleteTool(svc *service.MemoryTaskService) *MemoryDeleteTool {
	return &MemoryDeleteTool{svc: svc}
}

func (t *MemoryDeleteTool) Definition() service.ToolDefinition {
	return service.ToolDefinition{
		Name:        "memory_delete",
		Description: "删除一条记忆。当用户要求忘记某些信息时使用。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"memory_id": map[string]interface{}{
					"type":        "string",
					"description": "要删除的记忆ID",
				},
			},
			"required": []string{"memory_id"},
		},
	}
}

func (t *MemoryDeleteTool) Execute(ctx context.Context, userID string, args map[string]interface{}) (*service.ToolResult, error) {
	memoryID, _ := args["memory_id"].(string)
	if memoryID == "" {
		return &service.ToolResult{Success: false, Error: "memory_id is required"}, nil
	}

	if err := t.svc.DeleteMemory(ctx, userID, memoryID); err != nil {
		return &service.ToolResult{Success: false, Error: err.Error()}, nil
	}
	return &service.ToolResult{Success: true, Data: fmt.Sprintf("memory %s deleted", memoryID)}, nil
}

// ── helpers ──

func intFromArgs(args map[string]interface{}, key string, defaultVal int) int {
	if v, ok := args[key].(float64); ok {
		return int(v)
	}
	if v, ok := args[key].(int); ok {
		return v
	}
	return defaultVal
}
