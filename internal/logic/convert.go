package logic

import (
	"database/sql"
	"encoding/json"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/types"
)

// derefRawMessage converts *json.RawMessage to json.RawMessage
// Returns nil if the pointer is nil
func derefRawMessage(ptr *json.RawMessage) json.RawMessage {
	if ptr == nil {
		return nil
	}
	return *ptr
}

// refRawMessage converts json.RawMessage to *json.RawMessage
// Returns nil if the input is nil
func refRawMessage(msg json.RawMessage) *json.RawMessage {
	if msg == nil {
		return nil
	}
	return &msg
}

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
		AgentType:           agent.AgentType,
		IsDefault:           agent.IsDefault,
		Status:              agent.Status,
		ToolsConfig:         []byte(agent.ToolsConfig),
		MemoryConfig:        []byte(nullStringToString(agent.MemoryConfig)),
		CompactionConfig:    []byte(nullStringToString(agent.CompactionConfig)),
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
