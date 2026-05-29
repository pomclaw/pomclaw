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

type ListMemoryChunksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List memory chunks for agent
func NewListMemoryChunksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMemoryChunksLogic {
	return &ListMemoryChunksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMemoryChunksLogic) ListMemoryChunks(req *types.ListMemoryChunksReq) (resp *types.ListMemoryChunksResp, err error) {
	chunks, err := l.svcCtx.MemoryChunksModel.FindByAgentIdAndPath(l.ctx, req.AgentID, req.Path)
	if err != nil {
		l.Errorf("failed to list memory chunks: %v", err)
		return nil, err
	}

	resp = &types.ListMemoryChunksResp{
		Chunks: make([]types.MemoryChunk, 0),
	}

	if chunks == nil {
		return resp, nil
	}

	for _, chunk := range chunks {
		memChunk := types.MemoryChunk{
			ID:           fmt.Sprintf("%d", chunk.Id),
			StartLine:    chunk.StartLine,
			EndLine:      chunk.EndLine,
			TextPreview:  chunk.Text,
			HasEmbedding: chunk.Embedding.Valid,
		}

		resp.Chunks = append(resp.Chunks, memChunk)
	}

	return resp, nil
}
