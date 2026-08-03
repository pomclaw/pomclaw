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

type SetAgentFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Set agent bootstrap file content
func NewSetAgentFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetAgentFileLogic {
	return &SetAgentFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetAgentFileLogic) SetAgentFile(req *types.SetAgentFileReq) (resp *types.SetAgentFileResp, err error) {
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

	// Permission check: only owner can write agent files
	if agent.UserId != userID {
		return nil, &NotFoundError{Message: "agent not found"}
	}

	// Set/update the file in database
	dbFile := &model.AgentContextFiles{
		UserId:   userID,
		AgentId:  agentID,
		FileType: 0,
		FileName: fileName,
		Content:  req.Content,
	}

	if err := l.svcCtx.AgentContextFilesModel.SetAgentContextFile(l.ctx, dbFile); err != nil {
		l.Errorf("failed to set agent file: %v", err)
		return nil, fmt.Errorf("failed to save agent file")
	}

	// TODO: Implement propagate functionality if needed
	// For now, we just return 0 for propagated count
	propagated := 0

	return &types.SetAgentFileResp{
		AgentId:    agentID,
		File: types.BootstrapFile{
			Name:    fileName,
			Missing: false,
			Size:    len(req.Content),
			Content: req.Content,
		},
		Propagated: propagated,
	}, nil
}
