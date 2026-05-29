// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PutMemoryDocumentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create or update memory document
func NewPutMemoryDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PutMemoryDocumentLogic {
	return &PutMemoryDocumentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PutMemoryDocumentLogic) PutMemoryDocument(req *types.PutMemoryDocumentReq) (resp *types.PutMemoryDocumentResp, err error) {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(req.Content)))

	// Find existing document by ID
	existing, err := l.svcCtx.MemoryDocumentsModel.FindOne(l.ctx, req.DocumentID)
	if err != nil {
		l.Errorf("failed to find memory document: %v", err)
		return nil, err
	}

	// Update existing
	existing.Content = req.Content
	existing.Hash = hash
	existing.UpdatedAt = time.Now()
	if err := l.svcCtx.MemoryDocumentsModel.Update(l.ctx, existing); err != nil {
		l.Errorf("failed to update memory document: %v", err)
		return nil, err
	}

	resp = &types.PutMemoryDocumentResp{
		Document: types.MemoryDocument{
			DocumentID: existing.Id,
			Path:       existing.Path,
			Content:    existing.Content,
			AgentId:    existing.AgentId,
			UserId:     existing.UserId,
			UpdatedAt:  existing.UpdatedAt.UnixMilli(),
			CreatedAt:  existing.CreatedAt.UnixMilli(),
		},
	}

	return resp, nil
}
