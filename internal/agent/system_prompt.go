package agent

import (
	"fmt"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/pomclaw/pomclaw/internal/bootstrap"
	"github.com/pomclaw/pomclaw/prompt"
)

// SystemPromptConfig holds the data needed to build a system prompt.
// The caller is responsible for filtering ContextFiles by ModeAllowlist.
type SystemPromptConfig struct {
	AgentID      string
	DisplayName  string
	ContextFiles []bootstrap.ContextFile
	Workspace    string
	ToolsSection string
}

// BuildSystemPrompt is the shared system prompt builder used by both
// runtime (ContextBuilder) and preview API (GetSystemPromptPreviewLogic).
//
//  1. Renders the systemprompt.md Go template (core identity)
//  2. Appends context files (already filtered by the caller)
//  3. Joins sections with "---" separator
func BuildSystemPrompt(cfg SystemPromptConfig) string {
	var parts []string

	// Core identity section: render systemprompt.md template
	identity := renderIdentity(cfg.Workspace, cfg.ToolsSection)
	parts = append(parts, identity)

	// Context files (already filtered by ModeAllowlist)
	if len(cfg.ContextFiles) > 0 {
		var contextParts []string
		for _, f := range cfg.ContextFiles {
			contextParts = append(contextParts, fmt.Sprintf("## %s\n\n%s", f.Path, f.Content))
		}
		parts = append(parts, strings.Join(contextParts, "\n\n"))
	}

	return strings.Join(parts, "\n\n---\n\n")
}

// renderIdentity renders the systemprompt.md Go template with the given data.
func renderIdentity(workspace, toolsSection string) string {
	now := time.Now().Format("2006-01-02 15:04 (Monday)")
	runtimeInfo := fmt.Sprintf("%s %s, Go %s", runtime.GOOS, runtime.GOARCH, runtime.Version())

	tmpl, _ := template.New("systemprompt").Parse(prompt.SystemPrompt)

	data := map[string]interface{}{
		"Now":           now,
		"Runtime":       runtimeInfo,
		"WorkspacePath": workspace,
		"ToolsSection":  toolsSection,
	}

	var buf strings.Builder
	_ = tmpl.Execute(&buf, data)
	return buf.String()
}
