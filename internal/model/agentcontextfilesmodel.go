package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AgentContextFilesModel = (*customAgentContextFilesModel)(nil)

type (
	// AgentContextFilesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAgentContextFilesModel.
	AgentContextFilesModel interface {
		agentContextFilesModel
		withSession(session sqlx.Session) AgentContextFilesModel
		GetAgentContextFiles(ctx context.Context, agentID string) ([]*AgentContextFiles, error)
		SetAgentContextFile(ctx context.Context, agentID, fileName, content string) error
	}

	customAgentContextFilesModel struct {
		*defaultAgentContextFilesModel
	}
)

// NewAgentContextFilesModel returns a model for the database table.
func NewAgentContextFilesModel(conn sqlx.SqlConn) AgentContextFilesModel {
	return &customAgentContextFilesModel{
		defaultAgentContextFilesModel: newAgentContextFilesModel(conn),
	}
}

func (m *customAgentContextFilesModel) withSession(session sqlx.Session) AgentContextFilesModel {
	return NewAgentContextFilesModel(sqlx.NewSqlConnFromSession(session))
}

// GetAgentContextFiles retrieves all context files for an agent.
func (m *customAgentContextFilesModel) GetAgentContextFiles(ctx context.Context, agentID string) ([]*AgentContextFiles, error) {
	var files []*AgentContextFiles
	query := `SELECT id, user_id, agent_id, file_name, content, created_at, updated_at
	          FROM "public"."agent_context_files"
	          WHERE agent_id = $1
	          ORDER BY file_name`
	err := m.conn.QueryRowsCtx(ctx, &files, query, agentID)
	if err != nil {
		return nil, err
	}
	return files, nil
}

// SetAgentContextFile inserts or updates an agent-level context file.
// Uses INSERT ... ON CONFLICT to handle upsert.
func (m *customAgentContextFilesModel) SetAgentContextFile(ctx context.Context, agentID, fileName, content string) error {
	query := `INSERT INTO "public"."agent_context_files" (user_id, agent_id, file_name, content, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, NOW(), NOW())
	          ON CONFLICT (agent_id, file_name) DO UPDATE
	          SET content = EXCLUDED.content, updated_at = NOW()`
	_, err := m.conn.ExecCtx(ctx, query, "", agentID, fileName, content)
	return err
}
