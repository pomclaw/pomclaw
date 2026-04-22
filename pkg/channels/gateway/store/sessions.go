package store

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/pomclaw/pomclaw/pkg/providers"
)

// GatewaySession represents a row in pom_gateway_sessions.
type GatewaySession struct {
	ID        string
	UserID    string
	AgentID   string
	Title     string
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
		`INSERT INTO pom_gateway_sessions (id, user_id, agent_id, title)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, agent_id, title, created_at`,
		sessionID, userID, agentID, title,
	).Scan(&s.ID, &s.UserID, &s.AgentID, &s.Title, &s.CreatedAt)
	return s, err
}

// ListSessions returns all sessions owned by the given user.
func ListSessions(db *sql.DB, userID string) ([]*GatewaySession, error) {
	rows, err := db.Query(
		`SELECT id, user_id, agent_id, title, created_at
		 FROM pom_gateway_sessions WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*GatewaySession
	for rows.Next() {
		s := &GatewaySession{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.AgentID, &s.Title, &s.CreatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// GetSession returns the session with the given id owned by the given user.
func GetSession(db *sql.DB, id, userID string) (*GatewaySession, error) {
	s := &GatewaySession{}
	err := db.QueryRow(
		`SELECT id, user_id, agent_id, title, created_at
		 FROM pom_gateway_sessions WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&s.ID, &s.UserID, &s.AgentID, &s.Title, &s.CreatedAt)
	return s, err
}

// ListSessionsWithPagination returns sessions with pagination support.
func ListSessionsWithPagination(db *sql.DB, userID string, offset, limit int) ([]map[string]interface{}, error) {
	rows, err := db.Query(
		`SELECT id, title, created_at FROM pom_gateway_sessions
		 WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id, title string
		var createdAt time.Time
		if err := rows.Scan(&id, &title, &createdAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]interface{}{
			"id":            id,
			"title":         title,
			"preview":       title,
			"message_count": 0,
			"created":       createdAt.Format(time.RFC3339),
			"updated":       createdAt.Format(time.RFC3339),
		})
	}
	return items, rows.Err()
}

// GetSessionWithMessages returns the session with messages from POM_SESSIONS.
func GetSessionWithMessages(db *sql.DB, id, userID string) (map[string]interface{}, error) {
	s := &GatewaySession{}
	err := db.QueryRow(
		`SELECT id, user_id, agent_id, title, created_at
		 FROM pom_gateway_sessions WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&s.ID, &s.UserID, &s.AgentID, &s.Title, &s.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Fetch messages and summary from POM_SESSIONS
	var msgJSON sql.NullString
	var summary sql.NullString
	var updated time.Time
	err = db.QueryRow(
		`SELECT messages, summary, updated_at FROM POM_SESSIONS WHERE session_key = $1`,
		id,
	).Scan(&msgJSON, &summary, &updated)

	// Convert messages to simplified format
	messages := []SessionChatMessage{}
	if err == nil && msgJSON.Valid {
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

	updatedStr := s.CreatedAt.Format(time.RFC3339)
	if err == nil && !updated.IsZero() {
		updatedStr = updated.Format(time.RFC3339)
	}

	return map[string]interface{}{
		"id":       s.ID,
		"messages": messages,
		"summary":  summaryStr,
		"created":  s.CreatedAt.Format(time.RFC3339),
		"updated":  updatedStr,
	}, nil
}

// DeleteSession deletes a session owned by the given user.
func DeleteSession(db *sql.DB, id, userID string) error {
	result, err := db.Exec(
		`DELETE FROM pom_gateway_sessions WHERE id = $1 AND user_id = $2`,
		id, userID,
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
