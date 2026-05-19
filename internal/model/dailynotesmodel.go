package model

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ DailyNotesModel = (*customDailyNotesModel)(nil)

type (
	// DailyNotesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDailyNotesModel.
	DailyNotesModel interface {
		dailyNotesModel
		withSession(session sqlx.Session) DailyNotesModel
		ReadToday(ctx context.Context, agentID string) (string, error)
		GetRecentDailyNotes(ctx context.Context, agentID string, days int) ([]string, error)
		Upsert(ctx context.Context, agentID string, content string) error
	}

	customDailyNotesModel struct {
		*defaultDailyNotesModel
	}
)

// NewDailyNotesModel returns a model for the database table.
func NewDailyNotesModel(conn sqlx.SqlConn) DailyNotesModel {
	return &customDailyNotesModel{
		defaultDailyNotesModel: newDailyNotesModel(conn),
	}
}

func (m *customDailyNotesModel) withSession(session sqlx.Session) DailyNotesModel {
	return NewDailyNotesModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customDailyNotesModel) ReadToday(ctx context.Context, agentID string) (string, error) {
	var content *string
	query := fmt.Sprintf(`SELECT content FROM %s
	WHERE agent_id = $1 AND note_date = CURRENT_DATE
	ORDER BY updated_at DESC
	LIMIT 1`, m.table)
	err := m.conn.QueryRowCtx(ctx, &content, query, agentID)
	if err != nil || content == nil {
		return "", nil
	}
	return *content, nil
}

func (m *customDailyNotesModel) Upsert(ctx context.Context, agentID string, content string) error {
	query := fmt.Sprintf(`INSERT INTO %s (note_id, agent_id, note_date, content)
	VALUES (gen_random_uuid(), $1, CURRENT_DATE, $2)
	ON CONFLICT (agent_id, note_date) DO UPDATE
	SET content = CONCAT(excluded.content, '\n', $2), updated_at = CURRENT_TIMESTAMP`, m.table)
	_, err := m.conn.ExecCtx(ctx, query, agentID, content)
	return err
}

func (m *customDailyNotesModel) GetRecentDailyNotes(ctx context.Context, agentID string, days int) ([]string, error) {
	var results []string
	query := fmt.Sprintf(`SELECT content FROM %s
	WHERE agent_id = $1 AND note_date >= CURRENT_DATE - INTERVAL '1 day' * $2
	ORDER BY note_date DESC`, m.table)
	err := m.conn.QueryRowsCtx(ctx, &results, query, agentID, days)
	return results, err
}

func generateDailyNoteID() string {
	return uuid.New().String()[:8]
}
