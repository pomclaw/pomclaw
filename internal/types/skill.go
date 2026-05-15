package types

// CreateSkillReq - Create skill request
type CreateSkillReq struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

// GrantSkillReq - Grant skill to agent request
type GrantSkillReq struct {
	AgentID string `json:"agent_id"`
	Version int    `json:"version"`
}

// SkillResp - Skill response
type SkillResp struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Status      string `json:"status"`
	Version     int    `json:"version"`
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
