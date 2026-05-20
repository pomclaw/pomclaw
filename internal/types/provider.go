package types

// ListProvidersReq - List providers request
type ListProvidersReq struct {
}

// CreateProviderReq - Create provider request
type CreateProviderReq struct {
	Name         string `json:"name"`
	ProviderType string `json:"provider_type"`
	APIBase      string `json:"api_base"`
	APIKey       string `json:"api_key"`
	DisplayName  string `json:"display_name"`
	Enabled      bool   `json:"enabled"`
}

// UpdateProviderReq - Update provider request
type UpdateProviderReq struct {
	Name        string `json:"name,omitempty"`
	APIBase     string `json:"api_base,omitempty"`
	APIKey      string `json:"api_key,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

// ProviderResp - Provider response
type ProviderResp struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ProviderType string `json:"provider_type"`
	APIBase      string `json:"api_base"`
	APIKey       string `json:"api_key"` // Will be masked in handler
	DisplayName  string `json:"display_name"`
	Enabled      bool   `json:"enabled"`
}

// ProvidersResp - List providers response
type ProvidersResp struct {
	Providers []ProviderResp `json:"providers"`
}

type ProviderModelsReq struct {
	Id string `path:"id"`
}

type ProviderModelsRes struct {
	Models interface{} `json:"models"`
}

// VerifyProviderReq - Verify provider request
type VerifyProviderReq struct {
	ID    string `path:"id"`
	Model string `json:"model"`
}

// VerifyProviderResp - Verify provider response
type VerifyProviderResp struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
	Note  string `json:"note,omitempty"`
}

// GetProviderReq - Get provider request
type GetProviderReq struct {
	ID string `path:"id"`
}

// GetProviderResp - Get provider response
type GetProviderResp struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ProviderType string `json:"provider_type"`
	APIBase      string `json:"api_base"`
	APIKey       string `json:"api_key"` // Will be masked in handler
	DisplayName  string `json:"display_name"`
	Enabled      bool   `json:"enabled"`
}
