package types

import "encoding/json"

// ============ Builtin Tool Types ============

// BuiltinToolDef - Built-in tool definition
type BuiltinToolDef struct {
	Name     string          `json:"name"`
	Display  string          `json:"display,omitempty"`
	Desc     string          `json:"desc,omitempty"`
	Enabled  bool            `json:"enabled"`
	Settings json.RawMessage `json:"settings,omitempty"`
}

// ListBuiltinToolsResp - List all built-in tools response
type ListBuiltinToolsResp struct {
	Tools []BuiltinToolDef `json:"tools"`
}

// GetBuiltinToolReq - Get built-in tool request
type GetBuiltinToolReq struct {
	Name string `path:"name"`
}

// UpdateBuiltinToolReq - Update built-in tool request
type UpdateBuiltinToolReq struct {
	Name     string          `path:"name"`
	Enabled  *bool           `json:"enabled,optional"`
	Settings json.RawMessage `json:"settings,optional"`
}

// UpdateBuiltinToolResp - Update built-in tool response
type UpdateBuiltinToolResp struct {
	Status string `json:"status"`
}

// GetTenantConfigReq - Get tenant configuration request
type GetTenantConfigReq struct {
	Name string `path:"name"`
}

// TenantToolConfig - Tenant-specific tool configuration
type TenantToolConfig struct {
	ToolName string          `json:"tool_name"`
	Enabled  *bool           `json:"enabled,omitempty"`
	Settings json.RawMessage `json:"settings,omitempty"`
}

// SetTenantConfigReq - Set tenant configuration request
type SetTenantConfigReq struct {
	Name     string          `path:"name"`
	Enabled  *bool           `json:"enabled,optional"`
	Settings json.RawMessage `json:"settings,optional"`
}

// SetTenantConfigResp - Set tenant configuration response
type SetTenantConfigResp struct {
	Status string `json:"status"`
}

// DeleteTenantConfigReq - Delete tenant configuration request
type DeleteTenantConfigReq struct {
	Name string `path:"name"`
}

// DeleteTenantConfigResp - Delete tenant configuration response
type DeleteTenantConfigResp struct {
	Status string `json:"status"`
}
