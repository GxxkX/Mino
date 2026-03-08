package service

import (
	"context"
	"fmt"
)

// ToolDefinition describes a tool's schema for the LLM (OpenAI function calling format).
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolResult is the result of executing a tool.
type ToolResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Tool is the interface that all agent tools must implement.
type Tool interface {
	Definition() ToolDefinition
	Execute(ctx context.Context, userID string, args map[string]interface{}) (*ToolResult, error)
}

// ToolRegistry manages available tools and supports filtering by scope.
type ToolRegistry struct {
	tools  map[string]Tool
	scopes map[string][]string // tool name → scopes
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools:  make(map[string]Tool),
		scopes: make(map[string][]string),
	}
}

// Register adds a tool with the given scopes (e.g. "chat", "extract", "all").
func (r *ToolRegistry) Register(t Tool, scopes ...string) {
	name := t.Definition().Name
	r.tools[name] = t
	if len(scopes) == 0 {
		scopes = []string{"all"}
	}
	r.scopes[name] = scopes
}

// Get returns a tool by name.
func (r *ToolRegistry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// Definitions returns all tool definitions (unfiltered).
func (r *ToolRegistry) Definitions() []ToolDefinition {
	var defs []ToolDefinition
	for _, t := range r.tools {
		defs = append(defs, t.Definition())
	}
	return defs
}

// DefinitionsForScope returns tool definitions matching the given scope.
func (r *ToolRegistry) DefinitionsForScope(scope string) []ToolDefinition {
	var defs []ToolDefinition
	for name, t := range r.tools {
		for _, s := range r.scopes[name] {
			if s == scope || s == "all" {
				defs = append(defs, t.Definition())
				break
			}
		}
	}
	return defs
}

// Execute runs a tool by name with the given arguments.
func (r *ToolRegistry) Execute(ctx context.Context, userID, toolName string, args map[string]interface{}) (*ToolResult, error) {
	t, ok := r.tools[toolName]
	if !ok {
		return &ToolResult{Success: false, Error: fmt.Sprintf("unknown tool: %s", toolName)}, nil
	}
	return t.Execute(ctx, userID, args)
}
