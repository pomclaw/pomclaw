// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMemoryDocumentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get specific memory document by path
func NewGetMemoryDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMemoryDocumentLogic {
	return &GetMemoryDocumentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMemoryDocumentLogic) GetMemoryDocument(req *types.GetMemoryDocumentReq) (resp *types.GetMemoryDocumentResp, err error) {
	doc, err := l.svcCtx.MemoryDocumentsModel.FindOne(l.ctx, req.DocumentID)
	if err != nil {
		l.Errorf("failed to get memory document: %v", err)
		return nil, err
	}

	resp = &types.GetMemoryDocumentResp{
		Document: types.MemoryDocument{
			DocumentID: doc.Id,
			Path:       doc.Path,
			Content:    doc.Content,
			AgentId:    doc.AgentId,
			UserId:     doc.UserId,
			UpdatedAt:  doc.UpdatedAt.UnixMilli(),
			CreatedAt:  doc.CreatedAt.UnixMilli(),
		},
	}

	return resp, nil
}
