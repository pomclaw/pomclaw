package store

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/pomclaw/pomclaw/pkg/providers"
)

// GatewaySession represents a row in POM_SESSIONS.
type GatewaySession struct {
	ID        string
	AgentID   string
	CreatedAt time.Time
}

// SessionChatMessage is the simplified message format for API responses.
type SessionChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CreateSession inserts a new session and returns it.
func CreateSession(db *sql.DB, userID, agentID, title string) (*GatewaySession, error) {
	sessionID := GenerateID()
	s := &GatewaySession{}
	err := db.QueryRow(
		`INSERT INTO POM_SESSIONS (session_key, agent_id)
		 VALUES ($1, $2)
		 RETURNING session_key, agent_id, created_at`,
		sessionID, agentID,
	).Scan(&s.ID, &s.AgentID, &s.CreatedAt)
	return s, err
}

// ListSessions returns all sessions for the given agent.
func ListSessions(db *sql.DB, agentID string) ([]*GatewaySession, error) {
	rows, err := db.Query(
		`SELECT session_key, agent_id, created_at
		 FROM POM_SESSIONS WHERE agent_id = $1 ORDER BY created_at DESC`,
		agentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*GatewaySession
	for rows.Next() {
		s := &GatewaySession{}
		if err := rows.Scan(&s.ID, &s.AgentID, &s.CreatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// GetSession returns the session with the given id.
func GetSession(db *sql.DB, id string) (*GatewaySession, error) {
	s := &GatewaySession{}
	err := db.QueryRow(
		`SELECT session_key, agent_id, created_at
		 FROM POM_SESSIONS WHERE session_key = $1`,
		id,
	).Scan(&s.ID, &s.AgentID, &s.CreatedAt)
	return s, err
}

// ListSessionsWithPagination returns sessions with pagination support.
func ListSessionsWithPagination(db *sql.DB, agentID string, offset, limit int) ([]map[string]interface{}, error) {
	rows, err := db.Query(
		`SELECT session_key, created_at FROM POM_SESSIONS
		 WHERE agent_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		agentID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id string
		var createdAt time.Time
		if err := rows.Scan(&id, &createdAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]interface{}{
			"id":            id,
			"title":         "",
			"preview":       "",
			"message_count": 0,
			"created":       createdAt.Format(time.RFC3339),
			"updated":       createdAt.Format(time.RFC3339),
		})
	}
	return items, rows.Err()
}

// GetSessionWithMessages returns the session with messages from POM_SESSIONS.
func GetSessionWithMessages(db *sql.DB, id string) (map[string]interface{}, error) {
	// Fetch session with messages and summary from POM_SESSIONS
	var msgJSON sql.NullString
	var summary sql.NullString
	var createdAt, updated time.Time
	err := db.QueryRow(
		`SELECT messages, summary, created_at, updated_at FROM POM_SESSIONS WHERE session_key = $1`,
		id,
	).Scan(&msgJSON, &summary, &createdAt, &updated)
	if err != nil {
		return nil, err
	}

	// Convert messages to simplified format
	messages := []SessionChatMessage{}
	if msgJSON.Valid {
		var providerMsgs []providers.Message
		if err := json.Unmarshal([]byte(msgJSON.String), &providerMsgs); err == nil {
			for _, msg := range providerMsgs {
				messages = append(messages, SessionChatMessage{
					Role:    msg.Role,
					Content: msg.Content,
				})
			}
		}
	}

	summaryStr := ""
	if summary.Valid {
		summaryStr = summary.String
	}

	updatedStr := createdAt.Format(time.RFC3339)
	if !updated.IsZero() {
		updatedStr = updated.Format(time.RFC3339)
	}

	return map[string]interface{}{
		"id":       id,
		"messages": messages,
		"summary":  summaryStr,
		"created":  createdAt.Format(time.RFC3339),
		"updated":  updatedStr,
	}, nil
}

// DeleteSession deletes a session.
func DeleteSession(db *sql.DB, id string) error {
	result, err := db.Exec(
		`DELETE FROM POM_SESSIONS WHERE session_key = $1`,
		id,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
