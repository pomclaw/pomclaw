// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"

	"github.com/pomclaw/pomclaw/internal/model"
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
	now := time.Now()

	// Try to find existing document
	existing, err := l.svcCtx.MemoryDocumentsModel.FindOneByAgentIdPath(l.ctx, req.AgentID, req.Path)
	if err == nil && existing != nil {
		// Update existing
		existing.Content = req.Content
		existing.Hash = hash
		existing.UpdatedAt = now
		if err := l.svcCtx.MemoryDocumentsModel.Update(l.ctx, existing); err != nil {
			l.Errorf("failed to update memory document: %v", err)
			return nil, err
		}
	} else {
		// Insert new
		newDoc := &model.MemoryDocuments{
			AgentId:   req.AgentID,
			Path:      req.Path,
			Content:   req.Content,
			Hash:      hash,
			CreatedAt: now,
			UpdatedAt: now,
		}
		_, err := l.svcCtx.MemoryDocumentsModel.Insert(l.ctx, newDoc)
		if err != nil {
			l.Errorf("failed to insert memory document: %v", err)
			return nil, err
		}
		existing = newDoc
	}

	resp = &types.PutMemoryDocumentResp{
		Document: types.MemoryDocument{
			Path:      existing.Path,
			Content:   existing.Content,
			UpdatedAt: existing.UpdatedAt.Unix() * 1000,
			CreatedAt: existing.CreatedAt.Unix() * 1000,
		},
	}

	return resp, nil
}
