// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pomclaw/pomclaw/internal/store"
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

func (l *HandleDeleteSessionLogic) HandleDeleteSession(req *types.HandleDeleteSessionReq) (resp *types.HandleDeleteSessionResp, err error) {
	sessionID := req.Id
	if sessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}

	err = store.DeleteSession(l.svcCtx.Conn.DB(), sessionID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to delete session: %w", err)
	}

	return &types.HandleDeleteSessionResp{}, nil
}
