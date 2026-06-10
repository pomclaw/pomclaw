package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/pkg/contracts"
	"strings"
)

// MemoryStore implements contracts.MemoryStoreInterface backed by PostgreSQL models.
type MemoryStore struct {
	memoriesModel   model.MemoriesModel
	dailyNotesModel model.DailyNotesModel
}

// NewMemoryStore creates a new PostgreSQL-backed memory store.
func NewMemoryStore(memoriesModel model.MemoriesModel, dailyNotesModel model.DailyNotesModel) *MemoryStore {
	return &MemoryStore{
		memoriesModel:   memoriesModel,
		dailyNotesModel: dailyNotesModel,
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

// WriteLongTerm stores a new long-term memory with default importance.
func (ms *MemoryStore) WriteLongTerm(agentID string, content string) error {
	ctx := context.Background()
	_, err := ms.memoriesModel.InsertWithoutID(ctx, &model.Memories{
		AgentId:     agentID,
		Content:     sql.NullString{String: content, Valid: true},
		Embedding:   sql.NullString{},
		Importance:  0.7,
		Category:    sql.NullString{String: "long_term", Valid: true},
		AccessCount: 0,
	})
	return err
}

// ReadToday reads today's daily note.
func (ms *MemoryStore) ReadToday(agentID string) string {
	ctx := context.Background()
	content, _ := ms.dailyNotesModel.ReadToday(ctx, agentID)
	return content
}

// AppendToday appends content to today's daily note.
func (ms *MemoryStore) AppendToday(agentID string, content string) error {
	ctx := context.Background()
	return ms.dailyNotesModel.Upsert(ctx, agentID, content)
}

// GetRecentDailyNotes returns daily notes from the last N days.
func (ms *MemoryStore) GetRecentDailyNotes(agentID string, days int) string {
	ctx := context.Background()
	results, err := ms.dailyNotesModel.GetRecentDailyNotes(ctx, agentID, days)
	if err != nil || len(results) == 0 {
		return ""
	}
	var ret string
	for i, note := range results {
		if i > 0 {
			ret += "\n\n---\n\n"
		}
		ret += note
	}
	return ret
}

// GetMemoryContext returns formatted memory context for the agent prompt.
func (ms *MemoryStore) GetMemoryContext(agentID string) string {
	var parts []string

	longTerm := ms.ReadLongTerm(agentID)
	if longTerm != "" {
		parts = append(parts, "## Long-term Memory\n\n"+longTerm)
	}

	recentNotes := ms.GetRecentDailyNotes(agentID, 3)
	if recentNotes != "" {
		parts = append(parts, "## Recent Daily Notes\n\n"+recentNotes)
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

	_, err := ms.memoriesModel.InsertWithoutID(ctx, &model.Memories{
		AgentId:     agentID,
		Content:     sql.NullString{String: text, Valid: true},
		Embedding:   sql.NullString{},
		Importance:  importance,
		Category:    sql.NullString{String: category, Valid: true},
		AccessCount: 0,
	})

	return "", err
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
			MemoryID:   record.Id,
			Text:       record.Content.String,
			Importance: record.Importance,
			Category:   record.Category.String,
			Score:      0.0,
		})
	}
	return results, nil
}

// Forget removes a memory by ID.
func (ms *MemoryStore) Forget(agentID string, memoryID string) error {
	ctx := context.Background()
	return ms.memoriesModel.Delete(ctx, memoryID)
}
