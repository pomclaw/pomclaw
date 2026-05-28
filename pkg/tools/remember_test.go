package tools

import (
	"context"
	"encoding/json"
	"testing"

	einotool "github.com/cloudwego/eino/components/tool"
)

// mockRememberer implements the Rememberer interface for testing.
type mockRememberer struct {
	lastText       string
	lastImportance float64
	lastCategory   string
	shouldFail     bool
}

func (m *mockRememberer) Remember(agentID string, text string, importance float64, category string) (string, error) {
	m.lastText = text
	m.lastImportance = importance
	m.lastCategory = category
	if m.shouldFail {
		return "", context.DeadlineExceeded
	}
	return "mem-abc1", nil
}

func invokeRemember(t *testing.T, tool einotool.InvokableTool, input RememberInput) (RememberOutput, error) {
	t.Helper()
	b, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal input: %v", err)
	}
	resultStr, invokeErr := tool.InvokableRun(context.Background(), string(b))
	if invokeErr != nil {
		return RememberOutput{}, invokeErr
	}
	var out RememberOutput
	if jsonErr := json.Unmarshal([]byte(resultStr), &out); jsonErr != nil {
		t.Fatalf("failed to parse remember output: %v", jsonErr)
	}
	return out, nil
}

func TestRememberTool_Info(t *testing.T) {
	tool := NewRememberTool(&mockRememberer{})
	info, err := tool.Info(context.Background())
	if err != nil {
		t.Fatalf("Info() error: %v", err)
	}
	if info.Name != "remember" {
		t.Errorf("Name = %q, want %q", info.Name, "remember")
	}
}

func TestRememberTool_ExecuteBasic(t *testing.T) {
	store := &mockRememberer{}
	tool := NewRememberTool(store)

	_, err := invokeRemember(t, tool, RememberInput{Text: "My favorite color is blue"})

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if store.lastText != "My favorite color is blue" {
		t.Errorf("stored text = %q", store.lastText)
	}
	if store.lastImportance != 0.7 {
		t.Errorf("default importance = %f, want 0.7", store.lastImportance)
	}
}

func TestRememberTool_ExecuteWithAllParams(t *testing.T) {
	store := &mockRememberer{}
	tool := NewRememberTool(store)

	_, err := invokeRemember(t, tool, RememberInput{
		Text:       "User prefers dark mode",
		Importance: 0.9,
		Category:   "preference",
	})

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if store.lastImportance != 0.9 {
		t.Errorf("importance = %f, want 0.9", store.lastImportance)
	}
	if store.lastCategory != "preference" {
		t.Errorf("category = %q, want %q", store.lastCategory, "preference")
	}
}

func TestRememberTool_ExecuteMissingText(t *testing.T) {
	tool := NewRememberTool(&mockRememberer{})

	_, err := invokeRemember(t, tool, RememberInput{})
	if err == nil {
		t.Error("should fail when text is missing")
	}
}

func TestRememberTool_ExecuteStoreFailure(t *testing.T) {
	store := &mockRememberer{shouldFail: true}
	tool := NewRememberTool(store)

	_, err := invokeRemember(t, tool, RememberInput{Text: "test"})
	if err == nil {
		t.Error("should report error when store fails")
	}
}

func TestRememberTool_ImportanceClamping(t *testing.T) {
	store := &mockRememberer{}
	tool := NewRememberTool(store)

	// Out of range importance should use default
	invokeRemember(t, tool, RememberInput{Text: "test", Importance: 1.5})
	if store.lastImportance != 0.7 {
		t.Errorf("out-of-range importance should fallback to 0.7, got %f", store.lastImportance)
	}

	invokeRemember(t, tool, RememberInput{Text: "test", Importance: -0.5})
	if store.lastImportance != 0.7 {
		t.Errorf("negative importance should fallback to 0.7, got %f", store.lastImportance)
	}
}
