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

func (c *wsStreamer) SendEvent(ctx context.Context, agentEvent string, payload any) error {
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
