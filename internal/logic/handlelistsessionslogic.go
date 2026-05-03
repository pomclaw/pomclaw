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

	items, err := store.ListSessionsWithPagination(l.svcCtx.Conn.DB(), agentID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	sessions := make([]types.Session, 0, len(items))
	for _, item := range items {
		sessions = append(sessions, types.Session{
			Id:           item["id"].(string),
			Title:        item["title"].(string),
			Preview:      item["preview"].(string),
			MessageCount: item["message_count"].(int),
			Created:      item["created"].(string),
			Updated:      item["updated"].(string),
		})
	}

	return &types.ListSessionsResp{
		Total:    int64(len(sessions)),
		Sessions: sessions,
	}, nil
}
