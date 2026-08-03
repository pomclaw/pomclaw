package types

// ChatSendParams represents the parameters for chat.send method.
type ChatSendParams struct {
	Message   string `json:"message"`
	AgentID   string `json:"agentId"`
	SessionId int64  `json:"sessionId"`
	Stream    bool   `json:"stream"`
}

type ChatPayload struct {
	Content string `json:"content"`
	RunId   string `json:"runId"`
	Usage   Usage  `json:"usage"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatHistoryParams represents the parameters for chat.history method.
type ChatHistoryParams struct {
	SessionId int64 `json:"sessionId"`
	Limit     int   `json:"limit,omitempty"`
}

// ChatAbortParams represents the parameters for chat.abort method.
type ChatAbortParams struct {
	SessionKey string `json:"sessionKey"`
	RunID      string `json:"runId"`
}
