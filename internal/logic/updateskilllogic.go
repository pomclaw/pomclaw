// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"database/sql"
	"time"

	"github.com/pomclaw/pomclaw/internal/model"
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
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("UpdateSkill failed: %v", err)
		return nil, err
	}

	// Fetch existing skill
	skill, err := l.svcCtx.SkillsModel.FindOne(l.ctx, req.ID)
	if err == model.ErrNotFound || (err == nil && skill.UserId != userID) {
		l.Errorf("UpdateSkill failed: skill not found")
		return nil, model.ErrNotFound
	}
	if err != nil {
		l.Errorf("UpdateSkill failed: %v", err)
		return nil, err
	}

	// Apply updates
	if req.Enabled != false {
		skill.Enabled = req.Enabled
	}
	if req.Name != "" {
		skill.Name = req.Name
	}
	if req.Description != "" {
		skill.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.Status != "" {
		skill.Status = req.Status
	}

	skill.UpdatedAt = time.Now()

	// Update the skill
	if err := l.svcCtx.SkillsModel.Update(l.ctx, skill); err != nil {
		l.Errorf("UpdateSkill failed: %v", err)
		return nil, err
	}

	resp = &types.UpdateSkillResp{
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
