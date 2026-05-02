package config

import (
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf

	Auth     AuthConfig
	RocketMQ RocketMQConfig
	Postgres PostgresDBConfig `json:"postgres"`

	Agents      AgentsConfig    `json:"agents"`
	Channels    ChannelsConfig  `json:"channels"`
	Providers   ProvidersConfig `json:"providers"`
	Tools       ToolsConfig     `json:"tools,optional"`
	Heartbeat   HeartbeatConfig `json:"heartbeat,optional"`
	Devices     DevicesConfig   `json:"devices,optional"`
	Gateway     GatewayConfig   `json:"gateway,optional"`
	StorageType string          `json:"storage_type"` // "postgres"
}

type RocketMQConfig struct {
	Enable       bool
	NameSrv      string
	AccessKey    string
	SecretKey    string
	Topic        string
	GroupName    string
	InstanceName string
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
	Defaults      AgentDefaults `json:"defaults"`
	BaseWorkspace string        `json:"base_workspace"`
}

type AgentDefaults struct {
	RestrictToWorkspace       bool           `json:"restrict_to_workspace"`
	Provider                  string         `json:"provider"`
	Model                     string         `json:"model"`
	MaxTokens                 int            `json:"max_tokens"`
	Temperature               float64        `json:"temperature"`
	MaxToolIterations         int            `json:"max_tool_iterations"`
	SummarizeMessageThreshold int            `json:"summarize_message_threshold"`
	SummarizeTokenPercent     int            `json:"summarize_token_percent"`
	Routing                   *RoutingConfig `json:"routing,optional"`
	UseEino                   bool           `json:"use_eino,optional"` // Use eino framework for agent loop (default: false)
}

type RoutingConfig struct {
	Enabled    bool    `json:"enabled"`
	LightModel string  `json:"light_model"`
	Threshold  float64 `json:"threshold"`
}

type ChannelsConfig struct {
	WhatsApp   WhatsAppConfig   `json:"whatsapp,optional"`
	Telegram   TelegramConfig   `json:"telegram,optional"`
	Feishu     FeishuConfig     `json:"feishu,optional"`
	Discord    DiscordConfig    `json:"discord,optional"`
	MaixCam    MaixCamConfig    `json:"maixcam,optional"`
	QQ         QQConfig         `json:"qq,optional"`
	DingTalk   DingTalkConfig   `json:"dingtalk,optional"`
	Slack      SlackConfig      `json:"slack,optional"`
	LINE       LINEConfig       `json:"line,optional"`
	OneBot     OneBotConfig     `json:"onebot,optional"`
	Mattermost MattermostConfig `json:"mattermost,optional"`
}

type WhatsAppConfig struct {
	Enabled   bool                `json:"enabled"`
	BridgeURL string              `json:"bridge_url"`
	AllowFrom FlexibleStringSlice `json:"allow_from"`
}

type TelegramConfig struct {
	Enabled   bool                `json:"enabled"`
	Token     string              `json:"token"`
	Proxy     string              `json:"proxy"`
	AllowFrom FlexibleStringSlice `json:"allow_from"`
}

type FeishuConfig struct {
	Enabled           bool                `json:"enabled"`
	AppID             string              `json:"app_id"`
	AppSecret         string              `json:"app_secret"`
	EncryptKey        string              `json:"encrypt_key"`
	VerificationToken string              `json:"verification_token"`
	AllowFrom         FlexibleStringSlice `json:"allow_from"`
}

type DiscordConfig struct {
	Enabled   bool                `json:"enabled"`
	Token     string              `json:"token"`
	AllowFrom FlexibleStringSlice `json:"allow_from"`
}

type MaixCamConfig struct {
	Enabled   bool                `json:"enabled"`
	Host      string              `json:"host"`
	Port      int                 `json:"port"`
	AllowFrom FlexibleStringSlice `json:"allow_from"`
}

type QQConfig struct {
	Enabled   bool                `json:"enabled"`
	AppID     string              `json:"app_id"`
	AppSecret string              `json:"app_secret"`
	AllowFrom FlexibleStringSlice `json:"allow_from"`
}

type DingTalkConfig struct {
	Enabled      bool                `json:"enabled"`
	ClientID     string              `json:"client_id"`
	ClientSecret string              `json:"client_secret"`
	AllowFrom    FlexibleStringSlice `json:"allow_from"`
}

type SlackConfig struct {
	Enabled   bool                `json:"enabled"`
	BotToken  string              `json:"bot_token"`
	AppToken  string              `json:"app_token"`
	AllowFrom FlexibleStringSlice `json:"allow_from"`
}

type LINEConfig struct {
	Enabled            bool                `json:"enabled"`
	ChannelSecret      string              `json:"channel_secret"`
	ChannelAccessToken string              `json:"channel_access_token"`
	WebhookHost        string              `json:"webhook_host"`
	WebhookPort        int                 `json:"webhook_port"`
	WebhookPath        string              `json:"webhook_path"`
	AllowFrom          FlexibleStringSlice `json:"allow_from"`
}

type OneBotConfig struct {
	Enabled            bool                `json:"enabled"`
	WSUrl              string              `json:"ws_url"`
	AccessToken        string              `json:"access_token"`
	ReconnectInterval  int                 `json:"reconnect_interval"`
	GroupTriggerPrefix []string            `json:"group_trigger_prefix"`
	AllowFrom          FlexibleStringSlice `json:"allow_from"`
}

type MattermostConfig struct {
	Enabled   bool                `json:"enabled"`
	ServerURL string              `json:"server_url"`
	Token     string              `json:"token"`
	AllowFrom FlexibleStringSlice `json:"allow_from"`
}

type MySQLDBConfig struct {
	Enabled     bool   `json:"enabled"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Database    string `json:"database"`
	User        string `json:"user"`
	Password    string `json:"password"`
	PoolMaxOpen int    `json:"pool_max_open"`
	PoolMaxIdle int    `json:"pool_max_idle"`
}

type HeartbeatConfig struct {
	Enabled  bool `json:"enabled"`
	Interval int  `json:"interval"` // minutes, min 5
}

type DevicesConfig struct {
	Enabled    bool `json:"enabled"`
	MonitorUSB bool `json:"monitor_usb"`
}

type LoggingConfig struct {
	Level    string `json:"level"`              // DEBUG, INFO, WARN, ERROR, FATAL (default: INFO)
	FilePath string `json:"file_path,optional"` // Optional: path to log file for file logging
}

type PostgresDBConfig struct {
	Enabled           bool   `json:"enabled"`
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

type ProvidersConfig struct {
	Anthropic     ProviderConfig `json:"anthropic,optional"`
	OpenAI        ProviderConfig `json:"openai,optional"`
	OpenRouter    ProviderConfig `json:"openrouter,optional"`
	Groq          ProviderConfig `json:"groq,optional"`
	VLLM          ProviderConfig `json:"vllm,optional"`
	Gemini        ProviderConfig `json:"gemini,optional"`
	Nvidia        ProviderConfig `json:"nvidia,optional"`
	Ollama        ProviderConfig `json:"ollama,optional"`
	Moonshot      ProviderConfig `json:"moonshot,optional"`
	DeepSeek      ProviderConfig `json:"deepseek,optional"`
	GitHubCopilot ProviderConfig `json:"github_copilot,optional"`
}

type ProviderConfig struct {
	APIKey      string `json:"api_key"`
	APIBase     string `json:"api_base"`
	Proxy       string `json:"proxy,optional"`
	AuthMethod  string `json:"auth_method,optional"`
	ConnectMode string `json:"connect_mode,optional"` //only for Github Copilot, `stdio` or `grpc`
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
