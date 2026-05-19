package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"net/http"
	"strings"
)

// ModelInfo represents a model entry returned by the list-models endpoint
type ModelInfo struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type ListProviderModelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListProviderModelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProviderModelsLogic {
	return &ListProviderModelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListProviderModelsLogic) ListProviderModels(req types.ProviderModelsReq) (types.ProviderModelsRes, error) {
	p, err := l.svcCtx.ProvidersModel.FindOne(l.ctx, req.Id)
	if err == model.ErrNotFound || (err == nil && p.UserId != req.UserID) {
		logx.Errorf("ListProviderModels: provider not found")
		return types.ProviderModelsRes{}, model.ErrNotFound
	}
	if err != nil {
		logx.Errorf("ListProviderModels failed: %v", err)
		return types.ProviderModelsRes{}, err
	}

	return types.ProviderModelsRes{Models: nil}, nil
}

func fetchAnthropicModels(ctx context.Context, apiKey string) ([]ModelInfo, error) {
	base := "https://api.anthropic.com/v1"
	req, err := http.NewRequestWithContext(ctx, "GET", base+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("anthropic API returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			ID          string `json:"id"`
			DisplayName string `json:"display_name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode anthropic response: %w", err)
	}

	models := make([]ModelInfo, 0, len(result.Data))
	for _, m := range result.Data {
		models = append(models, ModelInfo{ID: m.ID, Name: m.DisplayName})
	}
	return models, nil
}

func fetchGeminiModels(ctx context.Context, apiKey string) ([]ModelInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://generativelanguage.googleapis.com/v1beta/models?key="+apiKey, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("gemini API returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Models []struct {
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode gemini response: %w", err)
	}

	models := make([]ModelInfo, 0, len(result.Models))
	for _, m := range result.Models {
		id := strings.TrimPrefix(m.Name, "models/")
		models = append(models, ModelInfo{ID: id, Name: m.DisplayName})
	}
	return models, nil
}

func fetchOpenAIModels(ctx context.Context, apiBase, apiKey string) ([]ModelInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", apiBase+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("provider API returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode provider response: %w", err)
	}

	models := make([]ModelInfo, 0, len(result.Data))
	for _, m := range result.Data {
		models = append(models, ModelInfo{ID: m.ID, Name: m.ID})
	}
	return models, nil
}
