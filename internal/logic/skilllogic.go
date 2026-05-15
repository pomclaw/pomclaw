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

type ListAgentSkillsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAgentSkillsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentSkillsLogic {
	return &ListAgentSkillsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAgentSkillsLogic) ListAgentSkills(userID, agentID string) (*types.SkillsWithGrantResp, error) {
	skills, err := l.svcCtx.SkillsModel.FindByUserID(l.ctx, userID)
	if err != nil {
		logx.Errorf("ListAgentSkills failed: %v", err)
		return nil, err
	}

	resp := make([]types.SkillWithGrantResp, 0, len(skills))
	for _, s := range skills {
		granted, _ := l.svcCtx.SkillGrantsModel.CheckSkillGranted(l.ctx, s.Id, agentID)
		resp = append(resp, types.SkillWithGrantResp{
			SkillResp: &types.SkillResp{
				ID:          s.Id,
				Name:        s.Name,
				Slug:        s.Slug,
				Description: nullStringToString(s.Description),
				Enabled:     s.Enabled,
				Status:      s.Status,
				Version:     int(s.Version),
			},
			Granted: granted,
		})
	}

	return &types.SkillsWithGrantResp{Skills: resp}, nil
}

type ListSkillsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListSkillsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSkillsLogic {
	return &ListSkillsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListSkillsLogic) ListSkills(userID string) (*types.SkillsResp, error) {
	skills, err := l.svcCtx.SkillsModel.FindByUserID(l.ctx, userID)
	if err != nil {
		logx.Errorf("ListSkills failed: %v", err)
		return nil, err
	}

	resp := make([]types.SkillResp, 0, len(skills))
	for _, s := range skills {
		resp = append(resp, types.SkillResp{
			ID:          s.Id,
			Name:        s.Name,
			Slug:        s.Slug,
			Description: nullStringToString(s.Description),
			Enabled:     s.Enabled,
			Status:      s.Status,
			Version:     int(s.Version),
		})
	}

	return &types.SkillsResp{Skills: resp}, nil
}

type GetSkillLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSkillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSkillLogic {
	return &GetSkillLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSkillLogic) GetSkill(userID, id string) (*types.SkillResp, error) {
	skill, err := l.svcCtx.SkillsModel.FindOne(l.ctx, id)
	if err == model.ErrNotFound || (err == nil && skill.UserId != userID) {
		logx.Errorf("GetSkill failed: skill not found")
		return nil, model.ErrNotFound
	}
	if err != nil {
		logx.Errorf("GetSkill failed: %v", err)
		return nil, err
	}

	return &types.SkillResp{
		ID:          skill.Id,
		Name:        skill.Name,
		Slug:        skill.Slug,
		Description: nullStringToString(skill.Description),
		Enabled:     skill.Enabled,
		Status:      skill.Status,
		Version:     int(skill.Version),
	}, nil
}

type CreateSkillLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateSkillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSkillLogic {
	return &CreateSkillLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSkillLogic) CreateSkill(userID string, req *types.CreateSkillReq) (*types.SkillResp, error) {
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
		logx.Errorf("CreateSkill failed: %v", err)
		return nil, err
	}

	return &types.SkillResp{
		ID:          skill.Id,
		Name:        skill.Name,
		Slug:        skill.Slug,
		Description: nullStringToString(skill.Description),
		Enabled:     skill.Enabled,
		Status:      skill.Status,
		Version:     int(skill.Version),
	}, nil
}

type GrantSkillLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGrantSkillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GrantSkillLogic {
	return &GrantSkillLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GrantSkillLogic) GrantSkill(skillID, agentID string, version int) error {
	if version <= 0 {
		version = 1
	}

	if err := l.svcCtx.SkillGrantsModel.GrantSkillToAgent(l.ctx, skillID, agentID, int64(version)); err != nil {
		logx.Errorf("GrantSkill failed: %v", err)
		return err
	}
	return nil
}

type RevokeSkillLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRevokeSkillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeSkillLogic {
	return &RevokeSkillLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RevokeSkillLogic) RevokeSkill(skillID, agentID string) error {
	if err := l.svcCtx.SkillGrantsModel.RevokeSkillFromAgent(l.ctx, skillID, agentID); err != nil {
		logx.Errorf("RevokeSkill failed: %v", err)
		return err
	}
	return nil
}
