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

	client     bus.Streamer
	runID      string
	sessionKey string
	channel    string
	chatID     string
}

func NewStreamCallback(client bus.Streamer, runID, sessionKey, channel, chatID string) callbacks.Handler {
	return &StreamCallback{
		client:     client,
		runID:      runID,
		sessionKey: sessionKey,
		channel:    channel,
		chatID:     chatID,
	}
}

func (cb *StreamCallback) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo, input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	logx.Infof("StreamCallback.OnStartWithStreamInput called: name=%s, component=%s, type=%s",
		info.Name, info.Component, info.Type)

	defer input.Close()
	return ctx
}

// OnEndWithStreamOutput handles streaming output from the agent
// This is the key method that receives incremental chunks in real-time
func (cb *StreamCallback) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo,
	output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {

	// Filter: only process top-level agent node to avoid duplicate outputs from nested graphs
	// In Eino's graph structure, nested nodes will trigger callbacks multiple times
	allowOutput := shouldOutputNode(info)
	if !allowOutput {
		// Still need to drain the stream to avoid blocking
		go func() {
			defer output.Close()
			for {
				_, err := output.Recv()
				if errors.Is(err, io.EOF) || err != nil {
					break
				}
			}
		}()
		return ctx
	}

	go func() {
		defer output.Close() // Always close the stream
		var finalContent string

		for {
			frame, err := output.Recv()
			if errors.Is(err, io.EOF) {
				// Stream ended normally - send run.completed event
				cb.client.PublishOutbound(bus.OutboundMessage{
					Type:       protocol.AgentEventRunCompleted,
					SessionKey: cb.sessionKey,
					RunID:      cb.runID,
					Channel:    cb.channel,
					ChatID:     cb.chatID,
					Content:    finalContent,
					Payload: map[string]interface{}{
						"content": finalContent,
						// TODO: Extract usage stats from Eino when available
					},
				})
				break
			}
			if err != nil {
				logx.Errorf("Stream recv error: %v", err)
				return
			}

			// Process different frame types
			switch v := frame.(type) {
			case *schema.Message:
				cb.handleMessage(ctx, v)
				// Accumulate content for final event
				if v.Role == schema.Assistant && v.Content != "" {
					finalContent += v.Content
				}
			case *ecmodel.CallbackOutput:
				if v.Message != nil {
					cb.handleMessage(ctx, v.Message)
					// Accumulate content for final event
					if v.Message.Role == schema.Assistant && v.Message.Content != "" {
						finalContent += v.Message.Content
					}
				}
			case []*schema.Message:
				for _, m := range v {
					cb.handleMessage(ctx, m)
					// Accumulate content for final event
					if m.Role == schema.Assistant && m.Content != "" {
						finalContent += m.Content
					}
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
			cb.client.PublishOutbound(bus.OutboundMessage{
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
		cb.client.PublishOutbound(bus.OutboundMessage{
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
		cb.client.PublishOutbound(bus.OutboundMessage{
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

// OnStart is called in non-streaming scenarios (not used in our streaming setup)
func (cb *StreamCallback) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	logx.Infof("StreamCallback.OnStart called: name=%s, component=%s, type=%s",
		info.Name, info.Component, info.Type)

	// Filter: only process nodes that should send output
	if info.Name == "pomclaw" && info.Component == "Agent" {
		// Send run.started event for non-streaming scenario
		cb.client.PublishOutbound(bus.OutboundMessage{
			Type:       protocol.AgentEventRunStarted,
			SessionKey: cb.sessionKey,
			RunID:      cb.runID,
			Channel:    cb.channel,
			ChatID:     cb.chatID,
			Payload: map[string]interface{}{
				"runId":      cb.runID,
				"sessionKey": cb.sessionKey,
			},
		})

		return ctx
	}

	return ctx
}

// OnEnd is called in non-streaming scenarios (not used in our streaming setup)
func (cb *StreamCallback) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	logx.Infof("StreamCallback.OnEnd called: name=%s, component=%s, type=%s",
		info.Name, info.Component, info.Type)

	// Filter: only process nodes that should send output
	if !shouldOutputNode(info) {
		return ctx
	}

	// Extract and handle messages from output
	var finalContent string

	switch v := output.(type) {
	case *schema.Message:
		cb.handleMessage(ctx, v)
		if v.Role == schema.Assistant && v.Content != "" {
			finalContent = v.Content
		}
	case *ecmodel.CallbackOutput:
		if v.Message != nil {
			cb.handleMessage(ctx, v.Message)
			if v.Message.Role == schema.Assistant && v.Message.Content != "" {
				finalContent = v.Message.Content
			}
		}
	case []*schema.Message:
		for _, m := range v {
			cb.handleMessage(ctx, m)
			if m.Role == schema.Assistant && m.Content != "" {
				finalContent += m.Content
			}
		}
	}

	return ctx
}

// OnError is automatically called by Eino when the agent run fails
func (cb *StreamCallback) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	logx.Errorf("StreamCallback.OnError: %v", err)

	// Send run.failed event
	cb.client.PublishOutbound(bus.OutboundMessage{
		Type:       protocol.AgentEventRunFailed,
		SessionKey: cb.sessionKey,
		RunID:      cb.runID,
		Channel:    cb.channel,
		ChatID:     cb.chatID,
		Payload: map[string]interface{}{
			"error": err.Error(),
		},
	})

	return ctx
}

// shouldOutputNode determines if a node's output should be sent to frontend
// Returns true only for top-level agent nodes to avoid duplicate outputs from nested graphs
func shouldOutputNode(info *callbacks.RunInfo) bool {
	// Only output from top-level agent named "pomclaw"
	// This filters out nested graph nodes which would cause duplicate outputs

	const (
		initNode_  = "Init"
		chatModel_ = "ChatModel"
		toolNode_  = "ToolNode"
		toolsNode_ = "ToolsNode"
	)

	// Accept ChatModel for streaming LLM output and tool calls/results
	// handleMessage will decide what events to send based on message content
	if info.Component == chatModel_ || info.Component == toolNode_ || info.Component == toolsNode_ {
		logx.Infof("shouldOutputNode true for node: name=%s, component=%s", info.Name, info.Component)
		return true
	}

	// Filter out other nodes: ToolsNode (container), Graph, Lambda, Tool
	logx.Infof("shouldOutputNode false for node: name=%s, component=%s (filtered)", info.Name, info.Component)
	return false
}
