/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	"io"
)

type LoggerCallback struct {
}

func NewLoggerCallback() callbacks.Handler {

	c := LoggerCallback{}

	handler := callbacks.NewHandlerBuilder().
		OnEndWithStreamOutputFn(c.OnEndWithStreamOutput).
		Build()
	return handler
}

func (cb *LoggerCallback) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo,
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

			case *ecmodel.CallbackOutput:
				fmt.Print(v.Message.Content)

				if len(v.Message.ToolCalls) > 0 {

					for i := range v.Message.ToolCalls {
						fmt.Printf("\n=%d===%s", i, v.Message.ToolCalls[i].Function.Name)
						fmt.Printf("\n=%d===%s", i, v.Message.ToolCalls[i].Function.Arguments)
					}
				}

			case []*schema.Message:
				for _, m := range v {
					fmt.Print(m.Content)
				}

			default:
			}
		}

		fmt.Println("\n=========[OnEndWithStreamOutput_Over]=========", msgID, info.Name, "|", info.Component, "|", info.Type)
	}()
	return ctx
}
