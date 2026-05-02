// Pomclaw - Ultra-lightweight personal AI agent
// Powered by Eino Framework
// License: MIT
//
// Copyright (c) 2026 Pomclaw contributors

package agent

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/pomclaw/pomclaw/internal/config"
	"github.com/pomclaw/pomclaw/pkg/bus"
	"github.com/pomclaw/pomclaw/pkg/constants"
	"github.com/pomclaw/pomclaw/pkg/contracts"
	"github.com/pomclaw/pomclaw/pkg/storage"
	"github.com/pomclaw/pomclaw/pkg/tools"
	"github.com/pomclaw/pomclaw/pkg/utils"
	"github.com/zeromicro/go-zero/core/logx"
)

// AgentLoop 使用 Eino 框架完全重写。
// 核心逻辑由 Eino 的 ChatModelAgent 驱动，处理 LLM 调用和工具执行。
type AgentLoop struct {
	agent                     *adk.ChatModelAgent
	bus                       *bus.MessageBus
	model                     string
	contextWindow             int
	maxIterations             int
	summarizeMessageThreshold int
	summarizeTokenPercent     int
	sessions                  contracts.SessionManagerInterface
	state                     contracts.StateManagerInterface
	contextBuilder            contracts.ContextBuilderInterface
	running                   atomic.Bool
	summarizing               sync.Map
	channelManager            channelManagerInterface
}

type channelManagerInterface interface {
	GetEnabledChannels() []string
	HasChannel(name string) bool
}

type processOptions struct {
	AgentID         string
	Workspace       string
	SessionKey      string
	Channel         string
	ChatID          string
	UserMessage     string
	DefaultResponse string
	EnableSummary   bool
	SendResponse    bool
	NoHistory       bool
}

// NewAgentLoop 创建使用 Eino 框架的 agent 循环。
func NewAgentLoop(cfg *config.Config, db *sql.DB, msgBus *bus.MessageBus) (*AgentLoop, error) {
	embSvc, err := storage.NewEmbeddingService(cfg, db)
	if err != nil {
		logx.Error("agent", "Failed to create embedding service", map[string]interface{}{"error": err.Error()})
		return nil, fmt.Errorf("failed to create embedding service: %w", err)
	}
	logx.Info("agent", "Using embedding service", map[string]interface{}{"type": cfg.StorageType})

	sessionStore := storage.NewSessionStore(cfg, db)
	stateStore := storage.NewStateStore(cfg, db)
	memoryStore := storage.NewMemoryStore(cfg, db, embSvc)
	promptStoreRaw := storage.NewPromptStore(cfg, db)

	// Build tool definitions
	restrict := cfg.Agents.Defaults.RestrictToWorkspace
	toolsNodeConfig := compose.ToolsNodeConfig{}
	toolsNodeConfig.Tools = append(toolsNodeConfig.Tools, []tool.BaseTool{

		tools.NewReadFileTool(restrict),
		tools.NewWriteFileTool(restrict),
		tools.NewListDirTool(restrict),
		tools.NewEditFileTool(restrict),
		tools.NewAppendFileTool(restrict),
		tools.NewExecTool(restrict),
		tools.NewRememberTool(&rememberAdapter{store: memoryStore}),
		tools.NewWriteDailyNoteTool(memoryStore),
		tools.NewRecallTool(&recallAdapter{store: memoryStore}),
	}...)

	llm, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		APIKey:  cfg.Providers.OpenAI.APIKey,
		BaseURL: cfg.Providers.OpenAI.APIBase,
		Model:   cfg.Agents.Defaults.Model,
	})
	if err != nil {
		return nil, err
	}

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
		bus:                       msgBus,
		model:                     cfg.Agents.Defaults.Model,
		contextWindow:             cfg.Agents.Defaults.MaxTokens,
		maxIterations:             cfg.Agents.Defaults.MaxToolIterations,
		summarizeMessageThreshold: summarizeMessageThreshold,
		summarizeTokenPercent:     summarizeTokenPercent,
		sessions:                  sessionStore,
		state:                     stateStore,
		contextBuilder:            contextBuilder,
		summarizing:               sync.Map{},
	}, nil

}

func (al *AgentLoop) SetChannelManager(cm channelManagerInterface) {
	al.channelManager = cm
}

// Start 主事件循环 - 使用 Eino 框架处理消息。
func (al *AgentLoop) Start() {
	al.running.Store(true)

	for al.running.Load() {
		msg, ok := al.bus.ConsumeInbound(context.Background())
		if !ok {
			continue
		}

		response, err := al.processMessage(context.Background(), msg)
		if err != nil {
			response = fmt.Sprintf("Error processing message: %v", err)
		}

		logx.Debug("agent", "Message processed", map[string]interface{}{
			"channel":      msg.Channel,
			"response_len": len(response),
			"has_error":    err != nil,
		})

		if response != "" {
			logx.Info("agent", "Response", map[string]interface{}{
				"channel":      msg.Channel,
				"response_len": len(response),
			})
		}
	}
}

func (al *AgentLoop) Stop() {
	al.running.Store(false)
}

func (al *AgentLoop) RecordLastChannel(agentID string, channel string) error {
	return al.state.SetLastChannel(agentID, channel)
}

func (al *AgentLoop) RecordLastChatID(agentID string, chatID string) error {
	return al.state.SetLastChatID(agentID, chatID)
}

func (al *AgentLoop) processMessage(ctx context.Context, msg bus.InboundMessage) (string, error) {
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

	var agentID = contracts.DefaultAgentID
	var workspace = contracts.DefaultWorkspace
	if v, ok := msg.Metadata[contracts.MetadataKey_AgentId]; ok && v != "" {
		agentID = v
	}
	if v, ok := msg.Metadata[contracts.MetadataKey_Workspace]; ok && v != "" {
		workspace = v
	}

	// 使用 Eino 处理消息
	return al.runEinoLoop(ctx, processOptions{
		AgentID:         agentID,
		Workspace:       workspace,
		SessionKey:      msg.SessionKey,
		Channel:         msg.Channel,
		ChatID:          msg.ChatID,
		UserMessage:     msg.Content,
		DefaultResponse: "I've completed processing but have no response to give.",
		EnableSummary:   true,
		SendResponse:    false,
		NoHistory:       false,
	})
}

// runEinoLoop 是核心 Eino 驱动的循环 - 处理 LLM 调用、工具执行等。
func (al *AgentLoop) runEinoLoop(ctx context.Context, opts processOptions) (string, error) {
	ctx = tools.WithAgentID(ctx, opts.AgentID)
	ctx = tools.WithWorkspace(ctx, opts.Workspace)

	// 记录最后频道
	if opts.Channel != "" && opts.ChatID != "" && !constants.IsInternalChannel(opts.Channel) {
		channelKey := fmt.Sprintf("%s:%s", opts.Channel, opts.ChatID)
		if err := al.RecordLastChannel(opts.AgentID, channelKey); err != nil {
			logx.Info("agent", "Failed to record channel", map[string]interface{}{"error": err.Error()})
		}
	}

	// 构建消息
	var history []schema.Message
	var summary string
	if !opts.NoHistory {
		history = al.sessions.GetHistory(opts.AgentID, opts.SessionKey)
		summary = al.sessions.GetSummary(opts.AgentID, opts.SessionKey)
	}

	msgValues := al.contextBuilder.BuildMessages(opts.AgentID, opts.Workspace,
		history, summary, opts.UserMessage, nil, opts.Channel, opts.ChatID)

	al.sessions.AddMessage(opts.AgentID, opts.SessionKey, "user", opts.UserMessage)

	messages := make([]*schema.Message, len(msgValues))
	for i := range msgValues {
		messages[i] = &msgValues[i]
	}

	// Get streamer for real-time updates
	streamer, hasStreamer := al.bus.GetStreamer(ctx, opts.Channel, opts.ChatID)

	iter := al.agent.Run(ctx, &adk.AgentInput{Messages: messages, EnableStreaming: true})

	var finalContent string
	for {
		i, ok := iter.Next()
		if !ok {
			break
		}
		if i.Err != nil {
			logx.Error("agent run failed. err:", i.Err)
			break
		}
		msg, err := i.Output.MessageOutput.GetMessage()
		if err != nil {
			break
		}

		// Stream each chunk to connected clients
		if hasStreamer && msg.Content != "" {
			if err := streamer.Update(ctx, msg.Content); err != nil {
				logx.Error("streamer update failed:", map[string]interface{}{"error": err.Error()})
			}
		}

		finalContent += msg.Content
	}

	// 处理空响应
	if finalContent == "" {
		finalContent = opts.DefaultResponse
	}

	// Finalize streaming
	if hasStreamer {
		if err := streamer.Finalize(ctx, finalContent); err != nil {
			logx.Error("streamer finalize failed:", map[string]interface{}{"error": err.Error()})
		}
	}

	// 保存最终消息
	al.sessions.AddMessage(opts.AgentID, opts.SessionKey, "assistant", finalContent)
	_ = al.sessions.Save(opts.AgentID, opts.SessionKey)

	// 可选的总结
	if opts.EnableSummary {
	}

	al.bus.PublishOutbound(bus.OutboundMessage{
		Channel: opts.Channel,
		ChatID:  opts.ChatID,
		Content: finalContent,
	})

	responsePreview := utils.Truncate(finalContent, 120)
	logx.Info("agent", fmt.Sprintf("Response: %s", responsePreview),
		map[string]interface{}{
			"length": len(finalContent),
		})

	return finalContent, nil
}
