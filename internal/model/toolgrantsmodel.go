package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ToolGrantsModel = (*customToolGrantsModel)(nil)

type (
	// ToolGrantsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customToolGrantsModel.
	ToolGrantsModel interface {
		toolGrantsModel
		withSession(session sqlx.Session) ToolGrantsModel
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
