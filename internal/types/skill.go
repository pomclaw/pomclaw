package types

// GetSkillReq - Get skill request
type GetSkillReq struct {
	ID string `path:"id"`
}

// CreateSkillReq - Create skill request
type CreateSkillReq struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

// GrantSkillReq - Grant skill to agent request
type GrantSkillReq struct {
	ID      string `path:"id"`
	AgentID string `path:"agent_id"`
	Version int    `json:"version,omitempty"`
}

// RevokeSkillReq - Revoke skill from agent request
type RevokeSkillReq struct {
	ID      string `path:"id"`
	AgentID string `path:"agent_id"`
}

// UpdateSkillReq - Update skill request
type UpdateSkillReq struct {
	ID          string  `path:"id"`
	Enabled     *bool   `json:"enabled,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
}

// SkillResp - Skill response
type SkillResp struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Description string   `json:"description"`
	Enabled     bool     `json:"enabled"`
	Status      string   `json:"status"`
	Version     int      `json:"version"`
	IsSystem    bool     `json:"is_system,omitempty"`
	Source      string   `json:"source,omitempty"`
	Visibility  string   `json:"visibility,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	MissingDeps []string `json:"missing_deps,omitempty"`
	Author      string   `json:"author,omitempty"`
}

// SkillWithGrantResp - Skill with grant status
type SkillWithGrantResp struct {
	*SkillResp
	Granted bool `json:"granted"`
}

// SkillsResp - List skills response
type SkillsResp struct {
	Skills []SkillResp `json:"skills"`
}

// SkillsWithGrantResp - List skills with grant status response
type SkillsWithGrantResp struct {
	Skills []SkillWithGrantResp `json:"skills"`
}
