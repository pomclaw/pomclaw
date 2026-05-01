package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func invokeExec(t *testing.T, tool interface{ InvokeV(context.Context, string) (string, error) }, ctx context.Context, input ExecInput) (ExecOutput, error) {
	t.Helper()
	b, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal input: %v", err)
	}
	resultStr, invokeErr := tool.InvokeV(ctx, string(b))
	if invokeErr != nil {
		return ExecOutput{}, invokeErr
	}
	var out ExecOutput
	if jsonErr := json.Unmarshal([]byte(resultStr), &out); jsonErr != nil {
		t.Fatalf("failed to parse exec output: %v", jsonErr)
	}
	return out, nil
}

func TestShellTool_Success(t *testing.T) {
	tool := NewExecTool(false)
	ctx := context.Background()

	out, err := invokeExec(t, tool, ctx, ExecInput{Command: "echo 'hello world'"})

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if !strings.Contains(out.Output, "hello world") {
		t.Errorf("Expected output to contain 'hello world', got: %s", out.Output)
	}
}

func TestShellTool_Failure(t *testing.T) {
	tool := NewExecTool(false)
	ctx := context.Background()

	out, err := invokeExec(t, tool, ctx, ExecInput{Command: "ls /nonexistent_directory_12345"})

	// Non-zero exit returns output (not error), with exit code in output
	if err != nil {
		t.Errorf("Expected non-zero exit to return output (not error), got: %v", err)
	}
	if out.Output == "" {
		t.Errorf("Expected output for failed command, got empty string")
	}
}

func TestShellTool_Timeout(t *testing.T) {
	tool := NewExecTool(false)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := invokeExec(t, tool, ctx, ExecInput{Command: "sleep 10"})

	if err == nil {
		t.Errorf("Expected error for timeout")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Expected timeout message, got: %v", err)
	}
}

func TestShellTool_WorkingDir(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	tool := NewExecTool(false)
	ctx := context.Background()

	out, err := invokeExec(t, tool, ctx, ExecInput{Command: "cat test.txt", WorkingDir: tmpDir})

	if err != nil {
		t.Errorf("Expected success in custom working dir, got error: %v", err)
	}
	if !strings.Contains(out.Output, "test content") {
		t.Errorf("Expected output from custom dir, got: %s", out.Output)
	}
}

func TestShellTool_DangerousCommand(t *testing.T) {
	tool := NewExecTool(false)
	ctx := context.Background()

	_, err := invokeExec(t, tool, ctx, ExecInput{Command: "rm -rf /"})

	if err == nil {
		t.Errorf("Expected dangerous command to be blocked")
	}
	if !strings.Contains(err.Error(), "blocked") {
		t.Errorf("Expected 'blocked' message, got: %v", err)
	}
}

func TestShellTool_MissingCommand(t *testing.T) {
	tool := NewExecTool(false)
	ctx := context.Background()

	_, err := invokeExec(t, tool, ctx, ExecInput{})

	if err == nil {
		t.Errorf("Expected error when command is missing")
	}
}

func TestShellTool_StderrCapture(t *testing.T) {
	tool := NewExecTool(false)
	ctx := context.Background()

	out, err := invokeExec(t, tool, ctx, ExecInput{Command: "sh -c 'echo stdout; echo stderr >&2'"})

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if !strings.Contains(out.Output, "stdout") {
		t.Errorf("Expected stdout in output, got: %s", out.Output)
	}
	if !strings.Contains(out.Output, "stderr") {
		t.Errorf("Expected stderr in output, got: %s", out.Output)
	}
}

func TestShellTool_OutputTruncation(t *testing.T) {
	tool := NewExecTool(false)
	ctx := context.Background()

	out, err := invokeExec(t, tool, ctx, ExecInput{
		Command: "echo " + strings.Repeat("x", 20000),
	})

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if len(out.Output) > 15000 {
		t.Errorf("Expected output to be truncated, got length: %d", len(out.Output))
	}
}

func TestShellTool_RestrictToWorkspace(t *testing.T) {
	tool := NewExecTool(true) // restrict=true
	ctx := context.Background()

	_, err := invokeExec(t, tool, ctx, ExecInput{Command: "cat ../../etc/passwd"})

	if err == nil {
		t.Errorf("Expected path traversal to be blocked with restrict=true")
	}
	if !strings.Contains(err.Error(), "blocked") {
		t.Errorf("Expected 'blocked' message for path traversal, got: %v", err)
	}
}
