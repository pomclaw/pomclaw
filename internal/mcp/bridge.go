package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	einojsonschema "github.com/eino-contrib/jsonschema"
	mcpclient "github.com/mark3labs/mcp-go/client"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// BridgeTool wraps an MCP tool as an Eino tool.BaseTool.
// Each BridgeTool holds a reference to the MCP client and calls
// CallTool on it when the tool is invoked by the agent.
type BridgeTool struct {
	info     *schema.ToolInfo
	origName string // original tool name (without prefix), used for MCP protocol calls
	cli      mcpclient.MCPClient
}

// NewBridgeTool creates an Eino tool.BaseTool from an MCP tool definition.
// The toolPrefix is prepended to the tool name (set to "" if no prefix needed).
func NewBridgeTool(cli mcpclient.MCPClient, toolDef mcpgo.Tool, toolPrefix string) (tool.BaseTool, error) {
	name := toolDef.Name
	if toolPrefix != "" {
		name = toolPrefix + "_" + name
	}

	// Convert invopop/jsonschema.Schema to eino-contrib/jsonschema.Schema
	// Both are JSON-compatible, so marshal/unmarshal works.
	schemaJSON, err := json.Marshal(toolDef.InputSchema)
	if err != nil {
		return nil, fmt.Errorf("marshal tool schema: %w", err)
	}

	var inputSchema einojsonschema.Schema
	if err := json.Unmarshal(schemaJSON, &inputSchema); err != nil {
		return nil, fmt.Errorf("unmarshal tool schema: %w", err)
	}

	// eino-contrib/jsonschema uses json:"-" for the Type field,
	// so we need to extract it from the raw JSON and set it manually.
	var rawSchema map[string]interface{}
	if err := json.Unmarshal(schemaJSON, &rawSchema); err == nil {
		if t, ok := rawSchema["type"].(string); ok {
			inputSchema.Type = t
		}
	}

	return &BridgeTool{
		info: &schema.ToolInfo{
			Name:        name,
			Desc:        toolDef.Description,
			ParamsOneOf: schema.NewParamsOneOfByJSONSchema(&inputSchema),
		},
		origName: toolDef.Name,
		cli:      cli,
	}, nil
}

func (bt *BridgeTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return bt.info, nil
}

func (bt *BridgeTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	result, err := bt.cli.CallTool(ctx, mcpgo.CallToolRequest{
		Params: mcpgo.CallToolParams{
			Name:      bt.origName,
			Arguments: json.RawMessage(argumentsInJSON),
		},
	})
	if err != nil {
		return "", fmt.Errorf("mcp tool call failed: %w", err)
	}

	// Extract text content from the result
	var textParts []string
	for _, c := range result.Content {
		if tc, ok := c.(mcpgo.TextContent); ok {
			textParts = append(textParts, tc.Text)
		}
	}

	output := strings.Join(textParts, "\n")

	if result.IsError {
		return "", fmt.Errorf("mcp tool returned error: %s", output)
	}

	return output, nil
}
