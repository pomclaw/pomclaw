package agent

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/pomclaw/pomclaw/internal/bootstrap"
	"github.com/pomclaw/pomclaw/internal/contracts"
	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/zeromicro/go-zero/core/logx"
	"strings"
)

type ContextBuilder struct {
	toolsNodeConfig        compose.ToolsNodeConfig
	skillsLoader           contracts.SkillsLoaderInterface
	memory                 contracts.MemoryStoreInterface
	agentContextFilesModel model.AgentContextFilesModel
}

func NewContextBuilder(agentContextFilesModel model.AgentContextFilesModel, memoryStore contracts.MemoryStoreInterface, toolsNodeConfig compose.ToolsNodeConfig, skillsLoader contracts.SkillsLoaderInterface) contracts.ContextBuilderInterface {
	return &ContextBuilder{
		toolsNodeConfig:        toolsNodeConfig,
		skillsLoader:           skillsLoader,
		memory:                 memoryStore,
		agentContextFilesModel: agentContextFilesModel,
	}
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
	// Build tools section dynamically
	toolsSection := cb.buildToolsSection()

	// Load context files from agent_context_files table
	var contextFiles []bootstrap.ContextFile
	if cb.agentContextFilesModel != nil {
		ctx := context.Background()
		contextFiles = bootstrap.LoadFromStore(ctx, cb.agentContextFilesModel, agentID)
	}

	// Use shared builder for core identity + context files
	cfg := SystemPromptConfig{
		AgentID:      agentID,
		Workspace:    workspace,
		ToolsSection: toolsSection,
		ContextFiles: contextFiles,
	}
	prompt := BuildSystemPrompt(cfg)

	// Skills - show summary, AI can read full content with read_file tool
	// Cull oversized skills sections to prevent system prompt bloat
	if cb.skillsLoader != nil {

		skillsSummary := cb.skillsLoader.BuildSkillsSummary()
		if skillsSummary != "" {

			if len(skillsSummary) > 8192 {
				logx.Info("agent", "Skills summary exceeds 8KB, truncating to skill names only",
					map[string]interface{}{"original_size": len(skillsSummary)})
				allSkills := cb.skillsLoader.ListSkills()
				var names []string
				for _, s := range allSkills {
					names = append(names, s.Name)
				}
				skillsSummary = "Available skills (use read_file to see details): " + strings.Join(names, ", ")
			}
			prompt += "\n\n---\n\n# Skills\n\nThe following skills extend your capabilities. To use a skill, call the use_skill tool with its name.\n\n" + skillsSummary
		}
	}

	// Memory context
	memoryContext := cb.memory.GetMemoryContext(agentID)
	if memoryContext != "" {
		prompt += "\n\n---\n\n# Memory\n\n" + memoryContext
	}

	return prompt
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

	allSkills := cb.skillsLoader.ListSkills()
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
