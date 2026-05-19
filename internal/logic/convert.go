package logic

import (
	"database/sql"
	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/types"
)

// ConvertModelAgentToType converts model.Agents to types.Agent
func ConvertModelAgentToType(agent *model.Agents) *types.Agent {
	return &types.Agent{
		Id:                  agent.Id,
		AgentKey:            agent.AgentKey,
		DisplayName:         nullStringToString(agent.DisplayName),
		Frontmatter:         nullStringToString(agent.Frontmatter),
		OwnerId:             agent.OwnerId,
		Provider:            agent.Provider,
		Model:               agent.Model,
		ContextWindow:       int(agent.ContextWindow),
		MaxToolIterations:   int(agent.MaxToolIterations),
		Workspace:           agent.Workspace,
		RestrictToWorkspace: agent.RestrictToWorkspace,
		AgentType:           "predefined", // 默认值
		IsDefault:           false,        // 默认值
		Status:              "active",     // 默认值
		ToolsConfig:         []byte(agent.ToolsConfig),
		MemoryConfig:        []byte(agent.MemoryConfig),
		CompactionConfig:    []byte(agent.CompactionConfig),
		OtherConfig:         []byte(agent.OtherConfig),
		Emoji:               nullStringToString(agent.Emoji),
		AgentDescription:    nullStringToString(agent.AgentDescription),
		ThinkingLevel:       nullStringToString(agent.ThinkingLevel),
		MaxTokens:           int(agent.MaxTokens),
		SelfEvolve:          agent.SelfEvolve,
		SkillEvolve:         agent.SkillEvolve,
		CreatedAt:           agent.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:           agent.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// nullStringToString converts sql.NullString to string
func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// jsonOrEmpty converts json.RawMessage to valid JSON string, defaults to "{}" if empty
func jsonOrEmpty(data []byte) string {
	if len(data) == 0 {
		return "{}"
	}
	return string(data)
}
