package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ SkillGrantsModel = (*customSkillGrantsModel)(nil)

type (
	SkillGrantsModel interface {
		skillGrantsModel
		withSession(session sqlx.Session) SkillGrantsModel
		ToggleSkillGrant(ctx context.Context, userId string, skillId int64, agentId string, enabled bool) error
		FindByAgentID(ctx context.Context, agentID string) ([]*SkillGrants, error)
		FindBySkillID(ctx context.Context, skillID int64) ([]*SkillGrants, error)
		CountBySkillIDs(ctx context.Context, skillIDs []int64) (map[int64]int64, error)
		InsertGrant(ctx context.Context, userID string, skillID int64, agentID string) error
		DeleteBySkillIDAgentID(ctx context.Context, skillID int64, agentID string) error
	}

	customSkillGrantsModel struct {
		*defaultSkillGrantsModel
	}
)

func NewSkillGrantsModel(conn sqlx.SqlConn) SkillGrantsModel {
	return &customSkillGrantsModel{
		defaultSkillGrantsModel: newSkillGrantsModel(conn),
	}
}

func (m *customSkillGrantsModel) withSession(session sqlx.Session) SkillGrantsModel {
	return NewSkillGrantsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customSkillGrantsModel) ToggleSkillGrant(ctx context.Context, userId string, skillId int64, agentId string, enabled bool) error {
	query := fmt.Sprintf(
		"INSERT INTO %s (user_id, skill_id, agent_id, version, enabled) VALUES ($1, $2, $3, 1, $4) "+
			"ON CONFLICT (skill_id, agent_id) DO UPDATE SET enabled = $4",
		m.table,
	)
	_, err := m.conn.ExecCtx(ctx, query, userId, skillId, agentId, enabled)
	return err
}

func (m *customSkillGrantsModel) FindByAgentID(ctx context.Context, agentID string) ([]*SkillGrants, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE agent_id = $1 AND enabled = true ORDER BY created_at DESC",
		skillGrantsRows, m.table,
	)
	var grants []*SkillGrants
	err := m.conn.QueryRowsCtx(ctx, &grants, query, agentID)
	return grants, err
}

func (m *customSkillGrantsModel) FindBySkillID(ctx context.Context, skillID int64) ([]*SkillGrants, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE skill_id = $1 ORDER BY created_at DESC",
		skillGrantsRows, m.table,
	)
	var grants []*SkillGrants
	err := m.conn.QueryRowsCtx(ctx, &grants, query, skillID)
	return grants, err
}

func (m *customSkillGrantsModel) CountBySkillIDs(ctx context.Context, skillIDs []int64) (map[int64]int64, error) {
	if len(skillIDs) == 0 {
		return nil, nil
	}
	query := fmt.Sprintf("SELECT skill_id, COUNT(*) FROM %s WHERE skill_id = ANY($1) AND enabled = true GROUP BY skill_id", m.table)
	type row struct {
		SkillID int64 `db:"skill_id"`
		Count   int64 `db:"count"`
	}
	var rows []row
	err := m.conn.QueryRowsCtx(ctx, &rows, query, skillIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[int64]int64, len(rows))
	for _, r := range rows {
		result[r.SkillID] = r.Count
	}
	return result, nil
}

func (m *customSkillGrantsModel) InsertGrant(ctx context.Context, userID string, skillID int64, agentID string) error {
	query := fmt.Sprintf(
		"INSERT INTO %s (user_id, skill_id, agent_id, version, enabled) VALUES ($1, $2, $3, 1, true) ON CONFLICT (skill_id, agent_id) DO UPDATE SET enabled = true",
		m.table,
	)
	_, err := m.conn.ExecCtx(ctx, query, userID, skillID, agentID)
	return err
}

func (m *customSkillGrantsModel) DeleteBySkillIDAgentID(ctx context.Context, skillID int64, agentID string) error {
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE skill_id = $1 AND agent_id = $2",
		m.table,
	)
	_, err := m.conn.ExecCtx(ctx, query, skillID, agentID)
	return err
}
