package handler

import (
	"context"
	"encoding/json"

	"github.com/pomclaw/pomclaw/internal/store"
	"github.com/pomclaw/pomclaw/pkg/protocol"
	"github.com/pomclaw/pomclaw/pkg/storage"
	"github.com/zeromicro/go-zero/core/logx"
)

// SessionsMethods handles sessions.list, sessions.preview, sessions.delete.
type SessionsMethods struct {
	db storage.ConnectionManager
}

func NewSessionsMethods(db storage.ConnectionManager) *SessionsMethods {
	return &SessionsMethods{db: db}
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
	items, err := store.ListSessionsWithPagination(m.db.DB(), params.AgentID, params.Offset, params.Limit)
	if err != nil {
		logx.Errorf("failed to list sessions: %v", err)
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInternal, "failed to list sessions"))
		return
	}

	// Convert to response format
	sessions := make([]map[string]interface{}, len(items))
	for i, item := range items {
		sessions[i] = map[string]interface{}{
			"key":           item["id"], // Frontend expects "key" field
			"messageCount":  item["message_count"],
			"created":       item["created"],
			"updated":       item["updated"],
			"label":         item["title"],
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
	sessionData, err := store.GetSessionWithMessages(m.db.DB(), params.Key)
	if err != nil {
		logx.Errorf("failed to get session %s: %v", params.Key, err)
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrNotFound, "session not found"))
		return
	}

	client.SendResponse(protocol.NewOKResponse(req.ID, map[string]interface{}{
		"key":      params.Key,
		"messages": sessionData["messages"],
		"summary":  sessionData["summary"],
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
	if err := store.DeleteSession(m.db.DB(), params.Key); err != nil {
		logx.Errorf("failed to delete session %s: %v", params.Key, err)
		client.SendResponse(protocol.NewErrorResponse(req.ID, protocol.ErrInternal, "failed to delete session"))
		return
	}

	client.SendResponse(protocol.NewOKResponse(req.ID, map[string]interface{}{
		"ok": true,
	}))
}
