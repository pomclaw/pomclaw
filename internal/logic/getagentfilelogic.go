// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"
	"slices"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAgentFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get specific agent bootstrap file content
func NewGetAgentFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAgentFileLogic {
	return &GetAgentFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAgentFileLogic) GetAgentFile(req *types.GetAgentFileReq) (resp *types.GetAgentFileResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	agentID := req.AgentId
	if agentID == "" {
		return nil, &ValidationError{Message: "agent_id is required"}
	}

	fileName := req.Name
	if fileName == "" {
		return nil, &ValidationError{Message: "name is required"}
	}

	// Validate file name is in allowed list
	if !slices.Contains(allowedBootstrapFiles, fileName) {
		return nil, &ValidationError{Message: fmt.Sprintf("file not allowed: %s", fileName)}
	}

	// Verify agent exists and user has permission
	agent, err := l.svcCtx.AgentsModel.FindByAgentID(l.ctx, agentID)
	if err == model.ErrNotFound {
		return nil, &NotFoundError{Message: "agent not found"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	// Permission check: only owner can read agent files
	if agent.UserId != userID {
		return nil, &NotFoundError{Message: "agent not found"}
	}

	// Try to find the file in database
	dbFile, err := l.svcCtx.AgentContextFilesModel.FindOneByAgentIdFileName(l.ctx, agentID, fileName)
	if err == model.ErrNotFound {
		// File doesn't exist in database, return missing flag
		return &types.GetAgentFileResp{
			AgentId: agentID,
			File: types.BootstrapFile{
				Name:    fileName,
				Missing: true,
			},
		}, nil
	}
	if err != nil {
		l.Errorf("failed to get agent file: %v", err)
		return nil, fmt.Errorf("failed to get agent file")
	}

	return &types.GetAgentFileResp{
		AgentId: agentID,
		File: types.BootstrapFile{
			Name:    fileName,
			Missing: false,
			Size:    len(dbFile.Content),
			Content: dbFile.Content,
		},
	}, nil
}
