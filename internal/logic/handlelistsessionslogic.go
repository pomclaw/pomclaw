// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleListSessionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List sessions
func NewHandleListSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleListSessionsLogic {
	return &HandleListSessionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleListSessionsLogic) HandleListSessions() (resp *types.ListSessionsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
