package agent

import (
	"context"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/pomclaw/pomclaw/pkg/callback"
	"github.com/pomclaw/pomclaw/pkg/tools"
	"testing"
	"time"
)

func TestEino_run(t *testing.T) {

	const restrict = false

	toolsNodeConfig := compose.ToolsNodeConfig{}
	toolsNodeConfig.Tools = append(toolsNodeConfig.Tools, []tool.BaseTool{

		tools.NewReadFileTool(restrict),
		tools.NewWriteFileTool(restrict),
		tools.NewListDirTool(restrict),
		tools.NewEditFileTool(restrict),
		tools.NewAppendFileTool(restrict),
		tools.NewExecTool(restrict),
	}...)

	llm, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		Model: "deepseek-v3.2",
	})
	if err != nil {
		t.Fatal(err)
	}

	agent, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:          "pomclaw",
		Description:   "A chat pomclaw agent",
		MaxIterations: 50,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: toolsNodeConfig,
		},
		Model: llm,
	})

	runner := adk.NewRunner(context.Background(), adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: true,
	})

	// Run with messages and callback
	// Callback methods (OnStart, OnEnd, OnError, OnEndWithStreamOutput) are called automatically by Eino
	iter := runner.Run(context.Background(), []*schema.Message{
		schema.UserMessage("使用 list_dir 工具看看我 /Users/zhengjm/product/pomclaw/pomclaw/docs/sql  目录下有哪些文件？？"),
	}, adk.WithCallbacks(callback.NewLoggerCallback()))

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			t.Fatal(err)
		}

		// Collect final content for session storage
		if event.Output != nil && event.Output.MessageOutput != nil {
			msg, err := event.Output.MessageOutput.GetMessage()
			if err != nil {
				t.Fatal(err)
			}
			_ = msg
			//t.Log(msg.Content)
		}
	}

	time.Sleep(time.Second * 10)
}
