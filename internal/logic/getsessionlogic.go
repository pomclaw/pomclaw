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

type GetSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get session details
func NewGetSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSessionLogic {
	return &GetSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSessionLogic) GetSession(req *types.GetSessionReq) (resp *types.GetSessionResp, err error) {
	// Note: GetUserIDFromContext would be used here for user ownership verification once Sessions table is updated (TODO)
	// userID, err := GetUserIDFromContext(l.ctx)
	// if err != nil {
	// 	l.Errorf("GetSession failed: %v", err)
	// 	return nil, err
	// }

	session, err := l.svcCtx.SessionsModel.FindOne(l.ctx, req.Id)
	if err == model.ErrNotFound {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		l.Errorf("GetSession failed: %v", err)
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Note: User ownership verification would require user_id field in Sessions table
	// For now, verify agent exists and user has access via other means (TODO)

	resp = &types.GetSessionResp{
		Session: types.Session{
			Id:      session.Id,
			AgentId: session.AgentId,
			Created: session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Updated: session.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}

	return
}
