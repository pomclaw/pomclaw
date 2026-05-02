package handler

import (
	"context"
	"strings"

	"github.com/pomclaw/pomclaw/pkg/bus"
	"github.com/pomclaw/pomclaw/pkg/protocol"
)

// WSStreamDelegate implements bus.StreamDelegate for WebSocket streaming.
type WSStreamDelegate struct {
	server *WSServer
}

// NewWSStreamDelegate creates a new stream delegate.
func NewWSStreamDelegate(server *WSServer) *WSStreamDelegate {
	return &WSStreamDelegate{
		server: server,
	}
}

// GetStreamer returns a streamer for the given channel and chatID.
func (d *WSStreamDelegate) GetStreamer(ctx context.Context, channel, chatID string) (bus.Streamer, bool) {
	if channel != "ws" {
		return nil, false
	}

	// Find client by userID (which is used as chatID)
	client := d.server.FindClientByUserID(chatID)
	if client == nil {
		return nil, false
	}

	return &WSStreamer{
		client: client,
		buffer: strings.Builder{},
	}, true
}

// WSStreamer implements bus.Streamer for WebSocket clients.
type WSStreamer struct {
	client  ClientInterface
	buffer  strings.Builder
	hasPost bool
}

// Update sends an incremental content update to the client.
func (s *WSStreamer) Update(ctx context.Context, content string) error {
	s.buffer.WriteString(content)
	s.hasPost = true

	// Send incremental chunk as chat event
	s.client.SendEvent(protocol.EventFrame{
		Type:  protocol.FrameTypeEvent,
		Event: protocol.EventChat,
		Payload: map[string]interface{}{
			"type":    protocol.ChatEventChunk,
			"content": content,
		},
	})

	return nil
}

// Finalize sends the final message to the client.
func (s *WSStreamer) Finalize(ctx context.Context, content string) error {
	// Send final message event
	s.client.SendEvent(protocol.EventFrame{
		Type:  protocol.FrameTypeEvent,
		Event: protocol.EventChat,
		Payload: map[string]interface{}{
			"type":    protocol.ChatEventMessage,
			"content": s.buffer.String(),
		},
	})

	return nil
}

// Cancel notifies the client that the operation was cancelled.
func (s *WSStreamer) Cancel(ctx context.Context) {
	s.client.SendEvent(protocol.EventFrame{
		Type:  protocol.FrameTypeEvent,
		Event: protocol.EventAgent,
		Payload: map[string]interface{}{
			"type": protocol.AgentEventRunCancelled,
		},
	})
}

// HasPosted returns whether any content has been streamed.
func (s *WSStreamer) HasPosted() bool {
	return s.hasPost
}
