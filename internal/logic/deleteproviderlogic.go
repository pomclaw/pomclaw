// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteProviderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Delete provider
func NewDeleteProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProviderLogic {
	return &DeleteProviderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteProviderLogic) DeleteProvider(req *types.DeleteProviderReq) (resp *types.DeleteProviderResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("DeleteProvider failed: %v", err)
		return nil, err
	}

	// Verify ownership first
	p, err := l.svcCtx.ProvidersModel.FindOne(l.ctx, req.Id)
	if err == model.ErrNotFound || (err == nil && p.UserId != userID) {
		l.Errorf("DeleteProvider failed: provider not found")
		return nil, model.ErrNotFound
	}
	if err != nil {
		l.Errorf("DeleteProvider failed: %v", err)
		return nil, err
	}

	if err := l.svcCtx.ProvidersModel.Delete(l.ctx, req.Id); err != nil {
		l.Errorf("DeleteProvider failed: %v", err)
		return nil, err
	}

	return &types.DeleteProviderResp{}, nil
}
