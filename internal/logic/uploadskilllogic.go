// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadSkillLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Upload skill zip package
func NewUploadSkillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadSkillLogic {
	return &UploadSkillLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadSkillLogic) UploadSkill(req *types.UploadSkillReq) (resp *types.UploadSkillResp, err error) {
	// todo: add your logic here and delete this line

	return
}
