package model

import (
	"context"
	"fmt"

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
		// Find all spans for a trace
		FindByTraceId(ctx context.Context, traceId string) ([]*Spans, error)
		// Batch insert multiple spans
		BatchInsert(ctx context.Context, spans []*Spans) error
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

// FindByTraceId finds all spans for a trace
func (m *customSpansModel) FindByTraceId(ctx context.Context, traceId string) ([]*Spans, error) {
	query := `SELECT * FROM "public"."spans" WHERE trace_id = $1 ORDER BY id ASC`
	var resp []*Spans
	err := m.conn.QueryRowsCtx(ctx, &resp, query, traceId)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return []*Spans{}, nil
	default:
		return nil, err
	}
}

// BatchInsert inserts multiple spans in a single query
func (m *customSpansModel) BatchInsert(ctx context.Context, spans []*Spans) error {
	if len(spans) == 0 {
		return nil
	}

	var args []interface{}
	placeholders := ""

	for i, span := range spans {
		offset := i * 23
		placeholders += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7, offset+8, offset+9, offset+10,
			offset+11, offset+12, offset+13, offset+14, offset+15, offset+16, offset+17, offset+18, offset+19, offset+20,
			offset+21, offset+22, offset+23)
		if i < len(spans)-1 {
			placeholders += ","
		}

		args = append(args, span.TraceId, span.ParentSpanId, span.AgentId, span.SpanType, span.Name, span.StartTime,
			span.EndTime, span.DurationMs, span.Status, span.Error, span.Level, span.Model, span.Provider,
			span.InputTokens, span.OutputTokens, span.TotalCost, span.FinishReason, span.ModelParams,
			span.ToolName, span.ToolCallId, span.InputPreview, span.OutputPreview, span.Metadata)
	}

	query := fmt.Sprintf(`INSERT INTO %s (trace_id, parent_span_id, agent_id, span_type, name, start_time,
		end_time, duration_ms, status, error, level, model, provider, input_tokens, output_tokens,
		total_cost, finish_reason, model_params, tool_name, tool_call_id, input_preview, output_preview, metadata)
		VALUES %s`, m.table, placeholders)

	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}
