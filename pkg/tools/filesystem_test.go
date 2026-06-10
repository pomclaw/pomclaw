package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFilesystemTools_Creation(t *testing.T) {
	// Test that tools can be created with the new dependency injection pattern
	tmpDir := t.TempDir()

	// Create a simple test file
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test ReadFileTool creation
	readTool := NewReadFileTool(false, nil, nil)
	if readTool == nil {
		t.Errorf("Failed to create ReadFileTool")
	}

	// Test WriteFileTool creation
	writeTool := NewWriteFileTool(false, nil, nil)
	if writeTool == nil {
		t.Errorf("Failed to create WriteFileTool")
	}

	// Test ListFilesTool creation
	listTool := NewListFilesTool(false, nil, nil)
	if listTool == nil {
		t.Errorf("Failed to create ListFilesTool")
	}

	// Test EditTool creation
	editTool := NewEditTool(false, nil, nil)
	if editTool == nil {
		t.Errorf("Failed to create EditTool")
	}

	// Test that tools have proper methods
	ctx := context.Background()

	// All tools should implement Name()
	if readTool.Name() != "read_file" {
		t.Errorf("ReadFileTool.Name() = %s, expected read_file", readTool.Name())
	}
	if writeTool.Name() != "write_file" {
		t.Errorf("WriteFileTool.Name() = %s, expected write_file", writeTool.Name())
	}
	if listTool.Name() != "list_files" {
		t.Errorf("ListFilesTool.Name() = %s, expected list_files", listTool.Name())
	}
	if editTool.Name() != "edit" {
		t.Errorf("EditTool.Name() = %s, expected edit", editTool.Name())
	}

	// All tools should implement Info()
	info, err := readTool.Info(ctx)
	if err != nil || info == nil {
		t.Errorf("ReadFileTool.Info() failed: %v", err)
	}
	info, err = writeTool.Info(ctx)
	if err != nil || info == nil {
		t.Errorf("WriteFileTool.Info() failed: %v", err)
	}
	info, err = listTool.Info(ctx)
	if err != nil || info == nil {
		t.Errorf("ListFilesTool.Info() failed: %v", err)
	}
	info, err = editTool.Info(ctx)
	if err != nil || info == nil {
		t.Errorf("EditTool.Info() failed: %v", err)
	}
}
