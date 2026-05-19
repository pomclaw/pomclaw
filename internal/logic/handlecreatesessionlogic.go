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
	"github.com/pomclaw/pomclaw/pkg/utils"

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
	if req.AgentId == "" {
		return nil, fmt.Errorf("agent_id is required")
	}

	sessionKey := utils.GenerateID()
	now := time.Now()
	sessionData := &model.Sessions{
		SessionKey: sessionKey,
		AgentId:    req.AgentId,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	_, err = l.svcCtx.SessionsModel.Insert(l.ctx, sessionData)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &types.Session{
		Id:      sessionKey,
		AgentId: req.AgentId,
		Created: now.Format("2006-01-02T15:04:05Z07:00"),
		Updated: now.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
