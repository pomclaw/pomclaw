package bus

type InboundMessage struct {
	MessageID  string            `json:"message_id"`
	Channel    string            `json:"channel"`
	SenderID   string            `json:"sender_id"`
	ChatID     string            `json:"chat_id"`
	Content    string            `json:"content"`
	Media      []string          `json:"media,omitempty"`
	SessionKey string            `json:"session_key"`
	AgentID    string            `json:"agent_id"`
	UserID     string            `json:"user_id"`
	RunID      string            `json:"run_id"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// OutboundMessage represents a message from agent to clients.
// Extended for Protocol v3 event streaming support.
type OutboundMessage struct {
	Type       string                 `json:"type"`        // Event type: run.started, chunk, tool.call, tool.result, run.completed, etc.
	SessionKey string                 `json:"session_key"` // For routing to correct WebSocket clients
	RunID      string                 `json:"run_id"`      // Agent run identifier
	Channel    string                 `json:"channel"`     // Original channel (ws, telegram, slack, etc.)
	ChatID     string                 `json:"chat_id"`     // Chat identifier
	Content    string                 `json:"content"`     // Message content (for chunk events)
	Payload    map[string]interface{} `json:"payload"`     // Event-specific data
}
