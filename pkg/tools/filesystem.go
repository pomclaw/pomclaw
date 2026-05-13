package tools

import (
	"context"
	"fmt"
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

type ReadFileTool struct {
	restrict bool
}

func NewReadFileTool(restrict bool) *ReadFileTool {
	return &ReadFileTool{restrict: restrict}
}

func (t *ReadFileTool) Name() string {
	return "read_file"
}

func (t *ReadFileTool) Description() string {
	return "Read the contents of a file"
}

func (t *ReadFileTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to read",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ReadFileTool) Execute(ctx context.Context, args map[string]interface{}) *ToolResult {
	path, ok := args["path"].(string)
	if !ok {
		return ErrorResult("path is required")
	}

	resolvedPath, err := validatePath(ctx, path, t.restrict)
	if err != nil {
		return ErrorResult(err.Error())
	}

	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return ErrorResult(fmt.Sprintf("failed to read file: %v", err))
	}

	return NewToolResult(string(content))
}

type WriteFileTool struct {
	restrict bool
}

func NewWriteFileTool(restrict bool) *WriteFileTool {
	return &WriteFileTool{restrict: restrict}
}

func (t *WriteFileTool) Name() string {
	return "write_file"
}

func (t *WriteFileTool) Description() string {
	return "Write content to a file"
}

func (t *WriteFileTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to write",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content to write to the file",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (t *WriteFileTool) Execute(ctx context.Context, args map[string]interface{}) *ToolResult {
	path, ok := args["path"].(string)
	if !ok {
		return ErrorResult("path is required")
	}

	content, ok := args["content"].(string)
	if !ok {
		return ErrorResult("content is required")
	}

	resolvedPath, err := validatePath(ctx, path, t.restrict)
	if err != nil {
		return ErrorResult(err.Error())
	}

	dir := filepath.Dir(resolvedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ErrorResult(fmt.Sprintf("failed to create directory: %v", err))
	}

	if err := os.WriteFile(resolvedPath, []byte(content), 0644); err != nil {
		return ErrorResult(fmt.Sprintf("failed to write file: %v", err))
	}

	return SilentResult(fmt.Sprintf("File written: %s", path))
}

type ListDirTool struct {
	restrict bool
}

func NewListDirTool(restrict bool) *ListDirTool {
	return &ListDirTool{restrict: restrict}
}

func (t *ListDirTool) Name() string {
	return "list_dir"
}

func (t *ListDirTool) Description() string {
	return "List files and directories in a path"
}

func (t *ListDirTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to list",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ListDirTool) Execute(ctx context.Context, args map[string]interface{}) *ToolResult {
	path, ok := args["path"].(string)
	if !ok {
		path = "."
	}

	resolvedPath, err := validatePath(ctx, path, t.restrict)
	if err != nil {
		return ErrorResult(err.Error())
	}

	entries, err := os.ReadDir(resolvedPath)
	if err != nil {
		return ErrorResult(fmt.Sprintf("failed to read directory: %v", err))
	}

	result := ""
	for _, entry := range entries {
		if entry.IsDir() {
			result += "DIR:  " + entry.Name() + "\n"
		} else {
			result += "FILE: " + entry.Name() + "\n"
		}
	}

	return NewToolResult(result)
}

// EditFileTool edits a file by replacing old_text with new_text.
// The old_text must exist exactly in the file.
type EditFileTool struct {
	restrict bool
}

// NewEditFileTool creates a new EditFileTool with optional directory restriction.
func NewEditFileTool(restrict bool) *EditFileTool {
	return &EditFileTool{
		restrict: restrict,
	}
}

func (t *EditFileTool) Name() string {
	return "edit_file"
}

func (t *EditFileTool) Description() string {
	return "Edit a file by replacing old_text with new_text. The old_text must exist exactly in the file."
}

func (t *EditFileTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The file path to edit",
			},
			"old_text": map[string]interface{}{
				"type":        "string",
				"description": "The exact text to find and replace",
			},
			"new_text": map[string]interface{}{
				"type":        "string",
				"description": "The text to replace with",
			},
		},
		"required": []string{"path", "old_text", "new_text"},
	}
}

func (t *EditFileTool) Execute(ctx context.Context, args map[string]interface{}) *ToolResult {
	path, ok := args["path"].(string)
	if !ok {
		return ErrorResult("path is required")
	}

	oldText, ok := args["old_text"].(string)
	if !ok {
		return ErrorResult("old_text is required")
	}

	newText, ok := args["new_text"].(string)
	if !ok {
		return ErrorResult("new_text is required")
	}

	resolvedPath, err := validatePath(ctx, path, t.restrict)
	if err != nil {
		return ErrorResult(err.Error())
	}

	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		return ErrorResult(fmt.Sprintf("file not found: %s", path))
	}

	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return ErrorResult(fmt.Sprintf("failed to read file: %v", err))
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, oldText) {
		return ErrorResult("old_text not found in file. Make sure it matches exactly")
	}

	count := strings.Count(contentStr, oldText)
	if count > 1 {
		return ErrorResult(fmt.Sprintf("old_text appears %d times. Please provide more context to make it unique", count))
	}

	newContent := strings.Replace(contentStr, oldText, newText, 1)

	if err := os.WriteFile(resolvedPath, []byte(newContent), 0644); err != nil {
		return ErrorResult(fmt.Sprintf("failed to write file: %v", err))
	}

	return SilentResult(fmt.Sprintf("File edited: %s", path))
}

type AppendFileTool struct {
	restrict bool
}

func NewAppendFileTool(restrict bool) *AppendFileTool {
	return &AppendFileTool{restrict: restrict}
}

func (t *AppendFileTool) Name() string {
	return "append_file"
}

func (t *AppendFileTool) Description() string {
	return "Append content to the end of a file"
}

func (t *AppendFileTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The file path to append to",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The content to append",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (t *AppendFileTool) Execute(ctx context.Context, args map[string]interface{}) *ToolResult {
	path, ok := args["path"].(string)
	if !ok {
		return ErrorResult("path is required")
	}

	content, ok := args["content"].(string)
	if !ok {
		return ErrorResult("content is required")
	}

	resolvedPath, err := validatePath(ctx, path, t.restrict)
	if err != nil {
		return ErrorResult(err.Error())
	}

	f, err := os.OpenFile(resolvedPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return ErrorResult(fmt.Sprintf("failed to open file: %v", err))
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return ErrorResult(fmt.Sprintf("failed to append to file: %v", err))
	}

	return SilentResult(fmt.Sprintf("Appended to %s", path))
}
