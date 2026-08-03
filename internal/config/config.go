package config

import (
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf

	Auth     AuthConfig
	Postgres PostgresDBConfig `json:"postgres"`
	Agents   AgentsConfig     `json:"agents"`
	Tools    ToolsConfig      `json:"tools,optional"`
	Oss      OssConfig        `json:"oss,optional"`
	Redis    redis.RedisConf  `json:"redis,optional"`
}

type OssConfig struct {
	Endpoint        string `json:"Endpoint"`
	AccessKeyId     string `json:"AccessKeyId"`
	AccessKeySecret string `json:"AccessKeySecret"`
	BucketName      string `json:"BucketName"`
	Directory       string `json:"Directory"`
	OssDomain       string `json:"OssDomain"`
}

type LLMConfig struct {
	APIKey         string
	BaseURL        string
	Model          string
	EmbeddingModel string //: "text-embedding-ada-002"
}

// FlexibleStringSlice is a []string that also accepts JSON numbers,
// so allow_from can contain both "123" and 123.
type FlexibleStringSlice []string

func (f *FlexibleStringSlice) UnmarshalJSON(data []byte) error {
	// Try []string first
	var ss []string
	if err := json.Unmarshal(data, &ss); err == nil {
		*f = ss
		return nil
	}

	// Try []interface{} to handle mixed types
	var raw []interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	result := make([]string, 0, len(raw))
	for _, v := range raw {
		switch val := v.(type) {
		case string:
			result = append(result, val)
		case float64:
			result = append(result, fmt.Sprintf("%.0f", val))
		default:
			result = append(result, fmt.Sprintf("%v", val))
		}
	}
	*f = result
	return nil
}

type AgentsConfig struct {
	Defaults AgentDefaults `json:"defaults"`
}

type AgentDefaults struct {
	RestrictToWorkspace       bool    `json:"restrict_to_workspace"`
	MaxTokens                 int     `json:"max_tokens"`
	Temperature               float64 `json:"temperature"`
	MaxToolIterations         int     `json:"max_tool_iterations"`
	SummarizeMessageThreshold int     `json:"summarize_message_threshold"`
	SummarizeTokenPercent     int     `json:"summarize_token_percent"`
}

type RoutingConfig struct {
	Enabled    bool    `json:"enabled"`
	LightModel string  `json:"light_model"`
	Threshold  float64 `json:"threshold"`
}

type PostgresDBConfig struct {
	Host              string `json:"host"`
	Port              int    `json:"port"`
	Database          string `json:"database"`
	User              string `json:"user"`
	Password          string `json:"password"`
	SSLMode           string `json:"ssl_mode"` // "disable", "require", "verify-full"
	PoolMaxOpen       int    `json:"pool_max_open"`
	PoolMaxIdle       int    `json:"pool_max_idle"`
	EmbeddingProvider string `json:"embedding_provider"` // "api" or "local"
	EmbeddingAPIBase  string `json:"embedding_api_base"`
	EmbeddingAPIKey   string `json:"embedding_api_key"`
	EmbeddingModel    string `json:"embedding_model"`
}

type BraveConfig struct {
	Enabled    bool   `json:"enabled"`
	APIKey     string `json:"api_key"`
	MaxResults int    `json:"max_results"`
}

type DuckDuckGoConfig struct {
	Enabled    bool `json:"enabled"`
	MaxResults int  `json:"max_results"`
}

type PerplexityConfig struct {
	Enabled    bool   `json:"enabled"`
	APIKey     string `json:"api_key"`
	MaxResults int    `json:"max_results"`
}

type WebToolsConfig struct {
	Brave      BraveConfig      `json:"brave"`
	DuckDuckGo DuckDuckGoConfig `json:"duckduckgo"`
	Perplexity PerplexityConfig `json:"perplexity"`
}

type CronToolsConfig struct {
	ExecTimeoutMinutes int `json:"exec_timeout_minutes"` // 0 means no timeout
}

type ToolsConfig struct {
	Web  WebToolsConfig  `json:"web,optional"`
	Cron CronToolsConfig `json:"cron,optional"`
}

type AuthConfig struct {
	AccessSecret  string
	AccessExpire  int64
	RefreshExpire int64 `json:",optional"`
}


type GatewayConfig struct {
	Host string `json:"host"` // Default: "0.0.0.0"
	Port int    `json:"port"` // Default: 8080
}
