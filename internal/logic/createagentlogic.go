// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pomclaw/pomclaw/internal/store"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateAgentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create agent
func NewCreateAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAgentLogic {
	return &CreateAgentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateAgentLogic) CreateAgent(req *types.CreateAgentReq) (resp *types.Agent, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	if req.Name == "" || req.Model == "" {
		return nil, fmt.Errorf("name and model are required")
	}

	tools := req.Tools
	if tools == nil {
		tools = json.RawMessage("[]")
	}

	agent, err := store.CreateAgent(l.svcCtx.Conn.DB(), userID, req.Name, req.Description, req.SystemPrompt, req.Model, tools)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return &types.Agent{
		Id:           agent.ID,
		UserId:       agent.UserID,
		Name:         agent.Name,
		Description:  agent.Description,
		SystemPrompt: agent.SystemPrompt,
		Model:        agent.Model,
		Tools:        agent.Tools,
		Status:       agent.Status,
		CreatedAt:    agent.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    agent.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
