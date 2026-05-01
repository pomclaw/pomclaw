// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleCreateSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create session
func NewHandleCreateSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleCreateSessionLogic {
	return &HandleCreateSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleCreateSessionLogic) HandleCreateSession(req *types.CreateSessionReq) (resp *types.Session, err error) {
	// todo: add your logic here and delete this line

	return
}
