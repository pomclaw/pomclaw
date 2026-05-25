// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"database/sql"
	"time"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/pomclaw/pomclaw/pkg/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateProviderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create a new provider
func NewCreateProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateProviderLogic {
	return &CreateProviderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateProviderLogic) CreateProvider(req *types.CreateProviderReq) (resp *types.CreateProviderResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("CreateProvider failed: %v", err)
		return nil, err
	}

	providerID := utils.GenerateID()
	now := time.Now()
	p := &model.Providers{
		Id:           providerID,
		UserId:       userID,
		Name:         req.Name,
		ProviderType: req.ProviderType,
		ApiBase:      sql.NullString{String: req.APIBase, Valid: req.APIBase != ""},
		ApiKey:       req.APIKey,
		DisplayName:  sql.NullString{String: req.DisplayName, Valid: req.DisplayName != ""},
		Enabled:      req.Enabled,
		Settings:     "{}",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if _, err := l.svcCtx.ProvidersModel.Insert(l.ctx, p); err != nil {
		l.Errorf("CreateProvider failed: %v", err)
		return nil, err
	}

	return &types.CreateProviderResp{
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
