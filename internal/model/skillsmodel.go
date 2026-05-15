package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ SkillsModel = (*customSkillsModel)(nil)

type (
	// SkillsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSkillsModel.
	SkillsModel interface {
		skillsModel
		withSession(session sqlx.Session) SkillsModel
		// 业务查询方法
		FindByUserID(ctx context.Context, userID string) ([]*Skills, error)
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

// FindByUserID 返回指定用户的所有技能
func (m *customSkillsModel) FindByUserID(ctx context.Context, userID string) ([]*Skills, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE user_id = $1 ORDER BY created_at DESC",
		skillsRows, m.table,
	)
	var skills []*Skills
	err := m.conn.QueryRowsCtx(ctx, &skills, query, userID)
	return skills, err
}
