// Pomclaw - Ultra-lightweight personal AI agent
// Powered by Eino Framework
// License: MIT
//
// Copyright (c) 2026 Pomclaw contributors

package agent

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/callbacks/apmplus"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	bus2 "github.com/pomclaw/pomclaw/internal/bus"
	callback2 "github.com/pomclaw/pomclaw/internal/callback"
	"github.com/pomclaw/pomclaw/internal/contracts"
	"github.com/pomclaw/pomclaw/internal/tools"
	"github.com/pomclaw/pomclaw/pkg/utils"
	"github.com/zeromicro/go-zero/core/logx"
	"strconv"
	"strings"
)

// AgentLoop 使用 Eino 框架完全重写。
// 核心逻辑由 Eino 的 ChatModelAgent 驱动，处理 LLM 调用和工具执行。
type AgentLoop struct {
	agent          *adk.ChatModelAgent
	sessions       contracts.SessionManagerInterface
	contextBuilder contracts.ContextBuilderInterface
}

type processOptions struct {
	AgentID         string
	UserID          string
	Workspace       string
	SessionId       int64
	Channel         string
	ChatID          string
	UserMessage     string
	DefaultResponse string
	EnableSummary   bool
	SendResponse    bool
	NoHistory       bool
}

// NewAgentLoop 创建使用 Eino 框架的 agent 循环。
// svcCtx 提供配置、存储、工具管理器等共享依赖；provider/modelName/userID/agentID
// 为本次请求相关参数。
func NewAgentLoop(adkAgent *adk.ChatModelAgent, contextBuilder contracts.ContextBuilderInterface, sessionManager contracts.SessionManagerInterface) (*AgentLoop, error) {
	return &AgentLoop{
		agent:          adkAgent,
		sessions:       sessionManager,
		contextBuilder: contextBuilder,
	}, nil
}

func (al *AgentLoop) ProcessMessage(ctx context.Context, client bus2.Streamer, msg bus2.InboundMessage) (string, error) {
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
		SessionId:       msg.SessionId,
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
func (al *AgentLoop) runEinoLoop(ctx context.Context, client bus2.Streamer, opts processOptions) (string, error) {
	ctx = tools.WithAgentID(ctx, opts.AgentID)
	ctx = tools.WithUserID(ctx, opts.UserID)
	ctx = tools.WithWorkspace(ctx, opts.Workspace)

	// 构建消息
	var history []schema.Message
	var summary string
	if !opts.NoHistory {
		history = al.sessions.GetHistory(opts.AgentID, opts.SessionId)
		summary = al.sessions.GetSummary(opts.AgentID, opts.SessionId)
	}

	msgValues := al.contextBuilder.BuildMessages(opts.AgentID, opts.Workspace,
		history, summary, opts.UserMessage, nil, opts.Channel, opts.ChatID)

	al.sessions.AddMessage(opts.AgentID, opts.SessionId, schema.User, opts.UserMessage)

	messages := make([]*schema.Message, len(msgValues))
	for i := range msgValues {
		messages[i] = &msgValues[i]
	}

	// Register StreamCallback to handle real-time streaming output
	streamCallback := callback2.NewStreamCallback(client, al.sessions, opts.SessionId, opts.AgentID)

	ctx = apmplus.SetSession(ctx, apmplus.WithSessionID(strconv.Itoa(int(opts.SessionId))), apmplus.WithUserID(opts.UserID))

	// Create Runner with streaming enabled (correct ADK pattern)
	// All ADK examples use Runner instead of calling agent.Run() directly
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           al.agent,
		EnableStreaming: true,
	})

	logx.Info("Runner created with streaming enabled")

	// Stream ended normally - send run.completed event
	_ = client.PublishRunStarted(ctx, &bus2.RunStartedPayload{})

	// Run with messages and callbacks
	// Callback methods (OnStart, OnEnd, OnError, OnEndWithStreamOutput) are called automatically by Eino
	iter := runner.Run(ctx, messages, adk.WithCallbacks(streamCallback, callback2.NewLoggerCallback()))

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
	// OnError callback already sent run.failed event
	if runErr != nil {
		return "", runErr
	}

	// 处理空响应
	if finalContent == "" {
		finalContent = opts.DefaultResponse
	}

	// 保存最终消息
	err := al.sessions.Save(opts.AgentID, opts.SessionId, opts.UserID)
	if err != nil {
		logx.Errorf("sessions.Save failed err: %v", err)
	}

	// Stream ended normally - send run.completed event
	_ = client.PublishRunCompleted(ctx, &bus2.RunCompletedPayload{Content: finalContent})
	return finalContent, nil
}
