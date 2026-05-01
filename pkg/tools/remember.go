package tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
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

type RememberOutput struct {
	Message string `json:"message"`
}

func NewRememberTool(store Rememberer) tool.InvokableTool {
	return utils.NewTool[RememberInput, RememberOutput](
		&schema.ToolInfo{
			Name: "remember",
			Desc: "Store a piece of information in long-term memory with vector embedding for later semantic recall. Use this to remember facts, preferences, or important context.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"text": {
					Type:     schema.String,
					Desc:     "The text content to remember",
					Required: true,
				},
				"importance": {
					Type: schema.Number,
					Desc: "Importance score from 0.0 to 1.0 (default: 0.7)",
				},
				"category": {
					Type: schema.String,
					Desc: "Optional category for organizing memories (e.g., 'preference', 'fact', 'context')",
				},
			}),
		},
		func(ctx context.Context, input RememberInput) (RememberOutput, error) {
			if input.Text == "" {
				return RememberOutput{}, fmt.Errorf("text parameter is required")
			}

			importance := 0.7
			if input.Importance >= 0 && input.Importance <= 1 && input.Importance != 0 {
				importance = input.Importance
			}

			memoryID, err := store.Remember(AgentIDFromContext(ctx), input.Text, importance, input.Category)
			if err != nil {
				return RememberOutput{}, fmt.Errorf("failed to remember: %w", err)
			}

			return RememberOutput{
				Message: fmt.Sprintf("Remembered (ID: %s, importance: %.1f, category: %s): %s",
					memoryID, importance, input.Category, truncate(input.Text, 100)),
			}, nil
		},
	)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
