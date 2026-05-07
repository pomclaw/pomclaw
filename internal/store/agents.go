package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Agent represents a row in pom_agents (complete pomclaw structure).
type Agent struct {
	// Basic fields
	ID           string    `json:"id"`
	AgentKey     string    `json:"agent_key"`
	DisplayName  string    `json:"display_name"`
	Frontmatter  string    `json:"frontmatter"`
	OwnerID      string    `json:"owner_id"`

	// LLM configuration
	Provider          string `json:"provider"`
	Model             string `json:"model"`
	ContextWindow     int    `json:"context_window"`
	MaxToolIterations int    `json:"max_tool_iterations"`

	// Workspace configuration
	Workspace           string `json:"workspace"`
	RestrictToWorkspace bool   `json:"restrict_to_workspace"`

	// Type and status
	AgentType string `json:"agent_type"`
	IsDefault bool   `json:"is_default"`
	Status    string `json:"status"`

	// Budget (optional)
	BudgetMonthlyCents *int `json:"budget_monthly_cents,omitempty"`

	// JSONB configs
	ToolsConfig      json.RawMessage  `json:"tools_config"`
	SandboxConfig    *json.RawMessage `json:"sandbox_config,omitempty"`
	SubagentsConfig  *json.RawMessage `json:"subagents_config,omitempty"`
	MemoryConfig     *json.RawMessage `json:"memory_config,omitempty"`
	CompactionConfig *json.RawMessage `json:"compaction_config,omitempty"`
	ContextPruning   *json.RawMessage `json:"context_pruning,omitempty"`
	OtherConfig      json.RawMessage  `json:"other_config"`

	// V3 fields (promoted from other_config)
	Emoji               string          `json:"emoji,omitempty"`
	AgentDescription    string          `json:"agent_description,omitempty"`
	ThinkingLevel       string          `json:"thinking_level,omitempty"`
	MaxTokens           int             `json:"max_tokens"`
	SelfEvolve          bool            `json:"self_evolve"`
	SkillEvolve         bool            `json:"skill_evolve"`
	SkillNudgeInterval  int             `json:"skill_nudge_interval"`
	ReasoningConfig     *json.RawMessage `json:"reasoning_config,omitempty"`
	WorkspaceSharing    *json.RawMessage `json:"workspace_sharing,omitempty"`
	ChatGPTOAuthRouting *json.RawMessage `json:"chatgpt_oauth_routing,omitempty"`
	ShellDenyGroups     *json.RawMessage `json:"shell_deny_groups,omitempty"`
	KGDedupConfig       *json.RawMessage `json:"kg_dedup_config,omitempty"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// ListAgents returns all agents owned by the given user.
func ListAgents(db *sql.DB, userID string) ([]*Agent, error) {
	query := `
		SELECT id, agent_key, display_name, frontmatter, owner_id,
		       provider, model, context_window, max_tool_iterations,
		       workspace, restrict_to_workspace, agent_type, is_default, status,
		       budget_monthly_cents,
		       tools_config, sandbox_config, subagents_config, memory_config,
		       compaction_config, context_pruning, other_config,
		       emoji, agent_description, thinking_level, max_tokens,
		       self_evolve, skill_evolve, skill_nudge_interval,
		       reasoning_config, workspace_sharing, chatgpt_oauth_routing,
		       shell_deny_groups, kg_dedup_config,
		       created_at, updated_at, deleted_at
		FROM pom_agents
		WHERE owner_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []*Agent
	for rows.Next() {
		a := &Agent{}
		err := rows.Scan(
			&a.ID, &a.AgentKey, &a.DisplayName, &a.Frontmatter, &a.OwnerID,
			&a.Provider, &a.Model, &a.ContextWindow, &a.MaxToolIterations,
			&a.Workspace, &a.RestrictToWorkspace, &a.AgentType, &a.IsDefault, &a.Status,
			&a.BudgetMonthlyCents,
			&a.ToolsConfig, &a.SandboxConfig, &a.SubagentsConfig, &a.MemoryConfig,
			&a.CompactionConfig, &a.ContextPruning, &a.OtherConfig,
			&a.Emoji, &a.AgentDescription, &a.ThinkingLevel, &a.MaxTokens,
			&a.SelfEvolve, &a.SkillEvolve, &a.SkillNudgeInterval,
			&a.ReasoningConfig, &a.WorkspaceSharing, &a.ChatGPTOAuthRouting,
			&a.ShellDenyGroups, &a.KGDedupConfig,
			&a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		agents = append(agents, a)
	}
	return agents, rows.Err()
}

// CreateAgent inserts a new agent with all provided fields.
func CreateAgent(db *sql.DB, a *Agent) error {
	if a.ID == "" {
		a.ID = GenerateShortID()
	}

	// Set defaults
	if a.ContextWindow <= 0 {
		a.ContextWindow = 200000
	}
	if a.MaxToolIterations <= 0 {
		a.MaxToolIterations = 20
	}
	if a.Workspace == "" {
		a.Workspace = "."
	}
	if a.AgentType == "" {
		a.AgentType = "predefined"
	}
	if a.Status == "" {
		a.Status = "active"
	}
	if len(a.ToolsConfig) == 0 {
		a.ToolsConfig = []byte("{}")
	}
	if len(a.OtherConfig) == 0 {
		a.OtherConfig = []byte("{}")
	}

	query := `
		INSERT INTO pom_agents (
			id, agent_key, display_name, frontmatter, owner_id,
			provider, model, context_window, max_tool_iterations,
			workspace, restrict_to_workspace, agent_type, is_default, status,
			budget_monthly_cents,
			tools_config, sandbox_config, subagents_config, memory_config,
			compaction_config, context_pruning, other_config,
			emoji, agent_description, thinking_level, max_tokens,
			self_evolve, skill_evolve, skill_nudge_interval,
			reasoning_config, workspace_sharing, chatgpt_oauth_routing,
			shell_deny_groups, kg_dedup_config
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
			$31, $32, $33, $34
		) RETURNING created_at, updated_at
	`

	return db.QueryRow(query,
		a.ID, a.AgentKey, a.DisplayName, a.Frontmatter, a.OwnerID,
		a.Provider, a.Model, a.ContextWindow, a.MaxToolIterations,
		a.Workspace, a.RestrictToWorkspace, a.AgentType, a.IsDefault, a.Status,
		a.BudgetMonthlyCents,
		a.ToolsConfig, a.SandboxConfig, a.SubagentsConfig, a.MemoryConfig,
		a.CompactionConfig, a.ContextPruning, a.OtherConfig,
		a.Emoji, a.AgentDescription, a.ThinkingLevel, a.MaxTokens,
		a.SelfEvolve, a.SkillEvolve, a.SkillNudgeInterval,
		a.ReasoningConfig, a.WorkspaceSharing, a.ChatGPTOAuthRouting,
		a.ShellDenyGroups, a.KGDedupConfig,
	).Scan(&a.CreatedAt, &a.UpdatedAt)
}

// GetAgent returns the agent with the given id or agent_key.
func GetAgent(db *sql.DB, idOrKey, userID string) (*Agent, error) {
	query := `
		SELECT id, agent_key, display_name, frontmatter, owner_id,
		       provider, model, context_window, max_tool_iterations,
		       workspace, restrict_to_workspace, agent_type, is_default, status,
		       budget_monthly_cents,
		       tools_config, sandbox_config, subagents_config, memory_config,
		       compaction_config, context_pruning, other_config,
		       emoji, agent_description, thinking_level, max_tokens,
		       self_evolve, skill_evolve, skill_nudge_interval,
		       reasoning_config, workspace_sharing, chatgpt_oauth_routing,
		       shell_deny_groups, kg_dedup_config,
		       created_at, updated_at, deleted_at
		FROM pom_agents
		WHERE (id = $1 OR agent_key = $1) AND owner_id = $2 AND deleted_at IS NULL
	`
	a := &Agent{}
	err := db.QueryRow(query, idOrKey, userID).Scan(
		&a.ID, &a.AgentKey, &a.DisplayName, &a.Frontmatter, &a.OwnerID,
		&a.Provider, &a.Model, &a.ContextWindow, &a.MaxToolIterations,
		&a.Workspace, &a.RestrictToWorkspace, &a.AgentType, &a.IsDefault, &a.Status,
		&a.BudgetMonthlyCents,
		&a.ToolsConfig, &a.SandboxConfig, &a.SubagentsConfig, &a.MemoryConfig,
		&a.CompactionConfig, &a.ContextPruning, &a.OtherConfig,
		&a.Emoji, &a.AgentDescription, &a.ThinkingLevel, &a.MaxTokens,
		&a.SelfEvolve, &a.SkillEvolve, &a.SkillNudgeInterval,
		&a.ReasoningConfig, &a.WorkspaceSharing, &a.ChatGPTOAuthRouting,
		&a.ShellDenyGroups, &a.KGDedupConfig,
		&a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// UpdateAgent updates agent fields dynamically.
func UpdateAgent(db *sql.DB, id, userID string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Build dynamic SET clause
	var setClauses []string
	var args []interface{}
	argPos := 3 // $1 = id, $2 = userID

	// Allowed fields (whitelist)
	allowedFields := map[string]bool{
		"agent_key": true, "display_name": true, "frontmatter": true,
		"provider": true, "model": true, "status": true,
		"context_window": true, "max_tool_iterations": true, "workspace": true,
		"restrict_to_workspace": true, "is_default": true, "budget_monthly_cents": true,
		"tools_config": true, "sandbox_config": true, "subagents_config": true,
		"memory_config": true, "compaction_config": true, "context_pruning": true,
		"other_config": true, "emoji": true, "agent_description": true,
		"thinking_level": true, "max_tokens": true, "self_evolve": true,
		"skill_evolve": true, "skill_nudge_interval": true,
		"reasoning_config": true, "workspace_sharing": true,
		"chatgpt_oauth_routing": true, "shell_deny_groups": true, "kg_dedup_config": true,
	}

	for field, value := range updates {
		if !allowedFields[field] {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argPos))
		args = append(args, value)
		argPos++
	}

	if len(setClauses) == 0 {
		return nil
	}

	// Always update updated_at
	setClauses = append(setClauses, "updated_at = NOW()")

	query := fmt.Sprintf(`
		UPDATE pom_agents
		SET %s
		WHERE id = $1 AND owner_id = $2 AND deleted_at IS NULL
	`, strings.Join(setClauses, ", "))

	allArgs := append([]interface{}{id, userID}, args...)
	result, err := db.Exec(query, allArgs...)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteAgent soft-deletes an agent.
func DeleteAgent(db *sql.DB, id, userID string) error {
	result, err := db.Exec(
		`UPDATE pom_agents SET deleted_at = NOW() WHERE id = $1 AND owner_id = $2 AND deleted_at IS NULL`,
		id, userID,
	)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
