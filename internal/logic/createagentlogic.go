// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/pomclaw/pomclaw/pkg/utils"

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
	provider := req.Provider
	if provider == "" {
		provider = "openrouter"
	}

	agent := &model.Agents{
		Id:                  utils.GenerateShortID(),
		AgentKey:            req.AgentKey,
		DisplayName:         sql.NullString{String: req.DisplayName, Valid: true},
		Frontmatter:         sql.NullString{String: req.Frontmatter, Valid: req.Frontmatter != ""},
		OwnerId:             userID,
		Provider:            provider,
		Model:               req.Model,
		AgentDescription:    sql.NullString{String: req.AgentDescription, Valid: req.AgentDescription != ""},
		ContextWindow:       int64(req.ContextWindow),
		MaxToolIterations:   int64(req.MaxToolIterations),
		Workspace:           req.Workspace,
		RestrictToWorkspace: true,
		ToolsConfig:         jsonOrEmpty(req.ToolsConfig),
		MemoryConfig:        jsonOrEmpty(req.MemoryConfig),
		CompactionConfig:    jsonOrEmpty(req.CompactionConfig),
		OtherConfig:         jsonOrEmpty(req.OtherConfig),
		Emoji:               sql.NullString{String: req.Emoji, Valid: req.Emoji != ""},
		ThinkingLevel:       sql.NullString{String: req.ThinkingLevel, Valid: req.ThinkingLevel != ""},
		MaxTokens:           int64(req.MaxTokens),
		SelfEvolve:          req.SelfEvolve,
		SkillEvolve:         req.SkillEvolve,
	}

	_, err = l.svcCtx.AgentsModel.Insert(l.ctx, agent)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return ConvertModelAgentToType(agent), nil
}
