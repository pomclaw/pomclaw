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
	if err == model.ErrNotFound {
		l.Errorf("GetProvider failed: provider not found")
		return nil, model.ErrNotFound
	}
	if err != nil {
		l.Errorf("GetProvider failed: %v", err)
		return nil, err
	}

	// 权限检查：用户只能访问自己的provider，或者被分享的provider
	if p.UserId != userID && !p.IsShared {
		l.Errorf("GetProvider failed: provider not found for user")
		return nil, model.ErrNotFound
	}

	// 分享的provider不返回API key
	apiKey := p.ApiKey
	if p.IsShared {
		apiKey = "***"
	}

	result := types.Provider{
		Id:           p.Id,
		Name:         p.Name,
		Description:  nullStringToString(p.Description),
		ProviderType: p.ProviderType,
		APIBase:      nullStringToString(p.ApiBase),
		APIKey:       apiKey,
		Enabled:      p.Enabled,
		IsShared:     p.IsShared,
		CreatedBy:    p.UserId,
	}
	if user, err := l.svcCtx.UsersModel.FindOneByUserId(l.ctx, p.UserId); err == nil {
		result.CreatedByName = user.Username
	}

	return &types.GetProviderResp{
		Provider: result,
	}, nil
}
