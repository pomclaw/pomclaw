// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pomclaw/pomclaw/internal/bootstrap"

	"github.com/google/uuid"
	"github.com/pomclaw/pomclaw/internal/model"
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

func (l *CreateAgentLogic) CreateAgent(req *types.CreateAgentReq) (resp *types.CreateAgentResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	if req.DisplayName == "" || req.Model == "" || req.ProviderID == 0 {
		return nil, fmt.Errorf("display_name, model, and provider_id are required")
	}

	agent := &model.Agents{
		AgentId:             uuid.New().String(),
		Name:                sql.NullString{String: req.DisplayName, Valid: true},
		Frontmatter:         sql.NullString{String: req.Frontmatter, Valid: req.Frontmatter != ""},
		UserId:              userID,
		ProviderId:          req.ProviderID,
		Model:               req.Model,
		AgentDescription:    sql.NullString{String: req.AgentDescription, Valid: req.AgentDescription != ""},
		ContextWindow:       int64(req.ContextWindow),
		MaxToolIterations:   int64(req.MaxToolIterations),
		Workspace:           req.Workspace,
		RestrictToWorkspace: true,
		ToolsConfig:         jsonOrEmpty(nil),
		MemoryConfig:        jsonOrEmpty(nil),
		CompactionConfig:    jsonOrEmpty(nil),
		OtherConfig:         jsonOrEmpty(nil),
		Emoji:               sql.NullString{String: req.Emoji, Valid: req.Emoji != ""},
		ThinkingLevel:       sql.NullString{String: req.ThinkingLevel, Valid: req.ThinkingLevel != ""},
		MaxTokens:           int64(req.MaxTokens),
		SelfEvolve:          req.SelfEvolve,
		SkillEvolve:         req.SkillEvolve,
		IsShared:            req.IsShared,
	}

	_, err = l.svcCtx.AgentsModel.Insert(l.ctx, agent)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// Sync agent_description to agent_context_files as AGENTS.md
	// so it automatically appears in the system prompt via context file loading
	if req.AgentDescription != "" {
		contextFile := &model.AgentContextFiles{
			UserId:   userID,
			AgentId:  agent.AgentId,
			FileType: model.AgentContextFiles_FileType_Agent,
			FileName: bootstrap.AgentsFile,
			Content:  req.AgentDescription,
		}
		if err := l.svcCtx.AgentContextFilesModel.SetAgentContextFile(l.ctx, contextFile); err != nil {
			l.Logger.Errorf("failed to sync agent_description to AGENTS.md: %v", err)
			// Don't fail the create — the agent is already created
		}
	}

	return &types.CreateAgentResp{
		Agent: *ConvertModelAgentToType(agent),
	}, nil
}
