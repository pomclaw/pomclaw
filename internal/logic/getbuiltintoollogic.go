// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBuiltinToolLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get built-in tool details
func NewGetBuiltinToolLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBuiltinToolLogic {
	return &GetBuiltinToolLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBuiltinToolLogic) GetBuiltinTool(req *types.GetBuiltinToolReq) (resp *types.GetBuiltinToolResp, err error) {
	l.Infof("GetBuiltinTool called with name: %s", req.Name)

	// TODO: Implement getting builtin tool from storage/service
	resp = &types.GetBuiltinToolResp{
		Tool: types.BuiltinToolDef{
			Name:    req.Name,
			Enabled: false,
		},
	}

	return
}
