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

type UpdateProviderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Update provider
func NewUpdateProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProviderLogic {
	return &UpdateProviderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateProviderLogic) UpdateProvider(req *types.UpdateProviderReq) (resp *types.UpdateProviderResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("UpdateProvider failed: %v", err)
		return nil, err
	}

	// Verify ownership first
	p, err := l.svcCtx.ProvidersModel.FindOne(l.ctx, req.Id)
	if err == model.ErrNotFound || (err == nil && p.UserId != userID) {
		l.Errorf("UpdateProvider failed: provider not found")
		return nil, model.ErrNotFound
	}
	if err != nil {
		l.Errorf("UpdateProvider failed: %v", err)
		return nil, err
	}

	// Update the provider object
	if req.Name != "" {
		p.Name = req.Name
	}
	if req.APIBase != "" {
		p.ApiBase = sql.NullString{String: req.APIBase, Valid: true}
	}
	if req.APIKey != "" && req.APIKey != "***" {
		p.ApiKey = req.APIKey
	}
	if req.DisplayName != "" {
		p.DisplayName = sql.NullString{String: req.DisplayName, Valid: true}
	}
	// Note: Enabled is a bool in the new API, so we always apply it from the request
	p.Enabled = req.Enabled
	p.UpdatedAt = time.Now()

	if err := l.svcCtx.ProvidersModel.Update(l.ctx, p); err != nil {
		l.Errorf("UpdateProvider failed: %v", err)
		return nil, err
	}

	return &types.UpdateProviderResp{
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
