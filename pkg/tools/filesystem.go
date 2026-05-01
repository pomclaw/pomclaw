package tools

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"os"
	"path/filepath"
	"strings"
)

// validatePath ensures the given path is within the workspace if restrict is true.
func validatePath(ctx context.Context, path string, restrict bool) (string, error) {
	workspace := WorkspaceFromContext(ctx)
	if workspace == "" {
		return path, nil
	}

	absWorkspace, err := filepath.Abs(workspace)
	if err != nil {
		return "", fmt.Errorf("failed to resolve workspace path: %w", err)
	}

	var absPath string
	if filepath.IsAbs(path) {
		absPath = filepath.Clean(path)
	} else {
		absPath, err = filepath.Abs(filepath.Join(absWorkspace, path))
		if err != nil {
			return "", fmt.Errorf("failed to resolve file path: %w", err)
		}
	}

	if restrict {
		if !isWithinWorkspace(absPath, absWorkspace) {
			return "", fmt.Errorf("access denied: path is outside the workspace")
		}

		workspaceReal := absWorkspace
		if resolved, err := filepath.EvalSymlinks(absWorkspace); err == nil {
			workspaceReal = resolved
		}

		if resolved, err := filepath.EvalSymlinks(absPath); err == nil {
			if !isWithinWorkspace(resolved, workspaceReal) {
				return "", fmt.Errorf("access denied: symlink resolves outside workspace")
			}
		} else if os.IsNotExist(err) {
			if parentResolved, err := resolveExistingAncestor(filepath.Dir(absPath)); err == nil {
				if !isWithinWorkspace(parentResolved, workspaceReal) {
					return "", fmt.Errorf("access denied: symlink resolves outside workspace")
				}
			} else if !os.IsNotExist(err) {
				return "", fmt.Errorf("failed to resolve path: %w", err)
			}
		} else {
			return "", fmt.Errorf("failed to resolve path: %w", err)
		}
	}

	return absPath, nil
}

func resolveExistingAncestor(path string) (string, error) {
	for current := filepath.Clean(path); ; current = filepath.Dir(current) {
		if resolved, err := filepath.EvalSymlinks(current); err == nil {
			return resolved, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		if filepath.Dir(current) == current {
			return "", os.ErrNotExist
		}
	}
}

func isWithinWorkspace(candidate, workspace string) bool {
	rel, err := filepath.Rel(filepath.Clean(workspace), filepath.Clean(candidate))
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

type ReadFileInput struct {
	Path string `json:"path"`
}

type ReadFileOutput struct {
	Content string `json:"content"`
}

func NewReadFileTool(restrict bool) tool.InvokableTool {
	return utils.WrapInvokableToolWithErrorHandler(utils.NewTool[ReadFileInput, ReadFileOutput](
		&schema.ToolInfo{
			Name: "read_file",
			Desc: "Read the contents of a file",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"path": {
					Type:     schema.String,
					Desc:     "Path to the file to read",
					Required: true,
				},
			}),
		},
		func(ctx context.Context, input ReadFileInput) (ReadFileOutput, error) {
			if input.Path == "" {
				return ReadFileOutput{}, fmt.Errorf("path is required")
			}

			resolvedPath, err := validatePath(ctx, input.Path, restrict)
			if err != nil {
				return ReadFileOutput{}, err
			}

			content, err := os.ReadFile(resolvedPath)
			if err != nil {
				return ReadFileOutput{}, fmt.Errorf("failed to read file: %w", err)
			}

			return ReadFileOutput{
				Content: string(content),
			}, nil
		},
	), func(ctx context.Context, err error) string { return err.Error() })
}

type WriteFileInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type WriteFileOutput struct {
	Message string `json:"message"`
}

func NewWriteFileTool(restrict bool) tool.InvokableTool {
	return utils.WrapInvokableToolWithErrorHandler(utils.NewTool[WriteFileInput, WriteFileOutput](
		&schema.ToolInfo{
			Name: "write_file",
			Desc: "Write content to a file",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"path": {
					Type:     schema.String,
					Desc:     "Path to the file to write",
					Required: true,
				},
				"content": {
					Type:     schema.String,
					Desc:     "Content to write to the file",
					Required: true,
				},
			}),
		},
		func(ctx context.Context, input WriteFileInput) (WriteFileOutput, error) {
			if input.Path == "" {
				return WriteFileOutput{}, fmt.Errorf("path is required")
			}
			if input.Content == "" {
				return WriteFileOutput{}, fmt.Errorf("content is required")
			}

			resolvedPath, err := validatePath(ctx, input.Path, restrict)
			if err != nil {
				return WriteFileOutput{}, err
			}

			dir := filepath.Dir(resolvedPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return WriteFileOutput{}, fmt.Errorf("failed to create directory: %w", err)
			}

			if err := os.WriteFile(resolvedPath, []byte(input.Content), 0644); err != nil {
				return WriteFileOutput{}, fmt.Errorf("failed to write file: %w", err)
			}

			return WriteFileOutput{Message: fmt.Sprintf("File written: %s", input.Path)}, nil
		},
	), func(ctx context.Context, err error) string { return err.Error() })
}

type ListDirInput struct {
	Path string `json:"path,omitempty"`
}

type ListDirOutput struct {
	Entries string `json:"entries"`
}

func NewListDirTool(restrict bool) tool.InvokableTool {
	return utils.WrapInvokableToolWithErrorHandler(utils.NewTool[ListDirInput, ListDirOutput](
		&schema.ToolInfo{
			Name: "list_dir",
			Desc: "List files and directories in a path",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"path": {
					Type: schema.String,
					Desc: "Path to list (defaults to current directory)",
				},
			}),
		},
		func(ctx context.Context, input ListDirInput) (ListDirOutput, error) {
			path := input.Path
			if path == "" {
				path = "."
			}

			resolvedPath, err := validatePath(ctx, path, restrict)
			if err != nil {
				return ListDirOutput{}, err
			}

			entries, err := os.ReadDir(resolvedPath)
			if err != nil {
				return ListDirOutput{}, fmt.Errorf("failed to read directory: %w", err)
			}

			result := ""
			for _, entry := range entries {
				if entry.IsDir() {
					result += "DIR:  " + entry.Name() + "\n"
				} else {
					result += "FILE: " + entry.Name() + "\n"
				}
			}

			return ListDirOutput{Entries: result}, nil
		},
	), func(ctx context.Context, err error) string { return err.Error() })
}

type EditFileInput struct {
	Path    string `json:"path"`
	OldText string `json:"old_text"`
	NewText string `json:"new_text"`
}

type EditFileOutput struct {
	Message string `json:"message"`
}

func NewEditFileTool(restrict bool) tool.InvokableTool {
	return utils.WrapInvokableToolWithErrorHandler(utils.NewTool[EditFileInput, EditFileOutput](
		&schema.ToolInfo{
			Name: "edit_file",
			Desc: "Edit a file by replacing old_text with new_text. The old_text must exist exactly in the file.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"path": {
					Type:     schema.String,
					Desc:     "The file path to edit",
					Required: true,
				},
				"old_text": {
					Type:     schema.String,
					Desc:     "The exact text to find and replace",
					Required: true,
				},
				"new_text": {
					Type:     schema.String,
					Desc:     "The text to replace with",
					Required: true,
				},
			}),
		},
		func(ctx context.Context, input EditFileInput) (EditFileOutput, error) {
			if input.Path == "" {
				return EditFileOutput{}, fmt.Errorf("path is required")
			}
			if input.OldText == "" {
				return EditFileOutput{}, fmt.Errorf("old_text is required")
			}

			resolvedPath, err := validatePath(ctx, input.Path, restrict)
			if err != nil {
				return EditFileOutput{}, err
			}

			if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
				return EditFileOutput{}, fmt.Errorf("file not found: %s", input.Path)
			}

			content, err := os.ReadFile(resolvedPath)
			if err != nil {
				return EditFileOutput{}, fmt.Errorf("failed to read file: %w", err)
			}

			contentStr := string(content)

			if !strings.Contains(contentStr, input.OldText) {
				return EditFileOutput{}, fmt.Errorf("old_text not found in file. Make sure it matches exactly")
			}

			count := strings.Count(contentStr, input.OldText)
			if count > 1 {
				return EditFileOutput{}, fmt.Errorf("old_text appears %d times. Please provide more context to make it unique", count)
			}

			newContent := strings.Replace(contentStr, input.OldText, input.NewText, 1)

			if err := os.WriteFile(resolvedPath, []byte(newContent), 0644); err != nil {
				return EditFileOutput{}, fmt.Errorf("failed to write file: %w", err)
			}

			return EditFileOutput{Message: fmt.Sprintf("File edited: %s", input.Path)}, nil
		},
	), func(ctx context.Context, err error) string { return err.Error() })
}

type AppendFileInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type AppendFileOutput struct {
	Message string `json:"message"`
}

func NewAppendFileTool(restrict bool) tool.InvokableTool {
	return utils.WrapInvokableToolWithErrorHandler(utils.NewTool[AppendFileInput, AppendFileOutput](
		&schema.ToolInfo{
			Name: "append_file",
			Desc: "Append content to the end of a file",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"path": {
					Type:     schema.String,
					Desc:     "The file path to append to",
					Required: true,
				},
				"content": {
					Type:     schema.String,
					Desc:     "The content to append",
					Required: true,
				},
			}),
		},
		func(ctx context.Context, input AppendFileInput) (AppendFileOutput, error) {
			if input.Path == "" {
				return AppendFileOutput{}, fmt.Errorf("path is required")
			}
			if input.Content == "" {
				return AppendFileOutput{}, fmt.Errorf("content is required")
			}

			resolvedPath, err := validatePath(ctx, input.Path, restrict)
			if err != nil {
				return AppendFileOutput{}, err
			}

			f, err := os.OpenFile(resolvedPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return AppendFileOutput{}, fmt.Errorf("failed to open file: %w", err)
			}
			defer f.Close()

			if _, err := f.WriteString(input.Content); err != nil {
				return AppendFileOutput{}, fmt.Errorf("failed to append to file: %w", err)
			}

			return AppendFileOutput{Message: fmt.Sprintf("Appended to %s", input.Path)}, nil
		},
	), func(ctx context.Context, err error) string { return err.Error() })
}
