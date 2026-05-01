package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// mockRecaller implements the Recaller interface for testing.
type mockRecaller struct {
	results    []RecallResult
	shouldFail bool
	lastQuery  string
	lastMax    int
}

func (m *mockRecaller) Recall(agentID string, query string, maxResults int) ([]RecallResult, error) {
	m.lastQuery = query
	m.lastMax = maxResults
	if m.shouldFail {
		return nil, fmt.Errorf("recall failed")
	}
	return m.results, nil
}

func invokeRecall(t *testing.T, tool interface{ InvokeV(context.Context, string) (string, error) }, input RecallInput) (RecallOutput, error) {
	t.Helper()
	b, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal input: %v", err)
	}
	resultStr, invokeErr := tool.InvokeV(context.Background(), string(b))
	if invokeErr != nil {
		return RecallOutput{}, invokeErr
	}
	var out RecallOutput
	if jsonErr := json.Unmarshal([]byte(resultStr), &out); jsonErr != nil {
		t.Fatalf("failed to parse recall output: %v", jsonErr)
	}
	return out, nil
}

func TestRecallTool_Info(t *testing.T) {
	tool := NewRecallTool(&mockRecaller{})
	info, err := tool.Info(context.Background())
	if err != nil {
		t.Fatalf("Info() error: %v", err)
	}
	if info.Name != "recall" {
		t.Errorf("Name = %q, want %q", info.Name, "recall")
	}
}

func TestRecallTool_ExecuteWithResults(t *testing.T) {
	store := &mockRecaller{
		results: []RecallResult{
			{MemoryID: "mem-1", Text: "Favorite color is blue", Score: 0.95, Importance: 0.8, Category: "preference"},
			{MemoryID: "mem-2", Text: "Lives in Madrid", Score: 0.72, Importance: 0.6, Category: "fact"},
		},
	}
	tool := NewRecallTool(store)

	out, err := invokeRecall(t, tool, RecallInput{Query: "what color do they like"})

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if store.lastQuery != "what color do they like" {
		t.Errorf("query = %q", store.lastQuery)
	}
	if store.lastMax != 5 {
		t.Errorf("default max_results = %d, want 5", store.lastMax)
	}
	if !strings.Contains(out.Results, "2 matching memories") {
		t.Errorf("result should mention 2 matches: %s", out.Results)
	}
	if !strings.Contains(out.Results, "95%") {
		t.Errorf("result should contain score percentage: %s", out.Results)
	}
	if !strings.Contains(out.Results, "preference") {
		t.Errorf("result should contain category: %s", out.Results)
	}
}

func TestRecallTool_ExecuteNoResults(t *testing.T) {
	store := &mockRecaller{results: []RecallResult{}}
	tool := NewRecallTool(store)

	out, err := invokeRecall(t, tool, RecallInput{Query: "something obscure"})

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if !strings.Contains(out.Results, "No matching memories") {
		t.Errorf("should indicate no matches: %s", out.Results)
	}
}

func TestRecallTool_ExecuteMissingQuery(t *testing.T) {
	tool := NewRecallTool(&mockRecaller{})

	_, err := invokeRecall(t, tool, RecallInput{})
	if err == nil {
		t.Error("should fail when query is missing")
	}
}

func TestRecallTool_ExecuteCustomMaxResults(t *testing.T) {
	store := &mockRecaller{results: []RecallResult{}}
	tool := NewRecallTool(store)

	invokeRecall(t, tool, RecallInput{Query: "test", MaxResults: 10})

	if store.lastMax != 10 {
		t.Errorf("max_results = %d, want 10", store.lastMax)
	}
}

func TestRecallTool_ExecuteStoreFailure(t *testing.T) {
	store := &mockRecaller{shouldFail: true}
	tool := NewRecallTool(store)

	_, err := invokeRecall(t, tool, RecallInput{Query: "test"})
	if err == nil {
		t.Error("should report error when store fails")
	}
}
