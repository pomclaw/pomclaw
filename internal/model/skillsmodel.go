package model

import (
	"context"
	"database/sql"
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
		FindByUserIDWithShared(ctx context.Context, userID string) ([]*Skills, error)
		InsertReturningID(ctx context.Context, data *Skills) (int64, error)
		FindEnabledByIDs(ctx context.Context, ids []int64) ([]*Skills, error)
		FindByName(ctx context.Context, userID, name string) (*Skills, error)
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

// FindByUserID 返回指定用户的所有技能（排除已删除）
func (m *customSkillsModel) FindByUserID(ctx context.Context, userID string) ([]*Skills, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE user_id = $1 AND status != 'deleted' ORDER BY created_at DESC",
		skillsRows, m.table,
	)
	var skills []*Skills
	err := m.conn.QueryRowsCtx(ctx, &skills, query, userID)
	return skills, err
}

// FindByUserIDWithShared 返回用户自己的 + 被分享的技能
func (m *customSkillsModel) FindByUserIDWithShared(ctx context.Context, userID string) ([]*Skills, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE (user_id = $1 OR is_shared = true) AND status != 'deleted' ORDER BY is_shared DESC, created_at DESC",
		skillsRows, m.table,
	)
	var skills []*Skills
	err := m.conn.QueryRowsCtx(ctx, &skills, query, userID)
	return skills, err
}

func (m *customSkillsModel) InsertReturningID(ctx context.Context, data *Skills) (int64, error) {
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
		m.table, skillsRowsExpectAutoSet,
	)
	var id int64
	err := m.conn.QueryRowCtx(ctx, &id, query,
		data.UserId, data.Name, data.Slug, data.Description,
		data.Enabled, data.Status, data.Version, data.IsShared,
	)
	return id, err
}

func (m *customSkillsModel) FindEnabledByIDs(ctx context.Context, ids []int64) ([]*Skills, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE id IN (%s) AND enabled = true",
		skillsRows, m.table, joinStrings(placeholders, ","),
	)
	var skills []*Skills
	err := m.conn.QueryRowsCtx(ctx, &skills, query, args...)
	return skills, err
}

func (m *customSkillsModel) FindByName(ctx context.Context, userID, name string) (*Skills, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE user_id = $1 AND name = $2 LIMIT 1",
		skillsRows, m.table,
	)
	var resp Skills
	err := m.conn.QueryRowCtx(ctx, &resp, query, userID, name)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

func NewNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
