package mcp

import (
	"fmt"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
)

// createClient creates the appropriate MCP client based on transport type.
func createClient(transportType, command string, args []string, env map[string]string, url string, headers map[string]string) (*mcpclient.Client, error) {
	switch transportType {
	case "stdio":
		envSlice := mapToEnvSlice(env)
		return mcpclient.NewStdioMCPClient(command, envSlice, args...)

	case "sse":
		var opts []transport.ClientOption
		if len(headers) > 0 {
			opts = append(opts, mcpclient.WithHeaders(headers))
		}
		return mcpclient.NewSSEMCPClient(url, opts...)

	case "streamable-http":
		var opts []transport.StreamableHTTPCOption
		if len(headers) > 0 {
			opts = append(opts, transport.WithHTTPHeaders(headers))
		}
		return mcpclient.NewStreamableHttpClient(url, opts...)

	default:
		return nil, fmt.Errorf("unsupported transport: %q", transportType)
	}
}