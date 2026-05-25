// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTenantConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Delete tenant-specific configuration for a tool
func NewDeleteTenantConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTenantConfigLogic {
	return &DeleteTenantConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTenantConfigLogic) DeleteTenantConfig(req *types.DeleteTenantConfigReq) (resp *types.DeleteTenantConfigResp, err error) {
	l.Infof("DeleteTenantConfig called with name: %s", req.Name)

	// TODO: Implement deleting tenant config from storage/service
	resp = &types.DeleteTenantConfigResp{
		Status: "deleted",
	}

	return
}
