// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GrantSkillAgentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Grant skill to agent
func NewGrantSkillAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GrantSkillAgentLogic {
	return &GrantSkillAgentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GrantSkillAgentLogic) GrantSkillAgent(req *types.GrantSkillAgentReq) (resp *types.GrantSkillAgentResp, err error) {
	// todo: add your logic here and delete this line

	return
}
