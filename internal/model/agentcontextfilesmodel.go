package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	_ AgentContextFilesModel = (*customAgentContextFilesModel)(nil)

	// agentContextFilesRowsNoContent is like agentContextFilesRows but excludes the content field
	// for efficient metadata-only queries
	agentContextFilesRowsNoContent = strings.Join(
		stringx.Remove(agentContextFilesFieldNames, "content"),
		",",
	)
)

type (
	// AgentContextFilesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAgentContextFilesModel.
	AgentContextFilesModel interface {
		agentContextFilesModel
		withSession(session sqlx.Session) AgentContextFilesModel
		GetAgentContextFiles(ctx context.Context, agentID string) ([]*AgentContextFiles, error)
		SetAgentContextFile(ctx context.Context, data *AgentContextFiles) error
		ListAllContextFilesMetadata(ctx context.Context, agentID string) ([]*AgentContextFiles, error)
		SearchContextFilesMetadata(ctx context.Context, agentID string, pathPrefix string) ([]*AgentContextFiles, error)
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
	query := fmt.Sprintf("select %s from %s where agent_id = $1 order by file_name", agentContextFilesRows, m.table)
	err := m.conn.QueryRowsCtx(ctx, &files, query, agentID)
	if err != nil {
		return nil, err
	}
	return files, nil
}

// SetAgentContextFile inserts or updates an agent-level context file.
// Uses INSERT ... ON CONFLICT to handle upsert.
func (m *customAgentContextFilesModel) SetAgentContextFile(ctx context.Context, data *AgentContextFiles) error {
	query := `INSERT INTO "public"."agent_context_files" (user_id, agent_id, file_type, file_name, content, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	          ON CONFLICT (agent_id, file_name) DO UPDATE
	          SET content = EXCLUDED.content, updated_at = NOW()`
	_, err := m.conn.ExecCtx(ctx, query, data.UserId, data.AgentId, data.FileType, data.FileName, data.Content)
	return err
}

// ListAllContextFilesMetadata retrieves all context files for an agent (metadata only, without content).
// Used for listing operations - significantly faster than loading all content.
func (m *customAgentContextFilesModel) ListAllContextFilesMetadata(ctx context.Context, agentID string) ([]*AgentContextFiles, error) {
	var files []*AgentContextFiles
	query := fmt.Sprintf("select %s from %s where agent_id = $1 order by file_name", agentContextFilesRowsNoContent, m.table)
	err := m.conn.QueryRowsCtx(ctx, &files, query, agentID)
	if err != nil {
		return nil, err
	}
	return files, nil
}

// SearchContextFilesMetadata searches context files by path prefix (metadata only, without content).
// Matches files starting with the given prefix pattern.
func (m *customAgentContextFilesModel) SearchContextFilesMetadata(ctx context.Context, agentID string, pathPrefix string) ([]*AgentContextFiles, error) {
	var files []*AgentContextFiles
	// Use LIKE for prefix matching: "SOUL%" matches "SOUL", "SOUL.md", etc.
	query := fmt.Sprintf("select %s from %s where agent_id = $1 and file_name like $2 order by file_name", agentContextFilesRowsNoContent, m.table)
	searchPattern := pathPrefix + "%"
	err := m.conn.QueryRowsCtx(ctx, &files, query, agentID, searchPattern)
	if err != nil {
		return nil, err
	}
	return files, nil
}
