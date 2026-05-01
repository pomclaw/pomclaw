package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
)

type BindingRecord struct {
	ID                  int
	UserID              string
	UserName            string
	FamilyID            int
	AgentID             string
	ImBotToken          string
	ImUserToken         string
	ImChannelID         string
	MattermostBotUserID string
	MattermostUserID    string
}

type AccountStore interface {
	ListActive(ctx context.Context) ([]BindingRecord, error)
	Close() error
}

type BindingStore struct {
	db    *sql.DB
	table string
}

var validTableName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func NewBindingStore(db *sql.DB, table string) (*BindingStore, error) {
	if table == "" {
		table = "user_binding_relations"
	}
	if !validTableName.MatchString(table) {
		return nil, fmt.Errorf("invalid table name: %q", table)
	}
	return &BindingStore{db: db, table: table}, nil
}

func (s *BindingStore) ListActive(ctx context.Context) ([]BindingRecord, error) {
	query := fmt.Sprintf(
		`SELECT id, user_id, user_name, family_id,
			COALESCE(agent_id, ''), COALESCE(im_bot_token, ''),
			COALESCE(im_user_token, ''), COALESCE(im_channel_id, ''),
			COALESCE(mattermost_bot_user_id, ''), COALESCE(mattermost_user_id, '')
		FROM %s
		WHERE is_deleted = 0 AND im_bot_token IS NOT NULL AND im_bot_token != ''`,
		s.table,
	)

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query bindings: %w", err)
	}
	defer rows.Close()

	var records []BindingRecord
	for rows.Next() {
		var r BindingRecord
		if err := rows.Scan(
			&r.ID, &r.UserID, &r.UserName, &r.FamilyID,
			&r.AgentID, &r.ImBotToken, &r.ImUserToken, &r.ImChannelID,
			&r.MattermostBotUserID, &r.MattermostUserID,
		); err != nil {
			return nil, fmt.Errorf("scan binding: %w", err)
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func (s *BindingStore) Close() error {
	return nil
}

type StaticStore struct {
	records []BindingRecord
}

func NewStaticStore(records []BindingRecord) *StaticStore {
	return &StaticStore{records: records}
}

func (s *StaticStore) ListActive(_ context.Context) ([]BindingRecord, error) {
	return s.records, nil
}

func (s *StaticStore) Close() error {
	return nil
}
