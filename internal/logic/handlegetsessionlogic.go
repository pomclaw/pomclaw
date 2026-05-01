// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleGetSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get session details
func NewHandleGetSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleGetSessionLogic {
	return &HandleGetSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleGetSessionLogic) HandleGetSession() (resp *types.Session, err error) {
	// todo: add your logic here and delete this line

	return
}
