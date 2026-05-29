package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MemoryDocumentsModel = (*customMemoryDocumentsModel)(nil)

type (
	// MemoryDocumentsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMemoryDocumentsModel.
	MemoryDocumentsModel interface {
		memoryDocumentsModel
		withSession(session sqlx.Session) MemoryDocumentsModel
		FindByAgentId(ctx context.Context, agentId string) ([]*MemoryDocuments, error)
		FindByAgentIdAndUserId(ctx context.Context, agentId string, userId string) ([]*MemoryDocuments, error)
		DeleteByAgentIdAndPath(ctx context.Context, agentId string, path string) error
	}

	customMemoryDocumentsModel struct {
		*defaultMemoryDocumentsModel
	}
)

// NewMemoryDocumentsModel returns a model for the database table.
func NewMemoryDocumentsModel(conn sqlx.SqlConn) MemoryDocumentsModel {
	return &customMemoryDocumentsModel{
		defaultMemoryDocumentsModel: newMemoryDocumentsModel(conn),
	}
}

func (m *customMemoryDocumentsModel) withSession(session sqlx.Session) MemoryDocumentsModel {
	return NewMemoryDocumentsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customMemoryDocumentsModel) FindByAgentId(ctx context.Context, agentId string) ([]*MemoryDocuments, error) {
	var resp []*MemoryDocuments
	query := `select id, agent_id, user_id, path, content, hash, custom_scope, created_at, updated_at from "public"."memory_documents" where agent_id = $1 order by updated_at desc`
	err := m.conn.QueryRowsCtx(ctx, &resp, query, agentId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customMemoryDocumentsModel) FindByAgentIdAndUserId(ctx context.Context, agentId string, userId string) ([]*MemoryDocuments, error) {
	var resp []*MemoryDocuments
	query := `select id, agent_id, user_id, path, content, hash, custom_scope, created_at, updated_at from "public"."memory_documents" where agent_id = $1 and user_id = $2 order by updated_at desc`
	err := m.conn.QueryRowsCtx(ctx, &resp, query, agentId, userId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customMemoryDocumentsModel) DeleteByAgentIdAndPath(ctx context.Context, agentId string, path string) error {
	query := `delete from "public"."memory_documents" where agent_id = $1 and path = $2`
	_, err := m.conn.ExecCtx(ctx, query, agentId, path)
	return err
}
