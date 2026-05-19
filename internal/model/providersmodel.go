package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ProvidersModel = (*customProvidersModel)(nil)

type (
	// ProvidersModel is an interface to be customized, add more methods here,
	// and implement the added methods in customProvidersModel.
	ProvidersModel interface {
		providersModel
		withSession(session sqlx.Session) ProvidersModel
		// 业务查询方法
		FindByUserID(ctx context.Context, userID string) ([]*Providers, error)
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

// FindByUserID 返回指定用户的所有提供商
func (m *customProvidersModel) FindByUserID(ctx context.Context, userID string) ([]*Providers, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE user_id = $1 ORDER BY created_at DESC",
		providersRows, m.table,
	)
	var providers []*Providers
	err := m.conn.QueryRowsCtx(ctx, &providers, query, userID)
	return providers, err
}
