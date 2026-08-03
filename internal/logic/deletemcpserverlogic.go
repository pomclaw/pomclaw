package logic

import (
	"context"
	"strconv"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteMCPServerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Delete MCP server
func NewDeleteMCPServerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMCPServerLogic {
	return &DeleteMCPServerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteMCPServerLogic) DeleteMCPServer(req *types.DeleteMCPServerReq) (resp *types.DeleteMCPServerResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("DeleteMCPServer failed: %v", err)
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
		l.Errorf("DeleteMCPServer failed: %v", err)
		return nil, err
	}

	if server.UserId != userID {
		return nil, &NotFoundError{Message: "MCP server not found"}
	}

	// DB cascade delete handles grants
	err = l.svcCtx.McpServersModel.Delete(l.ctx, id)
	if err != nil {
		l.Errorf("DeleteMCPServer failed: %v", err)
		return nil, err
	}

	return &types.DeleteMCPServerResp{}, nil
}