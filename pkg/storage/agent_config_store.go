package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pomclaw/pomclaw/pkg/agent"
	"github.com/pomclaw/pomclaw/pkg/logger"
)

// AgentConfig represents an agent configuration from the database
type AgentConfig struct {
	agent.AgentConfig
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AgentConfigStore manages agent configurations in the database
type AgentConfigStore struct {
	db          *sql.DB
	storageType string // "postgres" or "oracle"
}

// NewAgentConfigStore creates a new agent config store
func NewAgentConfigStore(db *sql.DB, storageType string) *AgentConfigStore {
	return &AgentConfigStore{
		db:          db,
		storageType: storageType,
	}
}

// GetByAgentID retrieves an agent config by agent_id
func (s *AgentConfigStore) GetByAgentID(agentID string) (*AgentConfig, error) {
	var cfg AgentConfig
	var query string

	if s.storageType == "postgres" {
		query = `
			SELECT config_id, user_id, agent_name, agent_id,
			       model, provider, max_tokens, temperature, max_iterations,
			       system_prompt, workspace, restrict_workspace,
			       is_active, created_at, updated_at
			FROM POM_AGENT_CONFIGS
			WHERE agent_id = $1 AND is_active = true
		`
	} else {
		query = `
			SELECT config_id, user_id, agent_name, agent_id,
			       model, provider, max_tokens, temperature, max_iterations,
			       system_prompt, workspace, restrict_workspace,
			       is_active, created_at, updated_at
			FROM POM_AGENT_CONFIGS
			WHERE agent_id = :1 AND is_active = 1
		`
	}

	var restrictWorkspace int
	var isActive int
	var systemPrompt sql.NullString
	var workspace sql.NullString

	err := s.db.QueryRow(query, agentID).Scan(
		&cfg.ConfigID, &cfg.UserID, &cfg.AgentName, &cfg.AgentID,
		&cfg.Model, &cfg.Provider, &cfg.MaxTokens, &cfg.Temperature, &cfg.MaxIterations,
		&systemPrompt, &workspace, &restrictWorkspace,
		&isActive, &cfg.CreatedAt, &cfg.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("agent config not found: %s", agentID)
	}
	if err != nil {
		logger.ErrorCF("storage", "Failed to query agent config", map[string]interface{}{
			"agent_id": agentID,
			"error":    err.Error(),
		})
		return nil, err
	}

	// Convert nullable fields
	if systemPrompt.Valid {
		cfg.SystemPrompt = systemPrompt.String
	}
	if workspace.Valid {
		cfg.Workspace = workspace.String
	}

	// Convert Oracle NUMBER(1) to bool
	if s.storageType == "oracle" {
		cfg.RestrictWorkspace = restrictWorkspace == 1
		cfg.IsActive = isActive == 1
	} else {
		cfg.RestrictWorkspace = restrictWorkspace != 0
		cfg.IsActive = isActive != 0
	}

	logger.DebugCF("storage", "Loaded agent config", map[string]interface{}{
		"agent_id":   agentID,
		"agent_name": cfg.AgentName,
		"model":      cfg.Model,
	})

	return &cfg, nil
}

// GetByUserID retrieves all agent configs for a user
func (s *AgentConfigStore) GetByUserID(userID string) ([]*AgentConfig, error) {
	var query string

	if s.storageType == "postgres" {
		query = `
			SELECT config_id, user_id, agent_name, agent_id,
			       model, provider, max_tokens, temperature, max_iterations,
			       system_prompt, workspace, restrict_workspace,
			       is_active, created_at, updated_at
			FROM POM_AGENT_CONFIGS
			WHERE user_id = $1
			ORDER BY created_at DESC
		`
	} else {
		query = `
			SELECT config_id, user_id, agent_name, agent_id,
			       model, provider, max_tokens, temperature, max_iterations,
			       system_prompt, workspace, restrict_workspace,
			       is_active, created_at, updated_at
			FROM POM_AGENT_CONFIGS
			WHERE user_id = :1
			ORDER BY created_at DESC
		`
	}

	rows, err := s.db.Query(query, userID)
	if err != nil {
		logger.ErrorCF("storage", "Failed to query agent configs", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	var configs []*AgentConfig
	for rows.Next() {
		var cfg AgentConfig
		var restrictWorkspace int
		var isActive int
		var systemPrompt sql.NullString
		var workspace sql.NullString

		if err := rows.Scan(
			&cfg.ConfigID, &cfg.UserID, &cfg.AgentName, &cfg.AgentID,
			&cfg.Model, &cfg.Provider, &cfg.MaxTokens, &cfg.Temperature, &cfg.MaxIterations,
			&systemPrompt, &workspace, &restrictWorkspace,
			&isActive, &cfg.CreatedAt, &cfg.UpdatedAt,
		); err != nil {
			logger.WarnCF("storage", "Failed to scan agent config row", map[string]interface{}{"error": err.Error()})
			continue
		}

		// Convert nullable fields
		if systemPrompt.Valid {
			cfg.SystemPrompt = systemPrompt.String
		}
		if workspace.Valid {
			cfg.Workspace = workspace.String
		}

		// Convert Oracle NUMBER(1) to bool
		if s.storageType == "oracle" {
			cfg.RestrictWorkspace = restrictWorkspace == 1
			cfg.IsActive = isActive == 1
		} else {
			cfg.RestrictWorkspace = restrictWorkspace != 0
			cfg.IsActive = isActive != 0
		}

		configs = append(configs, &cfg)
	}

	logger.InfoCF("storage", "Loaded agent configs for user", map[string]interface{}{
		"user_id": userID,
		"count":   len(configs),
	})

	return configs, nil
}

// Create creates a new agent config
func (s *AgentConfigStore) Create(cfg *AgentConfig) error {
	var query string

	restrictWorkspace := 0
	if cfg.RestrictWorkspace {
		restrictWorkspace = 1
	}
	isActive := 0
	if cfg.IsActive {
		isActive = 1
	}

	if s.storageType == "postgres" {
		query = `
			INSERT INTO POM_AGENT_CONFIGS (
				config_id, user_id, agent_name, agent_id,
				model, provider, max_tokens, temperature, max_iterations,
				system_prompt, workspace, restrict_workspace,
				is_active, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
			)
		`
	} else {
		query = `
			INSERT INTO POM_AGENT_CONFIGS (
				config_id, user_id, agent_name, agent_id,
				model, provider, max_tokens, temperature, max_iterations,
				system_prompt, workspace, restrict_workspace,
				is_active, created_at, updated_at
			) VALUES (
				:1, :2, :3, :4, :5, :6, :7, :8, :9, :10, :11, :12, :13, :14, :15
			)
		`
	}

	_, err := s.db.Exec(query,
		cfg.ConfigID, cfg.UserID, cfg.AgentName, cfg.AgentID,
		cfg.Model, cfg.Provider, cfg.MaxTokens, cfg.Temperature, cfg.MaxIterations,
		cfg.SystemPrompt, cfg.Workspace, restrictWorkspace,
		isActive, cfg.CreatedAt, cfg.UpdatedAt)

	if err != nil {
		logger.ErrorCF("storage", "Failed to create agent config", map[string]interface{}{
			"agent_id": cfg.AgentID,
			"error":    err.Error(),
		})
		return err
	}

	logger.InfoCF("storage", "Created agent config", map[string]interface{}{
		"agent_id":   cfg.AgentID,
		"agent_name": cfg.AgentName,
		"user_id":    cfg.UserID,
	})

	return nil
}

// Update updates an existing agent config
func (s *AgentConfigStore) Update(cfg *AgentConfig) error {
	var query string

	restrictWorkspace := 0
	if cfg.RestrictWorkspace {
		restrictWorkspace = 1
	}

	if s.storageType == "postgres" {
		query = `
			UPDATE POM_AGENT_CONFIGS
			SET agent_name = $1, model = $2, provider = $3,
			    max_tokens = $4, temperature = $5, max_iterations = $6,
			    system_prompt = $7, workspace = $8, restrict_workspace = $9,
			    updated_at = CURRENT_TIMESTAMP
			WHERE agent_id = $10
		`
	} else {
		query = `
			UPDATE POM_AGENT_CONFIGS
			SET agent_name = :1, model = :2, provider = :3,
			    max_tokens = :4, temperature = :5, max_iterations = :6,
			    system_prompt = :7, workspace = :8, restrict_workspace = :9,
			    updated_at = CURRENT_TIMESTAMP
			WHERE agent_id = :10
		`
	}

	_, err := s.db.Exec(query,
		cfg.AgentName, cfg.Model, cfg.Provider,
		cfg.MaxTokens, cfg.Temperature, cfg.MaxIterations,
		cfg.SystemPrompt, cfg.Workspace, restrictWorkspace,
		cfg.AgentID)

	if err != nil {
		logger.ErrorCF("storage", "Failed to update agent config", map[string]interface{}{
			"agent_id": cfg.AgentID,
			"error":    err.Error(),
		})
		return err
	}

	logger.InfoCF("storage", "Updated agent config", map[string]interface{}{
		"agent_id":   cfg.AgentID,
		"agent_name": cfg.AgentName,
	})

	return nil
}

// Delete soft-deletes an agent config
func (s *AgentConfigStore) Delete(agentID string) error {
	var query string

	if s.storageType == "postgres" {
		query = `
			UPDATE POM_AGENT_CONFIGS
			SET is_active = false, updated_at = CURRENT_TIMESTAMP
			WHERE agent_id = $1
		`
	} else {
		query = `
			UPDATE POM_AGENT_CONFIGS
			SET is_active = 0, updated_at = CURRENT_TIMESTAMP
			WHERE agent_id = :1
		`
	}

	_, err := s.db.Exec(query, agentID)
	if err != nil {
		logger.ErrorCF("storage", "Failed to delete agent config", map[string]interface{}{
			"agent_id": agentID,
			"error":    err.Error(),
		})
		return err
	}

	logger.InfoCF("storage", "Deleted agent config", map[string]interface{}{
		"agent_id": agentID,
	})

	return nil
}
