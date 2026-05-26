// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/pomclaw/pomclaw/pkg/contracts"
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
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	tools := l.svcCtx.ToolsManager.GetToolsToolDef(l.ctx, userID, contracts.DefaultAgentID)

	resp = &types.ListBuiltinToolsResp{
		Tools: make([]types.BuiltinToolDef, 0, len(tools)),
	}

	for i := range tools {
		resp.Tools = append(resp.Tools, types.BuiltinToolDef{
			Name:    tools[i].Name,
			Display: tools[i].Name,
			Desc:    tools[i].Desc,
			Enabled: tools[i].Enabled,
		})
	}

	return resp, nil
}
