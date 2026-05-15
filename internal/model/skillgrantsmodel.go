package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ SkillGrantsModel = (*customSkillGrantsModel)(nil)

type (
	// SkillGrantsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSkillGrantsModel.
	SkillGrantsModel interface {
		skillGrantsModel
		withSession(session sqlx.Session) SkillGrantsModel
		// 业务方法
		GrantSkillToAgent(ctx context.Context, skillID, agentID string, version int64) error
		RevokeSkillFromAgent(ctx context.Context, skillID, agentID string) error
		CheckSkillGranted(ctx context.Context, skillID, agentID string) (bool, error)
		FindByAgentID(ctx context.Context, agentID string) ([]*SkillGrants, error)
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

// GrantSkillToAgent 授权技能给 agent（如果已存在则更新版本）
func (m *customSkillGrantsModel) GrantSkillToAgent(ctx context.Context, skillID, agentID string, version int64) error {
	query := fmt.Sprintf(
		"INSERT INTO %s (skill_id, agent_id, version) VALUES ($1, $2, $3) ON CONFLICT (skill_id, agent_id) DO UPDATE SET version = $3",
		m.table,
	)
	_, err := m.conn.ExecCtx(ctx, query, skillID, agentID, version)
	return err
}

// RevokeSkillFromAgent 撤销 agent 的技能授权
func (m *customSkillGrantsModel) RevokeSkillFromAgent(ctx context.Context, skillID, agentID string) error {
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE skill_id = $1 AND agent_id = $2",
		m.table,
	)
	result, err := m.conn.ExecCtx(ctx, query, skillID, agentID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// CheckSkillGranted 检查技能是否被授予给 agent
func (m *customSkillGrantsModel) CheckSkillGranted(ctx context.Context, skillID, agentID string) (bool, error) {
	_, err := m.FindOneBySkillIdAgentId(ctx, skillID, agentID)
	if err == nil {
		return true, nil
	}
	if err == ErrNotFound {
		return false, nil
	}
	return false, err
}

// FindByAgentID 返回授予给指定 agent 的所有技能授权
func (m *customSkillGrantsModel) FindByAgentID(ctx context.Context, agentID string) ([]*SkillGrants, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE agent_id = $1 ORDER BY created_at DESC",
		skillGrantsRows, m.table,
	)
	var grants []*SkillGrants
	err := m.conn.QueryRowsCtx(ctx, &grants, query, agentID)
	return grants, err
}
