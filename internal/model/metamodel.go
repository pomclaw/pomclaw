package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ MetaModel = (*customMetaModel)(nil)

type (
	// MetaModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMetaModel.
	MetaModel interface {
		metaModel
		withSession(session sqlx.Session) MetaModel
	}

	customMetaModel struct {
		*defaultMetaModel
	}
)

// NewMetaModel returns a model for the database table.
func NewMetaModel(conn sqlx.SqlConn) MetaModel {
	return &customMetaModel{
		defaultMetaModel: newMetaModel(conn),
	}
}

func (m *customMetaModel) withSession(session sqlx.Session) MetaModel {
	return NewMetaModel(sqlx.NewSqlConnFromSession(session))
}
