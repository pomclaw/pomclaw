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

type GetProviderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get provider details
func NewGetProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProviderLogic {
	return &GetProviderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProviderLogic) GetProvider(req *types.GetProviderReq) (resp *types.GetProviderResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("GetProvider failed: %v", err)
		return nil, err
	}

	p, err := l.svcCtx.ProvidersModel.FindOne(l.ctx, req.Id)
	if err == model.ErrNotFound || (err == nil && p.UserId != userID) {
		l.Errorf("GetProvider failed: provider not found")
		return nil, model.ErrNotFound
	}
	if err != nil {
		l.Errorf("GetProvider failed: %v", err)
		return nil, err
	}

	return &types.GetProviderResp{
		Provider: types.Provider{
			Id:           p.Id,
			Name:         p.Name,
			ProviderType: p.ProviderType,
			APIBase:      nullStringToString(p.ApiBase),
			APIKey:       "***", // Mask API key
			DisplayName:  nullStringToString(p.DisplayName),
			Enabled:      p.Enabled,
		},
	}, nil
}
