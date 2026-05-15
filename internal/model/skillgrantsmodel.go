package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SkillGrantsModel = (*customSkillGrantsModel)(nil)

type (
	// SkillGrantsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSkillGrantsModel.
	SkillGrantsModel interface {
		skillGrantsModel
		withSession(session sqlx.Session) SkillGrantsModel
	}

	customSkillGrantsModel struct {
		*defaultSkillGrantsModel
	}
)

// NewSkillGrantsModel returns a model for the database table.
func NewSkillGrantsModel(conn sqlx.SqlConn) SkillGrantsModel {
	return &customSkillGrantsModel{
		defaultSkillGrantsModel: newSkillGrantsModel(conn),
	}
}

func (m *customSkillGrantsModel) withSession(session sqlx.Session) SkillGrantsModel {
	return NewSkillGrantsModel(sqlx.NewSqlConnFromSession(session))
}
