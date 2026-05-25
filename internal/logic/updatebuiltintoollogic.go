// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateBuiltinToolLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Update built-in tool (enabled, settings)
func NewUpdateBuiltinToolLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBuiltinToolLogic {
	return &UpdateBuiltinToolLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateBuiltinToolLogic) UpdateBuiltinTool(req *types.UpdateBuiltinToolReq) (resp *types.UpdateBuiltinToolResp, err error) {
	// TODO: Implement updating builtin tool in storage/service
	resp = &types.UpdateBuiltinToolResp{
		Status: "updated",
	}

	return
}
