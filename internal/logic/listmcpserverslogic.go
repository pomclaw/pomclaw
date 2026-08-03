package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/exp/slices"
)

type ListMCPServersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List all MCP servers
func NewListMCPServersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMCPServersLogic {
	return &ListMCPServersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMCPServersLogic) ListMCPServers(req *types.ListMCPServersReq) (resp *types.ListMCPServersResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("ListMCPServers failed: %v", err)
		return nil, err
	}

	servers, err := l.svcCtx.McpServersModel.FindByUserIDWithShared(l.ctx, userID)
	if err != nil {
		l.Errorf("ListMCPServers failed: %v", err)
		return nil, err
	}

	// Batch query grant counts for all servers
	ids := make([]int64, 0, len(servers))
	for _, s := range servers {
		ids = append(ids, s.Id)
	}
	counts, err := l.svcCtx.McpAgentGrantsModel.CountByServerIDs(l.ctx, ids)
	if err != nil {
		l.Errorf("ListMCPServers: failed to count grants: %v", err)
	}

	// Batch query usernames for all unique user IDs
	uidSet := make([]string, 0, len(servers))
	for _, s := range servers {
		if !slices.Contains(uidSet, s.UserId) {
			uidSet = append(uidSet, s.UserId)
		}
	}
	userMap := make(map[string]string, len(uidSet))
	if users, err := l.svcCtx.UsersModel.FindByUserIDs(l.ctx, uidSet); err == nil {
		for _, u := range users {
			userMap[u.UserId] = u.Username
		}
	}

	serverList := make([]types.MCPServer, 0, len(servers))
	for _, s := range servers {
		svr := convertModelMCPServerToType(s)
		svr.AgentCount = int(counts[s.Id])
		if name, ok := userMap[s.UserId]; ok {
			svr.CreatedByName = name
		}
		serverList = append(serverList, svr)
	}

	return &types.ListMCPServersResp{
		Total:   int64(len(serverList)),
		Servers: serverList,
	}, nil
}
