package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// ReadFileTool reads file contents, optionally through a sandbox container.
type ReadFileTool struct {
	restrict        bool
	contextFileIntc *ContextFileInterceptor // nil = no virtual FS routing
	memIntc         *MemoryInterceptor      // nil = no memory routing
}

func NewReadFileTool(restrict bool, contextFileIntc *ContextFileInterceptor, memIntc *MemoryInterceptor) tool.InvokableTool {
	return &ReadFileTool{
		restrict:        restrict,
		contextFileIntc: contextFileIntc,
		memIntc:         memIntc,
	}
}

func (t *ReadFileTool) Name() string { return "read_file" }

func (t *ReadFileTool) Description() string {
	return "Read the contents of a file. For large files, use offset and limit to read specific line ranges."
}

// Info implements eino's BaseTool interface
func (t *ReadFileTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.Name(),
		Desc: t.Description(),
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type:     schema.String,
				Desc:     "File path (relative to workspace, or absolute)",
				Required: true,
			},
			"offset": {
				Type:     schema.Integer,
				Desc:     "Start reading from this line number (0-indexed). Defaults to 0.",
				Required: false,
			},
			"limit": {
				Type:     schema.Integer,
				Desc:     "Maximum number of lines to return. Omit to read until output cap.",
				Required: false,
			},
		}),
	}, nil
}

// ReadFileInput defines the input parameters for ReadFileTool.
type ReadFileInput struct {
	Path   string `json:"path"`
	Offset int    `json:"offset,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

func (t *ReadFileTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input ReadFileInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if input.Path == "" {
		return "", fmt.Errorf("path is required")
	}

	// Virtual FS: route context files to DB
	if t.contextFileIntc != nil {
		if content, handled, err := t.contextFileIntc.ReadFile(ctx, input.Path); handled {
			if err != nil {
				return "", fmt.Errorf("failed to read context file: %v", err)
			}
			if content == "" {
				return "", fmt.Errorf("context file not found: %s", input.Path)
			}
			return content, nil
		}
	}

	// Virtual FS: route memory files to DB
	if t.memIntc != nil {
		if content, handled, err := t.memIntc.ReadFile(ctx, input.Path); handled {
			if err != nil {
				return "", fmt.Errorf("failed to read memory file: %v", err)
			}
			if content == "" {
				return fmt.Sprintf("(memory file %s does not exist yet — it will be created when memory is saved)", input.Path), nil
			}
			return content + "\n\n[Source: database, not filesystem]", nil
		}
	}

	return "", fmt.Errorf("no file")
}
