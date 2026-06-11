package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
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

// RecallTool searches long-term memory using semantic similarity.
type RecallTool struct {
	store Recaller
}

// NewRecallTool creates a recall tool.
func NewRecallTool(store Recaller) tool.InvokableTool {
	return &RecallTool{store: store}
}

func (t *RecallTool) Name() string { return "recall" }

func (t *RecallTool) Description() string {
	return "Search long-term memory using semantic similarity. Use this to find previously remembered information by describing what you're looking for."
}

// Info implements eino's BaseTool interface
func (t *RecallTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.Name(),
		Desc: t.Description(),
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type:     schema.String,
				Desc:     "Search query describing what to recall",
				Required: true,
			},
			"max_results": {
				Type:     schema.Integer,
				Desc:     "Maximum number of results to return (default: 5)",
				Required: false,
			},
		}),
	}, nil
}

func (t *RecallTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input RecallInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if input.Query == "" {
		return "", fmt.Errorf("query parameter is required")
	}

	maxResults := 5
	if input.MaxResults > 0 {
		maxResults = input.MaxResults
	}

	results, err := t.store.Recall(AgentIDFromContext(ctx), input.Query, maxResults)
	if err != nil {
		return "", fmt.Errorf("recall failed: %w", err)
	}

	if len(results) == 0 {
		return "No matching memories found for: " + input.Query, nil
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

	return sb.String(), nil
}
