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

	now := time.Now()
	p := &model.Providers{
		UserId:       userID,
		Name:         req.Name,
		Description:  sql.NullString{String: req.Description, Valid: req.Description != ""},
		ProviderType: req.ProviderType,
		ApiBase:      sql.NullString{String: req.APIBase, Valid: req.APIBase != ""},
		ApiKey:       req.APIKey,
		Enabled:      req.Enabled,
		Settings:     "{}",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err = l.svcCtx.ProvidersModel.Insert(l.ctx, p)
	if err != nil {
		l.Errorf("CreateProvider failed: %v", err)
		return nil, err
	}

	result := types.Provider{
		Id:           p.Id,
		Name:         p.Name,
		Description:  nullStringToString(p.Description),
		ProviderType: p.ProviderType,
		APIBase:      nullStringToString(p.ApiBase),
		APIKey:       "***", // Mask API key
		Enabled:      p.Enabled,
		IsShared:     p.IsShared,
		CreatedBy:    p.UserId,
	}
	if user, err := l.svcCtx.UsersModel.FindOneByUserId(l.ctx, p.UserId); err == nil {
		result.CreatedByName = user.Username
	}

	return &types.CreateProviderResp{
		Provider: result,
	}, nil
}
