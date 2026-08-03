// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"
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
	if err == model.ErrNotFound {
		l.Errorf("DeleteProvider failed: provider not found")
		return nil, model.ErrNotFound
	}
	if err != nil {
		l.Errorf("DeleteProvider failed: %v", err)
		return nil, err
	}

	// Check if provider is shared: shared providers cannot be deleted
	if p.IsShared {
		l.Errorf("DeleteProvider failed: cannot delete shared provider")
		return nil, fmt.Errorf("cannot delete shared provider")
	}

	// Check ownership: only allow user to delete their own provider
	if p.UserId != userID {
		l.Errorf("DeleteProvider failed: permission denied - provider not owned by user")
		return nil, fmt.Errorf("permission denied: cannot delete provider owned by another user")
	}

	if err := l.svcCtx.ProvidersModel.Delete(l.ctx, req.Id); err != nil {
		l.Errorf("DeleteProvider failed: %v", err)
		return nil, err
	}

	return &types.DeleteProviderResp{}, nil
}
