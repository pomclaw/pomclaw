// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleListSessionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List sessions
func NewHandleListSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleListSessionsLogic {
	return &HandleListSessionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleListSessionsLogic) HandleListSessions(req *types.HandleListSessionsReq) (resp *types.ListSessionsResp, err error) {
	agentID := req.AgentId
	offset := req.Offset
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	sessions, err := l.svcCtx.SessionsModel.FindByAgentIDWithPagination(l.ctx, agentID, offset, limit)
	if err != nil {
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

	return &types.ListSessionsResp{
		Total:    int64(len(sessionList)),
		Sessions: sessionList,
	}, nil
}
