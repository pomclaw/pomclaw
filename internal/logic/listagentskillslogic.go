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
			ID:          s.Id,
			Name:        s.Name,
			Slug:        s.Slug,
			Description: nullStringToString(s.Description),
			Enabled:     s.Enabled,
			Status:      s.Status,
			Version:     int(s.Version),
			IsSystem:    false,
			Source:      "file",
			Visibility:  "private",
			Granted:     granted,
		})
	}

	return &types.SkillsWithGrantResp{Skills: resp}, nil
}
