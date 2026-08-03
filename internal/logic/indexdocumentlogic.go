// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
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
	resp = &types.IndexDocumentResp{
		Status: "indexed",
		Count:  int64(0),
	}

	return resp, nil
}
