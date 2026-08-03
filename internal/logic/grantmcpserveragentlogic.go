package logic

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GrantMCPServerAgentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Grant MCP server access to an agent
func NewGrantMCPServerAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GrantMCPServerAgentLogic {
	return &GrantMCPServerAgentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GrantMCPServerAgentLogic) GrantMCPServerAgent(req *types.GrantMCPServerAgentReq) (resp *types.GrantMCPServerAgentResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("GrantMCPServerAgent failed: %v", err)
		return nil, err
	}

	if req.AgentID == "" {
		return nil, &ValidationError{Message: "agent_id is required"}
	}

	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return nil, &ValidationError{Message: "invalid server id"}
	}

	// Verify server ownership
	server, err := l.svcCtx.McpServersModel.FindOne(l.ctx, id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, &NotFoundError{Message: "MCP server not found"}
		}
		l.Errorf("GrantMCPServerAgent failed: %v", err)
		return nil, err
	}
	if server.UserId != userID {
		return nil, &NotFoundError{Message: "MCP server not found"}
	}

	// Check if grant already exists
	existing, err := l.svcCtx.McpAgentGrantsModel.FindOneByServerIdAgentId(l.ctx, id, req.AgentID)
	if err != nil && err != model.ErrNotFound {
		l.Errorf("GrantMCPServerAgent failed: %v", err)
		return nil, err
	}
	if existing != nil {
		// Update existing grant
		existing.ToolAllow = sql.NullString{String: req.ToolAllow, Valid: req.ToolAllow != ""}
		existing.ToolDeny = sql.NullString{String: req.ToolDeny, Valid: req.ToolDeny != ""}
		existing.Enabled = true
		err = l.svcCtx.McpAgentGrantsModel.Update(l.ctx, existing)
		if err != nil {
			l.Errorf("GrantMCPServerAgent failed to update: %v", err)
			return nil, err
		}
		return &types.GrantMCPServerAgentResp{
			Grant: convertModelMCPAgentGrantToType(existing),
		}, nil
	}

	// Create new grant
	now := time.Now()
	grant := &model.McpAgentGrants{
		UserId:    userID,
		ServerId:  id,
		AgentId:   req.AgentID,
		Enabled:   true,
		CreatedBy: userID,
		CreatedAt: now,
	}
	if req.ToolAllow != "" {
		grant.ToolAllow = sql.NullString{String: req.ToolAllow, Valid: true}
	}
	if req.ToolDeny != "" {
		grant.ToolDeny = sql.NullString{String: req.ToolDeny, Valid: true}
	}

	_, err = l.svcCtx.McpAgentGrantsModel.Insert(l.ctx, grant)
	if err != nil {
		l.Errorf("GrantMCPServerAgent failed: %v", err)
		return nil, err
	}

	// Read back to get the generated ID
	created, err := l.svcCtx.McpAgentGrantsModel.FindOneByServerIdAgentId(l.ctx, id, req.AgentID)
	if err != nil {
		l.Errorf("GrantMCPServerAgent failed to read back: %v", err)
		return nil, err
	}

	return &types.GrantMCPServerAgentResp{
		Grant: convertModelMCPAgentGrantToType(created),
	}, nil
}