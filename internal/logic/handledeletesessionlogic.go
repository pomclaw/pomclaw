// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleDeleteSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Delete session
func NewHandleDeleteSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleDeleteSessionLogic {
	return &HandleDeleteSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleDeleteSessionLogic) HandleDeleteSession() (resp *types.HandleDeleteSessionResp, err error) {
	// todo: add your logic here and delete this line

	return
}
