package handler

import (
	"github.com/pomclaw/pomclaw/pkg/bus"
	"github.com/pomclaw/pomclaw/pkg/protocol"
)

type wsStreamer struct {
	c *WSClient
}

func (c *wsStreamer) PublishOutbound(message bus.OutboundMessage) {
	c.c.SendEvent(convertToEvent(message))
}

// convertToEvent converts an OutboundMessage to a Protocol v3 EventFrame.
// Maps message types to Protocol v3 agent event format.
func convertToEvent(msg bus.OutboundMessage) protocol.EventFrame {
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
