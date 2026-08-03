package handler

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	bus2 "github.com/pomclaw/pomclaw/internal/bus"
	"github.com/pomclaw/pomclaw/internal/logic"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/pomclaw/pomclaw/pkg/protocol"
	"github.com/zeromicro/go-zero/core/logx"
)

// ChatHandlerV3 handles Protocol v3 chat methods: send, history, abort.
// Adapted from PomClaw's implementation but simplified for Pomclaw's architecture.
// Phase 1: No media handling, TTS, or team dispatch.
type ChatHandlerV3 struct {
	serverCtx   *svc.ServiceContext
	rateLimiter *RateLimiter
}

// NewChatHandlerV3 creates a new Protocol v3 chat handler.
func NewChatHandlerV3(svc *svc.ServiceContext, rateLimiter *RateLimiter) *ChatHandlerV3 {
	return &ChatHandlerV3{
		serverCtx:   svc,
		rateLimiter: rateLimiter,
	}
}

// Register adds chat methods to the router.
func (h *ChatHandlerV3) Register(router *WSMethodRouter) {
	router.Register(protocol.MethodChatSend, h.handleSend)
	router.Register(protocol.MethodChatHistory, h.handleHistory)
	router.Register(protocol.MethodChatAbort, h.handleAbort)
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

	var params types.ChatSendParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "invalid JSON"))
		return
	}
	if params.Message == "" {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "message is required"))
		return
	}
	if params.SessionId == 0 {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "session id is required"))
		return
	}
	if params.AgentID == "" {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "agent id is required"))
		return
	}

	userID := client.UserID()
	if userID == "" {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "user_id is required"))
		return
	}
	// Generate run ID
	chatID := uuid.NewString()

	streamer := newStreamer(client, options{
		AgentId:   params.AgentID,
		SessionId: params.SessionId,
		RunId:     chatID,
		UserId:    userID,
		Channel:   "ws",
	})

	l := logic.NewAgentLogic(ctx, h.serverCtx)
	resp, err := l.Chat(streamer, userID, chatID, params)
	if err != nil {
		client.sendError(req.ID, protocol.ErrInternal, err.Error())
		return
	}

	// Immediately send response (don't wait for agent completion)
	client.SendResponse(protocol.NewOKResponse(req.ID, resp))
}

// handleHistory retrieves conversation history for a session.
func (h *ChatHandlerV3) handleHistory(ctx context.Context, client *WSClient, req *protocol.RequestFrame) {
	var params types.ChatHistoryParams
	if err := json.Unmarshal(req.Params, &params); err != nil || params.SessionId == 0 {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "sessionKey is required"))
		return
	}

	// Extract agentID from sessionKey (format: agent:{agentId}:ws:direct:system:{uuid})
	// For Phase 1, use default agent
	agentID := "default"

	// Load conversation history from session store
	history := h.serverCtx.SessionManager.GetHistory(agentID, params.SessionId)

	client.SendResponse(protocol.NewOKResponse(req.ID, map[string]any{
		"sessionId": params.SessionId,
		"messages":  bus2.ConvertMessages(history),
	}))
}

// handleAbort cancels a running agent execution.
// Phase 1: Simple implementation - signals abort via context cancellation.
func (h *ChatHandlerV3) handleAbort(ctx context.Context, client *WSClient, req *protocol.RequestFrame) {
	var params types.ChatAbortParams
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

	//abortMsg := bus.InboundMessage{
	//	MessageID:  uuid.NewString(),
	//	SessionKey: params.SessionKey,
	//	UserID:     client.UserID(),
	//	Content:    "__ABORT__",
	//	Channel:    "ws",
	//	ChatID:     client.UserID(),
	//	RunID:      params.RunID,
	//	Metadata: map[string]string{
	//		"type": "abort",
	//	},
	//}
	//
	//h.msgBus.PublishInbound(abortMsg)

	logx.Infof("chat.abort published: sessionKey=%s, runId=%s", params.SessionKey, params.RunID)

	client.SendResponse(protocol.NewOKResponse(req.ID, map[string]any{
		"aborted": true,
	}))
}
