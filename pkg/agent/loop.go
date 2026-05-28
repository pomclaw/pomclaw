// Pomclaw - Ultra-lightweight personal AI agent
// Powered by Eino Framework
// License: MIT
//
// Copyright (c) 2026 Pomclaw contributors

package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/callbacks/apmplus"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/pomclaw/pomclaw/internal/config"
	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/pkg/bus"
	"github.com/pomclaw/pomclaw/pkg/callback"
	"github.com/pomclaw/pomclaw/pkg/contracts"
	"github.com/pomclaw/pomclaw/pkg/tools"
	"github.com/pomclaw/pomclaw/pkg/utils"
	"github.com/zeromicro/go-zero/core/logx"
)

// AgentLoop 使用 Eino 框架完全重写。
// 核心逻辑由 Eino 的 ChatModelAgent 驱动，处理 LLM 调用和工具执行。
type AgentLoop struct {
	agent                     *adk.ChatModelAgent
	model                     string
	contextWindow             int
	maxIterations             int
	summarizeMessageThreshold int
	summarizeTokenPercent     int
	sessions                  contracts.SessionManagerInterface
	contextBuilder            contracts.ContextBuilderInterface
	tracesModel               model.TracesModel
	spansModel                model.SpansModel
}

type processOptions struct {
	AgentID         string
	UserID          string
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
func NewAgentLoop(cfg config.Config, memoryStore contracts.SqlMemoryStore, promptStoreRaw contracts.PromptStoreInterface, sessionManager contracts.SessionManagerInterface, toolsManager contracts.ToolsManagerInterface,
	tracesModel model.TracesModel, spansModel model.SpansModel, userID string, agentID string) (*AgentLoop, error) {

	llm, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		APIKey:  cfg.Providers.OpenAI.APIKey,
		BaseURL: cfg.Providers.OpenAI.APIBase,
		Model:   cfg.Agents.Defaults.Model,
	})
	if err != nil {
		return nil, err
	}

	toolsNodeConfig := toolsManager.GetTools(context.Background(), userID, agentID)

	agent, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:          "pomclaw",
		Description:   "A chat pomclaw agent",
		MaxIterations: cfg.Agents.Defaults.MaxToolIterations,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: toolsNodeConfig,
		},
		Model: llm,
	})
	if err != nil {
		return nil, err
	}

	var skillsLoader contracts.SkillsLoaderInterface
	contextBuilder := NewContextBuilder(promptStoreRaw, memoryStore, toolsNodeConfig, skillsLoader)

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
		contextBuilder:            contextBuilder,
		tracesModel:               tracesModel,
		spansModel:                spansModel,
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
		UserID:          msg.UserID,
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
	streamCallback := callback.NewStreamCallback(client, al.sessions, opts.SessionKey, opts.AgentID)
	logx.Infof("StreamCallback registered for runID: %s", opts.RunID)

	ctx = apmplus.SetSession(ctx, apmplus.WithSessionID(opts.SessionKey), apmplus.WithUserID(opts.UserID))
	logx.Infof("TraceCallback registered for runID: %s", opts.RunID)

	// Create Runner with streaming enabled (correct ADK pattern)
	// All ADK examples use Runner instead of calling agent.Run() directly
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           al.agent,
		EnableStreaming: true,
	})

	logx.Info("Runner created with streaming enabled")

	// Stream ended normally - send run.completed event
	client.PublishRunStarted(ctx, &bus.RunStartedPayload{
		Message: "",
	})

	// Run with messages and callbacks
	// Callback methods (OnStart, OnEnd, OnError, OnEndWithStreamOutput) are called automatically by Eino
	iter := runner.Run(ctx, messages, adk.WithCallbacks(streamCallback))

	var finalContent string
	var runErr error

	// Simply iterate to get the final result
	// Streaming chunks are already sent by the callback
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			runErr = event.Err
			logx.Errorf("agent run failed: %v", event.Err)
			break
		}

		// Collect final content for session storage
		if event.Output != nil && event.Output.MessageOutput != nil {
			msg, err := event.Output.MessageOutput.GetMessage()
			if err != nil {
				runErr = err
				logx.Errorf("failed to get message: %v", err)
				break
			}
			finalContent = msg.Content
		}
	}

	// Handle run failure
	if runErr != nil {
		// OnError callback already sent run.failed event
		return "", runErr
	}

	// 处理空响应
	if finalContent == "" {
		finalContent = opts.DefaultResponse
	}

	// 保存最终消息
	err := al.sessions.Save(opts.AgentID, opts.SessionKey)
	if err != nil {
		logx.Errorf("sessions.Save failed err: %v", err)
	}

	// Stream ended normally - send run.completed event
	client.PublishRunCompleted(ctx, &bus.RunCompletedPayload{
		Content: finalContent,
	})

	responsePreview := utils.Truncate(finalContent, 120)
	logx.Info("agent", fmt.Sprintf("Response: %s", responsePreview),
		map[string]interface{}{
			"length": len(finalContent),
		})

	return finalContent, nil
}
