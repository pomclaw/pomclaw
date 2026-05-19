package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ StateModel = (*customStateModel)(nil)

type (
	// StateModel is an interface to be customized, add more methods here,
	// and implement the added methods in customStateModel.
	StateModel interface {
		stateModel
		withSession(session sqlx.Session) StateModel
	}

	customStateModel struct {
		*defaultStateModel
	}
)

// NewStateModel returns a model for the database table.
func NewStateModel(conn sqlx.SqlConn) StateModel {
	return &customStateModel{
		defaultStateModel: newStateModel(conn),
	}
}

func (m *customStateModel) withSession(session sqlx.Session) StateModel {
	return NewStateModel(sqlx.NewSqlConnFromSession(session))
}
