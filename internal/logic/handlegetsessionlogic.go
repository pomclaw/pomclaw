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

func (l *HandleGetSessionLogic) HandleGetSession(req *types.HandleGetSessionReq) (resp *types.Session, err error) {
	sessionID := req.Id
	if sessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}

	sessionData, err := store.GetSessionWithMessages(l.svcCtx.Conn.DB(), sessionID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &types.Session{
		Id:      sessionData["id"].(string),
		Created: sessionData["created"].(string),
		Updated: sessionData["updated"].(string),
	}, nil
}
