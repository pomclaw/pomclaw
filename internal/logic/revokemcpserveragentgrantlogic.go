package logic

import (
	"context"
	"strconv"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type RevokeMCPServerAgentGrantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Revoke MCP server access from an agent
func NewRevokeMCPServerAgentGrantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeMCPServerAgentGrantLogic {
	return &RevokeMCPServerAgentGrantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RevokeMCPServerAgentGrantLogic) RevokeMCPServerAgentGrant(req *types.RevokeMCPServerAgentGrantReq) (resp *types.RevokeMCPServerAgentGrantResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("RevokeMCPServerAgentGrant failed: %v", err)
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
		l.Errorf("RevokeMCPServerAgentGrant failed: %v", err)
		return nil, err
	}
	if server.UserId != userID {
		return nil, &NotFoundError{Message: "MCP server not found"}
	}

	err = l.svcCtx.McpAgentGrantsModel.DeleteByServerIDAgentID(l.ctx, id, req.AgentID)
	if err != nil {
		l.Errorf("RevokeMCPServerAgentGrant failed: %v", err)
		return nil, err
	}

	return &types.RevokeMCPServerAgentGrantResp{}, nil
}