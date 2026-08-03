// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/exp/slices"
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

	skills, err := l.svcCtx.SkillsModel.FindByUserIDWithShared(l.ctx, userID)
	if err != nil {
		l.Errorf("ListSkills failed: %v", err)
		return nil, fmt.Errorf("failed to list skills: %w", err)
	}

	// Batch query usernames for all unique user IDs
	uidSet := make([]string, 0, len(skills))
	for _, s := range skills {
		if !slices.Contains(uidSet, s.UserId) {
			uidSet = append(uidSet, s.UserId)
		}
	}
	userMap := make(map[string]string, len(uidSet))
	if users, err := l.svcCtx.UsersModel.FindByUserIDs(l.ctx, uidSet); err == nil {
		for _, u := range users {
			userMap[u.UserId] = u.Username
		}
	}

	skillList := make([]types.SkillResp, 0, len(skills))
	for _, s := range skills {
		svr := types.SkillResp{
			ID:          s.Id,
			Name:        s.Name,
			Slug:        s.Slug,
			Description: nullStringToString(s.Description),
			Enabled:     s.Enabled,
			Status:      s.Status,
			Version:     int(s.Version),
			IsSystem:    false,
			Source:      "",
			Visibility:  "private",
			Tags:        nil,
			MissingDeps: nil,
			IsShared:    s.IsShared,
			CreatedBy:   s.UserId,
		}
		if name, ok := userMap[s.UserId]; ok {
			svr.CreatedByName = name
		}
		skillList = append(skillList, svr)
	}

	return &types.SkillsResp{
		Skills: skillList,
	}, nil
}