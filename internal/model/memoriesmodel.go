package model

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zeromicro/go-zero/core/stringx"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MemoriesModel = (*customMemoriesModel)(nil)

type (
	// MemoriesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMemoriesModel.
	MemoriesModel interface {
		memoriesModel
		withSession(session sqlx.Session) MemoriesModel

		InsertWithoutID(ctx context.Context, data *Memories) (sql.Result, error)
		ReadLongTerm(ctx context.Context, agentID string) ([]string, error)
		Recall(ctx context.Context, agentID string, limit int) ([]*Memories, error)
	}

	customMemoriesModel struct {
		*defaultMemoriesModel
	}
)

// NewMemoriesModel returns a model for the database table.
func NewMemoriesModel(conn sqlx.SqlConn) MemoriesModel {
	return &customMemoriesModel{
		defaultMemoriesModel: newMemoriesModel(conn),
	}
}

func (m *customMemoriesModel) withSession(session sqlx.Session) MemoriesModel {
	return NewMemoriesModel(sqlx.NewSqlConnFromSession(session))
}

func (m *defaultMemoriesModel) InsertWithoutID(ctx context.Context, data *Memories) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values ($1, $2, $3, $4, $5, $6)", m.table, memoriesRowsExpectAutoSet2)
	ret, err := m.conn.ExecCtx(ctx, query, data.AgentId, data.Content, data.Embedding, data.Importance, data.Category, data.AccessCount)
	return ret, err
}

func (m *customMemoriesModel) ReadLongTerm(ctx context.Context, agentID string) ([]string, error) {
	query := fmt.Sprintf(`SELECT content FROM %s
	WHERE agent_id = $1
	ORDER BY (importance * (1.0 / (1.0 + EXTRACT(DAY FROM (CURRENT_TIMESTAMP - COALESCE(updated_at, created_at))) * 0.1))) DESC,
	         DATE(created_at) DESC
	LIMIT 50`, m.table)
	var results []string
	err := m.conn.QueryRowsCtx(ctx, &results, query, agentID)
	return results, err
}

func (m *customMemoriesModel) Recall(ctx context.Context, agentID string, limit int) ([]*Memories, error) {
	query := fmt.Sprintf("select %s from %s where agent_id = $1 ORDER BY created_at DESC limit $2", memoriesRows, m.table)
	var results []*Memories
	err := m.conn.QueryRowsCtx(ctx, &results, query, agentID, limit)
	return results, err
}

var (
	memoriesRowsExpectAutoSet2 = strings.Join(stringx.Remove(memoriesFieldNames, "id", "create_at", "create_time", "created_at", "update_at", "update_time", "updated_at"), ",")
)
