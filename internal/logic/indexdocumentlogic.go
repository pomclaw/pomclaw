// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type IndexDocumentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Index single document
func NewIndexDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IndexDocumentLogic {
	return &IndexDocumentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IndexDocumentLogic) IndexDocument(req *types.IndexDocumentReq) (resp *types.IndexDocumentResp, err error) {
	doc, err := l.svcCtx.MemoryDocumentsModel.FindOneByAgentIdPath(l.ctx, req.AgentID, req.Path)
	if err != nil {
		l.Errorf("failed to find document: %v", err)
		return nil, err
	}

	// Delete existing chunks for this path
	_ = l.svcCtx.MemoryChunksModel.DeleteByAgentIdAndPath(l.ctx, req.AgentID, req.Path)

	// Split content into chunks
	lines := strings.Split(doc.Content, "\n")
	chunkSize := 50 // lines per chunk
	chunkCount := 0
	now := time.Now()

	for i := 0; i < len(lines); i += chunkSize {
		endIdx := i + chunkSize
		if endIdx > len(lines) {
			endIdx = len(lines)
		}

		chunkText := strings.Join(lines[i:endIdx], "\n")
		hash := fmt.Sprintf("%x", md5.Sum([]byte(chunkText)))

		chunk := &model.MemoryChunks{
			AgentId:   req.AgentID,
			Path:      req.Path,
			StartLine: int64(i + 1),
			EndLine:   int64(endIdx),
			Hash:      hash,
			Text:      chunkText,
			CreatedAt: now,
			UpdatedAt: now,
		}

		_, err := l.svcCtx.MemoryChunksModel.Insert(l.ctx, chunk)
		if err != nil {
			l.Errorf("failed to insert chunk: %v", err)
			return nil, err
		}

		chunkCount++
	}

	resp = &types.IndexDocumentResp{
		Status: "indexed",
		Count:  int64(chunkCount),
	}

	return resp, nil
}
