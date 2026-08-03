// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"errors"
	"strings"

	"github.com/pomclaw/pomclaw/internal/agent"
	"github.com/pomclaw/pomclaw/internal/bootstrap"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSystemPromptPreviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get system prompt preview for agent
func NewGetSystemPromptPreviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSystemPromptPreviewLogic {
	return &GetSystemPromptPreviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSystemPromptPreviewLogic) GetSystemPromptPreview(req *types.GetSystemPromptPreviewReq) (resp *types.GetSystemPromptPreviewResp, err error) {
	// Get agent by agent_id
	a, err := l.svcCtx.AgentsModel.FindByAgentID(l.ctx, req.AgentId)
	if err != nil {
		l.Errorf("failed to get agent: %v", err)
		return nil, errors.New("agent not found")
	}

	// Validate mode parameter
	mode := req.Mode
	validModes := map[string]bool{
		types.SystemModeFull:    true,
		types.SystemModeTask:    true,
		types.SystemModeMinimal: true,
		types.SystemModeNone:    true,
	}
	if !validModes[mode] {
		return nil, errors.New("invalid mode: must be full, task, minimal, or none")
	}

	// Load context files from agent_context_files table and filter by mode
	allFiles := bootstrap.LoadFromStore(l.ctx, l.svcCtx.AgentContextFilesModel, req.AgentId)
	allowlist := bootstrap.ModeAllowlist(mode)
	var contextFiles []bootstrap.ContextFile
	if allowlist == nil {
		// full mode: include all files
		contextFiles = allFiles
	} else {
		for _, f := range allFiles {
			if allowlist[f.Path] {
				contextFiles = append(contextFiles, f)
			}
		}
	}

	// Build system prompt using shared builder
	displayName := a.AgentId
	if a.Name.Valid && a.Name.String != "" {
		displayName = a.Name.String
	}

	prompt := agent.BuildSystemPrompt(agent.SystemPromptConfig{
		AgentID:      a.AgentId,
		DisplayName:  displayName,
		ContextFiles: contextFiles,
		Workspace:    a.Workspace,
	})

	// Parse sections from markdown headers
	sections := parseSections(prompt)

	// Calculate token count (approximate - in real implementation use tokencount library)
	tokenCount := len(strings.Fields(prompt)) / 2 // rough approximation

	resp = &types.GetSystemPromptPreviewResp{
		Mode:       mode,
		Prompt:     prompt,
		TokenCount: tokenCount,
		Sections:   sections,
		Tools:      []types.ToolDefinition{}, // Can be populated from agent's tool registry
	}

	return resp, nil
}

// parseSections extracts section boundaries from markdown headers
func parseSections(prompt string) []types.PromptSection {
	var sections []types.PromptSection
	lines := strings.Split(prompt, "\n")
	pos := 0

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "# ") {
			name := strings.TrimPrefix(strings.TrimPrefix(line, "## "), "# ")
			sections = append(sections, types.PromptSection{
				Name:  name,
				Start: pos,
			})
			// Update previous section's end
			if len(sections) > 1 {
				sections[len(sections)-2].End = pos - 1
			}
		}
		pos += len(line) + 1
	}

	// Set last section's end
	if len(sections) > 0 {
		sections[len(sections)-1].End = len(prompt)
	}

	return sections
}
