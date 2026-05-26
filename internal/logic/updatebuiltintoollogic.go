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
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	// 使用 PostgreSQL UPSERT 语法：不存在则插入，存在则更新
	grant := &model.ToolGrants{
		UserId:    userID,
		ToolName:  req.Name,
		Enabled:   sql.NullBool{Valid: true, Bool: req.Enabled},
		UpdatedAt: time.Now(),
	}

	if err := l.svcCtx.ToolGrantsModel.Upsert(l.ctx, grant); err != nil {
		l.Logger.Errorf("failed to upsert tool grant: %v", err)
		return nil, err
	}

	resp = &types.UpdateBuiltinToolResp{
		Status: "updated",
	}

	return resp, nil
}
