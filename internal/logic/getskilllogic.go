// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetSkillLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get skill details
func NewGetSkillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSkillLogic {
	return &GetSkillLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSkillLogic) GetSkill(req *types.GetSkillReq) (resp *types.GetSkillResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("GetSkill failed: %v", err)
		return nil, err
	}

	skill, err := l.svcCtx.SkillsModel.FindOne(l.ctx, req.ID)
	if err == model.ErrNotFound || (err == nil && skill.UserId != userID) {
		l.Errorf("GetSkill failed: skill not found")
		return nil, model.ErrNotFound
	}
	if err != nil {
		l.Errorf("GetSkill failed: %v", err)
		return nil, err
	}

	resp = &types.GetSkillResp{
		Skill: types.SkillResp{
			ID:          skill.Id,
			Name:        skill.Name,
			Slug:        skill.Slug,
			Description: nullStringToString(skill.Description),
			Enabled:     skill.Enabled,
			Status:      skill.Status,
			Version:     int(skill.Version),
			IsSystem:    false,
			Source:      "file",
			Visibility:  "private",
		},
	}

	return
}
