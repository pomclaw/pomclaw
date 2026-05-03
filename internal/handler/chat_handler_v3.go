package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/pomclaw/pomclaw/internal/config"
	"github.com/pomclaw/pomclaw/pkg/agent"
	"github.com/pomclaw/pomclaw/pkg/bus"
	"github.com/pomclaw/pomclaw/pkg/contracts"
	"github.com/pomclaw/pomclaw/pkg/protocol"
)

// ChatHandlerV3 handles Protocol v3 chat methods: send, history, abort.
// Adapted from GoClaw's implementation but simplified for Pomclaw's architecture.
// Phase 1: No media handling, TTS, or team dispatch.
type ChatHandlerV3 struct {
	cfg         *config.Config
	agentLoop   *agent.AgentLoop
	sessions    contracts.SessionManagerInterface
	msgBus      *bus.MessageBus
	rateLimiter *RateLimiter
}

// NewChatHandlerV3 creates a new Protocol v3 chat handler.
func NewChatHandlerV3(cfg *config.Config, agentLoop *agent.AgentLoop, sessions contracts.SessionManagerInterface, msgBus *bus.MessageBus, rateLimiter *RateLimiter) *ChatHandlerV3 {
	return &ChatHandlerV3{
		cfg:         cfg,
		agentLoop:   agentLoop,
		sessions:    sessions,
		msgBus:      msgBus,
		rateLimiter: rateLimiter,
	}
}

// Register adds chat methods to the router.
func (h *ChatHandlerV3) Register(router *WSMethodRouter) {
	router.Register(protocol.MethodChatSend, h.handleSend)
	router.Register(protocol.MethodChatHistory, h.handleHistory)
	router.Register(protocol.MethodChatAbort, h.handleAbort)
}

// chatSendParams represents the parameters for chat.send method.
type chatSendParams struct {
	Message    string `json:"message"`
	AgentID    string `json:"agentId"`
	SessionKey string `json:"sessionKey"`
	Stream     bool   `json:"stream"`
}

// handleSend processes a chat.send request.
// Phase 1: Publishes message to bus, returns immediately (async agent execution).
func (h *ChatHandlerV3) handleSend(ctx context.Context, client *WSClient, req *protocol.RequestFrame) {
	// Rate limit check per user/client
	if h.rateLimiter != nil && h.rateLimiter.Enabled() {
		key := client.UserID()
		if key == "" {
			key = client.ID()
		}
		if !h.rateLimiter.Allow(key) {
			client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "rate limit exceeded"))
			return
		}
	}

	var params chatSendParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "invalid JSON"))
		return
	}

	if params.Message == "" {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "message is required"))
		return
	}

	// Default agent
	if params.AgentID == "" {
		params.AgentID = "default"
	}

	userID := client.UserID()
	if userID == "" {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "user_id is required"))
		return
	}

	// Generate session key if not provided
	sessionKey := params.SessionKey
	if sessionKey == "" {
		// Format: agent:{agentId}:ws:direct:system:{uuid}
		sessionKey = fmt.Sprintf("agent:%s:ws:direct:system:%s", params.AgentID, uuid.NewString()[:8])
	}

	// Generate run ID
	runID := uuid.NewString()

	// Set active session key on client for event routing
	client.SetActiveSessionKey(sessionKey)

	// Publish to message bus (inbound channel)
	// The agent loop will pick this up and process it asynchronously
	inboundMsg := bus.InboundMessage{
		MessageID:  uuid.NewString(),
		SessionKey: sessionKey,
		AgentID:    params.AgentID,
		UserID:     userID,
		Content:    params.Message,
		Channel:    "ws",
		ChatID:     userID,
		RunID:      runID,
		Metadata: map[string]string{
			"stream": fmt.Sprintf("%v", params.Stream),
		},
	}

	// Publish to bus
	h.msgBus.PublishInbound(inboundMsg)

	logx.Infof("chat.send published to bus: runId=%s, sessionKey=%s, user=%s", runID, sessionKey, userID)

	// Immediately send response (don't wait for agent completion)
	client.SendResponse(protocol.NewOKResponse(req.ID, map[string]any{
		"runId":      runID,
		"sessionKey": sessionKey,
		"status":     "processing",
	}))
}

// chatHistoryParams represents the parameters for chat.history method.
type chatHistoryParams struct {
	SessionKey string `json:"sessionKey"`
	Limit      int    `json:"limit,omitempty"`
}

// handleHistory retrieves conversation history for a session.
func (h *ChatHandlerV3) handleHistory(ctx context.Context, client *WSClient, req *protocol.RequestFrame) {
	var params chatHistoryParams
	if err := json.Unmarshal(req.Params, &params); err != nil || params.SessionKey == "" {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "sessionKey is required"))
		return
	}

	// Extract agentID from sessionKey (format: agent:{agentId}:ws:direct:system:{uuid})
	// For Phase 1, use default agent
	agentID := "default"

	// Load conversation history from session store
	history := h.sessions.GetHistory(agentID, params.SessionKey)

	// Convert to response format
	messages := make([]map[string]any, 0, len(history))
	for _, msg := range history {
		messages = append(messages, map[string]any{
			"role":    msg.Role,
			"content": msg.Content,
			// Add timestamp if available from metadata
			"timestamp": 0,
		})
	}

	client.SendResponse(protocol.NewOKResponse(req.ID, map[string]any{
		"sessionKey": params.SessionKey,
		"messages":   messages,
	}))
}

// chatAbortParams represents the parameters for chat.abort method.
type chatAbortParams struct {
	SessionKey string `json:"sessionKey"`
	RunID      string `json:"runId"`
}

// handleAbort cancels a running agent execution.
// Phase 1: Simple implementation - signals abort via context cancellation.
func (h *ChatHandlerV3) handleAbort(ctx context.Context, client *WSClient, req *protocol.RequestFrame) {
	var params chatAbortParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "invalid JSON"))
		return
	}

	if params.SessionKey == "" && params.RunID == "" {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "sessionKey or runId is required"))
		return
	}

	// TODO: Implement abort mechanism in AgentLoop
	// For Phase 1, we'll publish an abort message to the bus
	// The agent loop will need to handle this message type

	abortMsg := bus.InboundMessage{
		MessageID:  uuid.NewString(),
		SessionKey: params.SessionKey,
		UserID:     client.UserID(),
		Content:    "__ABORT__",
		Channel:    "ws",
		ChatID:     client.UserID(),
		RunID:      params.RunID,
		Metadata: map[string]string{
			"type": "abort",
		},
	}

	h.msgBus.PublishInbound(abortMsg)

	logx.Infof("chat.abort published: sessionKey=%s, runId=%s", params.SessionKey, params.RunID)

	client.SendResponse(protocol.NewOKResponse(req.ID, map[string]any{
		"aborted": true,
	}))
}
