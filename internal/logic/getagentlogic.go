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

type GetAgentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get agent details
func NewGetAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAgentLogic {
	return &GetAgentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAgentLogic) GetAgent(req *types.GetAgentReq) (resp *types.GetAgentResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	agentID := req.AgentId
	if agentID == "" {
		return nil, fmt.Errorf("agent_id is required")
	}

	agent, err := l.svcCtx.AgentsModel.FindByAgentID(l.ctx, agentID)
	if err == model.ErrNotFound {
		return nil, &NotFoundError{Message: "agent not found"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	// Permission check: only owner or shared agents can be viewed by other users
	if agent.UserId != userID && !agent.IsShared {
		return nil, &NotFoundError{Message: "agent not found"}
	}

	result := ConvertModelAgentToType(agent)
	if user, err := l.svcCtx.UsersModel.FindOneByUserId(l.ctx, agent.UserId); err == nil {
		result.CreatedByName = user.Username
	}
	return &types.GetAgentResp{
		Agent: *result,
	}, nil
}
