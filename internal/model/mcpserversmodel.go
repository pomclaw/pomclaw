package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ McpServersModel = (*customMcpServersModel)(nil)

type (
	// McpServersModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMcpServersModel.
	McpServersModel interface {
		mcpServersModel
		withSession(session sqlx.Session) McpServersModel
		FindByUserID(ctx context.Context, userID string) ([]*McpServers, error)
		FindByUserIDWithShared(ctx context.Context, userID string) ([]*McpServers, error)
		FindByUserIDName(ctx context.Context, userID string, name string) (*McpServers, error)
	}

	customMcpServersModel struct {
		*defaultMcpServersModel
	}
)

// NewMcpServersModel returns a model for the database table.
func NewMcpServersModel(conn sqlx.SqlConn) McpServersModel {
	return &customMcpServersModel{
		defaultMcpServersModel: newMcpServersModel(conn),
	}
}

func (m *customMcpServersModel) withSession(session sqlx.Session) McpServersModel {
	return NewMcpServersModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customMcpServersModel) FindByUserID(ctx context.Context, userID string) ([]*McpServers, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE user_id = $1 ORDER BY created_at DESC",
		mcpServersRows, m.table,
	)
	var servers []*McpServers
	err := m.conn.QueryRowsCtx(ctx, &servers, query, userID)
	return servers, err
}

// FindByUserIDWithShared 返回用户的 MCP servers + 被分享的 MCP servers
func (m *customMcpServersModel) FindByUserIDWithShared(ctx context.Context, userID string) ([]*McpServers, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE (user_id = $1 OR is_shared = true) AND enabled = true ORDER BY is_shared DESC, created_at DESC",
		mcpServersRows, m.table,
	)
	var servers []*McpServers
	err := m.conn.QueryRowsCtx(ctx, &servers, query, userID)
	return servers, err
}

func (m *customMcpServersModel) FindByUserIDName(ctx context.Context, userID string, name string) (*McpServers, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE user_id = $1 AND name = $2 LIMIT 1",
		mcpServersRows, m.table,
	)
	var resp McpServers
	err := m.conn.QueryRowCtx(ctx, &resp, query, userID, name)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}