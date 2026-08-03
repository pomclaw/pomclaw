package logic

import (
	"context"
	"strconv"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListMCPServerGrantsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List all grants for an MCP server
func NewListMCPServerGrantsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMCPServerGrantsLogic {
	return &ListMCPServerGrantsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMCPServerGrantsLogic) ListMCPServerGrants(req *types.ListMCPServerGrantsReq) (resp *types.ListMCPServerGrantsResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("ListMCPServerGrants failed: %v", err)
		return nil, err
	}

	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return nil, &ValidationError{Message: "invalid server id"}
	}

	// Verify ownership
	server, err := l.svcCtx.McpServersModel.FindOne(l.ctx, id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, &NotFoundError{Message: "MCP server not found"}
		}
		l.Errorf("ListMCPServerGrants failed: %v", err)
		return nil, err
	}
	if server.UserId != userID {
		return nil, &NotFoundError{Message: "MCP server not found"}
	}

	grants, err := l.svcCtx.McpAgentGrantsModel.FindByServerID(l.ctx, id)
	if err != nil {
		l.Errorf("ListMCPServerGrants failed: %v", err)
		return nil, err
	}

	grantList := make([]types.MCPAgentGrant, 0, len(grants))
	for _, g := range grants {
		grantList = append(grantList, convertModelMCPAgentGrantToType(g))
	}

	return &types.ListMCPServerGrantsResp{
		AgentGrants: grantList,
	}, nil
}
