package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// WriteFileTool writes content to a file, optionally through a sandbox container.
type WriteFileTool struct {
	restrict        bool
	contextFileIntc *ContextFileInterceptor // nil = no virtual FS routing
	memIntc         *MemoryInterceptor      // nil = no memory routing
}

func NewWriteFileTool(restrict bool, contextFileIntc *ContextFileInterceptor, memIntc *MemoryInterceptor) tool.InvokableTool {
	return &WriteFileTool{
		restrict:        restrict,
		contextFileIntc: contextFileIntc,
		memIntc:         memIntc,
	}
}

func (t *WriteFileTool) Name() string { return "write_file" }
func (t *WriteFileTool) Description() string {
	return "Write content to a file, creating directories as needed. " +
		"IMPORTANT: content longer than ~12000 characters may be truncated by the API. " +
		"For large files, use the edit tool to build the file in sections, or split into multiple write_file calls with append=true."
}

// Info implements eino's BaseTool interface
func (t *WriteFileTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.Name(),
		Desc: t.Description(),
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type:     schema.String,
				Desc:     "File path (relative to workspace, or absolute)",
				Required: true,
			},
			"content": {
				Type:     schema.String,
				Desc:     "Content to write",
				Required: true,
			},
			"append": {
				Type:     schema.Boolean,
				Desc:     "Append content to the file instead of overwriting. Use this to build large files in chunks.",
				Required: false,
			},
		}),
	}, nil
}

// WriteFileInput defines the input parameters for WriteFileTool.
type WriteFileInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Append  bool   `json:"append,omitempty"`
}

func (t *WriteFileTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input WriteFileInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if input.Path == "" {
		return "", fmt.Errorf("path is required")
	}

	// Virtual FS: route context files to DB
	if t.contextFileIntc != nil {
		if handled, err := t.contextFileIntc.WriteFile(ctx, input.Path, input.Content); handled {
			if err != nil {
				return "", fmt.Errorf("failed to write context file: %v", err)
			}
			return fmt.Sprintf("Context file written: %s (%d bytes)", input.Path, len(input.Content)), nil
		}
	}

	// Virtual FS: route memory files to DB
	if t.memIntc != nil {
		if mwr, err := t.memIntc.WriteFile(ctx, input.Path, input.Content, input.Append); mwr.Handled {
			if err != nil {
				return "", fmt.Errorf("failed to write memory file: %v", err)
			}
			msg := fmt.Sprintf("Memory file written: %s (%d bytes)", input.Path, len(input.Content))
			if mwr.PreviousContent != "" {
				prev := mwr.PreviousContent
				prevRunes := []rune(prev)
				if len(prevRunes) > 4000 {
					prev = string(prevRunes[:4000]) + "\n... (truncated)"
				}
				msg += fmt.Sprintf("\n\n⚠️ WARNING: This file had existing content (%d chars) that was replaced. "+
					"If the old content below contains information not present in your new version, "+
					"please re-write the file to merge both.\n\n"+
					"--- PREVIOUS CONTENT ---\n%s\n--- END PREVIOUS CONTENT ---",
					len([]rune(mwr.PreviousContent)), prev)
			}
			return msg, nil
		}
	}

	return "", fmt.Errorf("write file failed")
}
