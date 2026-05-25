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

func (l *UpdateAgentLogic) UpdateAgent(req *types.UpdateAgentReq) (resp *types.UpdateAgentResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	agentID := req.AgentId
	if agentID == "" {
		return nil, fmt.Errorf("agent_id is required")
	}

	// Build updates map from non-empty/non-zero fields
	updates := make(map[string]interface{})
	if req.AgentKey != "" {
		updates["agent_key"] = req.AgentKey
	}
	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	}
	if req.Frontmatter != "" {
		updates["frontmatter"] = req.Frontmatter
	}
	if req.Provider != "" {
		updates["provider"] = req.Provider
	}
	if req.Model != "" {
		updates["model"] = req.Model
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.ContextWindow > 0 {
		updates["context_window"] = req.ContextWindow
	}
	if req.MaxToolIterations > 0 {
		updates["max_tool_iterations"] = req.MaxToolIterations
	}
	if req.Workspace != "" {
		updates["workspace"] = req.Workspace
	}
	if len(req.ToolsConfig) > 0 {
		updates["tools_config"] = jsonOrEmpty(req.ToolsConfig)
	}
	if len(req.MemoryConfig) > 0 {
		updates["memory_config"] = jsonOrEmpty(req.MemoryConfig)
	}
	if len(req.CompactionConfig) > 0 {
		updates["compaction_config"] = jsonOrEmpty(req.CompactionConfig)
	}
	if len(req.OtherConfig) > 0 {
		updates["other_config"] = jsonOrEmpty(req.OtherConfig)
	}
	if req.AgentDescription != "" {
		updates["agent_description"] = req.AgentDescription
	}
	if req.Emoji != "" {
		updates["emoji"] = req.Emoji
	}
	if req.ThinkingLevel != "" {
		updates["thinking_level"] = req.ThinkingLevel
	}
	if req.MaxTokens > 0 {
		updates["max_tokens"] = req.MaxTokens
	}
	if req.SelfEvolve {
		updates["self_evolve"] = req.SelfEvolve
	}
	if req.SkillEvolve {
		updates["skill_evolve"] = req.SkillEvolve
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

	return &types.UpdateAgentResp{
		Agent: *ConvertModelAgentToType(agent),
	}, nil
}
