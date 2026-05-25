// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get tenant-specific configuration for a tool
func NewGetTenantConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantConfigLogic {
	return &GetTenantConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantConfigLogic) GetTenantConfig(req *types.GetTenantConfigReq) (resp *types.GetTenantConfigResp, err error) {
	l.Infof("GetTenantConfig called with name: %s", req.Name)

	// TODO: Implement getting tenant config from storage/service
	resp = &types.GetTenantConfigResp{
		Config: types.TenantToolConfig{
			ToolName: req.Name,
		},
	}

	return
}
