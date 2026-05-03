// Pomclaw - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 Pomclaw contributors

package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// EinoToolAdapter wraps a pomclaw Tool to be compatible with eino's tool interface.
// This adapter serves as a bridge between pomclaw's Tool interface and eino's
// tool system, allowing gradual migration to eino components.
type EinoToolAdapter struct {
	tool Tool
}

// NewEinoToolAdapter creates an adapter for a pomclaw Tool.
func NewEinoToolAdapter(tool Tool) *EinoToolAdapter {
	return &EinoToolAdapter{
		tool: tool,
	}
}

// Name returns the tool's name.
func (a *EinoToolAdapter) Name() string {
	return a.tool.Name()
}

// Description returns the tool's description.
func (a *EinoToolAdapter) Description() string {
	return a.tool.Description()
}

// GetInputSchema returns the tool's parameter schema.
// This is compatible with eino's schema expectations.
func (a *EinoToolAdapter) GetInputSchema() map[string]interface{} {
	return a.tool.Parameters()
}

// Execute calls the pomclaw tool and returns the result as a JSON string.
// eino expects tool results to be serializable strings.
func (a *EinoToolAdapter) Execute(ctx context.Context, input interface{}) (string, error) {
	// Convert input to map if necessary
	var args map[string]interface{}

	switch v := input.(type) {
	case map[string]interface{}:
		args = v
	case string:
		// Try to unmarshal JSON string
		if err := json.Unmarshal([]byte(v), &args); err != nil {
			return "", fmt.Errorf("failed to parse tool input as JSON: %w", err)
		}
	default:
		// Try to marshal to JSON and back
		data, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to marshal tool input: %w", err)
		}
		if err := json.Unmarshal(data, &args); err != nil {
			return "", fmt.Errorf("failed to unmarshal tool input: %w", err)
		}
	}

	// Execute the tool
	result := a.tool.Execute(ctx, args)
	if result == nil {
		return "", fmt.Errorf("tool returned nil result")
	}

	// Convert result to JSON for eino
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tool result: %w", err)
	}

	return string(resultJSON), nil
}

// ExecuteAsync is for tools that support async execution.
// Wraps the underlying AsyncExecutor if available.
func (a *EinoToolAdapter) ExecuteAsync(ctx context.Context, input interface{}, callback func(ctx context.Context, output string) error) error {
	// Check if the underlying tool supports async execution
	asyncTool, ok := a.tool.(AsyncExecutor)
	if !ok {
		// Fall back to synchronous execution
		result, err := a.Execute(ctx, input)
		if err != nil {
			return err
		}
		return callback(ctx, result)
	}

	// Convert input
	var args map[string]interface{}
	switch v := input.(type) {
	case map[string]interface{}:
		args = v
	case string:
		if err := json.Unmarshal([]byte(v), &args); err != nil {
			return fmt.Errorf("failed to parse async tool input: %w", err)
		}
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal async tool input: %w", err)
		}
		if err := json.Unmarshal(data, &args); err != nil {
			return fmt.Errorf("failed to unmarshal async tool input: %w", err)
		}
	}

	// Create async callback that wraps the provided callback
	asyncCallback := func(asyncCtx context.Context, toolResult *ToolResult) {
		resultJSON, err := json.Marshal(toolResult)
		if err != nil {
			// Log error but don't fail - async callback should be fire-and-forget
			fmt.Printf("Failed to marshal async result: %v\n", err)
			return
		}
		// Invoke the eino callback
		if err := callback(asyncCtx, string(resultJSON)); err != nil {
			// Log error but don't fail
			fmt.Printf("Failed to invoke async callback: %v\n", err)
		}
	}

	// Execute async
	_ = asyncTool.ExecuteAsync(ctx, args, asyncCallback)
	return nil
}

