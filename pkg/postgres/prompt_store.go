package postgres

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/zeromicro/go-zero/core/logx"
)

// PromptStore manages system prompts backed by PostgreSQL models.
type PromptStore struct {
	promptsModel model.PromptsModel
}

// NewPromptStore creates a new PostgreSQL-backed prompt store.
func NewPromptStore(promptsModel model.PromptsModel) *PromptStore {
	return &PromptStore{
		promptsModel: promptsModel,
	}
}

// SavePrompt upserts a prompt.
func (ps *PromptStore) SavePrompt(agentID string, name, content string) error {
	ctx := context.Background()
	return ps.promptsModel.SavePrompt(ctx, agentID, name, content)
}

// LoadBootstrapFiles returns a map of all prompts for context builder.
func (ps *PromptStore) LoadBootstrapFiles(agentID string) map[string]string {
	ctx := context.Background()
	result, err := ps.promptsModel.LoadBootstrapFiles(ctx, agentID)
	if err != nil {
		logx.Info("postgres", "Failed to load bootstrap prompts from PostgreSQL", map[string]interface{}{
			"agent_id": agentID,
			"error":    err.Error(),
		})
		return make(map[string]string)
	}
	return result
}

// SeedFromWorkspace reads workspace .md files and stores them as prompts.
func (ps *PromptStore) SeedFromWorkspace(workspacePath string) error {
	files := []string{"IDENTITY.md", "SOUL.md", "USER.md", "AGENT.md", "AGENTS.md"}

	seeded := 0
	for _, filename := range files {
		filePath := filepath.Join(workspacePath, filename)
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		promptName := filename[:len(filename)-3]
		if err := ps.SavePrompt("default", promptName, string(data)); err != nil {
			logx.Info("postgres", "Failed to seed prompt", map[string]interface{}{
				"file":  filename,
				"error": err.Error(),
			})
			continue
		}
		seeded++
	}

	if seeded > 0 {
		logx.Info("postgres", "Seeded prompts from workspace", map[string]interface{}{"count": seeded})
	}
	return nil
}
