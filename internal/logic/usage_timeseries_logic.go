package logic

import (
	"context"
	"time"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUsageTimeSeriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUsageTimeSeriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUsageTimeSeriesLogic {
	return &GetUsageTimeSeriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUsageTimeSeriesLogic) GetUsageTimeSeries(req *types.GetUsageTimeSeriesReq) (resp *types.GetUsageTimeSeriesResp, err error) {
	// Parse from/to timestamps
	from, err := time.Parse(time.RFC3339, req.From)
	if err != nil {
		l.Errorf("invalid from timestamp: %v", err)
		return nil, err
	}

	to, err := time.Parse(time.RFC3339, req.To)
	if err != nil {
		l.Errorf("invalid to timestamp: %v", err)
		return nil, err
	}

	// Normalize group_by
	groupBy := req.GroupBy
	if groupBy == "" {
		groupBy = "hour"
	}

	// For now, return empty points (no snapshot store in pomclaw yet)
	// In future, this should query the database for usage metrics
	points := []types.UsageTimeSeriesPoint{}

	// Generate zero-filled points for the requested time range
	// This provides a consistent response shape even with no data
	current := from.Truncate(parseGroupByDuration(groupBy))
	for current.Before(to) {
		points = append(points, types.UsageTimeSeriesPoint{
			BucketTime:    current.UTC().Format(time.RFC3339),
			RequestCount:  0,
			ErrorCount:    0,
			UniqueUsers:   0,
			InputTokens:   0,
			OutputTokens:  0,
			TotalCost:     0.0,
			LLMCallCount:  0,
			ToolCallCount: 0,
			AvgDurationMS: 0,
		})
		current = current.Add(parseGroupByDuration(groupBy))
	}

	return &types.GetUsageTimeSeriesResp{
		Points: points,
	}, nil
}

func parseGroupByDuration(groupBy string) time.Duration {
	switch groupBy {
	case "day":
		return 24 * time.Hour
	case "hour":
		return time.Hour
	case "minute":
		return time.Minute
	default:
		return time.Hour
	}
}
