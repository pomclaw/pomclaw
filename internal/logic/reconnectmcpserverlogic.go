package logic

import (
	"context"
	"strconv"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type ReconnectMCPServerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Reconnect MCP server (evict connection pool)
func NewReconnectMCPServerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReconnectMCPServerLogic {
	return &ReconnectMCPServerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReconnectMCPServerLogic) ReconnectMCPServer(req *types.ReconnectMCPServerReq) (resp *types.ReconnectMCPServerResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("ReconnectMCPServer failed: %v", err)
		return nil, err
	}

	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return nil, &ValidationError{Message: "invalid server id"}
	}

	server, err := l.svcCtx.McpServersModel.FindOne(l.ctx, id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, &NotFoundError{Message: "MCP server not found"}
		}
		l.Errorf("ReconnectMCPServer failed: %v", err)
		return nil, err
	}

	if server.UserId != userID {
		return nil, &NotFoundError{Message: "MCP server not found"}
	}

	// TODO: Evict connection pool entry when pool is implemented
	// In goclaw: poolEvictor.Evict(tenantID, server.Name)

	return &types.ReconnectMCPServerResp{}, nil
}