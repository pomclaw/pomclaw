// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"

	"github.com/pomclaw/pomclaw/internal/model"
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

	session, err := l.svcCtx.SessionsModel.FindOne(l.ctx, sessionID)
	if err == model.ErrNotFound {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &types.Session{
		Id:      session.SessionKey,
		AgentId: session.AgentId,
		Created: session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Updated: session.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
