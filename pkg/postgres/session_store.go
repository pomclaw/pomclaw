package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/pomclaw/pomclaw/pkg/logger"
	"github.com/pomclaw/pomclaw/pkg/providers"
)

// PostgresSession mirrors the file-based Session struct.
type PostgresSession struct {
	Key      string              `json:"key"`
	AgentID  string              `json:"agent_id"`
	Messages []providers.Message `json:"messages"`
	Summary  string              `json:"summary,omitempty"`
	Created  time.Time           `json:"created"`
	Updated  time.Time           `json:"updated"`
}

// SessionStore implements SessionManagerInterface backed by PostgreSQL.
type SessionStore struct {
	db       *sql.DB
	sessions map[string]*PostgresSession
	mu       sync.RWMutex
}

// NewSessionStore creates a new PostgreSQL-backed session store.
func NewSessionStore(db *sql.DB) *SessionStore {
	ss := &SessionStore{
		db:       db,
		sessions: make(map[string]*PostgresSession),
	}
	ss.loadAll()
	return ss
}

// AddMessage adds a simple role/content message to the session.
func (ss *SessionStore) AddMessage(agentID string, key, role, content string) {
	ss.AddFullMessage(agentID, key, providers.Message{
		Role:    role,
		Content: content,
	})
}

// AddFullMessage adds a complete message with tool calls.
func (ss *SessionStore) AddFullMessage(agentID string, key string, msg providers.Message) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	s, ok := ss.sessions[key]
	if !ok {
		s = &PostgresSession{
			Key:      key,
			AgentID:  agentID,
			Messages: []providers.Message{},
			Created:  time.Now(),
		}
		ss.sessions[key] = s
	}

	s.Messages = append(s.Messages, msg)
	s.Updated = time.Now()
}

// GetHistory returns a copy of the session's message history.
func (ss *SessionStore) GetHistory(agentID string, key string) []providers.Message {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	s, ok := ss.sessions[key]
	if !ok {
		return []providers.Message{}
	}

	history := make([]providers.Message, len(s.Messages))
	copy(history, s.Messages)
	return history
}

// SetHistory replaces the session's message history.
func (ss *SessionStore) SetHistory(agentID string, key string, history []providers.Message) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	s, ok := ss.sessions[key]
	if !ok {
		s = &PostgresSession{
			Key:      key,
			AgentID:  agentID,
			Messages: history,
			Created:  time.Now(),
			Updated:  time.Now(),
		}
		ss.sessions[key] = s
		return
	}

	s.Messages = history
	s.Updated = time.Now()
}

// GetSummary returns the session summary.
func (ss *SessionStore) GetSummary(agentID string, key string) string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	s, ok := ss.sessions[key]
	if !ok {
		return ""
	}
	return s.Summary
}

// SetSummary updates the session summary.
func (ss *SessionStore) SetSummary(agentID string, key, summary string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if s, ok := ss.sessions[key]; ok {
		s.Summary = summary
		s.Updated = time.Now()
	}
}

// TruncateHistory truncates the session history to keep the last N messages.
func (ss *SessionStore) TruncateHistory(agentID string, key string, keepLast int) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if s, ok := ss.sessions[key]; ok {
		if len(s.Messages) > keepLast {
			s.Messages = s.Messages[len(s.Messages)-keepLast:]
			s.Updated = time.Now()
		}
	}
}

// Save persists the session to the database.
func (ss *SessionStore) Save(agentID string, key string) error {
	ss.mu.RLock()
	s, ok := ss.sessions[key]
	ss.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", key)
	}

	// Encode messages and summary to JSON
	msgData, err := json.Marshal(s.Messages)
	if err != nil {
		return fmt.Errorf("failed to marshal messages: %w", err)
	}

	// Upsert into database using PostgreSQL ON CONFLICT syntax
	_, err = ss.db.Exec(`
		INSERT INTO POM_SESSIONS (session_key, agent_id, messages, summary, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (session_key) DO UPDATE
		SET messages = $3, summary = $4, updated_at = $6
	`, key, s.AgentID, string(msgData), s.Summary, s.Created, s.Updated)

	if err != nil {
		return fmt.Errorf("session save failed: %w", err)
	}

	return nil
}

// loadAll pre-populates sessions from PostgreSQL at startup.
// 全量agents加载 有风险
func (ss *SessionStore) loadAll() {
	rows, err := ss.db.Query(`
		SELECT session_key, agent_id, messages, summary, created_at, updated_at
		FROM POM_SESSIONS
	`)
	if err != nil {
		logger.WarnCF("postgres", "Failed to load sessions", map[string]interface{}{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		var agentID string
		var msgJSON string
		var summary sql.NullString
		var created, updated time.Time

		if err := rows.Scan(&key, &agentID, &msgJSON, &summary, &created, &updated); err != nil {
			logger.WarnCF("postgres", "Failed to scan session row", map[string]interface{}{"error": err.Error()})
			continue
		}

		var messages []providers.Message
		if err := json.Unmarshal([]byte(msgJSON), &messages); err != nil {
			logger.WarnCF("postgres", "Failed to unmarshal session messages", map[string]interface{}{"error": err.Error()})
			messages = []providers.Message{}
		}

		sum := ""
		if summary.Valid {
			sum = summary.String
		}

		ss.sessions[key] = &PostgresSession{
			Key:      key,
			AgentID:  agentID,
			Messages: messages,
			Summary:  sum,
			Created:  created,
			Updated:  updated,
		}
	}
}
