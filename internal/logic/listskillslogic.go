// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSkillsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List all skills
func NewListSkillsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSkillsLogic {
	return &ListSkillsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListSkillsLogic) ListSkills(req *types.ListSkillsReq) (resp *types.SkillsResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("ListSkills failed: %v", err)
		return nil, err
	}

	skills, err := l.svcCtx.SkillsModel.FindByUserID(l.ctx, userID)
	if err != nil {
		l.Errorf("ListSkills failed: %v", err)
		return nil, err
	}

	resp = &types.SkillsResp{}
	resp.Skills = make([]types.SkillResp, 0, len(skills))
	for _, s := range skills {
		resp.Skills = append(resp.Skills, types.SkillResp{
			ID:          s.Id,
			Name:        s.Name,
			Slug:        s.Slug,
			Description: nullStringToString(s.Description),
			Enabled:     s.Enabled,
			Status:      s.Status,
			Version:     int(s.Version),
			IsSystem:    false, // Default: not a system skill (set true only for built-in skills)
			Source:      "file",
			Visibility:  "private",
		})
	}

	return
}
