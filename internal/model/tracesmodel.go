package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TracesModel = (*customTracesModel)(nil)

type (
	// TracesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTracesModel.
	TracesModel interface {
		tracesModel
		withSession(session sqlx.Session) TracesModel

		// List traces with filters
		ListTraces(ctx context.Context, userID string, agentID, sessionKey, status, channel string, limit, offset int) ([]Traces, error)

		// Count traces with filters
		CountTraces(ctx context.Context, userID string, agentID, sessionKey, status, channel string) (int64, error)

		// Find trace by string ID (UUID)
		FindByUUID(ctx context.Context, id string) (*Traces, error)

		// Find child traces
		FindChildTraces(ctx context.Context, parentTraceID string) ([]Traces, error)
	}

	customTracesModel struct {
		*defaultTracesModel
	}
)

// NewTracesModel returns a model for the database table.
func NewTracesModel(conn sqlx.SqlConn) TracesModel {
	return &customTracesModel{
		defaultTracesModel: newTracesModel(conn),
	}
}

func (m *customTracesModel) withSession(session sqlx.Session) TracesModel {
	return NewTracesModel(sqlx.NewSqlConnFromSession(session))
}

// ListTraces lists traces with optional filters
func (m *customTracesModel) ListTraces(ctx context.Context, userID string, agentID, sessionKey, status, channel string, limit, offset int) ([]Traces, error) {
	query := "SELECT * FROM \"public\".\"traces\" WHERE 1=1"
	var args []interface{}
	argIndex := 1

	if userID != "" {
		query += " AND user_id = $" + string(rune('0'+argIndex))
		args = append(args, userID)
		argIndex++
	}
	if agentID != "" {
		query += " AND agent_id = $" + string(rune('0'+argIndex))
		args = append(args, agentID)
		argIndex++
	}
	if sessionKey != "" {
		query += " AND session_key = $" + string(rune('0'+argIndex))
		args = append(args, sessionKey)
		argIndex++
	}
	if status != "" {
		query += " AND status = $" + string(rune('0'+argIndex))
		args = append(args, status)
		argIndex++
	}
	if channel != "" {
		query += " AND channel = $" + string(rune('0'+argIndex))
		args = append(args, channel)
		argIndex++
	}

	query += " ORDER BY created_at DESC LIMIT $" + string(rune('0'+argIndex)) + " OFFSET $" + string(rune('0'+argIndex+1))
	args = append(args, limit, offset)

	var resp []Traces
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, err
}

// CountTraces counts traces with optional filters
func (m *customTracesModel) CountTraces(ctx context.Context, userID string, agentID, sessionKey, status, channel string) (int64, error) {
	query := "SELECT COUNT(*) FROM \"public\".\"traces\" WHERE 1=1"
	var args []interface{}
	argIndex := 1

	if userID != "" {
		query += " AND user_id = $" + string(rune('0'+argIndex))
		args = append(args, userID)
		argIndex++
	}
	if agentID != "" {
		query += " AND agent_id = $" + string(rune('0'+argIndex))
		args = append(args, agentID)
		argIndex++
	}
	if sessionKey != "" {
		query += " AND session_key = $" + string(rune('0'+argIndex))
		args = append(args, sessionKey)
		argIndex++
	}
	if status != "" {
		query += " AND status = $" + string(rune('0'+argIndex))
		args = append(args, status)
		argIndex++
	}
	if channel != "" {
		query += " AND channel = $" + string(rune('0'+argIndex))
		args = append(args, channel)
		argIndex++
	}

	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	return count, err
}

// FindByUUID finds a trace by UUID string
func (m *customTracesModel) FindByUUID(ctx context.Context, id string) (*Traces, error) {
	query := `SELECT * FROM "public"."traces" WHERE CAST(id AS varchar) = $1 LIMIT 1`
	var resp Traces
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

// FindChildTraces finds child traces by parent trace ID
func (m *customTracesModel) FindChildTraces(ctx context.Context, parentTraceID string) ([]Traces, error) {
	query := `SELECT * FROM "public"."traces" WHERE parent_trace_id = $1 ORDER BY created_at DESC`
	var resp []Traces
	err := m.conn.QueryRowsCtx(ctx, &resp, query, parentTraceID)
	return resp, err
}
