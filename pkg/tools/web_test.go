package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func invokeWeb(t *testing.T, tool interface{ InvokeV(context.Context, string) (string, error) }, input interface{}) (string, error) {
	t.Helper()
	b, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal input: %v", err)
	}
	return tool.InvokeV(context.Background(), string(b))
}

func TestWebTool_WebFetch_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>Test Page</h1><p>Content here</p></body></html>"))
	}))
	defer server.Close()

	tool := NewWebFetchTool(50000)
	resultStr, err := invokeWeb(t, tool, WebFetchInput{URL: server.URL})

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	var out WebFetchOutput
	if jsonErr := json.Unmarshal([]byte(resultStr), &out); jsonErr != nil {
		t.Fatalf("Failed to parse result: %v", jsonErr)
	}

	if !strings.Contains(out.Text, "Test Page") {
		t.Errorf("Expected text to contain 'Test Page', got: %s", out.Text)
	}
	if out.Extractor == "" {
		t.Errorf("Expected non-empty extractor field")
	}
}

func TestWebTool_WebFetch_JSON(t *testing.T) {
	testData := map[string]string{"key": "value", "number": "123"}
	expectedJSON, _ := json.MarshalIndent(testData, "", "  ")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(expectedJSON)
	}))
	defer server.Close()

	tool := NewWebFetchTool(50000)
	resultStr, err := invokeWeb(t, tool, WebFetchInput{URL: server.URL})

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	var out WebFetchOutput
	json.Unmarshal([]byte(resultStr), &out)

	if !strings.Contains(out.Text, "key") && !strings.Contains(out.Text, "value") {
		t.Errorf("Expected text to contain JSON data, got: %s", out.Text)
	}
}

func TestWebTool_WebFetch_InvalidURL(t *testing.T) {
	tool := NewWebFetchTool(50000)
	_, err := invokeWeb(t, tool, WebFetchInput{URL: "not-a-valid-url"})

	if err == nil {
		t.Errorf("Expected error for invalid URL")
	}
	if !strings.Contains(err.Error(), "URL") && !strings.Contains(err.Error(), "url") {
		t.Errorf("Expected URL error message, got: %v", err)
	}
}

func TestWebTool_WebFetch_UnsupportedScheme(t *testing.T) {
	tool := NewWebFetchTool(50000)
	_, err := invokeWeb(t, tool, WebFetchInput{URL: "ftp://example.com/file.txt"})

	if err == nil {
		t.Errorf("Expected error for unsupported URL scheme")
	}
	if !strings.Contains(err.Error(), "http/https") {
		t.Errorf("Expected scheme error message, got: %v", err)
	}
}

func TestWebTool_WebFetch_MissingURL(t *testing.T) {
	tool := NewWebFetchTool(50000)
	_, err := invokeWeb(t, tool, WebFetchInput{})

	if err == nil {
		t.Errorf("Expected error when URL is missing")
	}
	if !strings.Contains(err.Error(), "url is required") {
		t.Errorf("Expected 'url is required' message, got: %v", err)
	}
}

func TestWebTool_WebFetch_Truncation(t *testing.T) {
	longContent := strings.Repeat("x", 20000)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(longContent))
	}))
	defer server.Close()

	tool := NewWebFetchTool(1000)
	resultStr, err := invokeWeb(t, tool, WebFetchInput{URL: server.URL})

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	var out WebFetchOutput
	if jsonErr := json.Unmarshal([]byte(resultStr), &out); jsonErr != nil {
		t.Fatalf("Failed to parse result: %v", jsonErr)
	}

	if len(out.Text) > 1100 {
		t.Errorf("Expected content to be truncated to ~1000 chars, got: %d", len(out.Text))
	}
	if !out.Truncated {
		t.Errorf("Expected 'truncated' to be true in result")
	}
}

func TestWebTool_WebSearch_NoApiKey(t *testing.T) {
	tool := NewWebSearchTool(WebSearchToolOptions{BraveEnabled: true, BraveAPIKey: ""})
	if tool != nil {
		t.Errorf("Expected nil tool when Brave API key is empty")
	}

	tool = NewWebSearchTool(WebSearchToolOptions{})
	if tool != nil {
		t.Errorf("Expected nil tool when no provider is enabled")
	}
}

func TestWebTool_WebSearch_MissingQuery(t *testing.T) {
	tool := NewWebSearchTool(WebSearchToolOptions{BraveEnabled: true, BraveAPIKey: "test-key", BraveMaxResults: 5})
	_, err := invokeWeb(t, tool, WebSearchInput{})

	if err == nil {
		t.Errorf("Expected error when query is missing")
	}
}

func TestWebTool_WebFetch_HTMLExtraction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html><body><script>alert('test');</script><style>body{color:red;}</style><h1>Title</h1><p>Content</p></body></html>`))
	}))
	defer server.Close()

	tool := NewWebFetchTool(50000)
	resultStr, err := invokeWeb(t, tool, WebFetchInput{URL: server.URL})

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	var out WebFetchOutput
	json.Unmarshal([]byte(resultStr), &out)

	if !strings.Contains(out.Text, "Title") && !strings.Contains(out.Text, "Content") {
		t.Errorf("Expected text to contain extracted content, got: %s", out.Text)
	}
	if strings.Contains(out.Text, "<script>") || strings.Contains(out.Text, "<style>") {
		t.Errorf("Expected script/style tags to be removed, got: %s", out.Text)
	}
}

func TestWebTool_WebFetch_MissingDomain(t *testing.T) {
	tool := NewWebFetchTool(50000)
	_, err := invokeWeb(t, tool, WebFetchInput{URL: "https://"})

	if err == nil {
		t.Errorf("Expected error for URL without domain")
	}
	if !strings.Contains(err.Error(), "domain") {
		t.Errorf("Expected domain error message, got: %v", err)
	}
}
