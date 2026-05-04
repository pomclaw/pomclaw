// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"

	"github.com/pomclaw/pomclaw/internal/store"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAgentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List agents
func NewListAgentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentsLogic {
	return &ListAgentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAgentsLogic) ListAgents() (resp *types.ListAgentsResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	agents, err := store.ListAgents(l.svcCtx.Conn.DB(), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	agentList := make([]types.Agent, 0, len(agents))
	for _, a := range agents {
		agentList = append(agentList, *ConvertStoreAgentToType(a))
	}

	return &types.ListAgentsResp{
		Total:  int64(len(agentList)),
		Agents: agentList,
	}, nil
}
