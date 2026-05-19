package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ PromptsModel = (*customPromptsModel)(nil)

type (
	// PromptsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPromptsModel.
	PromptsModel interface {
		promptsModel
		withSession(session sqlx.Session) PromptsModel
		SavePrompt(ctx context.Context, agentID, name, content string) error
		LoadBootstrapFiles(ctx context.Context, agentID string) (map[string]string, error)
	}

	customPromptsModel struct {
		*defaultPromptsModel
	}
)

// NewPromptsModel returns a model for the database table.
func NewPromptsModel(conn sqlx.SqlConn) PromptsModel {
	return &customPromptsModel{
		defaultPromptsModel: newPromptsModel(conn),
	}
}

func (m *customPromptsModel) withSession(session sqlx.Session) PromptsModel {
	return NewPromptsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customPromptsModel) SavePrompt(ctx context.Context, agentID, name, content string) error {
	query := fmt.Sprintf(`INSERT INTO %s (prompt_name, agent_id, content)
	VALUES ($1, $2, $3)
	ON CONFLICT (prompt_name, agent_id) DO UPDATE
	SET content = $3, updated_at = CURRENT_TIMESTAMP`, m.table)
	_, err := m.conn.ExecCtx(ctx, query, name, agentID, content)
	return err
}

func (m *customPromptsModel) LoadBootstrapFiles(ctx context.Context, agentID string) (map[string]string, error) {
	result := make(map[string]string)
	var rows []map[string]interface{}

	query := fmt.Sprintf(`SELECT prompt_name, content FROM %s WHERE agent_id = $1`, m.table)
	err := m.conn.QueryRowsCtx(ctx, &rows, query, agentID)
	if err != nil {
		return result, nil
	}

	for _, row := range rows {
		if name, ok := row["prompt_name"].(string); ok {
			if content, ok := row["content"].(string); ok {
				result[name] = content
			}
		}
	}
	return result, nil
}
