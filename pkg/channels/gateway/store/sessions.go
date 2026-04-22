package store

import (
	"database/sql"
	"time"
)

// GatewaySession represents a row in pom_gateway_sessions.
type GatewaySession struct {
	ID        string
	UserID    string
	AgentID   string
	Title     string
	CreatedAt time.Time
}

// CreateSession inserts a new session and returns it.
func CreateSession(db *sql.DB, userID, agentID, title string) (*GatewaySession, error) {
	s := &GatewaySession{}
	err := db.QueryRow(
		`INSERT INTO pom_gateway_sessions (user_id, agent_id, title)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, agent_id, title, created_at`,
		userID, agentID, title,
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
