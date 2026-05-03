// Pomclaw - Ultra-lightweight personal AI agent
// Powered by Eino Framework
// License: MIT
//
// Copyright (c) 2026 Pomclaw contributors

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/cloudwego/eino/callbacks"
	ecmodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/pomclaw/pomclaw/pkg/bus"
	"github.com/pomclaw/pomclaw/pkg/protocol"
	"github.com/zeromicro/go-zero/core/logx"
)

// StreamCallback implements Eino's callback interface for streaming output
type StreamCallback struct {
	callbacks.HandlerBuilder

	bus        *bus.MessageBus
	runID      string
	sessionKey string
	channel    string
	chatID     string
}

func NewStreamCallback(msgBus *bus.MessageBus, runID, sessionKey, channel, chatID string) callbacks.Handler {
	return &StreamCallback{
		bus:        msgBus,
		runID:      runID,
		sessionKey: sessionKey,
		channel:    channel,
		chatID:     chatID,
	}
}

func (cb *StreamCallback) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo, input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	logx.Info("StreamCallback.OnStartWithStreamInput called")
	defer input.Close()
	return ctx
}

// OnEndWithStreamOutput handles streaming output from the agent
// This is the key method that receives incremental chunks in real-time
func (cb *StreamCallback) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo,
	output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {

	logx.Infof("StreamCallback.OnEndWithStreamOutput called: name=%s, component=%s, type=%s",
		info.Name, info.Component, info.Type)

	go func() {
		defer output.Close() // Always close the stream
		frameCount := 0

		for {
			frame, err := output.Recv()
			if errors.Is(err, io.EOF) {
				// Stream ended normally
				logx.Infof("Stream ended normally after %d frames", frameCount)
				break
			}
			if err != nil {
				logx.Errorf("Stream recv error: %v", err)
				return
			}

			frameCount++
			logx.Infof("Received frame #%d, type: %T", frameCount, frame)

			// Process different frame types
			switch v := frame.(type) {
			case *schema.Message:
				cb.handleMessage(ctx, v)
			case *ecmodel.CallbackOutput:
				if v.Message != nil {
					cb.handleMessage(ctx, v.Message)
				}
			case []*schema.Message:
				for _, m := range v {
					cb.handleMessage(ctx, m)
				}
			default:
				logx.Infof("Unknown frame type: %T", v)
			}
		}
	}()

	return ctx
}

func (cb *StreamCallback) handleMessage(ctx context.Context, msg *schema.Message) {
	if msg == nil {
		return
	}

	logx.Infof("handleMessage: role=%s, content_len=%d, tool_calls=%d",
		msg.Role, len(msg.Content), len(msg.ToolCalls))

	// Handle tool calls - send each tool call as a separate event
	if len(msg.ToolCalls) > 0 {
		for _, toolCall := range msg.ToolCalls {
			logx.Infof("Publishing tool call event: id=%s, name=%s", toolCall.ID, toolCall.Function.Name)

			// Parse arguments string to object
			var argsObj map[string]interface{}
			if toolCall.Function.Arguments != "" {
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &argsObj); err != nil {
					logx.Errorf("Failed to parse tool arguments: %v", err)
					argsObj = map[string]interface{}{"raw": toolCall.Function.Arguments}
				}
			}

			// Send in flattened format expected by frontend
			cb.bus.PublishOutbound(bus.OutboundMessage{
				Type:       protocol.AgentEventToolCall,
				SessionKey: cb.sessionKey,
				RunID:      cb.runID,
				Channel:    cb.channel,
				ChatID:     cb.chatID,
				Payload: map[string]interface{}{
					"id":        toolCall.ID,
					"name":      toolCall.Function.Name,
					"arguments": argsObj,
				},
			})
		}
		return
	}

	// Handle tool results (tool role)
	if msg.Role == schema.Tool {
		logx.Infof("Publishing tool result event: id=%s", msg.ToolCallID)

		// Send in format expected by frontend
		cb.bus.PublishOutbound(bus.OutboundMessage{
			Type:       protocol.AgentEventToolResult,
			SessionKey: cb.sessionKey,
			RunID:      cb.runID,
			Channel:    cb.channel,
			ChatID:     cb.chatID,
			Payload: map[string]interface{}{
				"id":       msg.ToolCallID,
				"result":   msg.Content,
				"content":  msg.Content, // Fallback for frontend
				"is_error": false,       // TODO: detect actual errors
			},
		})
		return
	}

	// Handle regular message chunks (assistant role)
	if msg.Content != "" {
		logx.Infof("Publishing chunk event: %d bytes", len(msg.Content))
		cb.bus.PublishOutbound(bus.OutboundMessage{
			Type:       protocol.ChatEventChunk,
			SessionKey: cb.sessionKey,
			RunID:      cb.runID,
			Channel:    cb.channel,
			ChatID:     cb.chatID,
			Content:    msg.Content,
			Payload: map[string]interface{}{
				"content": msg.Content,
			},
		})
	}
}

func (cb *StreamCallback) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	logx.Infof("StreamCallback.OnStart called: name=%s, component=%s, type=%s",
		info.Name, info.Component, info.Type)
	return ctx
}

func (cb *StreamCallback) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	logx.Infof("StreamCallback.OnEnd called: name=%s, component=%s, type=%s",
		info.Name, info.Component, info.Type)
	return ctx
}

func (cb *StreamCallback) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	logx.Errorf("StreamCallback.OnError: %v", err)
	return ctx
}
