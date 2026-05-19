package handler

import (
	"context"
	"encoding/json"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/pkg/protocol"
	"github.com/zeromicro/go-zero/core/logx"
)

// SessionsMethods handles sessions.list, sessions.preview, sessions.delete.
type SessionsMethods struct {
	serverCtx *svc.ServiceContext
}

func NewSessionsMethods(svc *svc.ServiceContext) *SessionsMethods {
	return &SessionsMethods{
		serverCtx: svc,
	}
}

func (m *SessionsMethods) Register(router *WSMethodRouter) {
	router.Register(protocol.MethodSessionsList, m.handleList)
	router.Register(protocol.MethodSessionsPreview, m.handlePreview)
	router.Register(protocol.MethodSessionsDelete, m.handleDelete)
	// TODO: Implement MethodSessionsPatch, MethodSessionsReset, MethodSessionsCompact
}

type sessionsListParams struct {
	AgentID string `json:"agentId"`
	Channel string `json:"channel"` // optional: filter by channel prefix ("ws", "telegram")
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
}

func (m *SessionsMethods) handleList(ctx context.Context, client *WSClient, req *protocol.RequestFrame) {
	var params sessionsListParams
	if req.Params != nil {
		json.Unmarshal(req.Params, &params)
	}

	if params.Limit <= 0 {
		params.Limit = 20
	}

	// Get sessions from database
	items, err := m.serverCtx.SessionsModel.FindByAgentIDWithPagination(ctx, params.AgentID, params.Offset, params.Limit)
	if err != nil {
		logx.Errorf("failed to list sessions: %v", err)
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInternal, "failed to list sessions"))
		return
	}

	// Convert to response format
	sessions := make([]map[string]interface{}, len(items))
	for i, item := range items {
		sessions[i] = map[string]interface{}{
			"key":          item.SessionKey,    // Frontend expects "key" field
			"messageCount": item.MessagesCount, //item.Messages.len["message_count"],
			"created":      item.CreatedAt.String(),
			"updated":      item.UpdatedAt.String(),
			"label":        item.Label.String,
		}
	}

	client.SendResponse(protocol.NewOKResponse(req.ID, map[string]interface{}{
		"sessions": sessions,
		"total":    len(sessions),
		"limit":    params.Limit,
		"offset":   params.Offset,
	}))
}

type sessionKeyParams struct {
	Key string `json:"key"`
}

func (m *SessionsMethods) handlePreview(ctx context.Context, client *WSClient, req *protocol.RequestFrame) {
	var params sessionKeyParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "invalid request"))
		return
	}

	if params.Key == "" {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "key is required"))
		return
	}

	// Get session with messages
	sessionData, err := m.serverCtx.SessionsModel.FindOne(ctx, params.Key)
	if err != nil {
		logx.Errorf("failed to get session %s: %v", params.Key, err)
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrNotFound, "session not found"))
		return
	}

	client.SendResponse(protocol.NewOKResponse(req.ID, map[string]interface{}{
		"key":      params.Key,
		"messages": sessionData.Messages.String,
		"summary":  sessionData.Summary,
	}))
}

func (m *SessionsMethods) handleDelete(ctx context.Context, client *WSClient, req *protocol.RequestFrame) {
	var params sessionKeyParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "invalid request"))
		return
	}

	if params.Key == "" {
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInvalidRequest, "key is required"))
		return
	}

	// Delete session
	if err := m.serverCtx.SessionsModel.Delete(ctx, params.Key); err != nil {
		logx.Errorf("failed to delete session %s: %v", params.Key, err)
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInternal, "failed to delete session"))
		return
	}

	client.SendResponse(protocol.NewOKResponse(req.ID, map[string]interface{}{
		"ok": true,
	}))
}
