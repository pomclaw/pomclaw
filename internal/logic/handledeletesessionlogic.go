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

type HandleDeleteSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Delete session
func NewHandleDeleteSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleDeleteSessionLogic {
	return &HandleDeleteSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleDeleteSessionLogic) HandleDeleteSession(req *types.DeleteSessionReq) (resp *types.DeleteSessionResp, err error) {
	// Note: GetUserIDFromContext would be used here for user ownership verification once Sessions table is updated (TODO)
	// userID, err := GetUserIDFromContext(l.ctx)
	// if err != nil {
	// 	l.Errorf("HandleDeleteSession failed: %v", err)
	// 	return nil, err
	// }

	sessionID := req.Id
	if sessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}

	// Note: User ownership verification would require user_id field in Sessions table
	// For now, verify session exists before deleting (TODO: add user_id field)

	err = l.svcCtx.SessionsModel.Delete(l.ctx, sessionID)
	if err == model.ErrNotFound {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		l.Errorf("HandleDeleteSession failed: %v", err)
		return nil, fmt.Errorf("failed to delete session: %w", err)
	}

	resp = &types.DeleteSessionResp{}

	return
}
