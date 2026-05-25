// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListBuiltinToolsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List all built-in tools
func NewListBuiltinToolsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListBuiltinToolsLogic {
	return &ListBuiltinToolsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListBuiltinToolsLogic) ListBuiltinTools(req *types.ListBuiltinToolsReq) (resp *types.ListBuiltinToolsResp, err error) {
	resp = &types.ListBuiltinToolsResp{
		Tools: []types.BuiltinToolDef{
			{
				Name:    "test",
				Display: "testtest",
				Desc:    "desc",
				Enabled: true,
			},
		},
	}

	return
}
