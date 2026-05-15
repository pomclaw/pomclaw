package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ AgentsModel = (*customAgentsModel)(nil)

type (
	// AgentsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAgentsModel.
	AgentsModel interface {
		agentsModel
		withSession(session sqlx.Session) AgentsModel
	}

	customAgentsModel struct {
		*defaultAgentsModel
	}
)

// NewAgentsModel returns a model for the database table.
func NewAgentsModel(conn sqlx.SqlConn) AgentsModel {
	return &customAgentsModel{
		defaultAgentsModel: newAgentsModel(conn),
	}
}

func (m *customAgentsModel) withSession(session sqlx.Session) AgentsModel {
	return NewAgentsModel(sqlx.NewSqlConnFromSession(session))
}
