// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"

	"github.com/pomclaw/pomclaw/internal/store"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateAgentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create agent
func NewCreateAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAgentLogic {
	return &CreateAgentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateAgentLogic) CreateAgent(req *types.CreateAgentReq) (resp *types.Agent, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	if req.AgentKey == "" || req.DisplayName == "" || req.Model == "" {
		return nil, fmt.Errorf("agent_key, display_name and model are required")
	}

	// Build agent from request
	agent := &store.Agent{
		AgentKey:            req.AgentKey,
		DisplayName:         req.DisplayName,
		Frontmatter:         req.Frontmatter,
		OwnerID:             userID,
		Provider:            req.Provider,
		Model:               req.Model,
		AgentDescription:    req.AgentDescription,
		ContextWindow:       req.ContextWindow,
		MaxToolIterations:   req.MaxToolIterations,
		Workspace:           req.Workspace,
		RestrictToWorkspace: true,
		ToolsConfig:         req.ToolsConfig,
		MemoryConfig:        refRawMessage(req.MemoryConfig),
		CompactionConfig:    refRawMessage(req.CompactionConfig),
		OtherConfig:         req.OtherConfig,
		Emoji:               req.Emoji,
		ThinkingLevel:       req.ThinkingLevel,
		MaxTokens:           req.MaxTokens,
		SelfEvolve:          req.SelfEvolve,
		SkillEvolve:         req.SkillEvolve,
	}

	// Set defaults
	if agent.Provider == "" {
		agent.Provider = "openrouter"
	}

	err = store.CreateAgent(l.svcCtx.Conn.DB(), agent)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return ConvertStoreAgentToType(agent), nil
}
