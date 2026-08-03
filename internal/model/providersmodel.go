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
		// 返回用户的providers + 被分享的providers
		FindByUserIDWithShared(ctx context.Context, userID string) ([]*Providers, error)
		// 查询指定名称的provider（用户自己的 或 被分享的）
		FindByNameWithShared(ctx context.Context, userID string, name string) (*Providers, error)
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

// FindByUserIDWithShared 返回用户的providers + 被分享的providers（enabled=true）
func (m *customProvidersModel) FindByUserIDWithShared(ctx context.Context, userID string) ([]*Providers, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE (user_id = $1 OR is_shared = true) AND enabled = true ORDER BY is_shared DESC, created_at DESC",
		providersRows, m.table,
	)
	var providers []*Providers
	err := m.conn.QueryRowsCtx(ctx, &providers, query, userID)
	return providers, err
}

// FindByNameWithShared 查询指定名称的provider（用户自己的 或 被分享的）
// 优先返回用户自己的provider，其次返回被分享的provider
func (m *customProvidersModel) FindByNameWithShared(ctx context.Context, userID string, name string) (*Providers, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE name = $1 AND (user_id = $2 OR is_shared = true) AND enabled = true ORDER BY user_id DESC LIMIT 1",
		providersRows, m.table,
	)
	var resp Providers
	err := m.conn.QueryRowCtx(ctx, &resp, query, name, userID)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
