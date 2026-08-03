package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AgentsModel = (*customAgentsModel)(nil)

type (
	// AgentsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAgentsModel.
	AgentsModel interface {
		agentsModel
		withSession(session sqlx.Session) AgentsModel
		// 业务查询方法
		FindByUserID(ctx context.Context, userID string) ([]*Agents, error)
		FindByAgentID(ctx context.Context, agentID string) (*Agents, error)
		// 软删除
		SoftDelete(ctx context.Context, agentID, userID string) error
		// 动态更新
		UpdateFields(ctx context.Context, agentID, userID string, updates map[string]interface{}) error
	}

	customAgentsModel struct {
		*defaultAgentsModel
	}
)

// NewAgentsModel returns a model for the database table.
func NewAgentsModel(conn sqlx.SqlConn) AgentsModel {
	return &customAgentsModel{
		defaultAgentsModel: newAgentsModel(conn),
	}
}

func (m *customAgentsModel) withSession(session sqlx.Session) AgentsModel {
	return NewAgentsModel(sqlx.NewSqlConnFromSession(session))
}

// FindByUserID 返回指定用户的所有 agents + 其他用户共享的 agents（不包括已删除的）
func (m *customAgentsModel) FindByUserID(ctx context.Context, userID string) ([]*Agents, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE (user_id = $1 OR is_shared = true) AND deleted_at IS NULL ORDER BY created_at DESC",
		agentsRows, m.table,
	)
	var agents []*Agents
	err := m.conn.QueryRowsCtx(ctx, &agents, query, userID)
	return agents, err
}

// FindByAgentID 通过 agent_id (UUID) 和 userID 获取 agent
func (m *customAgentsModel) FindByAgentID(ctx context.Context, agentID string) (*Agents, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE agent_id = $1 AND deleted_at IS NULL LIMIT 1",
		agentsRows, m.table,
	)
	var resp Agents
	err := m.conn.QueryRowCtx(ctx, &resp, query, agentID)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// SoftDelete 软删除 agent（设置 deleted_at）
func (m *customAgentsModel) SoftDelete(ctx context.Context, agentID, userID string) error {
	query := fmt.Sprintf(
		"UPDATE %s SET deleted_at = NOW() WHERE agent_id = $1 AND user_id = $2 AND deleted_at IS NULL",
		m.table,
	)
	result, err := m.conn.ExecCtx(ctx, query, agentID, userID)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateFields 动态更新指定字段
func (m *customAgentsModel) UpdateFields(ctx context.Context, agentID, userID string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// 字段白名单 (删除了 agent_key)
	allowedFields := map[string]bool{
		"display_name": true, "frontmatter": true,
		"provider": true, "model": true, "status": true,
		"context_window": true, "max_tool_iterations": true, "workspace": true,
		"restrict_to_workspace": true, "is_default": true, "budget_monthly_cents": true,
		"tools_config": true, "sandbox_config": true, "subagents_config": true,
		"memory_config": true, "compaction_config": true, "context_pruning": true,
		"other_config": true, "emoji": true, "agent_description": true,
		"thinking_level": true, "max_tokens": true, "self_evolve": true,
		"skill_evolve": true, "is_shared": true, "skill_nudge_interval": true,
		"reasoning_config": true, "workspace_sharing": true,
		"chatgpt_oauth_routing": true, "shell_deny_groups": true, "kg_dedup_config": true,
	}

	var setClauses []string
	var args []interface{}

	for field, value := range updates {
		if !allowedFields[field] {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, len(args)+1))
		args = append(args, value)
	}

	if len(setClauses) == 0 {
		return nil
	}

	// 始终更新 updated_at
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", len(args)+1))
	args = append(args, time.Now())

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE agent_id = $%d AND user_id = $%d AND deleted_at IS NULL",
		m.table,
		strings.Join(setClauses, ", "),
		len(args)+1,
		len(args)+2,
	)

	result, err := m.conn.ExecCtx(ctx, query, append(args, agentID, userID)...)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
