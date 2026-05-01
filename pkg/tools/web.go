package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

const (
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

type SearchProvider interface {
	Search(ctx context.Context, query string, count int) (string, error)
}

type BraveSearchProvider struct {
	apiKey string
}

func (p *BraveSearchProvider) Search(ctx context.Context, query string, count int) (string, error) {
	searchURL := fmt.Sprintf("https://api.search.brave.com/res/v1/web/search?q=%s&count=%d",
		url.QueryEscape(query), count)

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", p.apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var searchResp struct {
		Web struct {
			Results []struct {
				Title       string `json:"title"`
				URL         string `json:"url"`
				Description string `json:"description"`
			} `json:"results"`
		} `json:"web"`
	}

	if err := json.Unmarshal(body, &searchResp); err != nil {
		fmt.Printf("Brave API Error Body: %s\n", string(body))
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	results := searchResp.Web.Results
	if len(results) == 0 {
		return fmt.Sprintf("No results for: %s", query), nil
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Results for: %s", query))
	for i, item := range results {
		if i >= count {
			break
		}
		lines = append(lines, fmt.Sprintf("%d. %s\n   %s", i+1, item.Title, item.URL))
		if item.Description != "" {
			lines = append(lines, fmt.Sprintf("   %s", item.Description))
		}
	}

	return strings.Join(lines, "\n"), nil
}

type DuckDuckGoSearchProvider struct{}

func (p *DuckDuckGoSearchProvider) Search(ctx context.Context, query string, count int) (string, error) {
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return p.extractResults(string(body), count, query)
}

func (p *DuckDuckGoSearchProvider) extractResults(html string, count int, query string) (string, error) {
	reLink := regexp.MustCompile(`<a[^>]*class="[^"]*result__a[^"]*"[^>]*href="([^"]+)"[^>]*>([\s\S]*?)</a>`)
	matches := reLink.FindAllStringSubmatch(html, count+5)

	if len(matches) == 0 {
		return fmt.Sprintf("No results found or extraction failed. Query: %s", query), nil
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Results for: %s (via DuckDuckGo)", query))

	reSnippet := regexp.MustCompile(`<a class="result__snippet[^"]*".*?>([\s\S]*?)</a>`)
	snippetMatches := reSnippet.FindAllStringSubmatch(html, count+5)

	maxItems := min(len(matches), count)

	for i := 0; i < maxItems; i++ {
		urlStr := matches[i][1]
		title := stripTags(matches[i][2])
		title = strings.TrimSpace(title)

		if strings.Contains(urlStr, "uddg=") {
			if u, err := url.QueryUnescape(urlStr); err == nil {
				idx := strings.Index(u, "uddg=")
				if idx != -1 {
					urlStr = u[idx+5:]
				}
			}
		}

		lines = append(lines, fmt.Sprintf("%d. %s\n   %s", i+1, title, urlStr))

		if i < len(snippetMatches) {
			snippet := stripTags(snippetMatches[i][1])
			snippet = strings.TrimSpace(snippet)
			if snippet != "" {
				lines = append(lines, fmt.Sprintf("   %s", snippet))
			}
		}
	}

	return strings.Join(lines, "\n"), nil
}

func stripTags(content string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	return re.ReplaceAllString(content, "")
}

type PerplexitySearchProvider struct {
	apiKey string
}

func (p *PerplexitySearchProvider) Search(ctx context.Context, query string, count int) (string, error) {
	searchURL := "https://api.perplexity.ai/chat/completions"

	payload := map[string]interface{}{
		"model": "sonar",
		"messages": []map[string]string{
			{"role": "system", "content": "You are a search assistant. Provide concise search results with titles, URLs, and brief descriptions in the following format:\n1. Title\n   URL\n   Description\n\nDo not add extra commentary."},
			{"role": "user", "content": fmt.Sprintf("Search for: %s. Provide up to %d relevant results.", query, count)},
		},
		"max_tokens": 1000,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", searchURL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Perplexity API error: %s", string(body))
	}

	var searchResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &searchResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(searchResp.Choices) == 0 {
		return fmt.Sprintf("No results for: %s", query), nil
	}

	return fmt.Sprintf("Results for: %s (via Perplexity)\n%s", query, searchResp.Choices[0].Message.Content), nil
}

type WebSearchToolOptions struct {
	BraveAPIKey          string
	BraveMaxResults      int
	BraveEnabled         bool
	DuckDuckGoMaxResults int
	DuckDuckGoEnabled    bool
	PerplexityAPIKey     string
	PerplexityMaxResults int
	PerplexityEnabled    bool
}

type WebSearchInput struct {
	Query string `json:"query"`
	Count int    `json:"count,omitempty"`
}

type WebSearchOutput struct {
	Result string `json:"result"`
}

func NewWebSearchTool(opts WebSearchToolOptions) tool.InvokableTool {
	var provider SearchProvider
	maxResults := 5

	// Priority: Perplexity > Brave > DuckDuckGo
	if opts.PerplexityEnabled && opts.PerplexityAPIKey != "" {
		provider = &PerplexitySearchProvider{apiKey: opts.PerplexityAPIKey}
		if opts.PerplexityMaxResults > 0 {
			maxResults = opts.PerplexityMaxResults
		}
	} else if opts.BraveEnabled && opts.BraveAPIKey != "" {
		provider = &BraveSearchProvider{apiKey: opts.BraveAPIKey}
		if opts.BraveMaxResults > 0 {
			maxResults = opts.BraveMaxResults
		}
	} else if opts.DuckDuckGoEnabled {
		provider = &DuckDuckGoSearchProvider{}
		if opts.DuckDuckGoMaxResults > 0 {
			maxResults = opts.DuckDuckGoMaxResults
		}
	} else {
		return nil
	}

	return utils.WrapInvokableToolWithErrorHandler(utils.NewTool[WebSearchInput, WebSearchOutput](
		&schema.ToolInfo{
			Name: "web_search",
			Desc: "Search the web for current information. Returns titles, URLs, and snippets from search results.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"query": {
					Type:     schema.String,
					Desc:     "Search query",
					Required: true,
				},
				"count": {
					Type: schema.Integer,
					Desc: "Number of results (1-10)",
				},
			}),
		},
		func(ctx context.Context, input WebSearchInput) (WebSearchOutput, error) {
			if input.Query == "" {
				return WebSearchOutput{}, fmt.Errorf("query is required")
			}

			count := maxResults
			if input.Count > 0 && input.Count <= 10 {
				count = input.Count
			}

			result, err := provider.Search(ctx, input.Query, count)
			if err != nil {
				return WebSearchOutput{}, fmt.Errorf("search failed: %w", err)
			}

			return WebSearchOutput{Result: result}, nil
		},
	), func(ctx context.Context, err error) string { return err.Error() })
}

type WebFetchInput struct {
	URL      string `json:"url"`
	MaxChars int    `json:"max_chars,omitempty"`
}

type WebFetchOutput struct {
	URL       string `json:"url"`
	Status    int    `json:"status"`
	Extractor string `json:"extractor"`
	Truncated bool   `json:"truncated"`
	Length    int    `json:"length"`
	Text      string `json:"text"`
}

func NewWebFetchTool(maxChars int) tool.InvokableTool {
	if maxChars <= 0 {
		maxChars = 50000
	}

	return utils.WrapInvokableToolWithErrorHandler(utils.NewTool[WebFetchInput, WebFetchOutput](
		&schema.ToolInfo{
			Name: "web_fetch",
			Desc: "Fetch a URL and extract readable content (HTML to text). Use this to get weather info, news, articles, or any web content.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"url": {
					Type:     schema.String,
					Desc:     "URL to fetch",
					Required: true,
				},
				"max_chars": {
					Type: schema.Integer,
					Desc: "Maximum characters to extract (minimum 100)",
				},
			}),
		},
		func(ctx context.Context, input WebFetchInput) (WebFetchOutput, error) {
			if input.URL == "" {
				return WebFetchOutput{}, fmt.Errorf("url is required")
			}

			parsedURL, err := url.Parse(input.URL)
			if err != nil {
				return WebFetchOutput{}, fmt.Errorf("invalid URL: %w", err)
			}

			if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
				return WebFetchOutput{}, fmt.Errorf("only http/https URLs are allowed")
			}

			if parsedURL.Host == "" {
				return WebFetchOutput{}, fmt.Errorf("missing domain in URL")
			}

			limit := maxChars
			if input.MaxChars > 100 {
				limit = input.MaxChars
			}

			req, err := http.NewRequestWithContext(ctx, "GET", input.URL, nil)
			if err != nil {
				return WebFetchOutput{}, fmt.Errorf("failed to create request: %w", err)
			}

			req.Header.Set("User-Agent", userAgent)

			client := &http.Client{
				Timeout: 60 * time.Second,
				Transport: &http.Transport{
					MaxIdleConns:        10,
					IdleConnTimeout:     30 * time.Second,
					DisableCompression:  false,
					TLSHandshakeTimeout: 15 * time.Second,
				},
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					if len(via) >= 5 {
						return fmt.Errorf("stopped after 5 redirects")
					}
					return nil
				},
			}

			resp, err := client.Do(req)
			if err != nil {
				return WebFetchOutput{}, fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return WebFetchOutput{}, fmt.Errorf("failed to read response: %w", err)
			}

			contentType := resp.Header.Get("Content-Type")

			var text, extractor string

			if strings.Contains(contentType, "application/json") {
				var jsonData interface{}
				if err := json.Unmarshal(body, &jsonData); err == nil {
					formatted, _ := json.MarshalIndent(jsonData, "", "  ")
					text = string(formatted)
					extractor = "json"
				} else {
					text = string(body)
					extractor = "raw"
				}
			} else if strings.Contains(contentType, "text/html") || len(body) > 0 &&
				(strings.HasPrefix(string(body), "<!DOCTYPE") || strings.HasPrefix(strings.ToLower(string(body)), "<html")) {
				text = extractText(string(body))
				extractor = "text"
			} else {
				text = string(body)
				extractor = "raw"
			}

			truncated := len(text) > limit
			if truncated {
				text = text[:limit]
			}

			return WebFetchOutput{
				URL:       input.URL,
				Status:    resp.StatusCode,
				Extractor: extractor,
				Truncated: truncated,
				Length:    len(text),
				Text:      text,
			}, nil
		},
	), func(ctx context.Context, err error) string { return err.Error() })
}

func extractText(htmlContent string) string {
	re := regexp.MustCompile(`<script[\s\S]*?</script>`)
	result := re.ReplaceAllLiteralString(htmlContent, "")
	re = regexp.MustCompile(`<style[\s\S]*?</style>`)
	result = re.ReplaceAllLiteralString(result, "")
	re = regexp.MustCompile(`<[^>]+>`)
	result = re.ReplaceAllLiteralString(result, "")

	result = strings.TrimSpace(result)

	re = regexp.MustCompile(`\s+`)
	result = re.ReplaceAllLiteralString(result, " ")

	lines := strings.Split(result, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, "\n")
}
