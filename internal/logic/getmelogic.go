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

type GetMeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get current user info
func NewGetMeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMeLogic {
	return &GetMeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMeLogic) GetMe() (resp *types.UserResp, err error) {
	userId, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	user, err := l.svcCtx.UsersModel.FindOne(l.ctx, userId)
	if err == model.ErrNotFound {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &types.UserResp{
		Id:        user.Id,
		Email:     user.Email,
		Username:  user.Username,
		CreatedAt: user.CreatedAt.Unix(),
	}, nil
}
