// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// ModelInfo represents a model entry returned by the list-models endpoint
type ModelInfo struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type ListProviderModelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListProviderModelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProviderModelsLogic {
	return &ListProviderModelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListProviderModelsLogic) ListProviderModels(userID, providerID string) (types.ListProviderModelsResp, error) {
	_, err := l.svcCtx.ProvidersModel.FindOne(l.ctx, providerID)
	if err != nil {
		logx.Errorf("ListProviderModels failed: %v", err)
		return types.ListProviderModelsResp{}, err
	}

	return types.ListProviderModelsResp{Models: nil}, nil
}
