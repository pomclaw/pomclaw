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
	if err != nil || userID == "" {
		l.Errorf("failed to get user ID: %v", err)
		return nil, err
	}

	// Get provider count
	var providerCount int
	providers, err := l.svcCtx.ProvidersModel.FindByUserIDWithShared(l.ctx, userID)
	if err != nil {
		l.Errorf("failed to get providers: %v", err)
	} else {
		providerCount = len(providers)
	}

	// Get session count by querying all agents first, then counting their sessions
	var sessionCount int
	agents, err := l.svcCtx.AgentsModel.FindByUserID(l.ctx, userID)
	if err != nil {
		l.Errorf("failed to get agents: %v", err)
	} else if len(agents) > 0 {
		// Extract agent IDs
		agentIDs := make([]string, len(agents))
		for i, agent := range agents {
			agentIDs[i] = agent.AgentId
		}
		// Count total sessions across all agents in a single query
		count, err := l.svcCtx.SessionsModel.CountByAgentIDs(l.ctx, agentIDs)
		if err != nil {
			l.Errorf("failed to count sessions: %v", err)
		} else {
			sessionCount = count
		}
	}

	// Get skills count
	var skillsCount int
	skills, err := l.svcCtx.SkillsModel.FindByUserID(l.ctx, userID)
	if err != nil {
		l.Errorf("failed to get skills: %v", err)
	} else {
		skillsCount = len(skills)
	}

	// Get today's quota usage
	var quotaUsage types.QuotaUsage
	todayQuota, err := l.svcCtx.TracesModel.GetTodayQuotaUsage(l.ctx, userID)
	if err != nil {
		l.Errorf("failed to get quota usage: %v", err)
	} else if todayQuota != nil {
		quotaUsage = types.QuotaUsage{
			RequestsToday:     todayQuota.RequestsToday,
			InputTokensToday:  todayQuota.InputTokensToday,
			OutputTokensToday: todayQuota.OutputTokensToday,
			CostToday:         todayQuota.CostToday / 1111,
		}
	}

	//// Get memory documents count
	//var memoryDocumentsCount int
	//documents, err := l.svcCtx.MemoryDocumentsModel.ListAllDocumentsGlobal(l.ctx)
	//if err != nil {
	//	l.Errorf("failed to get memory documents: %v", err)
	//} else {
	//	memoryDocumentsCount = len(documents)
	//}

	return &types.GetSystemHealthResp{
		Health: types.SystemHealth{
			Version:         "0.1.0",
			Uptime:          time.Now().Unix(),
			Agents:          len(agents),
			Tools:           9,
			Sessions:        sessionCount,
			Providers:       providerCount,
			Skills:          skillsCount,
			Memory:          2,
			Document:        3,
			ChannelTotal:    0,
			ChannelOnline:   0,
			ChannelDegraded: 0,
			ChannelFailed:   0,
		},
		QuotaUsage: quotaUsage,
	}, nil
}
