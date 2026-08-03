// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAgentSkillsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List skills for agent with grant status
func NewListAgentSkillsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentSkillsLogic {
	return &ListAgentSkillsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAgentSkillsLogic) ListAgentSkills(req *types.ListAgentSkillsReq) (resp *types.SkillsWithGrantResp, err error) {
	// todo: add your logic here and delete this line

	return
}
