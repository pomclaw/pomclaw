package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// Rememberer is the interface the remember tool needs to store memories.
type Rememberer interface {
	Remember(agentID string, text string, importance float64, category string) (string, error)
}

type RememberInput struct {
	Text       string  `json:"text"`
	Importance float64 `json:"importance,omitempty"`
	Category   string  `json:"category,omitempty"`
}

// RememberTool stores information in long-term memory.
type RememberTool struct {
	store Rememberer
}

// NewRememberTool creates a remember tool.
func NewRememberTool(store Rememberer) tool.InvokableTool {
	return &RememberTool{store: store}
}

func (t *RememberTool) Name() string { return "remember" }

func (t *RememberTool) Description() string {
	return "Store a piece of information in long-term memory with vector embedding for later semantic recall. Use this to remember facts, preferences, or important context."
}

// Info implements eino's BaseTool interface
func (t *RememberTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.Name(),
		Desc: t.Description(),
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"text": {
				Type:     schema.String,
				Desc:     "The text content to remember",
				Required: true,
			},
			"importance": {
				Type:     schema.Number,
				Desc:     "Importance score from 0.0 to 1.0 (default: 0.7)",
				Required: false,
			},
			"category": {
				Type:     schema.String,
				Desc:     "Optional category for organizing memories (e.g., 'preference', 'fact', 'context')",
				Required: false,
			},
		}),
	}, nil
}

func (t *RememberTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input RememberInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if input.Text == "" {
		return "", fmt.Errorf("text parameter is required")
	}

	importance := 0.7
	if input.Importance >= 0 && input.Importance <= 1 && input.Importance != 0 {
		importance = input.Importance
	}

	memoryID, err := t.store.Remember(AgentIDFromContext(ctx), input.Text, importance, input.Category)
	if err != nil {
		return "", fmt.Errorf("failed to remember: %w", err)
	}

	return fmt.Sprintf("Remembered (ID: %s, importance: %.1f, category: %s): %s",
		memoryID, importance, input.Category, truncate(input.Text, 100)), nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
