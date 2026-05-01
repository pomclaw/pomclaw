package tools

import "context"

const DefaultAgentID = "default"

type contextKey string

const (
	agentIDKey  contextKey = "agent_id"
	workspaceKey contextKey = "workspace"
)

// WithAgentID returns a new context with the given agentID injected.
func WithAgentID(ctx context.Context, agentID string) context.Context {
	return context.WithValue(ctx, agentIDKey, agentID)
}

// AgentIDFromContext extracts the agentID from context.
// Falls back to DefaultAgentID if not set.
func AgentIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(agentIDKey).(string); ok && v != "" {
		return v
	}
	return DefaultAgentID
}

// WithWorkspace returns a new context with the given workspace injected.
func WithWorkspace(ctx context.Context, workspace string) context.Context {
	return context.WithValue(ctx, workspaceKey, workspace)
}

// WorkspaceFromContext extracts the workspace from context.
func WorkspaceFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(workspaceKey).(string); ok {
		return v
	}
	return ""
}
