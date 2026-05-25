// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"time"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSystemHealthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSystemHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSystemHealthLogic {
	return &GetSystemHealthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSystemHealthLogic) GetSystemHealth() (resp *types.GetSystemHealthResp, err error) {
	// Get user ID for counting
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("failed to get user ID: %v", err)
		userID = ""
	}

	// Get provider count
	var providerCount int
	if userID != "" {
		providers, err := l.svcCtx.ProvidersModel.FindByUserID(l.ctx, userID)
		if err != nil {
			l.Errorf("failed to get providers: %v", err)
		} else {
			providerCount = len(providers)
		}
	}

	// Get session count by querying all agents first, then counting their sessions
	var sessionCount int
	if userID != "" {
		agents, err := l.svcCtx.AgentsModel.FindByUserID(l.ctx, userID)
		if err != nil {
			l.Errorf("failed to get agents: %v", err)
		} else if len(agents) > 0 {
			// Extract agent IDs
			agentIDs := make([]string, len(agents))
			for i, agent := range agents {
				agentIDs[i] = agent.Id
			}
			// Count total sessions across all agents in a single query
			count, err := l.svcCtx.SessionsModel.CountByAgentIDs(l.ctx, agentIDs)
			if err != nil {
				l.Errorf("failed to count sessions: %v", err)
			} else {
				sessionCount = count
			}
		}
	}

	return &types.GetSystemHealthResp{
		Health: types.SystemHealth{
			Version:         "0.1.0",
			Uptime:          time.Now().Unix(),
			Database:        "ok",
			Tools:           9,
			Sessions:        sessionCount,
			Providers:       providerCount,
			ChannelTotal:    0,
			ChannelOnline:   0,
			ChannelDegraded: 0,
			ChannelFailed:   0,
		},
	}, nil
}
