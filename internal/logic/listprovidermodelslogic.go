// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"encoding/json"

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

func (l *ListProviderModelsLogic) ListProviderModels(userID string, req types.ListProviderModelsReq) (types.ListProviderModelsResp, error) {
	provider, err := l.svcCtx.ProvidersModel.FindOne(l.ctx, req.Id)
	if err != nil {
		logx.Errorf("ListProviderModels failed: %v", err)
		return types.ListProviderModelsResp{}, err
	}

	if len(provider.Settings) != 0 {
		res := types.ListProviderModelsResp{}
		err = json.Unmarshal([]byte(provider.Settings), &res)
		if err == nil {
			return res, nil
		}
	}

	return types.ListProviderModelsResp{Models: nil}, nil
}
