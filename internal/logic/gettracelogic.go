// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTraceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get trace details with spans
func NewGetTraceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTraceLogic {
	return &GetTraceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTraceLogic) GetTrace(req *types.GetTraceReq) (resp *types.GetTraceResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	trace, err := l.svcCtx.TracesModel.FindByUUID(l.ctx, req.TraceId)
	if err != nil {
		l.Errorf("trace not found: %s, error: %v", req.TraceId, err)
		return nil, fmt.Errorf("trace not found")
	}

	// Check permission: non-admin users can only access their own traces
	if trace.UserId.Valid && trace.UserId.String != userID {
		l.Errorf("permission denied: user %s trying to access trace %s owned by %s", userID, req.TraceId, trace.UserId.String)
		return nil, fmt.Errorf("trace not found")
	}

	// Load spans from spans table
	var spans []types.Span
	if trace.TraceId.Valid && trace.TraceId.String != "" {
		spanModels, err := l.svcCtx.SpansModel.FindByTraceId(l.ctx, trace.TraceId.String)
		if err != nil {
			l.Errorf("failed to load spans for trace %s: %v", req.TraceId, err)
			return nil, fmt.Errorf("failed to load spans")
		}

		spans = make([]types.Span, 0, len(spanModels))
		for _, spanModel := range spanModels {
			spans = append(spans, convertModelSpanToType(spanModel))
		}
	} else {
		spans = []types.Span{}
	}

	return &types.GetTraceResp{
		Trace: convertModelTraceToType(trace),
		Spans: spans,
	}, nil
}
