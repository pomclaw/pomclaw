package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUsageSummaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUsageSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUsageSummaryLogic {
	return &GetUsageSummaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUsageSummaryLogic) GetUsageSummary(req *types.GetUsageSummaryReq) (resp *types.GetUsageSummaryResp, err error) {
	// Parse period
	period := req.Period
	if period == "" {
		period = "24h"
	}

	// For now, return zero-filled summary (no snapshot store in pomclaw yet)
	// In future, this should aggregate usage data from the database
	emptySum := types.UsageSummary{
		Requests:      0,
		InputTokens:   0,
		OutputTokens:  0,
		Cost:          0.0,
		UniqueUsers:   0,
		Errors:        0,
		LLMCalls:      0,
		ToolCalls:     0,
		AvgDurationMS: 0,
	}

	// Log the period for debugging
	l.Infof("GetUsageSummary called with period: %s", period)

	return &types.GetUsageSummaryResp{
		Current:  emptySum,
		Previous: emptySum,
	}, nil
}
