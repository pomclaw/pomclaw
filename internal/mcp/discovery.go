package mcp

import (
	"context"
	"fmt"
	"io"

	mcpclient "github.com/mark3labs/mcp-go/client"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// ConnectedClient holds an initialized MCP client and its discovered tools.
// The caller is responsible for calling Close() on the Closer when done.
type ConnectedClient struct {
	Client    mcpclient.MCPClient
	Closer    io.Closer
	ToolDefs  []mcpgo.Tool
	Transport string
}

// ConnectAndListTools creates an MCP client, initializes the connection,
// and lists available tools. The caller is responsible for closing the client.
// This is used by the agent loop to discover MCP tools per session.
func ConnectAndListTools(ctx context.Context, transportType, command string, args []string, env map[string]string, url string, headers map[string]string) (*ConnectedClient, error) {
	rawClient, err := createClient(transportType, command, args, env, url, headers)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	if transportType != "stdio" {
		if err := rawClient.Start(ctx); err != nil {
			_ = rawClient.Close()
			return nil, fmt.Errorf("start transport: %w", err)
		}
	}

	initReq := mcpgo.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcpgo.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcpgo.Implementation{Name: "pomclaw-agent", Version: "1.0."}
	if _, err := rawClient.Initialize(ctx, initReq); err != nil {
		_ = rawClient.Close()
		return nil, fmt.Errorf("initialize: %w", err)
	}

	toolsResult, err := rawClient.ListTools(ctx, mcpgo.ListToolsRequest{})
	if err != nil {
		_ = rawClient.Close()
		return nil, fmt.Errorf("list tools: %w", err)
	}

	return &ConnectedClient{
		Client:    rawClient,
		Closer:    rawClient,
		ToolDefs:  toolsResult.Tools,
		Transport: transportType,
	}, nil
}