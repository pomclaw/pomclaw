package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListAgentMCPServersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List all MCP servers granted to an agent
func NewListAgentMCPServersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentMCPServersLogic {
	return &ListAgentMCPServersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAgentMCPServersLogic) ListAgentMCPServers(req *types.ListAgentMCPServersReq) (resp *types.ListAgentMCPServersResp, err error) {
	if req.AgentID == "" {
		return nil, &ValidationError{Message: "agent_id is required"}
	}

	grants, err := l.svcCtx.McpAgentGrantsModel.FindByAgentID(l.ctx, req.AgentID)
	if err != nil {
		l.Errorf("ListAgentMCPServers failed: %v", err)
		return nil, err
	}

	grantList := make([]types.MCPAgentGrant, 0, len(grants))
	for _, g := range grants {
		grantList = append(grantList, convertModelMCPAgentGrantToType(g))
	}

	return &types.ListAgentMCPServersResp{
		Grants: grantList,
	}, nil
}