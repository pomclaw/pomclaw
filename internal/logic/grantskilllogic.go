// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GrantSkillLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Grant skill to agent
func NewGrantSkillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GrantSkillLogic {
	return &GrantSkillLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GrantSkillLogic) GrantSkill(req *types.GrantSkillReq) (resp *types.GrantStatusResp, err error) {
	version := int64(req.Version)
	if version <= 0 {
		version = 1
	}

	if err := l.svcCtx.SkillGrantsModel.GrantSkillToAgent(l.ctx, req.ID, req.AgentID, version); err != nil {
		l.Errorf("GrantSkill failed: %v", err)
		return nil, err
	}

	resp = &types.GrantStatusResp{
		Status: "granted",
	}

	return
}
