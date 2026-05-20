package agent

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/pomclaw/pomclaw/prompt"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/pomclaw/pomclaw/pkg/contracts"
	"github.com/zeromicro/go-zero/core/logx"
)

type ContextBuilder struct {
	toolsNodeConfig compose.ToolsNodeConfig
	skillsLoader    contracts.SkillsLoaderInterface
	memory          contracts.MemoryStoreInterface
	promptStore     contracts.PromptStoreInterface // Optional Oracle prompt store
}

func NewContextBuilder(promptStore contracts.PromptStoreInterface, memoryStore contracts.MemoryStoreInterface, toolsNodeConfig compose.ToolsNodeConfig, skillsLoader contracts.SkillsLoaderInterface) contracts.ContextBuilderInterface {
	return &ContextBuilder{
		toolsNodeConfig: toolsNodeConfig,
		skillsLoader:    skillsLoader, // Will be set via SetSkillsLoader
		memory:          memoryStore,
		promptStore:     promptStore,
	}
}

func (cb *ContextBuilder) getIdentity(workspace string) string {
	now := time.Now().Format("2006-01-02 15:04 (Monday)")
	workspacePath, _ := filepath.Abs(filepath.Join(workspace))
	runtimeInfo := fmt.Sprintf("%s %s, Go %s", runtime.GOOS, runtime.GOARCH, runtime.Version())

	// Build tools section dynamically
	toolsSection := cb.buildToolsSection()

	// Use text/template with named placeholders
	tmpl, _ := template.New("systemprompt").Parse(prompt.SystemPrompt)

	data := map[string]interface{}{
		"Now":           now,
		"Runtime":       runtimeInfo,
		"WorkspacePath": workspacePath,
		"ToolsSection":  toolsSection,
	}

	var buf strings.Builder
	_ = tmpl.Execute(&buf, data)
	return buf.String()
}

func (cb *ContextBuilder) buildToolsSection() string {

	var sb strings.Builder
	for _, s := range cb.toolsNodeConfig.Tools {
		info, err := s.Info(context.Background())
		if err != nil {
			continue
		}
		sb.WriteString(fmt.Sprintf("- %s:%s\n", info.Name, info.Desc))
	}

	return sb.String()
}

func (cb *ContextBuilder) BuildSystemPrompt(agentID string, workspace string) string {
	var parts []string

	// Core identity section
	parts = append(parts, cb.getIdentity(workspace))

	// Bootstrap files
	bootstrapContent := cb.LoadBootstrapFiles(agentID, workspace)
	if bootstrapContent != "" {
		parts = append(parts, bootstrapContent)
	}

	// Skills - show summary, AI can read full content with read_file tool
	// Cull oversized skills sections to prevent system prompt bloat
	if cb.skillsLoader != nil {

		skillsSummary := cb.skillsLoader.BuildSkillsSummary(workspace)
		if skillsSummary != "" {

			if len(skillsSummary) > 8192 {
				logx.Info("agent", "Skills summary exceeds 8KB, truncating to skill names only",
					map[string]interface{}{"original_size": len(skillsSummary)})
				allSkills := cb.skillsLoader.ListSkills(workspace)
				var names []string
				for _, s := range allSkills {
					names = append(names, s.Name)
				}
				skillsSummary = "Available skills (use read_file to see details): " + strings.Join(names, ", ")
			}
			parts = append(parts, fmt.Sprintf(`# Skills

The following skills extend your capabilities. To use a skill, read its SKILL.md file using the read_file tool.

%s`, skillsSummary))
		}
	}

	// Memory context
	memoryContext := cb.memory.GetMemoryContext(agentID)
	if memoryContext != "" {
		parts = append(parts, "# Memory\n\n"+memoryContext)
	}

	// Join with "---" separator
	return strings.Join(parts, "\n\n---\n\n")
}

func (cb *ContextBuilder) LoadBootstrapFiles(agentID string, workspace string) string {
	// Try Oracle prompt store first
	if cb.promptStore != nil {
		prompts := cb.promptStore.LoadBootstrapFiles(agentID)
		if len(prompts) > 0 {
			var result string
			for name, content := range prompts {
				result += fmt.Sprintf("## %s\n\n%s\n\n", name, content)
			}
			return result
		}
	}

	return ""
}

func (cb *ContextBuilder) BuildMessages(agentID string, workspace string, history []schema.Message, summary string, currentMessage string, media []string, channel, chatID string) []schema.Message {
	var messages []schema.Message

	systemPrompt := cb.BuildSystemPrompt(agentID, workspace)

	// Add Current Session info if provided
	if channel != "" && chatID != "" {
		systemPrompt += fmt.Sprintf("\n\n## Current Session\nChannel: %s\nChat ID: %s", channel, chatID)
	}

	// Log system prompt summary for debugging (debug mode only)
	logx.Debug("agent", "System prompt built",
		map[string]interface{}{
			"total_chars":   len(systemPrompt),
			"total_lines":   strings.Count(systemPrompt, "\n") + 1,
			"section_count": strings.Count(systemPrompt, "\n\n---\n\n") + 1,
		})

	// Log preview of system prompt (avoid logging huge content)
	preview := systemPrompt
	if len(preview) > 500 {
		preview = preview[:500] + "... (truncated)"
	}
	logx.Debug("agent", "System prompt preview", preview)

	if summary != "" {
		systemPrompt += "\n\n## Summary of Previous Conversation\n\n" + summary
	}

	//This fix prevents the session memory from LLM failure due to elimination of toolu_IDs required from LLM
	for len(history) > 0 && (history[0].Role == "tool") {
		logx.Debug("agent", "Removing orphaned tool message from history to prevent LLM error",
			map[string]interface{}{"role": history[0].Role})
		history = history[1:]
	}

	messages = append(messages, schema.Message{
		Role:    "system",
		Content: systemPrompt,
	})

	messages = append(messages, history...)

	messages = append(messages, schema.Message{
		Role:    "user",
		Content: currentMessage,
	})

	return messages
}

func (cb *ContextBuilder) AddToolResult(messages []schema.Message, toolCallID, toolName, result string) []schema.Message {
	messages = append(messages, schema.Message{
		Role:       "tool",
		Content:    result,
		ToolCallID: toolCallID,
	})
	return messages
}

func (cb *ContextBuilder) AddAssistantMessage(messages []schema.Message, content string, toolCalls []map[string]interface{}) []schema.Message {
	msg := schema.Message{
		Role:    "assistant",
		Content: content,
	}
	// Always add assistant message, whether or not it has tool calls
	messages = append(messages, msg)
	return messages
}

// GetSkillsInfo returns information about loaded skills.
func (cb *ContextBuilder) GetSkillsInfo(workspace string) map[string]interface{} {
	if cb.skillsLoader == nil {
		return nil
	}

	allSkills := cb.skillsLoader.ListSkills(workspace)
	skillNames := make([]string, 0, len(allSkills))
	for _, s := range allSkills {
		skillNames = append(skillNames, s.Name)
	}
	return map[string]interface{}{
		"total":     len(allSkills),
		"available": len(allSkills),
		"names":     skillNames,
	}
}
