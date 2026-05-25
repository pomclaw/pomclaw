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

type HandleListSessionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List all sessions
func NewHandleListSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleListSessionsLogic {
	return &HandleListSessionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleListSessionsLogic) HandleListSessions(req *types.ListSessionsReq) (resp *types.ListSessionsResp, err error) {
	// Note: GetUserIDFromContext would be used here for user-scoped queries once Sessions table is updated (TODO)
	// userID, err := GetUserIDFromContext(l.ctx)
	// if err != nil {
	// 	l.Errorf("HandleListSessions failed: %v", err)
	// 	return nil, err
	// }

	agentID := req.AgentId
	offset := req.Offset
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	var sessions []*model.Sessions
	if agentID != "" {
		sessions, err = l.svcCtx.SessionsModel.FindByAgentIDWithPagination(l.ctx, agentID, offset, limit)
	} else {
		// TODO: Implement user sessions filtering when user_id is added to Sessions table
		sessions, err = l.svcCtx.SessionsModel.FindAll(l.ctx)
	}
	if err != nil {
		l.Errorf("HandleListSessions failed: %v", err)
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	sessionList := make([]types.Session, 0, len(sessions))
	for _, s := range sessions {
		sessionList = append(sessionList, types.Session{
			Id:      s.SessionKey,
			AgentId: s.AgentId,
			Created: s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Updated: s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	resp = &types.ListSessionsResp{
		Total:    int64(len(sessionList)),
		Sessions: sessionList,
	}

	return
}
