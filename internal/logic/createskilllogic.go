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
	"github.com/pomclaw/pomclaw/pkg/utils"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSkillLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create a new skill
func NewCreateSkillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSkillLogic {
	return &CreateSkillLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSkillLogic) CreateSkill(req *types.CreateSkillReq) (resp *types.SkillResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("CreateSkill failed: %v", err)
		return nil, err
	}

	skillID := utils.GenerateID()
	now := time.Now()
	skill := &model.Skills{
		Id:          skillID,
		UserId:      userID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		Enabled:     req.Enabled,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if _, err := l.svcCtx.SkillsModel.Insert(l.ctx, skill); err != nil {
		l.Errorf("CreateSkill failed: %v", err)
		return nil, err
	}

	resp = &types.SkillResp{
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
	}

	return
}
