package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/pomclaw/pomclaw/internal/model"
)

// MemorySearchTool implements the memory_search tool for hybrid semantic + FTS search.
type MemorySearchTool struct {
	memStore model.MemoryDocumentsModel // Postgres-backed
	hasKG    bool                       // knowledge_graph_search tool is available
}

func NewMemorySearchTool(memStore model.MemoryDocumentsModel, hasKG bool) tool.InvokableTool {
	return &MemorySearchTool{
		memStore: memStore,
		hasKG:    hasKG,
	}
}

func (t *MemorySearchTool) Name() string { return "memory_search" }

func (t *MemorySearchTool) Description() string {
	return "Mandatory recall step: semantically search MEMORY.md + memory/*.md before answering questions about prior work, decisions, dates, people, preferences, or todos; returns top snippets with path + lines. If response has disabled=true, memory retrieval is unavailable and should be surfaced to the user. IMPORTANT: Always query in the SAME language as the stored memory content. If the user speaks Vietnamese, search in Vietnamese. If memory was written in English, search in English. Matching the language dramatically improves search accuracy. If no relevant results found or confidence is low, tell the user you checked but found nothing — do not fabricate or guess memories."
}

// Info implements eino's BaseTool interface
func (t *MemorySearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.Name(),
		Desc: t.Description(),
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type:     schema.String,
				Desc:     "Natural language search query. Must be in the same language as the stored memory content (e.g., Vietnamese if memory is in Vietnamese).",
				Required: true,
			},
			"maxResults": {
				Type:     schema.Number,
				Desc:     "Maximum number of results to return (default: 6)",
				Required: false,
			},
			"minScore": {
				Type:     schema.Number,
				Desc:     "Minimum relevance score threshold (0-1)",
				Required: false,
			},
			"depth": {
				Type:     schema.String,
				Desc:     "Result depth: l0 (abstracts only), l1 (overview), l2 (full content). Default: l1. Only affects episodic memories.",
				Required: false,
			},
		}),
	}, nil
}

// MemorySearchInput defines the input parameters for MemorySearchTool.
type MemorySearchInput struct {
	Query      string  `json:"query"`
	MaxResults int     `json:"maxResults,omitempty"`
	MinScore   float64 `json:"minScore,omitempty"`
	Depth      string  `json:"depth,omitempty"`
}

func (t *MemorySearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input MemorySearchInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if input.Query == "" {
		return "", fmt.Errorf("query parameter is required")
	}

	agentID := AgentIDFromContext(ctx)
	if t.memStore == nil || agentID == "" {
		return "", fmt.Errorf("memory system not available")
	}

	userID := MemoryUserID(ctx)
	searchOpts := model.MemorySearchOptions{
		MaxResults: input.MaxResults,
		MinScore:   input.MinScore,
	}
	// Apply per-agent memory config overrides if set
	if mc := MemoryConfigFromCtx(ctx); mc != nil {
		if mc.MaxResults > 0 && searchOpts.MaxResults <= 0 {
			searchOpts.MaxResults = mc.MaxResults
		}
		if mc.VectorWeight > 0 {
			searchOpts.VectorWeight = mc.VectorWeight
		}
		if mc.TextWeight > 0 {
			searchOpts.TextWeight = mc.TextWeight
		}
		if mc.MinScore > 0 && searchOpts.MinScore <= 0 {
			searchOpts.MinScore = mc.MinScore
		}
	}

	results, err := t.memStore.Search(ctx, input.Query, agentID, userID, searchOpts)
	if err != nil {
		return "", fmt.Errorf("memory search failed: %v", err)
	}

	// Fallback: also search leader's memory for team members and merge results.
	if leaderID := LeaderAgentIDFromCtx(ctx); leaderID != "" && leaderID != agentID {
		leaderResults, lerr := t.memStore.Search(ctx, input.Query, leaderID, userID, searchOpts)
		if lerr != nil && userID != "" {
			leaderResults, _ = t.memStore.Search(ctx, input.Query, leaderID, "", searchOpts)
		}
		results = append(results, leaderResults...)
	}

	if len(results) == 0 {
		return "No memory results found for query: " + input.Query, nil
	}

	// Build output with tier labels
	type taggedResult struct {
		Tier string `json:"tier"`
		model.MemorySearchResult
	}
	var combined []taggedResult
	for _, r := range results {
		combined = append(combined, taggedResult{Tier: "document", MemorySearchResult: r})
	}

	output := map[string]any{
		"results": combined,
		"count":   len(combined),
	}
	if t.hasKG {
		output["hint"] = "Also run knowledge_graph_search if the query involves people, teams, projects, or connections between entities."
	}
	data, _ := json.MarshalIndent(output, "", "  ")

	return string(data), nil
}

// MemoryGetTool implements the memory_get tool for reading specific memory files.
type MemoryGetTool struct {
	memStore model.MemoryDocumentsModel // Postgres-backed
}

func NewMemoryGetTool(memStore model.MemoryDocumentsModel) tool.InvokableTool {
	return &MemoryGetTool{memStore: memStore}
}

func (t *MemoryGetTool) Name() string { return "memory_get" }

func (t *MemoryGetTool) Description() string {
	return "Safe snippet read from MEMORY.md or memory/*.md with optional from/lines; use after memory_search to pull only the needed lines and keep context small."
}

// Info implements eino's BaseTool interface
func (t *MemoryGetTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.Name(),
		Desc: t.Description(),
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type:     schema.String,
				Desc:     "Relative path to memory file (e.g., 'MEMORY.md' or 'memory/notes.md')",
				Required: true,
			},
			"from": {
				Type:     schema.Number,
				Desc:     "Start line number (1-indexed). Omit to read from beginning.",
				Required: false,
			},
			"lines": {
				Type:     schema.Number,
				Desc:     "Number of lines to read. Omit to read entire file.",
				Required: false,
			},
		}),
	}, nil
}

// MemoryGetInput defines the input parameters for MemoryGetTool.
type MemoryGetInput struct {
	Path  string `json:"path"`
	From  int    `json:"from,omitempty"`
	Lines int    `json:"lines,omitempty"`
}

func (t *MemoryGetTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input MemoryGetInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if input.Path == "" {
		return "", fmt.Errorf("path parameter is required")
	}

	agentID := AgentIDFromContext(ctx)
	if t.memStore == nil || agentID == "" {
		return "", fmt.Errorf("memory system not available")
	}

	userID := MemoryUserID(ctx)

	// Try per-user first, then global
	content, err := t.memStore.GetDocument(ctx, agentID, userID, input.Path)
	if (err != nil || content == "") && userID != "" {
		content, err = t.memStore.GetDocument(ctx, agentID, "", input.Path)
	}
	// Fallback: try leader's memory for team members.
	if err != nil || content == "" {
		if leaderID := LeaderAgentIDFromCtx(ctx); leaderID != "" && leaderID != agentID {
			content, err = t.memStore.GetDocument(ctx, leaderID, userID, input.Path)
			if (err != nil || content == "") && userID != "" {
				content, err = t.memStore.GetDocument(ctx, leaderID, "", input.Path)
			}
		}
	}
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %v", input.Path, err)
	}

	text := extractLines(content, input.From, input.Lines)
	if text == "" {
		return fmt.Sprintf("File %s is empty or the specified range has no content.", input.Path), nil
	}

	data, _ := json.MarshalIndent(map[string]any{
		"path": input.Path,
		"text": text,
	}, "", "  ")
	return string(data), nil
}

// extractLines extracts a range of lines from content.
// fromLine is 1-indexed. If 0, starts from beginning. If numLines is 0, returns all.
func extractLines(content string, fromLine, numLines int) string {
	if fromLine <= 0 && numLines <= 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	start := 0
	if fromLine > 0 {
		start = fromLine - 1
	}
	if start >= len(lines) {
		return ""
	}

	end := len(lines)
	if numLines > 0 && start+numLines < end {
		end = start + numLines
	}

	return strings.Join(lines[start:end], "\n")
}
