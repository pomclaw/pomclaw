package agent

import (
	"context"

	"github.com/pomclaw/pomclaw/pkg/providers"
	"github.com/pomclaw/pomclaw/pkg/tools"
)

const (
	DefaultAgentID = "default"
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
	ReadToday(agentID string) string
	AppendToday(agentID string, content string) error
	GetRecentDailyNotes(agentID string, days int) string
	GetMemoryContext(agentID string) string
}

// SessionManagerInterface defines the contract for session management backends.
type SessionManagerInterface interface {
	AddMessage(agentID string, key, role, content string)
	AddFullMessage(agentID string, key string, msg providers.Message)
	GetHistory(agentID string, key string) []providers.Message
	SetHistory(agentID string, key string, history []providers.Message)
	GetSummary(agentID string, key string) string
	SetSummary(agentID string, key, summary string)
	TruncateHistory(agentID string, key string, keepLast int)
	Save(agentID string, key string) error
}

// StateManagerInterface defines the contract for state management backends.
type StateManagerInterface interface {
	SetLastChannel(agentID string, channel string) error
	GetLastChannel(agentID string) string
	SetLastChatID(agentID string, chatID string) error
	GetLastChatID(agentID string) string
}

// OracleMemoryStore is an extended interface for Oracle-backed memory with vector search.
type OracleMemoryStore interface {
	MemoryStoreInterface
	Remember(agentID string, text string, importance float64, category string) (string, error)
	Recall(agentID string, query string, maxResults int) ([]MemoryRecallResult, error)
	Forget(agentID string, memoryID string) error
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

// SkillsLoaderInterface defines the contract for loading skills from various sources.
// Skills can be loaded from workspace, global (~/.pomclaw/skills), or builtin directories.
type SkillsLoaderInterface interface {
	// ListSkills returns all available skills across all sources (workspace, global, builtin).
	// Workspace skills take precedence over global, which take precedence over builtin.
	ListSkills(workspace string) []SkillInfo

	// LoadSkill loads the content of a single skill by name.
	// Returns the skill content (with frontmatter stripped) and true if found, or "" and false if not found.
	LoadSkill(workspace string, name string) (string, bool)

	// LoadSkillsForContext loads multiple skills and formats them for context inclusion.
	// Returns empty string if no skills are found or names is empty.
	LoadSkillsForContext(workspace string, skillNames []string) string

	// BuildSkillsSummary generates an XML-formatted summary of all available skills.
	// Used for including in system prompts.
	BuildSkillsSummary(workspace string) string
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
	// SetSkillsLoader sets the skills loader for dynamic skill loading.
	SetSkillsLoader(loader SkillsLoaderInterface)

	// SetToolsRegistry sets the tools registry for dynamic tool summary generation.
	SetToolsRegistry(registry *tools.ToolRegistry)

	// SetMemoryStore replaces the default file-based memory store with a custom implementation.
	SetMemoryStore(store MemoryStoreInterface)

	// SetPromptStore sets an optional Oracle-backed prompt store.
	SetPromptStore(store PromptStoreInterface)

	// GetMemoryStore returns the active memory store (file-based or Oracle-backed).
	GetMemoryStore() MemoryStoreInterface

	// BuildSystemPrompt assembles the system prompt from identity, bootstrap files, skills, and memory context.
	BuildSystemPrompt(agentID string, workspace string) string

	// LoadBootstrapFiles loads and formats bootstrap files from workspace or Oracle store.
	LoadBootstrapFiles(agentID string, workspace string) string

	// BuildMessages constructs the message list for LLM call with system prompt, history, and current message.
	BuildMessages(agentID string, workspace string, history []providers.Message, summary string, currentMessage string, media []string, channel, chatID string) []providers.Message

	// AddToolResult appends a tool result message to the message list.
	AddToolResult(messages []providers.Message, toolCallID, toolName, result string) []providers.Message

	// AddAssistantMessage appends an assistant message to the message list.
	AddAssistantMessage(messages []providers.Message, content string, toolCalls []map[string]interface{}) []providers.Message

	// GetSkillsInfo returns information about loaded skills.
	GetSkillsInfo(workspace string) map[string]interface{}
}
