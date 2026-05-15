package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SkillsModel = (*customSkillsModel)(nil)

type (
	// SkillsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSkillsModel.
	SkillsModel interface {
		skillsModel
		withSession(session sqlx.Session) SkillsModel
	}

	customSkillsModel struct {
		*defaultSkillsModel
	}
)

// NewSkillsModel returns a model for the database table.
func NewSkillsModel(conn sqlx.SqlConn) SkillsModel {
	return &customSkillsModel{
		defaultSkillsModel: newSkillsModel(conn),
	}
}

func (m *customSkillsModel) withSession(session sqlx.Session) SkillsModel {
	return NewSkillsModel(sqlx.NewSqlConnFromSession(session))
}
