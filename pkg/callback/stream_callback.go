// Pomclaw - Ultra-lightweight personal AI agent
// Powered by Eino Framework
// License: MIT
//
// Copyright (c) 2026 Pomclaw contributors

package callback

import (
	"context"
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
)

// StreamCallback implements Eino's callback interface for streaming output
type StreamCallback struct {
	callbacks.HandlerBuilder

	client     bus.Streamer
	sessions   contracts.SessionManagerInterface
	sessionKey string
	agentID    string
}

func NewStreamCallback(client bus.Streamer, sessions contracts.SessionManagerInterface, sessionKey, agentID string) callbacks.Handler {
	cb := &StreamCallback{
		sessions:   sessions,
		client:     client,
		sessionKey: sessionKey,
		agentID:    agentID,
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

				//fmt.Print(v.Message.Content)

				// Send tool result event
				_ = cb.client.PublishChunk(ctx, &bus.ChunkPayload{
					Content: v.Message.Content,
				})

				//if len(v.Message.ToolCalls) > 0 {
				//	for i := range v.Message.ToolCalls {
				//		fmt.Printf("\n=%d===%s", i, v.Message.ToolCalls[i].Function.Name)
				//		fmt.Printf("\n=%d===%s", i, v.Message.ToolCalls[i].Function.Arguments)
				//	}
				//}

				// tools result
			case []*schema.Message:

				for _, m := range v {
					//fmt.Print(m.Content)

					// Send tool result event
					_ = cb.client.PublishToolResult(ctx, &bus.ToolResultPayload{
						Id:      m.ToolCallID,
						Result:  m.Content,
						IsError: false, // TODO: detect actual errors
					})

					cb.sessions.AddFullMessage(cb.agentID, cb.sessionKey, *m)
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
					cb.sessions.AddFullMessage(cb.agentID, cb.sessionKey, *msg)

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
