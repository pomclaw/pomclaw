// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RevokeSkillAgentGrantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Revoke skill grant from agent
func NewRevokeSkillAgentGrantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeSkillAgentGrantLogic {
	return &RevokeSkillAgentGrantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RevokeSkillAgentGrantLogic) RevokeSkillAgentGrant(req *types.RevokeSkillAgentGrantReq) (resp *types.RevokeSkillAgentGrantResp, err error) {
	// todo: add your logic here and delete this line

	return
}
