// Pomclaw - Ultra-lightweight personal AI agent
// Powered by Eino Framework
// License: MIT
//
// Copyright (c) 2026 Pomclaw contributors

package callback

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	ecmodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/pomclaw/pomclaw/pkg/bus"
	"github.com/pomclaw/pomclaw/pkg/contracts"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"time"
)

// StreamCallback implements Eino's callback interface for streaming output
type StreamCallback struct {
	callbacks.HandlerBuilder

	client     bus.Streamer
	sessions   contracts.SessionManagerInterface
	runID      string
	sessionKey string
	channel    string
	chatID     string
}

func NewStreamCallback(client bus.Streamer, sessions contracts.SessionManagerInterface, runID, sessionKey, channel, chatID string) callbacks.Handler {
	cb := &StreamCallback{
		sessions:   sessions,
		client:     client,
		runID:      runID,
		sessionKey: sessionKey,
		channel:    channel,
		chatID:     chatID,
	}

	handler := callbacks.NewHandlerBuilder().
		OnEndWithStreamOutputFn(cb.OnEndWithStreamOutput).
		Build()
	return handler
}

func (cb *StreamCallback) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo,
	output *schema.StreamReader[callbacks.CallbackOutput],
) context.Context {
	msgID := uuid.New().String()

	fmt.Println("\n=========[OnEndWithStreamOutput]=========", msgID, info.Name, "|", info.Component, "|", info.Type)

	go func() {
		defer output.Close() // remember to close the stream in defer
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("[OnEndStream]panic_recover", "msgID", msgID, "err", err)
			}
		}()

		if info.Component != components.ComponentOfChatModel && info.Component != compose.ComponentOfToolsNode {
			return
		}

		var msgs []*schema.Message

		for {
			frame, err := output.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				fmt.Println(err, "[OnEndStream] recv_error")
				return
			}

			switch v := frame.(type) {

			// ai output
			case *ecmodel.CallbackOutput:
				msgs = append(msgs, v.Message)

				fmt.Print(v.Message.Content)

				// Send tool result event
				_ = cb.client.PublishChunk(ctx, &bus.ChunkPayload{
					Content: v.Message.Content,
				})

				if len(v.Message.ToolCalls) > 0 {
					for i := range v.Message.ToolCalls {
						fmt.Printf("\n=%d===%s", i, v.Message.ToolCalls[i].Function.Name)
						fmt.Printf("\n=%d===%s", i, v.Message.ToolCalls[i].Function.Arguments)
					}
				}

				// tools result
			case []*schema.Message:

				for _, m := range v {
					fmt.Print(m.Content)

					// Send tool result event
					_ = cb.client.PublishToolResult(ctx, &bus.ToolResultPayload{
						Id:      m.ToolCallID,
						Result:  m.Content,
						IsError: false, // TODO: detect actual errors
					})

					cb.sessions.AddFullMessage(cb.chatID, cb.sessionKey, bus.Message{
						Role:       m.Role,
						Content:    m.Content,
						ToolCallId: m.ToolCallID,
						CreatedAt:  time.Now(),
					})
				}

			default:
			}
		}

		if len(msgs) > 0 {
			msg, err := schema.ConcatMessages(msgs)
			if err != nil {
				logx.Errorf("schema.ConcatMessages failed,err:", err)
			} else {

				if msg.Content != "" || len(msg.ToolCalls) > 0 {
					cb.sessions.AddFullMessage(cb.chatID, cb.sessionKey, bus.Message{
						Role:       msg.Role,
						Content:    msg.Content,
						ToolCalls:  convertToolCalls(msg.ToolCalls),
						ToolCallId: msg.ToolCallID,
						CreatedAt:  time.Now(),
					})

					if len(msg.ToolCalls) > 0 {

						for i := range msg.ToolCalls {
							_ = cb.client.PublishToolCall(ctx, &bus.ToolCallPayload{
								Id:        msg.ToolCalls[i].ID,
								Arguments: msg.ToolCalls[i].Function.Arguments,
								Name:      msg.ToolCalls[i].Function.Name,
							})
						}
					}
				}
			}
		}

		fmt.Println("\n=========[OnEndWithStreamOutput_Over]=========", msgID, info.Name, "|", info.Component, "|", info.Type)
	}()
	return ctx
}

func convertToolCalls(calls []schema.ToolCall) []bus.ToolCallPayload {
	payloads := make([]bus.ToolCallPayload, 0, len(calls))
	for _, call := range calls {
		payloads = append(payloads, bus.ToolCallPayload{
			Id:        call.ID,
			Name:      call.Function.Name,
			Arguments: call.Function.Arguments,
		})
	}
	return payloads
}

func ConvertHistory(history []bus.Message) []schema.Message {
	if len(history) == 0 {
		return []schema.Message{}
	}

	messages := make([]schema.Message, 0, len(history))

	for _, msg := range history {
		schemaMsg := schema.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}

		// Convert tool calls from bus format to schema format
		if len(msg.ToolCalls) > 0 {
			schemaMsg.ToolCalls = make([]schema.ToolCall, 0, len(msg.ToolCalls))
			for _, call := range msg.ToolCalls {
				// Skip malformed tool calls (ID or Name missing)
				if call.Id == "" || call.Name == "" {
					logx.Infof("Skipping malformed tool call: id=%s, name=%s", call.Id, call.Name)
					continue
				}

				// Convert Arguments to JSON string if needed
				argsStr := ""
				if call.Arguments != nil {
					if argBytes, err := json.Marshal(call.Arguments); err == nil {
						argsStr = string(argBytes)
					}
				}

				toolCall := schema.ToolCall{
					ID:   call.Id,
					Type: "function",
					Function: schema.FunctionCall{
						Name:      call.Name,
						Arguments: argsStr,
					},
				}
				schemaMsg.ToolCalls = append(schemaMsg.ToolCalls, toolCall)
			}
		}

		// Handle tool result messages
		if msg.ToolCallId != "" {
			schemaMsg.ToolCallID = msg.ToolCallId
		}

		messages = append(messages, schemaMsg)
	}

	return messages
}
