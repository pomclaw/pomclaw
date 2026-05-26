package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ToolGrantsModel = (*customToolGrantsModel)(nil)

type (
	// ToolGrantsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customToolGrantsModel.
	ToolGrantsModel interface {
		toolGrantsModel
		withSession(session sqlx.Session) ToolGrantsModel
		Upsert(ctx context.Context, data *ToolGrants) error
		FindByUserIdAndToolName(ctx context.Context, userId string, toolName string) (*ToolGrants, error)
		FindAllByUserId(ctx context.Context, userId string) ([]*ToolGrants, error)
	}

	customToolGrantsModel struct {
		*defaultToolGrantsModel
	}
)

// NewToolGrantsModel returns a model for the database table.
func NewToolGrantsModel(conn sqlx.SqlConn) ToolGrantsModel {
	return &customToolGrantsModel{
		defaultToolGrantsModel: newToolGrantsModel(conn),
	}
}

func (m *customToolGrantsModel) withSession(session sqlx.Session) ToolGrantsModel {
	return NewToolGrantsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customToolGrantsModel) Upsert(ctx context.Context, data *ToolGrants) error {
	query := fmt.Sprintf("insert into %s (user_id, tool_name, enabled, settings, updated_at) values ($1, $2, $3, $4, $5) on conflict (user_id, tool_name) do update set enabled = excluded.enabled, settings = excluded.settings, updated_at = excluded.updated_at",
		m.table)
	_, err := m.conn.ExecCtx(ctx, query, data.UserId, data.ToolName, data.Enabled, data.Settings, data.UpdatedAt)
	return err
}

func (m *customToolGrantsModel) FindByUserIdAndToolName(ctx context.Context, userId string, toolName string) (*ToolGrants, error) {
	query := fmt.Sprintf("select %s from %s where user_id = $1 and tool_name = $2 limit 1", toolGrantsRows, m.table)
	var resp ToolGrants
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId, toolName)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, nil
	default:
		return nil, err
	}
}

func (m *customToolGrantsModel) FindAllByUserId(ctx context.Context, userId string) ([]*ToolGrants, error) {
	query := fmt.Sprintf("select %s from %s where user_id = $1", toolGrantsRows, m.table)
	var resp []*ToolGrants
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	return resp, err
}
