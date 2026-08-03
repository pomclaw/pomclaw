package contracts

import (
	"context"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	MetadataKey_AgentId   = "agent_id"
	MetadataKey_Workspace = "workspace"

	DefaultAgentID   = "default"
	DefaultWorkspace = "workspace"
)

// PromptStoreInterface is an optional interface for Oracle-backed prompt storage.
type PromptStoreInterface interface {
	LoadBootstrapFiles(agentID string) map[string]string
}

// MemoryStoreInterface defines the contract for memory storage backends.
// Both file-based (MemoryStore) and Oracle-backed implementations satisfy this.
type MemoryStoreInterface interface {
	ReadLongTerm(agentID string) string
	WriteLongTerm(agentID string, content string) error
	GetMemoryContext(agentID string) string
}

// SessionManagerInterface defines the contract for session management backends.
type SessionManagerInterface interface {
	AddMessage(agentID string, sessionId int64, role schema.RoleType, content string)
	AddFullMessage(agentID string, sessionId int64, msg schema.Message)
	GetHistory(agentID string, sessionId int64) []schema.Message
	SetHistory(agentID string, sessionId int64, history []schema.Message)
	GetSummary(agentID string, sessionId int64) string
	SetSummary(agentID string, sessionId int64, summary string)
	TruncateHistory(agentID string, sessionId int64, keepLast int)
	Save(agentID string, sessionId int64, userID string) error
}

type ToolsManagerInterface interface {
	GetToolsToolDef(ctx context.Context, userId, agentID string) []ToolDef
	GetTools(ctx context.Context, userId, agentID string) compose.ToolsNodeConfig
}

// SqlMemoryStore is an extended interface for Oracle-backed memory with vector search.
type SqlMemoryStore interface {
	MemoryStoreInterface
	Remember(agentID string, text string, importance float64, category string) (string, error)
	Recall(agentID string, query string, maxResults int) ([]MemoryRecallResult, error)
	Forget(agentID string, memoryID int64) error
}

// MemoryRecallResult represents a single recalled memory with similarity score.
type MemoryRecallResult struct {
	MemoryID   string  `json:"memory_id"`
	Text       string  `json:"text"`
	Importance float64 `json:"importance"`
	Category   string  `json:"category"`
	Score      float64 `json:"score"`
}

// --- Skills Management Interfaces ---

// SkillInfo represents metadata about an installed skill.
type SkillInfo struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Source      string `json:"source"` // "workspace", "global", or "builtin"
	Description string `json:"description"`
}

// AvailableSkill represents a skill available from the remote registry.
type AvailableSkill struct {
	Name        string   `json:"name"`
	Repository  string   `json:"repository"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Tags        []string `json:"tags"`
}

type ToolDef struct {
	Name    string `json:"name"`
	Display string `json:"display,optional"`
	Desc    string `json:"desc,optional"`
	Enabled bool   `json:"enabled"`
}

// SkillsLoaderInterface defines the contract for loading skills.
// Skills are stored as zip packages in OSS and addressed by the loader's
// bound agentID/userID; no workspace scoping is needed.
type SkillsLoaderInterface interface {
	// ListSkills returns all skills granted to the loader's agent.
	ListSkills() []SkillInfo

	// LoadSkill loads the content of a single skill by name.
	// Returns the skill content (with frontmatter stripped) and true if found, or "" and false if not found.
	LoadSkill(name string) (string, bool)

	// BuildSkillsSummary generates an XML-formatted summary of all available skills.
	// Used for including in system prompts.
	BuildSkillsSummary() string

	// GetSkillFile reads a single file from a skill's OSS zip by relative path.
	// filePath is relative to the skill root (e.g. "references/image-rules.md",
	// "scripts/image_client.py"); entries with or without a top-level directory
	// prefix are matched automatically.
	// Returns (bytes, true, nil) on success; (nil, false, nil) if the skill or
	// file is not found; (nil, false, err) on storage/zip errors.
	GetSkillFile(name string, filePath string) ([]byte, bool, error)
}

// SkillInstallerInterface defines the contract for installing and managing skills.
type SkillInstallerInterface interface {
	// InstallFromGitHub installs a skill from GitHub.
	// repo format: "owner/repo" or "owner/repo/subdir"
	InstallFromGitHub(ctx context.Context, repo string) error

	// Uninstall removes an installed skill by name.
	Uninstall(skillName string) error

	// ListAvailableSkills fetches the list of skills available from the remote registry.
	ListAvailableSkills(ctx context.Context) ([]AvailableSkill, error)
}

// ContextBuilderInterface defines the contract for building agent context.
// Implementations should handle system prompt assembly, message construction,
// memory management, and skill loading.
type ContextBuilderInterface interface {
	// BuildSystemPrompt assembles the system prompt from identity, bootstrap files, skills, and memory context.
	BuildSystemPrompt(agentID string, workspace string) string

	// BuildMessages constructs the message list for LLM call with system prompt, history, and current message.
	BuildMessages(agentID string, workspace string, history []schema.Message, summary string, currentMessage string, media []string, channel, chatID string) []schema.Message

	// AddToolResult appends a tool result message to the message list.
	AddToolResult(messages []schema.Message, toolCallID, toolName, result string) []schema.Message

	// AddAssistantMessage appends an assistant message to the message list.
	AddAssistantMessage(messages []schema.Message, content string, toolCalls []map[string]interface{}) []schema.Message

	// GetSkillsInfo returns information about loaded skills.
	GetSkillsInfo(workspace string) map[string]interface{}
}
