package channels

import "time"

// OpenClaw协议消息类型定义

// RequestFrame 客户端请求消息
type RequestFrame struct {
	Type   string                 `json:"type"`   // "req"
	ID     string                 `json:"id"`     // UUID
	Method string                 `json:"method"` // "chat.send"
	Params map[string]interface{} `json:"params"`
}

// ResponseFrame 服务器响应消息
type ResponseFrame struct {
	Type    string      `json:"type"`    // "res"
	ID      string      `json:"id"`      // 对应请求的ID
	OK      bool        `json:"ok"`      // 是否成功
	Payload interface{} `json:"payload,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// EventFrame 服务器事件消息
type EventFrame struct {
	Type         string                 `json:"type"`    // "event"
	Event        string                 `json:"event"`   // 事件类型
	Payload      interface{}            `json:"payload"` // 事件数据
	Seq          int                    `json:"seq,omitempty"`
	StateVersion map[string]int         `json:"stateVersion,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// HelloOkFrame 连接成功消息
type HelloOkFrame struct {
	Type     string                 `json:"type"`     // "hello-ok"
	Protocol int                    `json:"protocol"` // 协议版本
	Server   *ServerInfo            `json:"server,omitempty"`
	Features *Features              `json:"features,omitempty"`
	Auth     *AuthInfo              `json:"auth,omitempty"`
	Policy   map[string]interface{} `json:"policy,omitempty"`
}

// ServerInfo 服务器信息
type ServerInfo struct {
	Version string `json:"version,omitempty"`
	ConnID  string `json:"connId,omitempty"`
}

// Features 支持的功能
type Features struct {
	Methods []string `json:"methods,omitempty"`
	Events  []string `json:"events,omitempty"`
}

// AuthInfo 认证信息
type AuthInfo struct {
	DeviceToken string   `json:"deviceToken,omitempty"`
	Role        string   `json:"role,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
	IssuedAtMs  int64    `json:"issuedAtMs,omitempty"`
}

// Client WebSocket客户端连接
type Client struct {
	ID       string
	UserID   string
	Conn     interface{} // *websocket.Conn
	Send     chan []byte
	SeqNum   int
	Sessions map[string]bool // 当前活跃的session
	mu       interface{}     // sync.Mutex
}

// SessionInfo 会话信息
type SessionInfo struct {
	SessionID string
	AgentID   string
	UserID    string
	CreatedAt time.Time
	UpdatedAt time.Time
	Messages  []Message
}

// Message 消息
type Message struct {
	Role      string    `json:"role"`      // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
