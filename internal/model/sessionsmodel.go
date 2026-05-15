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
