package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ProvidersModel = (*customProvidersModel)(nil)

type (
	// ProvidersModel is an interface to be customized, add more methods here,
	// and implement the added methods in customProvidersModel.
	ProvidersModel interface {
		providersModel
		withSession(session sqlx.Session) ProvidersModel
	}

	customProvidersModel struct {
		*defaultProvidersModel
	}
)

// NewProvidersModel returns a model for the database table.
func NewProvidersModel(conn sqlx.SqlConn) ProvidersModel {
	return &customProvidersModel{
		defaultProvidersModel: newProvidersModel(conn),
	}
}

func (m *customProvidersModel) withSession(session sqlx.Session) ProvidersModel {
	return NewProvidersModel(sqlx.NewSqlConnFromSession(session))
}
