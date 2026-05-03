package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func invokeEdit(t *testing.T, tool interface{ InvokeV(context.Context, string) (string, error) }, ctx context.Context, input interface{}) (string, error) {
	t.Helper()
	b, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal input: %v", err)
	}
	return tool.InvokeV(ctx, string(b))
}

func TestEditTool_EditFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("Hello World\nThis is a test"), 0644)

	tool := NewEditFileTool(true)
	ctx := WithWorkspace(context.Background(), tmpDir)
	_, err := invokeEdit(t, tool, ctx, EditFileInput{
		Path:    testFile,
		OldText: "World",
		NewText: "Universe",
	})

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read edited file: %v", err)
	}
	contentStr := string(content)
	if !strings.Contains(contentStr, "Hello Universe") {
		t.Errorf("Expected file to contain 'Hello Universe', got: %s", contentStr)
	}
	if strings.Contains(contentStr, "Hello World") {
		t.Errorf("Expected 'Hello World' to be replaced, got: %s", contentStr)
	}
}

func TestEditTool_EditFile_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "nonexistent.txt")

	tool := NewEditFileTool(true)
	ctx := WithWorkspace(context.Background(), tmpDir)
	_, err := invokeEdit(t, tool, ctx, EditFileInput{
		Path:    testFile,
		OldText: "old",
		NewText: "new",
	})

	if err == nil {
		t.Errorf("Expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' message, got: %v", err)
	}
}

func TestEditTool_EditFile_OldTextNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("Hello World"), 0644)

	tool := NewEditFileTool(true)
	ctx := WithWorkspace(context.Background(), tmpDir)
	_, err := invokeEdit(t, tool, ctx, EditFileInput{
		Path:    testFile,
		OldText: "Goodbye",
		NewText: "Hello",
	})

	if err == nil {
		t.Errorf("Expected error when old_text not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' message, got: %v", err)
	}
}

func TestEditTool_EditFile_MultipleMatches(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test test test"), 0644)

	tool := NewEditFileTool(true)
	ctx := WithWorkspace(context.Background(), tmpDir)
	_, err := invokeEdit(t, tool, ctx, EditFileInput{
		Path:    testFile,
		OldText: "test",
		NewText: "done",
	})

	if err == nil {
		t.Errorf("Expected error when old_text appears multiple times")
	}
	if !strings.Contains(err.Error(), "times") {
		t.Errorf("Expected 'multiple times' message, got: %v", err)
	}
}

func TestEditTool_EditFile_OutsideAllowedDir(t *testing.T) {
	tmpDir := t.TempDir()
	otherDir := t.TempDir()
	testFile := filepath.Join(otherDir, "test.txt")
	os.WriteFile(testFile, []byte("content"), 0644)

	tool := NewEditFileTool(true)
	ctx := WithWorkspace(context.Background(), tmpDir) // workspace is tmpDir, not otherDir
	_, err := invokeEdit(t, tool, ctx, EditFileInput{
		Path:    testFile,
		OldText: "content",
		NewText: "new",
	})

	if err == nil {
		t.Errorf("Expected error when path is outside allowed directory")
	}
	if !strings.Contains(err.Error(), "outside") {
		t.Errorf("Expected 'outside' message, got: %v", err)
	}
}

func TestEditTool_EditFile_MissingPath(t *testing.T) {
	tool := NewEditFileTool(false)
	ctx := context.Background()
	_, err := invokeEdit(t, tool, ctx, EditFileInput{
		OldText: "old",
		NewText: "new",
	})

	if err == nil {
		t.Errorf("Expected error when path is missing")
	}
}

func TestEditTool_EditFile_MissingOldText(t *testing.T) {
	tool := NewEditFileTool(false)
	ctx := context.Background()
	_, err := invokeEdit(t, tool, ctx, EditFileInput{
		Path:    "/tmp/test.txt",
		NewText: "new",
	})

	if err == nil {
		t.Errorf("Expected error when old_text is missing")
	}
}

func TestEditTool_AppendFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("Initial content"), 0644)

	tool := NewAppendFileTool(false)
	ctx := context.Background()
	result, err := invokeEdit(t, tool, ctx, AppendFileInput{
		Path:    testFile,
		Content: "\nAppended content",
	})

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	var out AppendFileOutput
	if jsonErr := json.Unmarshal([]byte(result), &out); jsonErr != nil {
		t.Fatalf("Failed to parse result: %v", jsonErr)
	}
	if !strings.Contains(out.Message, "Appended") {
		t.Errorf("Expected 'Appended' message, got: %s", out.Message)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	contentStr := string(content)
	if !strings.Contains(contentStr, "Initial content") {
		t.Errorf("Expected original content to remain, got: %s", contentStr)
	}
	if !strings.Contains(contentStr, "Appended content") {
		t.Errorf("Expected appended content, got: %s", contentStr)
	}
}

func TestEditTool_AppendFile_MissingPath(t *testing.T) {
	tool := NewAppendFileTool(false)
	ctx := context.Background()
	_, err := invokeEdit(t, tool, ctx, AppendFileInput{Content: "test"})

	if err == nil {
		t.Errorf("Expected error when path is missing")
	}
}

func TestEditTool_AppendFile_MissingContent(t *testing.T) {
	tool := NewAppendFileTool(false)
	ctx := context.Background()
	_, err := invokeEdit(t, tool, ctx, AppendFileInput{Path: "/tmp/test.txt"})

	if err == nil {
		t.Errorf("Expected error when content is missing")
	}
}
