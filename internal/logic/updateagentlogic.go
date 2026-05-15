// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAgentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Update agent
func NewUpdateAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAgentLogic {
	return &UpdateAgentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAgentLogic) UpdateAgent(req *types.UpdateAgentReq) (resp *types.Agent, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	agentID := req.AgentId
	if agentID == "" {
		return nil, fmt.Errorf("agent_id is required")
	}

	// Build updates map from non-nil fields
	updates := make(map[string]interface{})
	if req.AgentKey != nil {
		updates["agent_key"] = *req.AgentKey
	}
	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
	}
	if req.Frontmatter != nil {
		updates["frontmatter"] = *req.Frontmatter
	}
	if req.Provider != nil {
		updates["provider"] = *req.Provider
	}
	if req.Model != nil {
		updates["model"] = *req.Model
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.ContextWindow != nil {
		updates["context_window"] = *req.ContextWindow
	}
	if req.MaxToolIterations != nil {
		updates["max_tool_iterations"] = *req.MaxToolIterations
	}
	if req.Workspace != nil {
		updates["workspace"] = *req.Workspace
	}
	if len(req.ToolsConfig) > 0 {
		updates["tools_config"] = req.ToolsConfig
	}
	if len(req.MemoryConfig) > 0 {
		updates["memory_config"] = refRawMessage(req.MemoryConfig)
	}
	if len(req.CompactionConfig) > 0 {
		updates["compaction_config"] = refRawMessage(req.CompactionConfig)
	}
	if len(req.OtherConfig) > 0 {
		updates["other_config"] = req.OtherConfig
	}
	if req.AgentDescription != nil {
		updates["agent_description"] = *req.AgentDescription
	}
	if req.Emoji != nil {
		updates["emoji"] = *req.Emoji
	}
	if req.ThinkingLevel != nil {
		updates["thinking_level"] = *req.ThinkingLevel
	}
	if req.MaxTokens != nil {
		updates["max_tokens"] = *req.MaxTokens
	}
	if req.SelfEvolve != nil {
		updates["self_evolve"] = *req.SelfEvolve
	}
	if req.SkillEvolve != nil {
		updates["skill_evolve"] = *req.SkillEvolve
	}

	err = l.svcCtx.AgentsModel.UpdateFields(l.ctx, agentID, userID, updates)
	if err == model.ErrNotFound {
		return nil, fmt.Errorf("agent not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update agent: %w", err)
	}

	// Fetch updated agent
	agent, err := l.svcCtx.AgentsModel.FindByUserAndIDOrKey(l.ctx, agentID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated agent: %w", err)
	}

	return ConvertModelAgentToType(agent), nil
}
