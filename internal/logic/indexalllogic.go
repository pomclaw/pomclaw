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

type IndexAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Index all documents
func NewIndexAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IndexAllLogic {
	return &IndexAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IndexAllLogic) IndexAll(req *types.IndexAllReq) (resp *types.IndexAllResp, err error) {
	// Get all documents for the agent
	docs, err := l.svcCtx.MemoryDocumentsModel.FindByAgentId(l.ctx, req.AgentID)
	if err != nil {
		l.Errorf("failed to list documents: %v", err)
		return nil, err
	}

	totalChunks := int64(0)
	processedDocs := int64(0)

	for _, doc := range docs {
		// Delete existing chunks for this path
		_ = l.svcCtx.MemoryChunksModel.DeleteByAgentIdAndPath(l.ctx, req.AgentID, doc.Path)

		// Split content into chunks
		lines := strings.Split(doc.Content, "\n")
		chunkSize := 50 // lines per chunk
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
				Path:      doc.Path,
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
				continue
			}

			totalChunks++
		}

		processedDocs++
	}

	resp = &types.IndexAllResp{
		Status:    "indexed_all",
		Count:     totalChunks,
		Processed: processedDocs,
	}

	return resp, nil
}
