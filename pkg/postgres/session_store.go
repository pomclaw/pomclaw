package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/schema"
	"github.com/pomclaw/pomclaw/internal/model"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// PostgresSession mirrors the file-based Session struct.
type PostgresSession struct {
	SessionId int64            `json:"sessionId"`
	AgentID   string           `json:"agent_id"`
	Messages  []schema.Message `json:"messages"`
	Summary   string           `json:"summary,omitempty"`
	Created   time.Time        `json:"created"`
	Updated   time.Time        `json:"updated"`
}

// SessionStore implements contracts.SessionManagerInterface backed by PostgreSQL.
type SessionStore struct {
	sessionsModel model.SessionsModel
	sessions      map[int64]*PostgresSession
	mu            sync.RWMutex
}

// NewSessionStore creates a new PostgreSQL-backed session store.
func NewSessionStore(sessionsModel model.SessionsModel) *SessionStore {
	ss := &SessionStore{
		sessionsModel: sessionsModel,
		sessions:      make(map[int64]*PostgresSession),
	}
	ss.loadAll()
	return ss
}

// AddMessage adds a simple role/content message to the session.
func (ss *SessionStore) AddMessage(agentID string, sessionId int64, role schema.RoleType, content string) {
	ss.AddFullMessage(agentID, sessionId, schema.Message{
		Role:    schema.RoleType(role),
		Content: content,
	})
}

// AddFullMessage adds a complete message with tool calls.
func (ss *SessionStore) AddFullMessage(agentID string, sessionId int64, msg schema.Message) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	s, ok := ss.sessions[sessionId]
	if !ok {
		s = &PostgresSession{
			SessionId: sessionId,
			AgentID:   agentID,
			Messages:  []schema.Message{},
			Created:   time.Now(),
		}
		ss.sessions[sessionId] = s
	}

	s.Messages = append(s.Messages, msg)
	s.Updated = time.Now()
}

// GetHistory returns a copy of the session's message history.
func (ss *SessionStore) GetHistory(agentID string, sessionId int64) []schema.Message {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	s, ok := ss.sessions[sessionId]
	if !ok {
		return []schema.Message{}
	}

	history := make([]schema.Message, len(s.Messages))
	copy(history, s.Messages)
	return history
}

// SetHistory replaces the session's message history.
func (ss *SessionStore) SetHistory(agentID string, sessionId int64, history []schema.Message) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	s, ok := ss.sessions[sessionId]
	if !ok {
		s = &PostgresSession{
			SessionId: sessionId,
			AgentID:   agentID,
			Messages:  history,
			Created:   time.Now(),
			Updated:   time.Now(),
		}
		ss.sessions[sessionId] = s
		return
	}

	s.Messages = history
	s.Updated = time.Now()
}

// GetSummary returns the session summary.
func (ss *SessionStore) GetSummary(agentID string, sessionId int64) string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	s, ok := ss.sessions[sessionId]
	if !ok {
		return ""
	}
	return s.Summary
}

// SetSummary updates the session summary.
func (ss *SessionStore) SetSummary(agentID string, sessionId int64, summary string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if s, ok := ss.sessions[sessionId]; ok {
		s.Summary = summary
		s.Updated = time.Now()
	}
}

// TruncateHistory truncates the session history to keep the last N messages.
func (ss *SessionStore) TruncateHistory(agentID string, sessionId int64, keepLast int) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if s, ok := ss.sessions[sessionId]; ok {
		if len(s.Messages) > keepLast {
			s.Messages = s.Messages[len(s.Messages)-keepLast:]
			s.Updated = time.Now()
		}
	}
}

// Save persists the session to the database.
func (ss *SessionStore) Save(agentID string, sessionId int64, userID string) error {
	ss.mu.RLock()
	s, ok := ss.sessions[sessionId]
	ss.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %d", sessionId)
	}

	// Encode messages and summary to JSON
	msgData, err := json.Marshal(s.Messages)
	if err != nil {
		return fmt.Errorf("failed to marshal messages: %w", err)
	}
	var label string
	if len(s.Messages) > 0 {
		label = s.Messages[0].Content
	}

	// Save session data (without SessionKey field)
	ctx := context.Background()
	sessionData := &model.Sessions{
		Id:            sessionId,
		UserId:        userID,
		AgentId:       s.AgentID,
		Messages:      sql.NullString{String: string(msgData), Valid: true},
		Summary:       sql.NullString{String: s.Summary, Valid: s.Summary != ""},
		Label:         sql.NullString{String: label, Valid: true},
		MessagesCount: int64(len(s.Messages)),
		InputTokens:   0,
		OutputTokens:  0,
		CreatedAt:     s.Created,
		UpdatedAt:     s.Updated,
	}
	err = ss.sessionsModel.Update(ctx, sessionData)
	if err != nil {
		return fmt.Errorf("session insert failed: %w", err)
	}

	return nil
}

// loadAll pre-populates sessions from PostgreSQL at startup.
// 全量agents加载 有风险
func (ss *SessionStore) loadAll() {
	ctx := context.Background()
	sessions, err := ss.sessionsModel.FindAll(ctx)
	if err != nil {
		logx.Info("postgres", "Failed to load sessions", map[string]interface{}{"error": err.Error()})
		return
	}

	for _, s := range sessions {
		var messages []schema.Message
		if s.Messages.Valid {
			if err := json.Unmarshal([]byte(s.Messages.String), &messages); err != nil {
				logx.Info("postgres", "Failed to unmarshal session messages", map[string]interface{}{
					"agent_id": s.AgentId,
					"error":    err.Error(),
				})
				messages = []schema.Message{}
			}
		}

		sum := ""
		if s.Summary.Valid {
			sum = s.Summary.String
		}

		// Use session_id as key for session indexing
		ss.sessions[s.Id] = &PostgresSession{
			SessionId: s.Id,
			AgentID:   s.AgentId,
			Messages:  messages,
			Summary:   sum,
			Created:   s.CreatedAt,
			Updated:   s.UpdatedAt,
		}
	}
}
