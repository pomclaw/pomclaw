package agent

import "github.com/pomclaw/pomclaw/pkg/providers"

const (
	DefaultAgentID = "default"
)

// PromptStoreInterface is an optional interface for Oracle-backed prompt storage.
type PromptStoreInterface interface {
	LoadBootstrapFiles(agentID string) map[string]string
}

// MemoryStoreInterface defines the contract for memory storage backends.
// Both file-based (MemoryStore) and Oracle-backed implementations satisfy this.
type MemoryStoreInterface interface {
	ReadLongTerm(agentID string) string
	WriteLongTerm(agentID string, content string) error
	ReadToday(agentID string) string
	AppendToday(agentID string, content string) error
	GetRecentDailyNotes(agentID string, days int) string
	GetMemoryContext(agentID string) string
}

// SessionManagerInterface defines the contract for session management backends.
type SessionManagerInterface interface {
	AddMessage(key, role, content string)
	AddFullMessage(key string, msg providers.Message)
	GetHistory(key string) []providers.Message
	SetHistory(key string, history []providers.Message)
	GetSummary(key string) string
	SetSummary(key, summary string)
	TruncateHistory(key string, keepLast int)
	Save(key string) error
}

// StateManagerInterface defines the contract for state management backends.
type StateManagerInterface interface {
	SetLastChannel(agentID string, channel string) error
	GetLastChannel(agentID string) string
	SetLastChatID(agentID string, chatID string) error
	GetLastChatID(agentID string) string
}

// OracleMemoryStore is an extended interface for Oracle-backed memory with vector search.
type OracleMemoryStore interface {
	MemoryStoreInterface
	Remember(agentID string, text string, importance float64, category string) (string, error)
	Recall(agentID string, query string, maxResults int) ([]MemoryRecallResult, error)
	Forget(agentID string, memoryID string) error
}

// MemoryRecallResult represents a single recalled memory with similarity score.
type MemoryRecallResult struct {
	MemoryID   string  `json:"memory_id"`
	Text       string  `json:"text"`
	Importance float64 `json:"importance"`
	Category   string  `json:"category"`
	Score      float64 `json:"score"`
}
