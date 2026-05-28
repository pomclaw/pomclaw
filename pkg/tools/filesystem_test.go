package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	einotool "github.com/cloudwego/eino/components/tool"
)

func invokeFS(t *testing.T, tool einotool.InvokableTool, input interface{}) (string, error) {
	t.Helper()
	b, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal input: %v", err)
	}
	return tool.InvokableRun(context.Background(), string(b))
}

func invokeFSCtx(t *testing.T, tool einotool.InvokableTool, ctx context.Context, input interface{}) (string, error) {
	t.Helper()
	b, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal input: %v", err)
	}
	return tool.InvokableRun(ctx, string(b))
}

func TestFilesystemTool_ReadFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	tool := NewReadFileTool(false)
	result, err := invokeFS(t, tool, ReadFileInput{Path: testFile})

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	var out ReadFileOutput
	if jsonErr := json.Unmarshal([]byte(result), &out); jsonErr != nil {
		t.Fatalf("Failed to parse result: %v", jsonErr)
	}

	if !strings.Contains(out.Content, "test content") {
		t.Errorf("Expected content 'test content', got: %s", out.Content)
	}
}

func TestFilesystemTool_ReadFile_NotFound(t *testing.T) {
	tool := NewReadFileTool(false)
	_, err := invokeFS(t, tool, ReadFileInput{Path: "/nonexistent_file_12345.txt"})

	if err == nil {
		t.Errorf("Expected error for missing file")
	}
	if !strings.Contains(err.Error(), "failed to read") {
		t.Errorf("Expected 'failed to read' message, got: %v", err)
	}
}

func TestFilesystemTool_ReadFile_MissingPath(t *testing.T) {
	tool := NewReadFileTool(false)
	_, err := invokeFS(t, tool, ReadFileInput{})

	if err == nil {
		t.Errorf("Expected error when path is missing")
	}
	if !strings.Contains(err.Error(), "path is required") {
		t.Errorf("Expected 'path is required' message, got: %v", err)
	}
}

func TestFilesystemTool_WriteFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "newfile.txt")

	tool := NewWriteFileTool(false)
	result, err := invokeFS(t, tool, WriteFileInput{Path: testFile, Content: "hello world"})

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	var out WriteFileOutput
	if jsonErr := json.Unmarshal([]byte(result), &out); jsonErr != nil {
		t.Fatalf("Failed to parse result: %v", jsonErr)
	}
	if !strings.Contains(out.Message, "File written") {
		t.Errorf("Expected 'File written' message, got: %s", out.Message)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}
	if string(content) != "hello world" {
		t.Errorf("Expected file content 'hello world', got: %s", string(content))
	}
}

func TestFilesystemTool_WriteFile_CreateDir(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "subdir", "newfile.txt")

	tool := NewWriteFileTool(false)
	_, err := invokeFS(t, tool, WriteFileInput{Path: testFile, Content: "test"})

	if err != nil {
		t.Fatalf("Expected success with directory creation, got error: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}
	if string(content) != "test" {
		t.Errorf("Expected file content 'test', got: %s", string(content))
	}
}

func TestFilesystemTool_WriteFile_MissingPath(t *testing.T) {
	tool := NewWriteFileTool(false)
	_, err := invokeFS(t, tool, WriteFileInput{Content: "test"})

	if err == nil {
		t.Errorf("Expected error when path is missing")
	}
}

func TestFilesystemTool_WriteFile_MissingContent(t *testing.T) {
	tool := NewWriteFileTool(false)
	_, err := invokeFS(t, tool, WriteFileInput{Path: "/tmp/test.txt"})

	if err == nil {
		t.Errorf("Expected error when content is missing")
	}
	if !strings.Contains(err.Error(), "content is required") {
		t.Errorf("Expected 'content is required' message, got: %v", err)
	}
}

func TestFilesystemTool_ListDir_Success(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	tool := NewListDirTool(false)
	result, err := invokeFS(t, tool, ListDirInput{Path: tmpDir})

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	var out ListDirOutput
	if jsonErr := json.Unmarshal([]byte(result), &out); jsonErr != nil {
		t.Fatalf("Failed to parse result: %v", jsonErr)
	}

	if !strings.Contains(out.Entries, "file1.txt") || !strings.Contains(out.Entries, "file2.txt") {
		t.Errorf("Expected files in listing, got: %s", out.Entries)
	}
	if !strings.Contains(out.Entries, "subdir") {
		t.Errorf("Expected subdir in listing, got: %s", out.Entries)
	}
}

func TestFilesystemTool_ListDir_NotFound(t *testing.T) {
	tool := NewListDirTool(false)
	_, err := invokeFS(t, tool, ListDirInput{Path: "/nonexistent_directory_12345"})

	if err == nil {
		t.Errorf("Expected error for non-existent directory")
	}
	if !strings.Contains(err.Error(), "failed to read") {
		t.Errorf("Expected error message, got: %v", err)
	}
}

func TestFilesystemTool_ListDir_DefaultPath(t *testing.T) {
	tool := NewListDirTool(false)
	_, err := invokeFS(t, tool, ListDirInput{})

	if err != nil {
		t.Errorf("Expected success with default path '.', got error: %v", err)
	}
}
