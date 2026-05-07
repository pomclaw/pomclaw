package logic

import (
	"encoding/json"

	"github.com/pomclaw/pomclaw/internal/store"
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

// ConvertStoreAgentToType converts store.Agent to types.Agent
func ConvertStoreAgentToType(agent *store.Agent) *types.Agent {
	return &types.Agent{
		Id:                  agent.ID,
		AgentKey:            agent.AgentKey,
		DisplayName:         agent.DisplayName,
		Frontmatter:         agent.Frontmatter,
		OwnerId:             agent.OwnerID,
		Provider:            agent.Provider,
		Model:               agent.Model,
		ContextWindow:       agent.ContextWindow,
		MaxToolIterations:   agent.MaxToolIterations,
		Workspace:           agent.Workspace,
		RestrictToWorkspace: agent.RestrictToWorkspace,
		AgentType:           agent.AgentType,
		IsDefault:           agent.IsDefault,
		Status:              agent.Status,
		ToolsConfig:         agent.ToolsConfig,
		MemoryConfig:        derefRawMessage(agent.MemoryConfig),
		CompactionConfig:    derefRawMessage(agent.CompactionConfig),
		OtherConfig:         agent.OtherConfig,
		Emoji:               agent.Emoji,
		AgentDescription:    agent.AgentDescription,
		ThinkingLevel:       agent.ThinkingLevel,
		MaxTokens:           agent.MaxTokens,
		SelfEvolve:          agent.SelfEvolve,
		SkillEvolve:         agent.SkillEvolve,
		CreatedAt:           agent.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:           agent.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
