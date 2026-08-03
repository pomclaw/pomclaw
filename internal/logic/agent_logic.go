package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/pomclaw/pomclaw/internal/agent"
	"github.com/pomclaw/pomclaw/internal/bus"
	"github.com/pomclaw/pomclaw/internal/mcp"
	"github.com/pomclaw/pomclaw/internal/skills"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/tools"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type AgentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AgentLogic {
	return &AgentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AgentLogic) Chat(streamer bus.Streamer, userID string, chatID string, req types.ChatSendParams) (types.ChatPayload, error) {
	// Publish to message bus (inbound channel)
	// The agent loop will pick this up and process it asynchronously
	inboundMsg := bus.InboundMessage{
		SessionId: req.SessionId,
		AgentID:   req.AgentID,
		UserID:    userID,
		Content:   req.Message,
		Channel:   "ws",
		ChatID:    chatID,
		Metadata:  map[string]string{},
	}

	// Fetch agent details from database
	agentRecord, err := l.svcCtx.AgentsModel.FindByAgentID(l.ctx, req.AgentID)
	if err != nil {
		return types.ChatPayload{}, errors.New("agent not found")
	}

	// Fetch provider by ID
	provider, err := l.svcCtx.ProvidersModel.FindOne(l.ctx, agentRecord.ProviderId)
	if err != nil {
		return types.ChatPayload{}, errors.New(fmt.Sprintf("Provider not found or disabled"))
	}

	// Initialize LLM with provider credentials and agent's model
	llm, err := tools.ResolveLLM(l.ctx, provider, agentRecord.Model)
	if err != nil {
		return types.ChatPayload{}, fmt.Errorf("failed to initialize LLM: %w", err)
	}

	toolsNodeConfig := l.svcCtx.ToolsManager.GetTools(l.ctx, userID, agentRecord.AgentId)

	// Discover MCP tools for this agent and append to tool config
	mcpTools, mcpClosers := l.discoverMCPTools(l.ctx, agentRecord.AgentId)
	toolsNodeConfig.Tools = append(toolsNodeConfig.Tools, mcpTools...)
	defer func() {
		for _, c := range mcpClosers {
			_ = c.Close()
		}
	}()

	skillsLoader := skills.NewDBSkillsLoader(l.svcCtx.SkillsModel, l.svcCtx.SkillGrantsModel, l.svcCtx.OssClient, l.svcCtx.Redis, agentRecord.AgentId, userID)
	contextBuilder := agent.NewContextBuilder(l.svcCtx.AgentContextFilesModel, l.svcCtx.MemoryStore, toolsNodeConfig, skillsLoader)

	adkAgent, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:          "pomclaw",
		MaxIterations: l.svcCtx.Config.Agents.Defaults.MaxToolIterations,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: toolsNodeConfig,
		},
		Model: llm,
	})
	if err != nil {
		return types.ChatPayload{}, err
	}

	a, err := agent.NewAgentLoop(adkAgent, contextBuilder, l.svcCtx.SessionManager)
	if err != nil {
		return types.ChatPayload{}, errors.New(fmt.Sprintf("Unable to initialize agent: %v", err))
	}

	finalContent, err := a.ProcessMessage(l.ctx, streamer, inboundMsg)
	if err != nil {
		return types.ChatPayload{}, err
	}

	logx.Infof("chat.send published to bus: chatId=%s, sessionKey=%s, user=%s", chatID, req.SessionId, userID)

	return types.ChatPayload{
		Content: finalContent,
		RunId:   chatID,
		Usage:   types.Usage{},
	}, nil
}

// discoverMCPTools queries MCP agent grants and connects to each granted server
// to discover tools. Returns the Eino tool.BaseTool list and closers for cleanup.
func (l *AgentLogic) discoverMCPTools(ctx context.Context, agentID string) ([]tool.BaseTool, []io.Closer) {
	grants, err := l.svcCtx.McpAgentGrantsModel.FindByAgentID(ctx, agentID)
	if err != nil {
		logx.Errorf("discoverMCPTools: failed to query grants: %v", err)
		return nil, nil
	}

	var mcpTools []tool.BaseTool
	var closers []io.Closer

	for _, grant := range grants {
		if !grant.Enabled {
			continue
		}

		server, err := l.svcCtx.McpServersModel.FindOne(ctx, grant.ServerId)
		if err != nil {
			logx.Errorf("discoverMCPTools: failed to find server %d: %v", grant.ServerId, err)
			continue
		}

		// Parse connection parameters
		var args []string
		if server.Args != "" {
			_ = json.Unmarshal([]byte(server.Args), &args)
		}
		var env map[string]string
		if server.Env != "" {
			_ = json.Unmarshal([]byte(server.Env), &env)
		}
		var headers map[string]string
		if server.Headers != "" {
			_ = json.Unmarshal([]byte(server.Headers), &headers)
		}

		command := ""
		if server.Command.Valid {
			command = server.Command.String
		}
		url := ""
		if server.Url.Valid {
			url = server.Url.String
		}
		toolPrefix := ""
		if server.ToolPrefix.Valid {
			toolPrefix = server.ToolPrefix.String
		}

		cc, err := mcp.ConnectAndListTools(ctx, server.Transport, command, args, env, url, headers)
		if err != nil {
			logx.Errorf("discoverMCPTools: failed to connect to server %q: %v", server.Name, err)
			continue
		}
		closers = append(closers, cc.Closer)

		// Parse allow/deny filters from grant
		allowSet := parseStringJSONSet(grant.ToolAllow.String)
		denySet := parseStringJSONSet(grant.ToolDeny.String)

		for _, toolDef := range cc.ToolDefs {
			// Apply allow/deny filtering
			if len(denySet) > 0 {
				if _, denied := denySet[toolDef.Name]; denied {
					continue
				}
			}
			if len(allowSet) > 0 {
				if _, allowed := allowSet[toolDef.Name]; !allowed {
					continue
				}
			}

			bt, err := mcp.NewBridgeTool(cc.Client, toolDef, toolPrefix)
			if err != nil {
				logx.Errorf("discoverMCPTools: failed to create bridge tool for %q: %v", toolDef.Name, err)
				continue
			}
			mcpTools = append(mcpTools, bt)
		}
	}

	return mcpTools, closers
}

// parseStringJSONSet parses a JSON array string into a set for lookup.
func parseStringJSONSet(s string) map[string]struct{} {
	if s == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(s), &items); err != nil {
		return nil
	}
	if len(items) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(items))
	for _, item := range items {
		set[item] = struct{}{}
	}
	return set
}
