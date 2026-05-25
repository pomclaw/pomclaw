// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListProvidersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List all providers
func NewListProvidersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProvidersLogic {
	return &ListProvidersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListProvidersLogic) ListProviders(req *types.ListProvidersReq) (resp *types.ListProvidersResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("ListProviders failed: %v", err)
		return nil, err
	}

	providers, err := l.svcCtx.ProvidersModel.FindByUserID(l.ctx, userID)
	if err != nil {
		l.Errorf("ListProviders failed: %v", err)
		return nil, err
	}

	providerList := make([]types.Provider, 0, len(providers))
	for _, p := range providers {
		providerList = append(providerList, types.Provider{
			Id:           p.Id,
			Name:         p.Name,
			ProviderType: p.ProviderType,
			APIBase:      nullStringToString(p.ApiBase),
			APIKey:       "***", // Mask API key
			DisplayName:  nullStringToString(p.DisplayName),
			Enabled:      p.Enabled,
		})
	}

	return &types.ListProvidersResp{
		Total:     int64(len(providerList)),
		Providers: providerList,
	}, nil
}
