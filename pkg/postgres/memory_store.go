package postgres

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/pomclaw/pomclaw/internal/contracts"
	"github.com/pomclaw/pomclaw/internal/model"
	"strings"
	"time"
)

// MemoryStore implements contracts.MemoryStoreInterface backed by PostgreSQL models.
type MemoryStore struct {
	memoriesModel model.MemoryDocumentsModel
}

// NewMemoryStore creates a new PostgreSQL-backed memory store.
func NewMemoryStore(memoriesModel model.MemoryDocumentsModel) *MemoryStore {
	return &MemoryStore{
		memoriesModel: memoriesModel,
	}
}

// ReadLongTerm reads all long-term memories, joined with "---" separator.
func (ms *MemoryStore) ReadLongTerm(agentID string) string {
	ctx := context.Background()
	results, err := ms.memoriesModel.ReadLongTerm(ctx, agentID)
	if err != nil || len(results) == 0 {
		return ""
	}
	var ret string
	for i, r := range results {
		if i > 0 {
			ret += "\n\n---\n\n"
		}
		ret += r
	}
	return ret
}

// WriteLongTerm stores a new long-term memory.
func (ms *MemoryStore) WriteLongTerm(agentID string, content string) error {
	ctx := context.Background()
	_, err := ms.memoriesModel.Insert(ctx, &model.MemoryDocuments{
		AgentId: agentID,
		Path:    fmt.Sprintf("long_term/%d", time.Now().UnixNano()),
		Content: content,
		Hash:    hashContent(content),
	})
	return err
}

// GetMemoryContext returns formatted memory context for the agent prompt.
func (ms *MemoryStore) GetMemoryContext(agentID string) string {
	var parts []string

	longTerm := ms.ReadLongTerm(agentID)
	if longTerm != "" {
		parts = append(parts, "## Long-term Memory\n\n"+longTerm)
	}

	if len(parts) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, part := range parts {
		if i > 0 {
			sb.WriteString("\n\n---\n\n")
		}
		sb.WriteString(part)
	}
	return fmt.Sprintf("# Memory\n\n%s", sb.String())
}

// Remember stores a new memory.
func (ms *MemoryStore) Remember(agentID string, text string, importance float64, category string) (string, error) {
	ctx := context.Background()
	path := fmt.Sprintf("memory/%d", time.Now().UnixNano())
	_, err := ms.memoriesModel.Insert(ctx, &model.MemoryDocuments{
		AgentId: agentID,
		Path:    path,
		Content: text,
		Hash:    hashContent(text),
	})
	return path, err
}

// Recall searches for memories.
func (ms *MemoryStore) Recall(agentID string, query string, maxResults int) ([]contracts.MemoryRecallResult, error) {
	ctx := context.Background()
	records, err := ms.memoriesModel.Recall(ctx, agentID, maxResults)
	if err != nil {
		return []contracts.MemoryRecallResult{}, err
	}

	var results []contracts.MemoryRecallResult
	for _, record := range records {
		results = append(results, contracts.MemoryRecallResult{
			MemoryID: fmt.Sprintf("%d", record.Id),
			Text:     record.Content,
			Category: record.CustomScope.String,
			Score:    0.0,
		})
	}
	return results, nil
}

// Forget removes a memory document by ID.
func (ms *MemoryStore) Forget(agentID string, memoryID int64) error {
	ctx := context.Background()
	return ms.memoriesModel.Delete(ctx, memoryID)
}

func hashContent(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h)
}
