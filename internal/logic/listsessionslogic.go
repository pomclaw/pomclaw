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

type ListSessionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List all sessions
func NewListSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSessionsLogic {
	return &ListSessionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListSessionsLogic) ListSessions(req *types.ListSessionsReq) (resp *types.ListSessionsResp, err error) {
	offset := req.Offset
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	// 查询当前用户的 session（分页 + 倒序）
	sessions, err := l.svcCtx.SessionsModel.FindByUserIDWithPagination(l.ctx, userID, offset, limit)
	if err != nil {
		l.Errorf("ListSessions failed: %v", err)
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	sessionList := make([]types.Session, 0, len(sessions))
	for _, s := range sessions {
		sessionList = append(sessionList, types.Session{
			Id:           s.Id,
			AgentId:      s.AgentId,
			Title:        s.Label.String,
			Preview:      s.Summary.String,
			MessageCount: int(s.MessagesCount),
			Created:      s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Updated:      s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	resp = &types.ListSessionsResp{
		Total:    int64(len(sessionList)),
		Sessions: sessionList,
	}

	return
}
