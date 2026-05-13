// Pomclaw - Ultra-lightweight personal AI agent
// Powered by Eino Framework
// License: MIT
//
// Copyright (c) 2026 Pomclaw contributors

package agent

import (
	"context"
	"fmt"
	"github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/pomclaw/pomclaw/internal/config"
	"github.com/pomclaw/pomclaw/pkg/bus"
	"github.com/pomclaw/pomclaw/pkg/contracts"
	"github.com/pomclaw/pomclaw/pkg/tools"
	"github.com/pomclaw/pomclaw/pkg/utils"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/zeromicro/go-zero/core/logx"
	"strings"
)

// AgentLoop 使用 Eino 框架完全重写。
// 核心逻辑由 Eino 的 ChatModelAgent 驱动，处理 LLM 调用和工具执行。
type AgentLoop struct {
	model                     string
	contextWindow             int
	maxIterations             int
	summarizeMessageThreshold int
	summarizeTokenPercent     int
	agent                     *agents.OneShotZeroAgent
	sessions                  contracts.SessionManagerInterface
	state                     contracts.StateManagerInterface
	contextBuilder            contracts.ContextBuilderInterface
}

type processOptions struct {
	AgentID         string
	Workspace       string
	SessionKey      string
	Channel         string
	ChatID          string
	RunID           string
	UserMessage     string
	DefaultResponse string
	EnableSummary   bool
	SendResponse    bool
	NoHistory       bool
}

// NewAgentLoop 创建使用 Eino 框架的 agent 循环。
func NewAgentLoop(cfg *config.Config, stateStore contracts.StateManagerInterface, memoryStore contracts.SqlMemoryStore, promptStoreRaw contracts.PromptStoreInterface, sessionManager contracts.SessionManagerInterface) (*AgentLoop, error) {

	// Build tool definitions
	restrict := cfg.Agents.Defaults.RestrictToWorkspace
	toolsNodeConfig := compose.ToolsNodeConfig{}
	toolsNodeConfig.Tools = append(toolsNodeConfig.Tools, []tool.BaseTool{}...)

	llm, err := openai.New(
		openai.WithBaseURL(cfg.Providers.OpenAI.APIBase),
		openai.WithToken(cfg.Providers.OpenAI.APIKey),
		openai.WithModel(cfg.Agents.Defaults.Model),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	agent := agents.NewOneShotAgent(llm,
		convertTools([]tools.Tool{

			tools.NewReadFileTool(restrict),
			tools.NewWriteFileTool(restrict),
			tools.NewListDirTool(restrict),
			tools.NewEditFileTool(restrict),
			tools.NewAppendFileTool(restrict),
			tools.NewExecTool("", restrict),
			tools.NewRememberTool(&rememberAdapter{store: memoryStore}),
			tools.NewWriteDailyNoteTool(memoryStore),
			tools.NewRecallTool(&recallAdapter{store: memoryStore}),
		}),
		agents.WithMaxIterations(50))

	var skillsLoader contracts.SkillsLoaderInterface
	contextBuilder := NewContextBuilder(promptStoreRaw, memoryStore, skillsLoader)

	logx.Info("agent", "Agent loop initialized with Eino framework", nil)

	summarizeMessageThreshold := cfg.Agents.Defaults.SummarizeMessageThreshold
	if summarizeMessageThreshold == 0 {
		summarizeMessageThreshold = 20
	}
	summarizeTokenPercent := cfg.Agents.Defaults.SummarizeTokenPercent
	if summarizeTokenPercent == 0 {
		summarizeTokenPercent = 75
	}

	return &AgentLoop{
		agent:                     agent,
		model:                     cfg.Agents.Defaults.Model,
		contextWindow:             cfg.Agents.Defaults.MaxTokens,
		maxIterations:             cfg.Agents.Defaults.MaxToolIterations,
		summarizeMessageThreshold: summarizeMessageThreshold,
		summarizeTokenPercent:     summarizeTokenPercent,
		sessions:                  sessionManager,
		state:                     stateStore,
		contextBuilder:            contextBuilder,
	}, nil

}

func (al *AgentLoop) ProcessMessage(ctx context.Context, client bus.Streamer, msg bus.InboundMessage) (string, error) {
	var logContent string
	if strings.Contains(msg.Content, "Error:") || strings.Contains(msg.Content, "error") {
		logContent = msg.Content
	} else {
		logContent = utils.Truncate(msg.Content, 80)
	}

	logx.Info("agent", fmt.Sprintf("Processing: %s", logContent),
		map[string]interface{}{
			"channel": msg.Channel,
			"chat_id": msg.ChatID,
		})

	var agentID = msg.AgentID
	if agentID == "" {
		agentID = contracts.DefaultAgentID
	}
	var workspace = contracts.DefaultWorkspace
	if v, ok := msg.Metadata[contracts.MetadataKey_AgentId]; ok && v != "" {
		agentID = v
	}
	if v, ok := msg.Metadata[contracts.MetadataKey_Workspace]; ok && v != "" {
		workspace = v
	}

	// 使用 Eino 处理消息（传递 runID 和 sessionKey 用于事件发射）
	return al.runEinoLoop(ctx, client, processOptions{
		AgentID:         agentID,
		Workspace:       workspace,
		SessionKey:      msg.SessionKey,
		Channel:         msg.Channel,
		ChatID:          msg.ChatID,
		RunID:           msg.RunID,
		UserMessage:     msg.Content,
		DefaultResponse: "I've completed processing but have no response to give.",
		EnableSummary:   true,
		SendResponse:    false,
		NoHistory:       false,
	})
}

// runEinoLoop 是核心 Eino 驱动的循环 - 处理 LLM 调用、工具执行等。
// 现已支持 Protocol v3 事件发射，使用 Callback 实现真正的流式输出。
func (al *AgentLoop) runEinoLoop(ctx context.Context, client bus.Streamer, opts processOptions) (string, error) {
	ctx = tools.WithAgentID(ctx, opts.AgentID)
	ctx = tools.WithWorkspace(ctx, opts.Workspace)

	// 构建消息
	var history []schema.Message
	var summary string
	if !opts.NoHistory {
		history = al.sessions.GetHistory(opts.AgentID, opts.SessionKey)
		summary = al.sessions.GetSummary(opts.AgentID, opts.SessionKey)
	}

	msgValues := al.contextBuilder.BuildMessages(opts.AgentID, opts.Workspace,
		history, summary, opts.UserMessage, nil, opts.Channel, opts.ChatID)

	al.sessions.AddMessage(opts.AgentID, opts.SessionKey, schema.User, opts.UserMessage)

	messages := make([]*schema.Message, len(msgValues))
	for i := range msgValues {
		messages[i] = &msgValues[i]
	}

	// Register StreamCallback to handle real-time streaming output
	executor := agents.NewExecutor(al.agent, agents.WithCallbacksHandler(NewHandler(client)))

	logx.Info("Runner created with streaming enabled")

	inputMap := make(map[string]any)
	inputMap["input"] = opts.UserMessage
	result, err := chains.Call(ctx, executor, inputMap)
	if err != nil {
		return "", fmt.Errorf("run chain: %w", err)
	}
	finalContent := result["output"].(string)

	// Create a proper event for the final output
	messageID := events.GenerateMessageID()
	finalMessage := events.NewTextMessageContentEvent(messageID, finalContent)
	if jsonData, err := finalMessage.ToJSON(); err == nil {
		client.SendEvent(ctx, string(finalMessage.EventType), string(jsonData))
	}

	// 处理空响应
	if finalContent == "" {
		finalContent = opts.DefaultResponse
	}

	// 保存最终消息
	al.sessions.AddMessage(opts.AgentID, opts.SessionKey, schema.Assistant, finalContent)
	err = al.sessions.Save(opts.AgentID, opts.SessionKey)
	if err != nil {
		logx.Errorf("sessions.Save failed err: %v", err)
	}

	responsePreview := utils.Truncate(finalContent, 120)
	logx.Info("agent", fmt.Sprintf("Response: %s", responsePreview),
		map[string]interface{}{
			"length": len(finalContent),
		})

	return finalContent, nil
}
