package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ PromptsModel = (*customPromptsModel)(nil)

type (
	// PromptsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPromptsModel.
	PromptsModel interface {
		promptsModel
		withSession(session sqlx.Session) PromptsModel
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
