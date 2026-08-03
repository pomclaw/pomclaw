// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"

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
		return nil, err
	}

	skill, err := l.svcCtx.SkillsModel.FindOne(l.ctx, req.ID)
	if err == model.ErrNotFound {
		return nil, &NotFoundError{Message: "skill not found"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get skill: %w", err)
	}

	// Permission check: only owner or shared skills can be viewed by other users
	if skill.UserId != userID && !skill.IsShared {
		return nil, &NotFoundError{Message: "skill not found"}
	}

	result := types.SkillResp{
		ID:          skill.Id,
		Name:        skill.Name,
		Slug:        skill.Slug,
		Description: nullStringToString(skill.Description),
		Enabled:     skill.Enabled,
		Status:      skill.Status,
		Version:     int(skill.Version),
		IsSystem:    false,
		Source:      "",
		Visibility:  "private",
		Tags:        nil,
		MissingDeps: nil,
		IsShared:    skill.IsShared,
		CreatedBy:   skill.UserId,
	}
	if user, err := l.svcCtx.UsersModel.FindOneByUserId(l.ctx, skill.UserId); err == nil {
		result.CreatedByName = user.Username
	}

	return &types.GetSkillResp{
		Skill: result,
	}, nil
}