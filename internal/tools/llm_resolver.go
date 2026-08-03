package tools

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/openai"
	model2 "github.com/cloudwego/eino/components/model"
	"github.com/pomclaw/pomclaw/internal/model"
)

// ResolveLLM creates an LLM client from provider credentials.
// Supports OpenAI-compatible providers (OpenAI, OpenRouter, Groq, etc.)
func ResolveLLM(ctx context.Context, provider *model.Providers, modelName string) (model2.ChatModel, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider is nil")
	}

	if !provider.Enabled {
		return nil, fmt.Errorf("provider '%s' is disabled", provider.Name)
	}

	// Extract API credentials
	apiKey := provider.ApiKey
	if apiKey == "" {
		return nil, fmt.Errorf("provider '%s' has no API key configured", provider.Name)
	}

	baseURL := ""
	if provider.ApiBase.Valid {
		baseURL = provider.ApiBase.String
	}

	// Initialize based on provider type
	//switch provider.ProviderType {
	//case "openai", "openrouter", "groq", "anthropic":
	// All OpenAI-compatible providers use the same initialization
	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   modelName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LLM for provider '%s': %w", provider.Name, err)
	}
	return llm, nil

	//default:
	//	return nil, fmt.Errorf("unsupported provider type: %s", provider.ProviderType)
	//}
}
