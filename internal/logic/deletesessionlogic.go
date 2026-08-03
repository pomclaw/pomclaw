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

type DeleteSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Delete session
func NewDeleteSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSessionLogic {
	return &DeleteSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteSessionLogic) DeleteSession(req *types.DeleteSessionReq) (resp *types.DeleteSessionResp, err error) {

	// Note: User ownership verification would require user_id field in Sessions table
	// For now, verify session exists before deleting (TODO: add user_id field)

	err = l.svcCtx.SessionsModel.Delete(l.ctx, req.Id)
	if err == model.ErrNotFound {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		l.Errorf("DeleteSession failed: %v", err)
		return nil, fmt.Errorf("failed to delete session: %w", err)
	}

	resp = &types.DeleteSessionResp{}

	return
}
