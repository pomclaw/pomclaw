package logic

import (
	"context"
	"strconv"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetMCPServerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get MCP server details
func NewGetMCPServerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMCPServerLogic {
	return &GetMCPServerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMCPServerLogic) GetMCPServer(req *types.GetMCPServerReq) (resp *types.GetMCPServerResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("GetMCPServer failed: %v", err)
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
		l.Errorf("GetMCPServer failed: %v", err)
		return nil, err
	}

	if server.UserId != userID && !server.IsShared {
		return nil, &NotFoundError{Message: "MCP server not found"}
	}

	result := convertModelMCPServerToType(server)
	if user, err := l.svcCtx.UsersModel.FindOneByUserId(l.ctx, server.UserId); err == nil {
		result.CreatedByName = user.Username
	}
	return &types.GetMCPServerResp{
		Server: result,
	}, nil
}