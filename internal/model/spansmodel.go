package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SpansModel = (*customSpansModel)(nil)

type (
	// SpansModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSpansModel.
	SpansModel interface {
		spansModel
		withSession(session sqlx.Session) SpansModel
	}

	customSpansModel struct {
		*defaultSpansModel
	}
)

// NewSpansModel returns a model for the database table.
func NewSpansModel(conn sqlx.SqlConn) SpansModel {
	return &customSpansModel{
		defaultSpansModel: newSpansModel(conn),
	}
}

func (m *customSpansModel) withSession(session sqlx.Session) SpansModel {
	return NewSpansModel(sqlx.NewSqlConnFromSession(session))
}
