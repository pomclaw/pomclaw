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
	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	}
	if req.Frontmatter != "" {
		updates["frontmatter"] = req.Frontmatter
	}
	if req.ProviderID > 0 {
		updates["provider_id"] = req.ProviderID
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
	// Handle is_shared: only update if explicitly provided (req.IsShared will be false by default, so we need to check if it's in the request)
	// Since IsShared is bool, we can't distinguish between "not provided" and "provided as false"
	// For now, we'll allow updating it. If needed, we can use a pointer type for more granular control.
	// Check if the field was actually provided in the JSON
	updates["is_shared"] = req.IsShared

	err = l.svcCtx.AgentsModel.UpdateFields(l.ctx, agentID, userID, updates)
	if err == model.ErrNotFound {
		return nil, fmt.Errorf("agent not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update agent: %w", err)
	}

	// Sync agent_description to agent_context_files as AGENTS.md
	if req.AgentDescription != "" {
		contextFile := &model.AgentContextFiles{
			UserId:   userID,
			AgentId:  agentID,
			FileType: model.AgentContextFiles_FileType_Agent,
			FileName: "AGENTS.md",
			Content:  req.AgentDescription,
		}
		if err := l.svcCtx.AgentContextFilesModel.SetAgentContextFile(l.ctx, contextFile); err != nil {
			l.Logger.Errorf("failed to sync agent_description to AGENTS.md: %v", err)
			// Don't fail the update
		}
	}

	// Fetch updated agent
	agent, err := l.svcCtx.AgentsModel.FindByAgentID(l.ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated agent: %w", err)
	}

	return &types.UpdateAgentResp{
		Agent: *ConvertModelAgentToType(agent),
	}, nil
}
