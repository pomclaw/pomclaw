package logic

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAgentAuthorizationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get agent authorizations (granted skills, MCP servers, tool policy)
func NewGetAgentAuthorizationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAgentAuthorizationsLogic {
	return &GetAgentAuthorizationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAgentAuthorizationsLogic) GetAgentAuthorizations(req *types.GetAgentAuthorizationsReq) (resp *types.GetAgentAuthorizationsResp, err error) {
	// 1. Get agent to access tools_config
	agent, err := l.svcCtx.AgentsModel.FindByAgentID(l.ctx, req.AgentId)
	if err != nil {
		return nil, err
	}

	// 2. Get granted skills (only enabled grants)
	grants, err := l.svcCtx.SkillGrantsModel.FindByAgentID(l.ctx, req.AgentId)
	if err != nil {
		return nil, err
	}

	skills := make([]types.AgentAuthorizationSkill, 0, len(grants))
	if len(grants) > 0 {
		skillIDs := make([]int64, 0, len(grants))
		for _, g := range grants {
			skillIDs = append(skillIDs, g.SkillId)
		}
		skillModels, err := l.svcCtx.SkillsModel.FindEnabledByIDs(l.ctx, skillIDs)
		if err != nil {
			return nil, err
		}

		// Collect user IDs for batch name lookup
		userIDs := make([]string, 0, len(skillModels))
		for _, s := range skillModels {
			userIDs = append(userIDs, s.UserId)
		}
		userNameMap := make(map[string]string, len(userIDs))
		if users, err := l.svcCtx.UsersModel.FindByUserIDs(l.ctx, userIDs); err == nil {
			for _, u := range users {
				userNameMap[u.UserId] = u.Username
			}
		}

		for _, s := range skillModels {
			visibility := "private"
			if s.IsShared {
				visibility = "shared"
			}
			desc := ""
			if s.Description.Valid {
				desc = s.Description.String
			}
			skills = append(skills, types.AgentAuthorizationSkill{
				Id:            s.Id,
				Name:          s.Name,
				Slug:          s.Slug,
				Description:   desc,
				Version:       int(s.Version),
				Visibility:    visibility,
				CreatedBy:     s.UserId,
				CreatedByName: userNameMap[s.UserId],
			})
		}
	}

	// 3. Get MCP server grants
	mcpGrants, err := l.svcCtx.McpAgentGrantsModel.FindByAgentID(l.ctx, req.AgentId)
	if err != nil {
		return nil, err
	}

	mcpServers := make([]types.AgentAuthorizationMCP, 0, len(mcpGrants))
	for _, g := range mcpGrants {
		server, err := l.svcCtx.McpServersModel.FindOne(l.ctx, g.ServerId)
		if err != nil {
			continue // skip if server not found/deleted
		}
		mcpServers = append(mcpServers, types.AgentAuthorizationMCP{
			ServerID:   strconv.FormatInt(server.Id, 10),
			ServerName: server.Name,
			Transport:  server.Transport,
			Enabled:    g.Enabled,
		})
	}

	// 4. Parse tools_config JSONB into tool policy
	var toolPolicy *types.AgentAuthorizationToolPolicy
	if agent.ToolsConfig != "" {
		var policy types.AgentAuthorizationToolPolicy
		if err := json.Unmarshal([]byte(agent.ToolsConfig), &policy); err == nil {
			toolPolicy = &policy
		}
	}

	return &types.GetAgentAuthorizationsResp{
		Skills:     skills,
		MCPServers: mcpServers,
		ToolPolicy: toolPolicy,
	}, nil
}
