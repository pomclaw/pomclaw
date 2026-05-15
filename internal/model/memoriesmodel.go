package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ MemoriesModel = (*customMemoriesModel)(nil)

type (
	// MemoriesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMemoriesModel.
	MemoriesModel interface {
		memoriesModel
		withSession(session sqlx.Session) MemoriesModel
	}

	customMemoriesModel struct {
		*defaultMemoriesModel
	}
)

// NewMemoriesModel returns a model for the database table.
func NewMemoriesModel(conn sqlx.SqlConn) MemoriesModel {
	return &customMemoriesModel{
		defaultMemoriesModel: newMemoriesModel(conn),
	}
}

func (m *customMemoriesModel) withSession(session sqlx.Session) MemoriesModel {
	return NewMemoriesModel(sqlx.NewSqlConnFromSession(session))
}
