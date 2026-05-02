package handler

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/pomclaw/pomclaw/pkg/bus"
	"github.com/pomclaw/pomclaw/pkg/protocol"
)

// WSStreamer bridges the message bus to WebSocket clients.
// It subscribes to outbound messages from the agent loop and routes them
// as Protocol v3 events to the appropriate WebSocket clients by session key.
//
// Architecture:
//   Eino Agent Loop
//     ↓ PublishOutbound(OutboundMessage)
//   MessageBus (outbound channel)
//     ↓ WSStreamer.run() subscribes
//   WSStreamer
//     ↓ convertToEvent(OutboundMessage) → EventFrame
//     ↓ FindClientsBySessionKey(sessionKey)
//   WSClient.SendEvent(EventFrame)
//     ↓ JSON + WebSocket write
//   Frontend receives agent events
type WSStreamer struct {
	server *WSServer
	msgBus *bus.MessageBus
	ctx    context.Context
	cancel context.CancelFunc
}

// NewWSStreamer creates a new WebSocket event streamer.
func NewWSStreamer(server *WSServer, msgBus *bus.MessageBus) *WSStreamer {
	ctx, cancel := context.WithCancel(context.Background())
	return &WSStreamer{
		server: server,
		msgBus: msgBus,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start begins the event streaming goroutine.
// Implements the service.Service interface for go-zero service groups.
func (s *WSStreamer) Start() {
	go s.run()
	logx.Info("WebSocket event streamer started")
}

// Stop terminates the event streaming goroutine.
func (s *WSStreamer) Stop() {
	s.cancel()
	logx.Info("WebSocket event streamer stopped")
}

// run is the main event loop that subscribes to outbound messages
// and routes them to WebSocket clients.
func (s *WSStreamer) run() {
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		// Subscribe to outbound message from bus
		msg, ok := s.msgBus.SubscribeOutbound(s.ctx)
		if !ok {
			// Channel closed or context cancelled
			return
		}

		// Convert to Protocol v3 EventFrame
		eventFrame := s.convertToEvent(msg)

		// Route to clients by session key
		clients := s.server.FindClientsBySessionKey(msg.SessionKey)
		if len(clients) == 0 {
			logx.Debugf("no clients for session %s, event type: %s", msg.SessionKey, msg.Type)
			continue
		}

		// Send event to all matching clients
		for _, client := range clients {
			client.SendEvent(eventFrame)
		}

		logx.Debugf("event routed: type=%s, sessionKey=%s, clients=%d", msg.Type, msg.SessionKey, len(clients))
	}
}

// convertToEvent converts an OutboundMessage to a Protocol v3 EventFrame.
// Maps message types to Protocol v3 agent event format.
func (s *WSStreamer) convertToEvent(msg bus.OutboundMessage) protocol.EventFrame {
	// Build payload based on event type
	payload := make(map[string]interface{})

	// Copy all payload fields from message
	for k, v := range msg.Payload {
		payload[k] = v
	}

	// Add type to payload (Protocol v3 agent events have type in payload)
	payload["type"] = msg.Type

	// Add standard fields
	if msg.RunID != "" {
		payload["runId"] = msg.RunID
	}
	if msg.SessionKey != "" {
		payload["sessionKey"] = msg.SessionKey
	}

	// Type-specific payload handling
	switch msg.Type {
	case "run.started":
		// payload already contains runId, sessionKey

	case "chunk":
		// Content goes into payload.content for streaming text
		if msg.Content != "" {
			payload["content"] = msg.Content
		}

	case "thinking":
		// Thinking text goes into payload.content
		if msg.Content != "" {
			payload["content"] = msg.Content
		}

	case "tool.call":
		// payload should contain: id, name, arguments
		// Already in payload from agent loop

	case "tool.result":
		// payload should contain: id, result, is_error
		// Already in payload from agent loop

	case "run.completed":
		// payload should contain: content, usage, media
		// Content goes into payload if not already there
		if msg.Content != "" && payload["content"] == nil {
			payload["content"] = msg.Content
		}

	case "run.failed":
		// payload should contain: error
		// Already in payload from agent loop

	case "activity":
		// payload should contain: phase, tool, iteration
		// Already in payload from agent loop
	}

	// Create Protocol v3 EventFrame
	// Event name is "agent" for all agent-related events
	return protocol.EventFrame{
		Type:    protocol.FrameTypeEvent,
		Event:   "agent", // Protocol v3 uses "agent" as event name with type in payload
		Payload: payload,
	}
}
