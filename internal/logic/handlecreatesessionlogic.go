// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"

	"github.com/pomclaw/pomclaw/internal/store"
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
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	if req.AgentId == "" {
		return nil, fmt.Errorf("agent_id is required")
	}

	session, err := store.CreateSession(l.svcCtx.Conn.DB(), userID, req.AgentId, req.Title)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &types.Session{
		Id:           session.ID,
		AgentId:      session.AgentID,
		Title:        "",
		Preview:      "",
		MessageCount: 0,
		Created:      session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Updated:      session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
