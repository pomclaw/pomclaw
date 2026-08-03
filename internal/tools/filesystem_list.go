package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/pomclaw/pomclaw/internal/model"
)

// ListFilesTool lists files in a directory, optionally through a sandbox container.
type ListFilesTool struct {
	restrict        bool
	contextFileIntc *ContextFileInterceptor // unused, satisfies InterceptorAware
}

func NewListFilesTool(restrict bool, contextFileIntc *ContextFileInterceptor) tool.InvokableTool {
	return &ListFilesTool{
		restrict:        restrict,
		contextFileIntc: contextFileIntc,
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

// ListFilesOutput represents a single file in the listing.
type ListFilesOutput struct {
	Path    string `json:"path"`
	Size    int64  `json:"size,omitempty"`
	Updated string `json:"updated,omitempty"`
}

func (t *ListFilesTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input ListFilesInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	agentID := AgentIDFromContext(ctx)
	if agentID == "" {
		return "", fmt.Errorf("agent context required")
	}

	// Virtual FS: route context files to DB
	if t.contextFileIntc == nil {
		return "", fmt.Errorf("context file interceptor not available")
	}

	var results []ListFilesOutput
	var err error
	var files []*model.AgentContextFiles

	if input.Path == "" || input.Path == "." || input.Path == "/" {
		// 全量查询：查询所有上下文文件元数据
		files, err = t.contextFileIntc.ListAllContextFilesMetadata(ctx, agentID)
	} else {
		// 模糊查询：按路径前缀查询
		files, err = t.contextFileIntc.SearchContextFilesMetadata(ctx, agentID, input.Path)
	}
	if err != nil {
		return "", fmt.Errorf("failed to search context files: %w", err)
	}

	// Build response from metadata (no content loaded)
	for _, f := range files {
		results = append(results, ListFilesOutput{
			Path:    f.FileName,
			Updated: f.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// Convert to JSON response
	respBytes, err := json.Marshal(results)
	if err != nil {
		return "", fmt.Errorf("failed to marshal results: %w", err)
	}

	return string(respBytes), nil
}
