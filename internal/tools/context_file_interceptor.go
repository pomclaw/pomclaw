package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/pomclaw/pomclaw/internal/bootstrap"
	"github.com/pomclaw/pomclaw/internal/model"
	"path/filepath"
	"strings"
)

// protectedFileSet defines files that require group file writer permission in group chats.
// These files control the agent's identity and behavior — only allowlisted users can modify them.
var protectedFileSet = map[string]bool{
	bootstrap.SoulFile:           true,
	bootstrap.IdentityFile:       true,
	bootstrap.AgentsFile:         true,
	bootstrap.UserFile:           true,
	bootstrap.UserPredefinedFile: true,
	bootstrap.CapabilitiesFile:   true,
}

// contextFileSet is the set of filenames routed to the DB store.
// TOOLS.md excluded — not applicable.
var contextFileSet = map[string]bool{
	bootstrap.SoulFile:           true,
	bootstrap.AgentsFile:         true,
	bootstrap.IdentityFile:       true,
	bootstrap.UserFile:           true,
	bootstrap.UserPredefinedFile: true,
	bootstrap.BootstrapFile:      true, // first-run file (deleted after completion)
	bootstrap.HeartbeatFile:      true, // agent-level heartbeat checklist
	bootstrap.CapabilitiesFile:   true, // domain expertise (evolvable when self_evolve=true)
}

// isContextFile checks if a path refers to a workspace-root context file.
// Handles both relative ("SOUL.md") and absolute ("/workspace/SOUL.md") paths.
func isContextFile(path string) (fileName string, ok bool) {
	base := filepath.Base(path)
	if !contextFileSet[base] {
		return "", false
	}

	// Relative root-level: "SOUL.md", "./SOUL.md"
	dir := filepath.Dir(path)
	if dir == "." || dir == "/" || dir == "" {
		return base, true
	}

	return "", false
}

// ContextFileInterceptor routes context file reads/writes to the agent store.
// Keeps SOUL.md, IDENTITY.md etc. in Postgres.
type ContextFileInterceptor struct {
	agentContextFilesModel model.AgentContextFilesModel
}

// NewContextFileInterceptor creates an interceptor backed by the given agent store.
func NewContextFileInterceptor(
	agentContextFilesModel model.AgentContextFilesModel,
) *ContextFileInterceptor {
	return &ContextFileInterceptor{
		agentContextFilesModel: agentContextFilesModel,
	}
}

// ReadFile attempts to read a context file from the DB.
// Returns (content, true, nil) if handled, or ("", false, nil) if not a context file.
func (b *ContextFileInterceptor) ReadFile(ctx context.Context, path string) (string, bool, error) {
	fileName, ok := isContextFile(path)
	if !ok {
		fileName = path
	}

	agentID := AgentIDFromContext(ctx)
	if agentID == "" {
		return "", false, nil // no agent context
	}

	// Agent-level context files only
	return b.readAgentFile(ctx, agentID, fileName)
}

func (b *ContextFileInterceptor) readAgentFile(ctx context.Context, agentID, fileName string) (string, bool, error) {
	files, err := b.agentContextFilesModel.FindOneByAgentIdFileName(ctx, agentID, fileName)
	if err != nil {
		return "", true, err
	}
	return files.Content, true, nil
}

// WriteFile attempts to write a context file to the DB.
// Returns (true, nil) if handled, or (false, nil) if not a context file.
func (b *ContextFileInterceptor) WriteFile(ctx context.Context, path, content string) (bool, error) {
	filetype := model.AgentContextFiles_FileType_User
	fileName, ok := isContextFile(path)
	if ok {
		filetype = model.AgentContextFiles_FileType_Agent
	} else {
		fileName = path
	}

	agentID := AgentIDFromContext(ctx)
	if agentID == "" {
		return false, nil // no agent context
	}

	userID := UserIDFromContext(ctx)
	if userID == "" {
		return false, fmt.Errorf("user ID not found in context")
	}

	// Agent-level write only
	data := &model.AgentContextFiles{
		UserId:    userID,
		AgentId:   agentID,
		FileType:  int64(filetype),
		FileName:  fileName,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := b.agentContextFilesModel.SetAgentContextFile(ctx, data)
	return true, err
}

// LoadContextFiles loads context files for an agent.
// Used by the agent loop to dynamically resolve context files for system prompt.
func (b *ContextFileInterceptor) LoadContextFiles(ctx context.Context, agentID string) []bootstrap.ContextFile {
	agentFiles, err := b.agentContextFilesModel.GetAgentContextFiles(ctx, agentID)
	if err != nil {
		return nil
	}
	var result []bootstrap.ContextFile
	for _, f := range agentFiles {
		if f.Content == "" {
			continue
		}
		result = append(result, bootstrap.ContextFile{
			Path:    f.FileName,
			Content: f.Content,
		})
	}
	return result
}

// ListAllContextFilesMetadata returns metadata for all context files (no content).
// Used by list_files tool for efficient listing.
func (b *ContextFileInterceptor) ListAllContextFilesMetadata(ctx context.Context, agentID string) ([]*model.AgentContextFiles, error) {
	return b.agentContextFilesModel.ListAllContextFilesMetadata(ctx, agentID)
}

// SearchContextFilesMetadata returns metadata for context files matching a path prefix (no content).
// Used by list_files tool for efficient prefix-based search.
func (b *ContextFileInterceptor) SearchContextFilesMetadata(ctx context.Context, agentID string, pathPrefix string) ([]*model.AgentContextFiles, error) {
	return b.agentContextFilesModel.SearchContextFilesMetadata(ctx, agentID, pathPrefix)
}

// normalizeToRelative strips the workspace prefix from an absolute path,
// returning a workspace-relative path for consistent DB storage.
// e.g. "/home/user/workspace/SOUL.md" → "SOUL.md"
func normalizeToRelative(path, workspace string) string {
	if workspace == "" || !filepath.IsAbs(path) {
		return path
	}
	rel, err := filepath.Rel(filepath.Clean(workspace), filepath.Clean(path))
	if err != nil || strings.HasPrefix(rel, "..") {
		return path // outside workspace, return as-is
	}
	return rel
}
