// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCostSummaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get cost summary by date, agent, model, or provider
func NewGetCostSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCostSummaryLogic {
	return &GetCostSummaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCostSummaryLogic) GetCostSummary(req *types.GetCostSummaryReq) (resp *types.GetCostSummaryResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	// Parse timestamp filters
	var fromTime, toTime *time.Time
	if req.From != "" {
		if t, err := time.Parse(time.RFC3339, req.From); err == nil {
			fromTime = &t
		}
	}
	if req.To != "" {
		if t, err := time.Parse(time.RFC3339, req.To); err == nil {
			toTime = &t
		}
	}

	// TODO: Implement cost summary aggregation when traces table is fully populated
	// For now, return empty summary
	rows := []types.CostSummaryRow{}

	// Build summary by agent if agent_id is specified
	if req.AgentId != "" {
		// Query traces for this agent and aggregate costs
		traces, err := l.svcCtx.TracesModel.ListTraces(l.ctx, userID, req.AgentId, "", "", "", 10000, 0)
		if err != nil {
			l.Errorf("failed to list traces for cost summary: %v", err)
			return &types.GetCostSummaryResp{Rows: rows}, nil
		}

		// Aggregate by date
		costByDate := make(map[string]types.CostSummaryRow)
		for _, t := range traces {
			// Filter by time range if specified
			if fromTime != nil && t.StartTime.Before(*fromTime) {
				continue
			}
			if toTime != nil && t.StartTime.After(*toTime) {
				continue
			}

			dateKey := t.StartTime.Format("2006-01-02")
			row, exists := costByDate[dateKey]
			if !exists {
				row = types.CostSummaryRow{
					AgentId: fmt.Sprintf("%d", t.Id), // Using trace ID as placeholder
				}
			}

			row.TotalCost += t.TotalCost
			row.TotalInputTokens += int(t.TotalInputTokens)
			row.TotalOutputTokens += int(t.TotalOutputTokens)
			row.TraceCount++

			costByDate[dateKey] = row
		}

		for _, row := range costByDate {
			rows = append(rows, row)
		}
	}

	return &types.GetCostSummaryResp{Rows: rows}, nil
}
