package handler

import (
	"context"
	"github.com/pomclaw/pomclaw/pkg/bus"
	"github.com/pomclaw/pomclaw/pkg/protocol"
)

type wsStreamer struct {
	c   *WSClient
	opt options
}

func (c *wsStreamer) PublishRunStarted(ctx context.Context, payload *bus.RunStartedPayload) error {
	return c.publishOutbound(ctx, protocol.AgentEventRunStarted, payload)
}

func (c *wsStreamer) PublishRunCompleted(ctx context.Context, payload *bus.RunCompletedPayload) error {
	return c.publishOutbound(ctx, protocol.AgentEventRunCompleted, payload)
}

func (c *wsStreamer) PublishToolCall(ctx context.Context, payload *bus.ToolCallPayload) error {
	return c.publishOutbound(ctx, protocol.AgentEventToolCall, payload)
}

func (c *wsStreamer) PublishToolResult(ctx context.Context, payload *bus.ToolResultPayload) error {
	return c.publishOutbound(ctx, protocol.AgentEventToolResult, payload)
}

func (c *wsStreamer) PublishChunk(ctx context.Context, payload *bus.ChunkPayload) error {
	return c.publishOutbound(ctx, protocol.ChatEventChunk, payload)
}

func (c *wsStreamer) publishOutbound(ctx context.Context, agentEvent string, payload any) error {
	c.c.SendEvent(protocol.EventFrame{
		Type: protocol.FrameTypeEvent,
		// Protocol v3 uses "agent" as event name with type in payload
		Event: "agent",
		Payload: Payload{
			Type:       agentEvent,
			AgentId:    c.opt.AgentId,
			RunId:      c.opt.RunId,
			Payload:    payload,
			UserId:     c.opt.UserId,
			Channel:    c.opt.Channel,
			SessionKey: c.opt.SessionKey,
		},
	})
	return nil
}

func newStreamer(c *WSClient, opt options) bus.Streamer {
	return &wsStreamer{
		c:   c,
		opt: opt,
	}
}

type options struct {
	AgentId    string `json:"agentId"`
	RunId      string `json:"runId"`
	UserId     string `json:"userId"`
	Channel    string `json:"channel"`
	SessionKey string `json:"sessionKey"`
}

type Payload struct {
	Type       string      `json:"type"`
	AgentId    string      `json:"agentId"`
	RunId      string      `json:"runId"`
	Payload    interface{} `json:"payload"`
	UserId     string      `json:"userId"`
	Channel    string      `json:"channel"`
	SessionKey string      `json:"sessionKey"`
}
