package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// EditTool performs search-and-replace edits on files.
// Supports context file interceptor and sandbox routing.
type EditTool struct {
	restrict        bool
	contextFileIntc *ContextFileInterceptor
}

func NewEditTool(restrict bool, contextFileIntc *ContextFileInterceptor) tool.InvokableTool {
	return &EditTool{
		restrict:        restrict,
		contextFileIntc: contextFileIntc,
	}
}

func (t *EditTool) Name() string { return "edit" }
func (t *EditTool) Description() string {
	return "Edit a file by replacing exact text matches. Use old_string/new_string for precise edits without rewriting the entire file."
}

// Info implements eino's BaseTool interface
func (t *EditTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.Name(),
		Desc: t.Description(),
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type:     schema.String,
				Desc:     "File path (relative to workspace, or absolute)",
				Required: true,
			},
			"old_string": {
				Type:     schema.String,
				Desc:     "Exact text to find (must match uniquely unless replace_all is true)",
				Required: true,
			},
			"new_string": {
				Type:     schema.String,
				Desc:     "Replacement text",
				Required: true,
			},
			"replace_all": {
				Type:     schema.Boolean,
				Desc:     "Replace all occurrences (default: false, requires unique match)",
				Required: false,
			},
		}),
	}, nil
}

// EditInput defines the input parameters for EditTool.
type EditInput struct {
	Path       string `json:"path"`
	OldString  string `json:"old_string"`
	NewString  string `json:"new_string"`
	ReplaceAll bool   `json:"replace_all,omitempty"`
}

func (t *EditTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input EditInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if input.Path == "" {
		return "", fmt.Errorf("path is required")
	}
	if input.OldString == "" {
		return "", fmt.Errorf("old_string is required")
	}
	if input.OldString == input.NewString {
		return "", fmt.Errorf("old_string and new_string are identical")
	}

	// Virtual FS: context files
	if t.contextFileIntc != nil {
		if content, handled, err := t.contextFileIntc.ReadFile(ctx, input.Path); handled {
			if err != nil {
				return "", fmt.Errorf("failed to read context file: %v", err)
			}
			if content == "" {
				return "", fmt.Errorf("context file not found: %s", input.Path)
			}
			newContent, err := applyEdit(content, input.OldString, input.NewString, input.ReplaceAll)
			if err != nil {
				return "", err
			}
			if _, err := t.contextFileIntc.WriteFile(ctx, input.Path, newContent); err != nil {
				return "", fmt.Errorf("failed to write context file: %v", err)
			}
			return fmt.Sprintf("Context file edited: %s", input.Path), nil
		}
	}

	return "", fmt.Errorf("write memory failed")
}

// applyEdit performs the search-and-replace. Returns (newContent, nil) on success
// or ("", error) on failure.
func applyEdit(content, oldStr, newStr string, replaceAll bool) (string, error) {
	count := strings.Count(content, oldStr)
	if count == 0 {
		return "", fmt.Errorf("old_string not found in file")
	}
	if !replaceAll && count > 1 {
		return "", fmt.Errorf("old_string found %d times — use replace_all=true or provide a more specific match", count)
	}

	if replaceAll {
		return strings.ReplaceAll(content, oldStr, newStr), nil
	}
	return strings.Replace(content, oldStr, newStr, 1), nil
}
