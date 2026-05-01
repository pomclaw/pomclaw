package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

// RecallResult represents a recalled memory.
type RecallResult struct {
	MemoryID   string
	Text       string
	Importance float64
	Category   string
	Score      float64
}

// Recaller is the interface the recall tool needs for semantic memory search.
type Recaller interface {
	Recall(agentID string, query string, maxResults int) ([]RecallResult, error)
}

type RecallInput struct {
	Query      string `json:"query"`
	MaxResults int    `json:"max_results,omitempty"`
}

type RecallOutput struct {
	Results string `json:"results"`
}

func NewRecallTool(store Recaller) tool.InvokableTool {
	return utils.WrapInvokableToolWithErrorHandler(utils.NewTool[RecallInput, RecallOutput](
		&schema.ToolInfo{
			Name: "recall",
			Desc: "Search long-term memory using semantic similarity. Use this to find previously remembered information by describing what you're looking for.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"query": {
					Type:     schema.String,
					Desc:     "Search query describing what to recall",
					Required: true,
				},
				"max_results": {
					Type: schema.Integer,
					Desc: "Maximum number of results to return (default: 5)",
				},
			}),
		},
		func(ctx context.Context, input RecallInput) (RecallOutput, error) {
			if input.Query == "" {
				return RecallOutput{}, fmt.Errorf("query parameter is required")
			}

			maxResults := 5
			if input.MaxResults > 0 {
				maxResults = input.MaxResults
			}

			results, err := store.Recall(AgentIDFromContext(ctx), input.Query, maxResults)
			if err != nil {
				return RecallOutput{}, fmt.Errorf("recall failed: %w", err)
			}

			if len(results) == 0 {
				return RecallOutput{Results: "No matching memories found for: " + input.Query}, nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Found %d matching memories:\n\n", len(results)))
			for i, r := range results {
				sb.WriteString(fmt.Sprintf("%d. [%.0f%% match] (ID: %s", i+1, r.Score*100, r.MemoryID))
				if r.Category != "" {
					sb.WriteString(fmt.Sprintf(", category: %s", r.Category))
				}
				sb.WriteString(fmt.Sprintf(", importance: %.1f)\n", r.Importance))
				sb.WriteString(fmt.Sprintf("   %s\n\n", r.Text))
			}

			return RecallOutput{Results: sb.String()}, nil
		},
	), func(ctx context.Context, err error) string { return err.Error() })
}
