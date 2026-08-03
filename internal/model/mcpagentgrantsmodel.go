package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ McpAgentGrantsModel = (*customMcpAgentGrantsModel)(nil)

type (
	// McpAgentGrantsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMcpAgentGrantsModel.
	McpAgentGrantsModel interface {
		mcpAgentGrantsModel
		withSession(session sqlx.Session) McpAgentGrantsModel
		FindByServerID(ctx context.Context, serverID int64) ([]*McpAgentGrants, error)
		FindByAgentID(ctx context.Context, agentID string) ([]*McpAgentGrants, error)
		CountByServerIDs(ctx context.Context, serverIDs []int64) (map[int64]int64, error)
		DeleteByServerIDAgentID(ctx context.Context, serverID int64, agentID string) error
	}

	customMcpAgentGrantsModel struct {
		*defaultMcpAgentGrantsModel
	}
)

// NewMcpAgentGrantsModel returns a model for the database table.
func NewMcpAgentGrantsModel(conn sqlx.SqlConn) McpAgentGrantsModel {
	return &customMcpAgentGrantsModel{
		defaultMcpAgentGrantsModel: newMcpAgentGrantsModel(conn),
	}
}

func (m *customMcpAgentGrantsModel) withSession(session sqlx.Session) McpAgentGrantsModel {
	return NewMcpAgentGrantsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customMcpAgentGrantsModel) FindByServerID(ctx context.Context, serverID int64) ([]*McpAgentGrants, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE server_id = $1 ORDER BY created_at DESC",
		mcpAgentGrantsRows, m.table,
	)
	var grants []*McpAgentGrants
	err := m.conn.QueryRowsCtx(ctx, &grants, query, serverID)
	return grants, err
}

func (m *customMcpAgentGrantsModel) FindByAgentID(ctx context.Context, agentID string) ([]*McpAgentGrants, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE agent_id = $1 ORDER BY created_at DESC",
		mcpAgentGrantsRows, m.table,
	)
	var grants []*McpAgentGrants
	err := m.conn.QueryRowsCtx(ctx, &grants, query, agentID)
	return grants, err
}

func (m *customMcpAgentGrantsModel) CountByServerIDs(ctx context.Context, serverIDs []int64) (map[int64]int64, error) {
	if len(serverIDs) == 0 {
		return nil, nil
	}
	query := fmt.Sprintf("SELECT server_id, COUNT(*) FROM %s WHERE server_id = ANY($1) GROUP BY server_id", m.table)
	type row struct {
		ServerID int64 `db:"server_id"`
		Count    int64 `db:"count"`
	}
	var rows []row
	err := m.conn.QueryRowsCtx(ctx, &rows, query, serverIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[int64]int64, len(rows))
	for _, r := range rows {
		result[r.ServerID] = r.Count
	}
	return result, nil
}

func (m *customMcpAgentGrantsModel) DeleteByServerIDAgentID(ctx context.Context, serverID int64, agentID string) error {
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE server_id = $1 AND agent_id = $2",
		m.table,
	)
	_, err := m.conn.ExecCtx(ctx, query, serverID, agentID)
	return err
}