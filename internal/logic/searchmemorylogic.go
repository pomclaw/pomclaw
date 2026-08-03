// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"strings"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchMemoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Search memory documents
func NewSearchMemoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchMemoryLogic {
	return &SearchMemoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchMemoryLogic) SearchMemory(req *types.SearchMemoryReq) (resp *types.SearchMemoryResp, err error) {
	limit := req.Limit
	if limit == 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// Simple in-memory search using basic string matching
	results := make([]types.SearchResult, 0)

	// Query documents containing search terms
	docs, err := l.svcCtx.MemoryDocumentsModel.FindByAgentId(l.ctx, req.AgentID)
	if err != nil {
		l.Errorf("failed to list documents: %v", err)
		return nil, err
	}

	if docs != nil {
		for _, doc := range docs {
			if strings.Contains(strings.ToLower(doc.Content), strings.ToLower(req.Query)) {
				results = append(results, types.SearchResult{
					Path:    doc.Path,
					Content: truncateContent(doc.Content, 200),
					Score:   1.0,
				})
				if len(results) >= limit {
					break
				}
			}
		}
	}

	resp = &types.SearchMemoryResp{
		Results: results,
		Total:   int64(len(results)),
	}

	return resp, nil
}

func truncateContent(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
