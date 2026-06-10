package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// ListFilesTool lists files in a directory, optionally through a sandbox container.
type ListFilesTool struct {
	restrict        bool
	contextFileIntc *ContextFileInterceptor // unused, satisfies InterceptorAware
	memIntc         *MemoryInterceptor      // nil = no memory routing
}

func NewListFilesTool(restrict bool, contextFileIntc *ContextFileInterceptor, memIntc *MemoryInterceptor) tool.InvokableTool {
	return &ListFilesTool{
		restrict:        restrict,
		contextFileIntc: contextFileIntc,
		memIntc:         memIntc,
	}
}

func (t *ListFilesTool) Name() string { return "list_files" }

func (t *ListFilesTool) Description() string { return "List files and directories in a path" }

// Info implements eino's BaseTool interface
func (t *ListFilesTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.Name(),
		Desc: t.Description(),
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type:     schema.String,
				Desc:     "Directory path (relative to workspace; omit for workspace root)",
				Required: false,
			},
		}),
	}, nil
}

// ListFilesInput defines the input parameters for ListFilesTool.
type ListFilesInput struct {
	Path string `json:"path,omitempty"`
}

func (t *ListFilesTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input ListFilesInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if input.Path == "" {
		input.Path = "."
	}

	// Virtual FS: route memory directory listing to DB
	if t.memIntc != nil {
		if listing, handled, err := t.memIntc.ListFiles(ctx, input.Path); handled {
			if err != nil {
				return "", fmt.Errorf("failed to list memory files: %v", err)
			}
			if listing == "" {
				return "No memory files stored yet", nil
			}
			return listing + "\n[Source: database, not filesystem]", nil
		}
	}

	return "", fmt.Errorf("failed to list memory")
}
