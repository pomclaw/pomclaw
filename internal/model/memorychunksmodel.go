package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MemoryChunksModel = (*customMemoryChunksModel)(nil)

type (
	// MemoryChunksModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMemoryChunksModel.
	MemoryChunksModel interface {
		memoryChunksModel
		withSession(session sqlx.Session) MemoryChunksModel
		FindByAgentIdAndPath(ctx context.Context, agentId string, path string) ([]*MemoryChunks, error)
		DeleteByAgentIdAndPath(ctx context.Context, agentId string, path string) error
	}

	customMemoryChunksModel struct {
		*defaultMemoryChunksModel
	}
)

// NewMemoryChunksModel returns a model for the database table.
func NewMemoryChunksModel(conn sqlx.SqlConn) MemoryChunksModel {
	return &customMemoryChunksModel{
		defaultMemoryChunksModel: newMemoryChunksModel(conn),
	}
}

func (m *customMemoryChunksModel) withSession(session sqlx.Session) MemoryChunksModel {
	return NewMemoryChunksModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customMemoryChunksModel) FindByAgentIdAndPath(ctx context.Context, agentId string, path string) ([]*MemoryChunks, error) {
	var resp []*MemoryChunks
	query := `select id, document_id, agent_id, user_id, path, start_line, end_line, hash, text, embedding, tsv, custom_scope, created_at, updated_at from "public"."memory_chunks" where agent_id = $1 and path = $2 order by start_line asc`
	err := m.conn.QueryRowsCtx(ctx, &resp, query, agentId, path)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customMemoryChunksModel) DeleteByAgentIdAndPath(ctx context.Context, agentId string, path string) error {
	query := `delete from "public"."memory_chunks" where agent_id = $1 and path = $2`
	_, err := m.conn.ExecCtx(ctx, query, agentId, path)
	return err
}
