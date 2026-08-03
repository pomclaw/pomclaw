package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UsersModel = (*customUsersModel)(nil)

type (
	// UsersModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUsersModel.
	UsersModel interface {
		usersModel
		withSession(session sqlx.Session) UsersModel
		FindOneByUserId(ctx context.Context, userID string) (*Users, error)
		FindByUserIDs(ctx context.Context, userIDs []string) ([]*Users, error)
	}

	customUsersModel struct {
		*defaultUsersModel
	}
)

// NewUsersModel returns a model for the database table.
func NewUsersModel(conn sqlx.SqlConn) UsersModel {
	return &customUsersModel{
		defaultUsersModel: newUsersModel(conn),
	}
}

func (m *customUsersModel) withSession(session sqlx.Session) UsersModel {
	return NewUsersModel(sqlx.NewSqlConnFromSession(session))
}

// FindByUserIDs 批量通过 user_id 查询用户
func (m *customUsersModel) FindByUserIDs(ctx context.Context, userIDs []string) ([]*Users, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE user_id = ANY($1)",
		usersRows, m.table,
	)
	var users []*Users
	err := m.conn.QueryRowsCtx(ctx, &users, query, userIDs)
	return users, err
}

// FindOneByUserId 通过 user_id (UUID) 获取用户
func (m *customUsersModel) FindOneByUserId(ctx context.Context, userID string) (*Users, error) {
	query := fmt.Sprintf("select %s from %s where user_id = $1 limit 1", usersRows, m.table)
	var resp Users
	err := m.conn.QueryRowCtx(ctx, &resp, query, userID)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
