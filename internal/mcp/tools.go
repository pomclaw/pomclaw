package mcp

import (
	"context"
	"fmt"
	"time"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// ToolInfo holds a tool's name and description for API responses.
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// DiscoverTools connects temporarily to an MCP server, lists its tools, and disconnects.
// Used for on-demand discovery (connection testing, tool listing).
func DiscoverTools(ctx context.Context, transportType, command string, args []string, env map[string]string, url string, headers map[string]string) ([]ToolInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	client, err := createClient(transportType, command, args, env, url, headers)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	if transportType != "stdio" {
		if err := client.Start(ctx); err != nil {
			return nil, fmt.Errorf("start transport: %w", err)
		}
	}

	initReq := mcpgo.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcpgo.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcpgo.Implementation{Name: "pomclaw-discovery", Version: "1.0.0"}
	if _, err := client.Initialize(ctx, initReq); err != nil {
		return nil, fmt.Errorf("initialize: %w", err)
	}

	toolsResult, err := client.ListTools(ctx, mcpgo.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("list tools: %w", err)
	}

	result := make([]ToolInfo, 0, len(toolsResult.Tools))
	for _, t := range toolsResult.Tools {
		result = append(result, ToolInfo{Name: t.Name, Description: t.Description})
	}
	return result, nil
}