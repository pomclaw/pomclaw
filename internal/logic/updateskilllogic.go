// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateSkillLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Update skill
func NewUpdateSkillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSkillLogic {
	return &UpdateSkillLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateSkillLogic) UpdateSkill(req *types.UpdateSkillReq) (resp *types.UpdateSkillResp, err error) {
	// todo: add your logic here and delete this line

	return
}
