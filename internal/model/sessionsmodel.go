package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ SessionsModel = (*customSessionsModel)(nil)

type (
	// SessionsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSessionsModel.
	SessionsModel interface {
		sessionsModel
		withSession(session sqlx.Session) SessionsModel
		// 业务查询方法
		FindByAgentID(ctx context.Context, agentID string) ([]*Sessions, error)
		FindByAgentIDWithPagination(ctx context.Context, agentID string, offset, limit int) ([]*Sessions, error)
		FindAll(ctx context.Context) ([]*Sessions, error)
		CountByAgentIDs(ctx context.Context, agentIDs []string) (int, error)
		Upsert(ctx context.Context, data *Sessions) error
	}

	customSessionsModel struct {
		*defaultSessionsModel
	}
)

// NewSessionsModel returns a model for the database table.
func NewSessionsModel(conn sqlx.SqlConn) SessionsModel {
	return &customSessionsModel{
		defaultSessionsModel: newSessionsModel(conn),
	}
}

func (m *customSessionsModel) withSession(session sqlx.Session) SessionsModel {
	return NewSessionsModel(sqlx.NewSqlConnFromSession(session))
}

// FindByAgentID 返回指定 agent 的所有会话
func (m *customSessionsModel) FindByAgentID(ctx context.Context, agentID string) ([]*Sessions, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE agent_id = $1 ORDER BY updated_at DESC",
		sessionsRows, m.table,
	)
	var sessions []*Sessions
	err := m.conn.QueryRowsCtx(ctx, &sessions, query, agentID)
	return sessions, err
}

// FindByAgentIDWithPagination 返回指定 agent 的分页会话列表
func (m *customSessionsModel) FindByAgentIDWithPagination(ctx context.Context, agentID string, offset, limit int) ([]*Sessions, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE agent_id = $1 ORDER BY updated_at DESC LIMIT $2 OFFSET $3",
		sessionsRows, m.table,
	)
	var sessions []*Sessions
	err := m.conn.QueryRowsCtx(ctx, &sessions, query, agentID, limit, offset)
	return sessions, err
}

// FindAll 返回所有会话
func (m *customSessionsModel) FindAll(ctx context.Context) ([]*Sessions, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s ORDER BY updated_at DESC",
		sessionsRows, m.table,
	)
	var sessions []*Sessions
	err := m.conn.QueryRowsCtx(ctx, &sessions, query)
	return sessions, err
}

// CountByAgentIDs 统计指定 agent IDs 的会话总数
func (m *customSessionsModel) CountByAgentIDs(ctx context.Context, agentIDs []string) (int, error) {
	if len(agentIDs) == 0 {
		return 0, nil
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE agent_id = ANY($1)", m.table)
	var count int
	err := m.conn.QueryRowCtx(ctx, &count, query, agentIDs)
	return count, err
}

// Upsert 使用 PostgreSQL ON CONFLICT 进行 insert 或 update
func (m *customSessionsModel) Upsert(ctx context.Context, data *Sessions) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (session_key) DO UPDATE
		SET agent_id=excluded.agent_id, messages=excluded.messages, summary=excluded.summary, label=excluded.label, messages_count=excluded.messages_count, input_tokens=excluded.input_tokens, output_tokens=excluded.output_tokens, updated_at=excluded.updated_at
	`, m.table, sessionsRowsExpectAutoSet)
	_, err := m.conn.ExecCtx(ctx, query, data.SessionKey, data.AgentId, data.Messages, data.Summary, data.Label, data.MessagesCount, data.InputTokens, data.OutputTokens)
	return err
}
