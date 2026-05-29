// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAgentMemoryDocumentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List memory documents for specific agent
func NewGetAgentMemoryDocumentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAgentMemoryDocumentsLogic {
	return &GetAgentMemoryDocumentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAgentMemoryDocumentsLogic) GetAgentMemoryDocuments(req *types.ListAgentMemoryDocumentsReq) (resp *types.ListMemoryDocumentsResp, err error) {
	docs, err := l.svcCtx.MemoryDocumentsModel.FindByAgentId(l.ctx, req.AgentID)
	if err != nil {
		l.Errorf("failed to list agent memory documents: %v", err)
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
			Path:      doc.Path,
			Content:   doc.Content,
			UpdatedAt: doc.UpdatedAt.Unix() * 1000,
			CreatedAt: doc.CreatedAt.Unix() * 1000,
		})
	}

	return resp, nil
}
