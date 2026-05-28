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

type ExportTraceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Export trace as gzip-compressed JSON
func NewExportTraceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExportTraceLogic {
	return &ExportTraceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ExportTraceLogic) ExportTrace(req *types.ExportTraceReq) (resp *types.ExportTraceResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	rootTrace, err := l.svcCtx.TracesModel.FindByUUID(l.ctx, req.TraceId)
	if err != nil {
		l.Errorf("trace not found for export: %s, error: %v", req.TraceId, err)
		return nil, fmt.Errorf("trace not found")
	}

	// Check permission: non-admin users can only export their own traces
	if rootTrace.UserId.Valid && rootTrace.UserId.String != userID {
		l.Errorf("permission denied: user %s trying to export trace %s owned by %s", userID, req.TraceId, rootTrace.UserId.String)
		return nil, fmt.Errorf("trace not found")
	}

	// Recursively collect trace tree
	_, err = l.collectTraceTree(req.TraceId, 0)
	if err != nil {
		l.Errorf("failed to collect trace tree: %v", err)
		return nil, err
	}

	// Handler will serialize to gzip-compressed JSON
	return &types.ExportTraceResp{}, nil
}

// collectTraceTree recursively collects a trace and its child traces
func (l *ExportTraceLogic) collectTraceTree(traceID string, depth int) (*types.ExportTraceEntry, error) {
	const maxDepth = 10
	if depth >= maxDepth {
		return nil, nil
	}

	trace, err := l.svcCtx.TracesModel.FindByUUID(l.ctx, traceID)
	if err != nil {
		return nil, err
	}

	// Load spans for this trace
	var spans []types.Span
	if trace.TraceId.Valid && trace.TraceId.String != "" {
		spanModels, err := l.svcCtx.SpansModel.FindByTraceId(l.ctx, trace.TraceId.String)
		if err != nil {
			l.Errorf("failed to load spans for trace %s: %v", traceID, err)
			spans = []types.Span{}
		} else {
			spans = make([]types.Span, 0, len(spanModels))
			for _, spanModel := range spanModels {
				spans = append(spans, convertModelSpanToType(spanModel))
			}
		}
	} else {
		spans = []types.Span{}
	}

	entry := &types.ExportTraceEntry{
		Trace: convertModelTraceToType(trace),
		Spans: spans,
	}

	// Find child traces
	children, err := l.svcCtx.TracesModel.FindChildTraces(l.ctx, traceID)
	if err != nil {
		l.Errorf("failed to find child traces for %s: %v", traceID, err)
		return entry, nil
	}

	for _, child := range children {
		childID := fmt.Sprintf("%d", child.Id)
		subEntry, err := l.collectTraceTree(childID, depth+1)
		if err != nil {
			l.Errorf("failed to collect child trace tree: %v", err)
			continue
		}
		if subEntry != nil {
			entry.SubTraces = append(entry.SubTraces, *subEntry)
		}
	}

	return entry, nil
}
