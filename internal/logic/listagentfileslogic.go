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

// allowedBootstrapFiles is the list of files exposed via agents.files.* endpoints
var allowedBootstrapFiles = []string{
	"AGENTS.md",
	"SOUL.md",
	"IDENTITY.md",
	"USER.md",
	"USER_PREDEFINED.md",
	"CAPABILITIES.md",
	"BOOTSTRAP.md",
	"MEMORY.json",
	"HEARTBEAT.md",
}

type ListAgentFilesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List agent bootstrap files (SOUL.md, AGENTS.md, etc.)
func NewListAgentFilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentFilesLogic {
	return &ListAgentFilesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAgentFilesLogic) ListAgentFiles(req *types.ListAgentFilesReq) (resp *types.ListAgentFilesResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	agentID := req.AgentId
	if agentID == "" {
		return nil, &ValidationError{Message: "agent_id is required"}
	}

	// Verify agent exists and user has permission
	agent, err := l.svcCtx.AgentsModel.FindByAgentID(l.ctx, agentID)
	if err == model.ErrNotFound {
		return nil, &NotFoundError{Message: "agent not found"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	// Permission check: only owner can manage agent files
	if agent.UserId != userID {
		return nil, &NotFoundError{Message: "agent not found"}
	}

	// Get all files for this agent from database
	dbFiles, err := l.svcCtx.AgentContextFilesModel.GetAgentContextFiles(l.ctx, agentID)
	if err != nil {
		l.Errorf("failed to list agent files: %v", err)
		return nil, fmt.Errorf("failed to list agent files")
	}

	// Build a map for quick lookup
	dbMap := make(map[string]*model.AgentContextFiles)
	for _, f := range dbFiles {
		dbMap[f.FileName] = f
	}

	// Build response with all allowed files, marking missing ones
	files := make([]types.BootstrapFile, 0, len(allowedBootstrapFiles))
	for _, name := range allowedBootstrapFiles {
		if f, ok := dbMap[name]; ok {
			files = append(files, types.BootstrapFile{
				Name:    name,
				Missing: false,
				Size:    len(f.Content),
			})
		} else {
			files = append(files, types.BootstrapFile{
				Name:    name,
				Missing: true,
			})
		}
	}

	return &types.ListAgentFilesResp{
		AgentId: agentID,
		Files:   files,
	}, nil
}
