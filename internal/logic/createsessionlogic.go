// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create a new session
func NewCreateSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSessionLogic {
	return &CreateSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSessionLogic) CreateSession(req *types.CreateSessionReq) (resp *types.CreateSessionResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("CreateSession failed: %v", err)
		return nil, err
	}

	if req.AgentId == "" {
		return nil, fmt.Errorf("agent_id is required")
	}

	now := time.Now()
	sessionData := &model.Sessions{
		UserId:    userID,
		AgentId:   req.AgentId,
		CreatedAt: now,
		UpdatedAt: now,
	}

	id, err := l.svcCtx.SessionsModel.InsertWithReturning(l.ctx, sessionData)
	if err != nil {
		l.Errorf("CreateSession failed: %v", err)
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	resp = &types.CreateSessionResp{
		Session: types.Session{
			Id:      id,
			AgentId: req.AgentId,
			Created: now.Format("2006-01-02T15:04:05Z07:00"),
			Updated: now.Format("2006-01-02T15:04:05Z07:00"),
		},
	}

	return
}
