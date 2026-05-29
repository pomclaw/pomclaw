// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMemoryDocumentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List all memory documents (global)
func NewListMemoryDocumentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMemoryDocumentsLogic {
	return &ListMemoryDocumentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMemoryDocumentsLogic) ListMemoryDocuments(req *types.ListMemoryDocumentsReq) (resp *types.ListMemoryDocumentsResp, err error) {
	docs, err := l.svcCtx.MemoryDocumentsModel.Finds(l.ctx)
	if err != nil {
		l.Errorf("failed to list memory documents: %v", err)
		return nil, err
	}

	resp = &types.ListMemoryDocumentsResp{
		Documents: make([]types.MemoryDocument, 0),
	}

	if docs == nil {
		return resp, nil
	}

	for _, doc := range docs {
		resp.Documents = append(resp.Documents, types.MemoryDocument{
			DocumentID: doc.Id,
			Path:       doc.Path,
			Content:    doc.Content,
			AgentId:    doc.AgentId,
			UserId:     doc.UserId,
			UpdatedAt:  doc.UpdatedAt.UnixMilli(),
			CreatedAt:  doc.CreatedAt.UnixMilli(),
		})
	}

	return resp, nil
}
