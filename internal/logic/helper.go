package logic

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/types"
)

// NotFoundError is a custom error for not found resources
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

func (e *NotFoundError) Code() int {
	return http.StatusNotFound
}

// ValidationError is a custom error for validation failures
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func (e *ValidationError) Code() int {
	return http.StatusBadRequest
}

// GetUserIDFromContext extracts the authenticated user ID from context.
// go-zero JWT middleware injects claims into context with their original types.
func GetUserIDFromContext(ctx context.Context) (string, error) {
	v := ctx.Value("userId")
	if v == nil {
		return "", fmt.Errorf("unauthorized: missing user context")
	}

	// Try string first
	if userId, ok := v.(string); ok {
		if userId == "" {
			return "", fmt.Errorf("unauthorized: empty user id")
		}
		return userId, nil
	}

	// If not string, convert to string
	userId := fmt.Sprintf("%v", v)
	if userId == "" {
		return "", fmt.Errorf("unauthorized: empty user id")
	}
	return userId, nil
}

// ConvertModelAgentToType converts model.Agents to types.Agent
func ConvertModelAgentToType(agent *model.Agents) *types.Agent {
	return &types.Agent{
		Id:                  agent.AgentId,
		DisplayName:         nullStringToString(agent.Name),
		Frontmatter:         nullStringToString(agent.Frontmatter),
		OwnerId:             agent.UserId,
		ProviderID:          agent.ProviderId,
		Model:               agent.Model,
		ContextWindow:       int(agent.ContextWindow),
		MaxToolIterations:   int(agent.MaxToolIterations),
		Workspace:           agent.Workspace,
		RestrictToWorkspace: agent.RestrictToWorkspace,
		AgentType:           "predefined", // 默认值
		IsDefault:           false,        // 默认值
		Status:              "active",     // 默认值
		Emoji:               nullStringToString(agent.Emoji),
		AgentDescription:    nullStringToString(agent.AgentDescription),
		ThinkingLevel:       nullStringToString(agent.ThinkingLevel),
		MaxTokens:           int(agent.MaxTokens),
		SelfEvolve:          agent.SelfEvolve,
		SkillEvolve:         agent.SkillEvolve,
		IsShared:            agent.IsShared,
		CreatedBy:           agent.UserId,
		CreatedAt:           agent.CreatedAt.Unix(),
		UpdatedAt:           agent.UpdatedAt.Unix(),
	}
}

// nullStringToString converts sql.NullString to string
func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// jsonOrEmpty converts json.RawMessage to valid JSON string, defaults to "{}" if empty
func jsonOrEmpty(data []byte) string {
	if len(data) == 0 {
		return "{}"
	}
	return string(data)
}

// convertModelTraceToType converts model.Traces to types.Trace
func convertModelTraceToType(t *model.Traces) types.Trace {
	var endTime int64
	if t.EndTime.Valid {
		endTime = t.EndTime.Time.Unix()
	}

	var durationMs int
	if t.DurationMs.Valid {
		durationMs = int(t.DurationMs.Int64)
	}

	return types.Trace{
		Id:                fmt.Sprintf("%d", t.Id),
		ParentTraceId:     "",
		AgentId:           t.AgentId.String,
		UserId:            t.UserId.String,
		SessionId:         t.SessionId.Int64,
		RunId:             t.RunId.String,
		StartTime:         t.StartTime.Unix(),
		EndTime:           endTime,
		DurationMs:        durationMs,
		Name:              t.Name.String,
		Channel:           t.Channel.String,
		InputPreview:      t.InputPreview.String,
		OutputPreview:     t.OutputPreview.String,
		TotalInputTokens:  int(t.TotalInputTokens),
		TotalOutputTokens: int(t.TotalOutputTokens),
		TotalCost:         t.TotalCost,
		SpanCount:         int(t.SpanCount),
		LLMCallCount:      int(t.LlmCallCount),
		ToolCallCount:     int(t.ToolCallCount),
		Status:            t.Status,
		Error:             t.Error.String,
		Metadata:          t.Metadata.String,
		Tags:              fmt.Sprintf("%v", t.Tags),
		CreatedAt:         t.CreatedAt.Unix(),
	}
}

// convertModelSpanToType converts model.Spans to types.Span
func convertModelSpanToType(s *model.Spans) types.Span {
	var endTime int64
	if s.EndTime.Valid {
		endTime = s.EndTime.Time.Unix()
	}

	var durationMs int
	if s.DurationMs.Valid {
		durationMs = int(s.DurationMs.Int64)
	}

	var inputTokens int
	if s.InputTokens.Valid {
		inputTokens = int(s.InputTokens.Int64)
	}

	var outputTokens int
	if s.OutputTokens.Valid {
		outputTokens = int(s.OutputTokens.Int64)
	}

	var totalCost float64
	if s.TotalCost.Valid {
		totalCost = s.TotalCost.Float64
	}

	return types.Span{
		Id:            fmt.Sprintf("%d", s.Id),
		TraceId:       s.TraceId,
		AgentId:       s.AgentId.String,
		SpanType:      s.SpanType,
		Name:          s.Name.String,
		StartTime:     s.StartTime.Unix(),
		EndTime:       endTime,
		DurationMs:    durationMs,
		Status:        s.Status,
		Error:         s.Error.String,
		Level:         s.Level,
		Model:         s.Model.String,
		Provider:      s.Provider.String,
		InputTokens:   inputTokens,
		OutputTokens:  outputTokens,
		TotalCost:     totalCost,
		FinishReason:  s.FinishReason.String,
		ToolName:      s.ToolName.String,
		ToolCallId:    s.ToolCallId.String,
		InputPreview:  s.InputPreview.String,
		OutputPreview: s.OutputPreview.String,
		Metadata:      s.Metadata.String,
		CreatedAt:     s.CreatedAt.Unix(),
	}
}

// convertModelMCPServerToType converts model.McpServers to types.MCPServer
func convertModelMCPServerToType(s *model.McpServers) types.MCPServer {
	return types.MCPServer{
		Id:          fmt.Sprintf("%d", s.Id),
		Name:        s.Name,
		Description: nullStringToString(s.Description),
		Transport:   s.Transport,
		Command:     nullStringToString(s.Command),
		Args:        s.Args,
		URL:         nullStringToString(s.Url),
		Headers:     s.Headers,
		Env:         s.Env,
		APIKey:      "***",
		ToolPrefix:  nullStringToString(s.ToolPrefix),
		TimeoutSec:  int(s.TimeoutSec),
		Settings:    s.Settings,
		Enabled:     s.Enabled,
		IsShared:    s.IsShared,
		CreatedBy:   s.UserId,
		CreatedAt:   s.CreatedAt.Unix(),
		UpdatedAt:   s.UpdatedAt.Unix(),
	}
}

// convertModelMCPAgentGrantToType converts model.McpAgentGrants to types.MCPAgentGrant
func convertModelMCPAgentGrantToType(g *model.McpAgentGrants) types.MCPAgentGrant {
	return types.MCPAgentGrant{
		Id:        fmt.Sprintf("%d", g.Id),
		ServerID:  fmt.Sprintf("%d", g.ServerId),
		AgentID:   g.AgentId,
		Enabled:   g.Enabled,
		ToolAllow: nullStringToString(g.ToolAllow),
		ToolDeny:  nullStringToString(g.ToolDeny),
		GrantedBy: g.CreatedBy,
		CreatedAt: g.CreatedAt.Unix(),
	}
}
