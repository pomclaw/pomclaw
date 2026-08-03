// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/exp/slices"
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

	// 获取用户的providers + 被分享的providers
	providers, err := l.svcCtx.ProvidersModel.FindByUserIDWithShared(l.ctx, userID)
	if err != nil {
		l.Errorf("ListProviders failed: %v", err)
		return nil, err
	}

	// Batch query usernames for all unique user IDs
	uidSet := make([]string, 0, len(providers))
	for _, p := range providers {
		if !slices.Contains(uidSet, p.UserId) {
			uidSet = append(uidSet, p.UserId)
		}
	}
	userMap := make(map[string]string, len(uidSet))
	if users, err := l.svcCtx.UsersModel.FindByUserIDs(l.ctx, uidSet); err == nil {
		for _, u := range users {
			userMap[u.UserId] = u.Username
		}
	}

	providerList := make([]types.Provider, 0, len(providers))
	for _, p := range providers {
		svr := types.Provider{
			Id:           p.Id,
			Name:         p.Name,
			Description:  nullStringToString(p.Description),
			ProviderType: p.ProviderType,
			APIBase:      nullStringToString(p.ApiBase),
			APIKey:       "***",
			Enabled:      p.Enabled,
			IsShared:     p.IsShared,
			CreatedBy:    p.UserId,
		}
		if name, ok := userMap[p.UserId]; ok {
			svr.CreatedByName = name
		}
		providerList = append(providerList, svr)
	}

	return &types.ListProvidersResp{
		Total:     int64(len(providerList)),
		Providers: providerList,
	}, nil
}
