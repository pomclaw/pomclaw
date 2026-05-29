// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteMemoryDocumentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Delete memory document
func NewDeleteMemoryDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMemoryDocumentLogic {
	return &DeleteMemoryDocumentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteMemoryDocumentLogic) DeleteMemoryDocument(req *types.DeleteMemoryDocumentReq) (resp *types.DeleteMemoryDocumentResp, err error) {
	// Delete document and associated chunks
	if err := l.svcCtx.MemoryDocumentsModel.DeleteByAgentIdAndPath(l.ctx, req.AgentID, req.Path); err != nil {
		l.Errorf("failed to delete memory document: %v", err)
		return nil, err
	}

	if err := l.svcCtx.MemoryChunksModel.DeleteByAgentIdAndPath(l.ctx, req.AgentID, req.Path); err != nil {
		l.Errorf("failed to delete memory chunks: %v", err)
		return nil, err
	}

	resp = &types.DeleteMemoryDocumentResp{
		Status: "deleted",
	}

	return resp, nil
}
