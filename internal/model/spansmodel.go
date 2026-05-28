package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ SpansModel = (*customSpansModel)(nil)

type (
	// SpansModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSpansModel.
	SpansModel interface {
		spansModel
		withSession(session sqlx.Session) SpansModel

		// Find span by string ID (UUID)
		FindByUUID(ctx context.Context, id string) (*Spans, error)
	}

	customSpansModel struct {
		*defaultSpansModel
	}
)

// NewSpansModel returns a model for the database table.
func NewSpansModel(conn sqlx.SqlConn) SpansModel {
	return &customSpansModel{
		defaultSpansModel: newSpansModel(conn),
	}
}

func (m *customSpansModel) withSession(session sqlx.Session) SpansModel {
	return NewSpansModel(sqlx.NewSqlConnFromSession(session))
}

// FindByUUID finds a span by UUID string
func (m *customSpansModel) FindByUUID(ctx context.Context, id string) (*Spans, error) {
	query := `SELECT * FROM "public"."spans" WHERE CAST(id AS varchar) = $1 LIMIT 1`
	var resp Spans
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
