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

type ListProvidersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListProvidersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProvidersLogic {
	return &ListProvidersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListProvidersLogic) ListProviders(userID string) (*types.ProvidersResp, error) {
	providers, err := l.svcCtx.ProvidersModel.FindByUserID(l.ctx, userID)
	if err != nil {
		l.Errorf("ListProviders failed: %v", err)
		return nil, err
	}

	resp := make([]types.ProviderResp, 0, len(providers))
	for _, p := range providers {
		resp = append(resp, types.ProviderResp{
			ID:           p.Id,
			Name:         p.Name,
			ProviderType: p.ProviderType,
			APIBase:      nullStringToString(p.ApiBase),
			APIKey:       "***", // Mask API key
			DisplayName:  nullStringToString(p.DisplayName),
			Enabled:      p.Enabled,
		})
	}

	return &types.ProvidersResp{Providers: resp}, nil
}

type CreateProviderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateProviderLogic {
	return &CreateProviderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateProviderLogic) CreateProvider(userID string, req *types.CreateProviderReq) (*types.ProviderResp, error) {
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

	return &types.ProviderResp{
		ID:           p.Id,
		Name:         p.Name,
		ProviderType: p.ProviderType,
		APIBase:      nullStringToString(p.ApiBase),
		APIKey:       "***", // Mask API key
		DisplayName:  nullStringToString(p.DisplayName),
		Enabled:      p.Enabled,
	}, nil
}

type GetProviderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProviderLogic {
	return &GetProviderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetProviderLogic) GetProvider(req types.GetProviderReq) (*types.GetProviderResp, error) {
	p, err := l.svcCtx.ProvidersModel.FindOne(l.ctx, req.ID)
	if err == model.ErrNotFound || (err == nil && p.UserId != req.UserID) {
		l.Errorf("GetProvider failed: provider not found")
		return nil, model.ErrNotFound
	}
	if err != nil {
		l.Errorf("GetProvider failed: %v", err)
		return nil, err
	}

	return &types.GetProviderResp{
		ID:           p.Id,
		Name:         p.Name,
		ProviderType: p.ProviderType,
		APIBase:      nullStringToString(p.ApiBase),
		APIKey:       "***", // Mask API key
		DisplayName:  nullStringToString(p.DisplayName),
		Enabled:      p.Enabled,
	}, nil
}

type UpdateProviderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProviderLogic {
	return &UpdateProviderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateProviderLogic) UpdateProvider(userID, id string, req *types.UpdateProviderReq) error {
	// Verify ownership first
	p, err := l.svcCtx.ProvidersModel.FindOne(l.ctx, id)
	if err == model.ErrNotFound || (err == nil && p.UserId != userID) {
		return model.ErrNotFound
	}
	if err != nil {
		l.Errorf("UpdateProvider failed: %v", err)
		return err
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.APIBase != "" {
		updates["api_base"] = req.APIBase
	}
	if req.APIKey != "" && req.APIKey != "***" {
		updates["api_key"] = req.APIKey
	}
	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	updates["updated_at"] = time.Now()

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
	if req.Enabled != nil {
		p.Enabled = *req.Enabled
	}
	p.UpdatedAt = time.Now()

	if err := l.svcCtx.ProvidersModel.Update(l.ctx, p); err != nil {
		l.Errorf("UpdateProvider failed: %v", err)
		return err
	}
	return nil
}

type DeleteProviderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProviderLogic {
	return &DeleteProviderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteProviderLogic) DeleteProvider(userID, id string) error {
	// Verify ownership first
	p, err := l.svcCtx.ProvidersModel.FindOne(l.ctx, id)
	if err == model.ErrNotFound || (err == nil && p.UserId != userID) {
		return model.ErrNotFound
	}
	if err != nil {
		l.Errorf("DeleteProvider failed: %v", err)
		return err
	}

	if err := l.svcCtx.ProvidersModel.Delete(l.ctx, id); err != nil {
		l.Errorf("DeleteProvider failed: %v", err)
		return err
	}
	return nil
}
