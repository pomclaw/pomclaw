package postgres

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pomclaw/pomclaw/pkg/logger"
)

// PromptStore manages system prompts in PostgreSQL POM_PROMPTS.
type PromptStore struct {
	db      *sql.DB
	agentID string
}

// NewPromptStore creates a new PostgreSQL-backed prompt store.
func NewPromptStore(db *sql.DB, agentID string) *PromptStore {
	return &PromptStore{
		db:      db,
		agentID: agentID,
	}
}

// LoadPrompt retrieves a named prompt from PostgreSQL.
func (ps *PromptStore) LoadPrompt(name string) (string, error) {
	var content sql.NullString
	err := ps.db.QueryRow(
		"SELECT content FROM POM_PROMPTS WHERE prompt_name = $1 AND agent_id = $2",
		name, ps.agentID,
	).Scan(&content)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("failed to load prompt %s: %w", name, err)
	}
	if !content.Valid {
		return "", nil
	}
	return content.String, nil
}

// SavePrompt upserts a prompt using PostgreSQL ON CONFLICT syntax.
func (ps *PromptStore) SavePrompt(name, content string) error {
	_, err := ps.db.Exec(`
		INSERT INTO POM_PROMPTS (prompt_name, agent_id, content)
		VALUES ($1, $2, $3)
		ON CONFLICT (prompt_name, agent_id) DO UPDATE
		SET content = $3, updated_at = CURRENT_TIMESTAMP
	`, name, ps.agentID, content)
	if err != nil {
		return fmt.Errorf("failed to save prompt %s: %w", name, err)
	}
	return nil
}

// LoadBootstrapFiles returns a map of all prompts for context builder.
func (ps *PromptStore) LoadBootstrapFiles() map[string]string {
	result := make(map[string]string)

	rows, err := ps.db.Query(
		"SELECT prompt_name, content FROM POM_PROMPTS WHERE agent_id = $1",
		ps.agentID,
	)
	if err != nil {
		logger.WarnCF("postgres", "Failed to load bootstrap prompts from PostgreSQL", map[string]interface{}{
			"agent_id": ps.agentID,
			"error":    err.Error(),
		})
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var content sql.NullString
		if err := rows.Scan(&name, &content); err == nil && content.Valid {
			result[name] = content.String
		}
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
			continue // File doesn't exist, skip
		}

		promptName := filename[:len(filename)-3] // Remove .md extension
		if err := ps.SavePrompt(promptName, string(data)); err != nil {
			logger.WarnCF("postgres", "Failed to seed prompt", map[string]interface{}{
				"file":  filename,
				"error": err.Error(),
			})
			continue
		}
		seeded++
	}

	if seeded > 0 {
		logger.InfoCF("postgres", "Seeded prompts from workspace", map[string]interface{}{"count": seeded})
	}
	return nil
}
