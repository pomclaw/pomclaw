package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pomclaw/pomclaw/pkg/tools"
	tools2 "github.com/tmc/langchaingo/tools"
)

// toolAdapter wraps tools.Tool to implement tools2.Tool interface for langchaingo
type toolAdapter struct {
	tool tools.Tool
}

// Name returns the tool name
func (ta *toolAdapter) Name() string {
	return ta.tool.Name()
}

// Description returns the tool description
func (ta *toolAdapter) Description() string {
	return ta.tool.Description()
}

// Call implements the langchaingo Tool interface
// It converts the string input to a parameters map and calls Execute
func (ta *toolAdapter) Call(ctx context.Context, input string) (string, error) {
	// Parse the input string as JSON into a parameters map
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		// If JSON parsing fails, treat the input as a single parameter
		args = map[string]interface{}{"input": input}
	}

	// Execute the tool
	result := ta.tool.Execute(ctx, args)
	if result == nil {
		return "", fmt.Errorf("tool execution returned nil result")
	}

	// Return the ForLLM content as a string
	// If there was an error, include it in the response
	if result.IsError {
		if result.Err != nil {
			return fmt.Sprintf("Error: %s (%v)", result.ForLLM, result.Err), nil
		}
		return fmt.Sprintf("Error: %s", result.ForLLM), nil
	}

	return result.ForLLM, nil
}

// convertTools converts a slice of tools.Tool to tools2.Tool for langchaingo agents
func convertTools(i []tools.Tool) []tools2.Tool {
	result := make([]tools2.Tool, len(i))
	for idx, t := range i {
		result[idx] = &toolAdapter{tool: t}
	}
	return result
}
