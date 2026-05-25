// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RevokeSkillLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Revoke skill from agent
func NewRevokeSkillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeSkillLogic {
	return &RevokeSkillLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RevokeSkillLogic) RevokeSkill(req *types.RevokeSkillReq) (resp *types.RevokeStatusResp, err error) {
	if err := l.svcCtx.SkillGrantsModel.RevokeSkillFromAgent(l.ctx, req.ID, req.AgentID); err != nil {
		l.Errorf("RevokeSkill failed: %v", err)
		return nil, err
	}

	resp = &types.RevokeStatusResp{
		Status: "revoked",
	}

	return
}
