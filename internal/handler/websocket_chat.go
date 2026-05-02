package handler

import (
	"context"
	"encoding/json"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/pomclaw/pomclaw/pkg/bus"
	"github.com/pomclaw/pomclaw/pkg/contracts"
	"github.com/pomclaw/pomclaw/pkg/protocol"
)

// WSChatHandler handles chat-related RPC methods.
type WSChatHandler struct {
	msgBus   *bus.MessageBus
	sessions contracts.SessionManagerInterface
}

// NewWSChatHandler creates a new chat handler.
func NewWSChatHandler(msgBus *bus.MessageBus, sessions contracts.SessionManagerInterface) *WSChatHandler {
	return &WSChatHandler{
		msgBus:   msgBus,
		sessions: sessions,
	}
}

// Register adds chat methods to the router.
func (h *WSChatHandler) Register(router *WSMethodRouter) {
	router.Register(protocol.MethodChatSend, h.handleSend)
	router.Register(protocol.MethodChatHistory, h.handleHistory)
	router.Register(protocol.MethodChatAbort, h.handleAbort)
}

type chatSendParams struct {
	Message    string `json:"message"`
	AgentID    string `json:"agentId"`
	SessionKey string `json:"sessionKey"`
}

// handleSend processes chat.send requests.
func (h *WSChatHandler) handleSend(ctx context.Context, client ClientInterface, req *protocol.RequestFrame) {
	var params chatSendParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "invalid JSON"))
		return
	}

	if params.Message == "" {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "message is required"))
		return
	}

	if params.AgentID == "" {
		params.AgentID = contracts.DefaultAgentID
	}

	if params.SessionKey == "" {
		// Generate a new session key if not provided
		params.SessionKey = "ws:default:" + client.UserID()
	}

	userID := client.UserID()
	if userID == "" {
		userID = "anonymous"
	}

	logx.Info("chat.send:", map[string]interface{}{
		"client":      client.ID(),
		"user_id":     userID,
		"agent_id":    params.AgentID,
		"session_key": params.SessionKey,
		"message_len": len(params.Message),
	})

	// Publish message to bus for agent processing
	msg := bus.InboundMessage{
		Channel:    "ws",
		SenderID:   client.ID(),
		ChatID:     userID,
		Content:    params.Message,
		SessionKey: params.SessionKey,
		Metadata: map[string]string{
			contracts.MetadataKey_AgentId:   params.AgentID,
			contracts.MetadataKey_Workspace: contracts.DefaultWorkspace,
		},
	}

	h.msgBus.PublishInbound(msg)

	// Acknowledge receipt immediately
	// Actual response will come via streaming events
	client.SendResponse(protocol.NewOKResponse(req.ID, map[string]any{
		"acknowledged": true,
		"sessionKey":   params.SessionKey,
	}))
}

type chatHistoryParams struct {
	AgentID    string `json:"agentId"`
	SessionKey string `json:"sessionKey"`
}

// handleHistory retrieves conversation history.
func (h *WSChatHandler) handleHistory(ctx context.Context, client ClientInterface, req *protocol.RequestFrame) {
	var params chatHistoryParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "invalid JSON"))
		return
	}

	if params.SessionKey == "" {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "sessionKey is required"))
		return
	}

	history := h.sessions.GetHistory(params.AgentID, params.SessionKey)

	client.SendResponse(protocol.NewOKResponse(req.ID, map[string]any{
		"messages": history,
	}))
}

type chatAbortParams struct {
	SessionKey string `json:"sessionKey"`
}

// handleAbort cancels a running conversation.
func (h *WSChatHandler) handleAbort(ctx context.Context, client ClientInterface, req *protocol.RequestFrame) {
	var params chatAbortParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "invalid JSON"))
		return
	}

	if params.SessionKey == "" {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "sessionKey is required"))
		return
	}

	// For Phase 1, we'll implement a simple abort mechanism
	// In the future, this should signal the agent loop to stop processing
	logx.Info("chat.abort:", map[string]interface{}{
		"client":      client.ID(),
		"session_key": params.SessionKey,
	})

	client.SendResponse(protocol.NewOKResponse(req.ID, map[string]any{
		"ok":      true,
		"aborted": true,
	}))
}
