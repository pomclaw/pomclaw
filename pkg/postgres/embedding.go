package postgres

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/pomclaw/pomclaw/pkg/logger"
)

// EmbeddingService handles text-to-vector embedding for PostgreSQL.
// PostgreSQL doesn't have built-in ONNX support, so it primarily uses API-based embedding.
type EmbeddingService struct {
	db         *sql.DB
	dims       int
	dimsOnce   sync.Once
	mode       string // "api" only for PostgreSQL
	apiBase    string
	apiKey     string
	apiModel   string
	httpClient *http.Client
}

// NewEmbeddingService creates a new EmbeddingService using API mode for PostgreSQL.
// PostgreSQL doesn't have built-in embedding, so API mode is required.
func NewEmbeddingService(db *sql.DB) *EmbeddingService {
	return &EmbeddingService{
		db:   db,
		dims: 384, // default embedding dimension
		mode: "api",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewAPIEmbeddingService creates an EmbeddingService that calls an external API.
// The API must be OpenAI-compatible (POST /v1/embeddings).
func NewAPIEmbeddingService(db *sql.DB, apiBase, apiKey, apiModel string) *EmbeddingService {
	return &EmbeddingService{
		db:       db,
		mode:     "api",
		apiBase:  apiBase,
		apiKey:   apiKey,
		apiModel: apiModel,
		dims:     0, // determined on first call
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// embeddingAPIRequest is the OpenAI-compatible request body.
type embeddingAPIRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// embeddingAPIResponse is the OpenAI-compatible response body.
type embeddingAPIResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// Embed generates an embedding vector for the given text.
func (es *EmbeddingService) Embed(text string) ([]float32, error) {
	return es.EmbedText(text)
}

// EmbedText generates an embedding vector for the given text (API mode).
func (es *EmbeddingService) EmbedText(text string) ([]float32, error) {
	if text == "" {
		if es.dims > 0 {
			return make([]float32, es.dims), nil
		}
		return make([]float32, 384), nil
	}

	// Truncate to a safe max input length
	if len(text) > 512 {
		text = text[:512]
	}

	return es.embedViaAPI(text)
}

// embedViaAPI calls the external embedding API.
func (es *EmbeddingService) embedViaAPI(text string) ([]float32, error) {
	if es.apiBase == "" || es.apiKey == "" || es.apiModel == "" {
		return nil, fmt.Errorf("API embedding not configured")
	}

	reqBody := embeddingAPIRequest{
		Model: es.apiModel,
		Input: text,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	req, err := http.NewRequest("POST", es.apiBase+"/v1/embeddings", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+es.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := es.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding API returned status %d: %s", resp.StatusCode, string(body))
	}

	var respBody embeddingAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("failed to decode embedding response: %w", err)
	}

	if len(respBody.Data) == 0 {
		return nil, fmt.Errorf("embedding API returned empty data")
	}

	emb := respBody.Data[0].Embedding
	if es.dims == 0 && len(emb) > 0 {
		es.dims = len(emb)
	}

	return emb, nil
}

// Mode returns the embedding mode ("api" for PostgreSQL).
func (es *EmbeddingService) Mode() string {
	return es.mode
}

// TestEmbedding tests the embedding service with a simple text.
func (es *EmbeddingService) TestEmbedding() bool {
	_, err := es.EmbedText("test")
	return err == nil
}

// CheckONNXLoaded returns true if ONNX model is loaded (not applicable for PostgreSQL).
func (es *EmbeddingService) CheckONNXLoaded() (bool, error) {
	return false, nil
}

// LoadONNXModel is a no-op for PostgreSQL (not supported).
func (es *EmbeddingService) LoadONNXModel(onnxDir, onnxFile string) error {
	logger.WarnC("postgres", "ONNX model loading not supported for PostgreSQL")
	return nil
}
