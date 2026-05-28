// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTracesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List traces with optional filtering
func NewListTracesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTracesLogic {
	return &ListTracesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTracesLogic) ListTraces(req *types.ListTracesReq) (resp *types.ListTracesResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	// Validate limit (max 200)
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	// Validate offset
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	traces, err := l.svcCtx.TracesModel.ListTraces(l.ctx, userID, req.AgentId, req.SessionKey, req.Status, req.Channel, limit, offset)
	if err != nil {
		l.Errorf("failed to list traces: %v", err)
		return nil, err
	}

	total, err := l.svcCtx.TracesModel.CountTraces(l.ctx, userID, req.AgentId, req.SessionKey, req.Status, req.Channel)
	if err != nil {
		l.Errorf("failed to count traces: %v", err)
		total = 0
	}

	// Convert model traces to API types
	traceList := make([]types.Trace, 0, len(traces))
	for _, t := range traces {
		traceList = append(traceList, convertModelTraceToType(&t))
	}

	return &types.ListTracesResp{
		Traces: traceList,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}
