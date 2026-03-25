package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pomclaw/pomclaw/pkg/agent"
	"github.com/pomclaw/pomclaw/pkg/logger"
)

// MemoryStore implements MemoryStoreInterface and OracleMemoryStore backed by PostgreSQL.
type MemoryStore struct {
	db        *sql.DB
	agentID   string
	embedding *EmbeddingService
}

// NewMemoryStore creates a new PostgreSQL-backed memory store.
func NewMemoryStore(db *sql.DB, agentID string, embedding interface{}) *MemoryStore {
	embSvc, ok := embedding.(*EmbeddingService)
	if !ok {
		embSvc = NewEmbeddingService(db)
	}
	return &MemoryStore{
		db:        db,
		agentID:   agentID,
		embedding: embSvc,
	}
}

// ReadLongTerm reads all long-term memories, joined with "---" separator.
func (ms *MemoryStore) ReadLongTerm() string {
	// Order by importance with time-decay: recently accessed memories rank higher
	rows, err := ms.db.Query(`
		SELECT content FROM PICO_MEMORIES
		WHERE agent_id = $1
		ORDER BY (importance * (1.0 / (1.0 + EXTRACT(DAY FROM (CURRENT_TIMESTAMP - COALESCE(accessed_at, created_at))) * 0.1))) DESC,
		         DATE(created_at) DESC
		LIMIT 50`,
		ms.agentID,
	)
	if err != nil {
		logger.WarnCF("postgres", "Failed to read long-term memories", map[string]interface{}{"error": err.Error()})
		return ""
	}
	defer rows.Close()

	var parts []string
	for rows.Next() {
		var content sql.NullString
		if err := rows.Scan(&content); err == nil && content.Valid {
			parts = append(parts, content.String)
		}
	}

	return strings.Join(parts, "\n\n---\n\n")
}

// WriteLongTerm stores a new long-term memory with embedding.
func (ms *MemoryStore) WriteLongTerm(content string) error {
	_, err := ms.Remember(content, 0.7, "long_term")
	return err
}

// ReadToday reads today's daily note.
func (ms *MemoryStore) ReadToday() string {
	var content sql.NullString
	err := ms.db.QueryRow(`
		SELECT content FROM PICO_DAILY_NOTES
		WHERE agent_id = $1 AND note_date = CURRENT_DATE
		ORDER BY updated_at DESC
		LIMIT 1`,
		ms.agentID,
	).Scan(&content)
	if err != nil || !content.Valid {
		return ""
	}
	return content.String
}

// AppendToday appends content to today's daily note.
func (ms *MemoryStore) AppendToday(content string) error {
	existing := ms.ReadToday()

	if existing == "" {
		// Insert new daily note
		header := fmt.Sprintf("# %s\n\n", time.Now().Format("2006-01-02"))
		fullContent := header + content

		noteID := uuid.New().String()[:8]

		if ms.embedding != nil && ms.embedding.Mode() == "api" {
			emb, err := ms.embedding.EmbedText(fullContent)
			if err != nil {
				logger.WarnCF("postgres", "Embedding failed, storing without vector", map[string]interface{}{"error": err.Error()})
				_, err = ms.db.Exec(`
					INSERT INTO PICO_DAILY_NOTES (note_id, agent_id, note_date, content)
					VALUES ($1, $2, CURRENT_DATE, $3)`,
					noteID, ms.agentID, fullContent,
				)
				return err
			}
			vecStr := float32SliceToString(emb)
			_, err = ms.db.Exec(`
				INSERT INTO PICO_DAILY_NOTES (note_id, agent_id, note_date, content, embedding)
				VALUES ($1, $2, CURRENT_DATE, $3, $4::vector)`,
				noteID, ms.agentID, fullContent, vecStr,
			)
			return err
		}

		_, err := ms.db.Exec(`
			INSERT INTO PICO_DAILY_NOTES (note_id, agent_id, note_date, content)
			VALUES ($1, $2, CURRENT_DATE, $3)`,
			noteID, ms.agentID, fullContent,
		)
		return err
	}

	// Append to existing
	newContent := existing + "\n" + content

	if ms.embedding != nil && ms.embedding.Mode() == "api" {
		emb, err := ms.embedding.EmbedText(newContent)
		if err != nil {
			logger.WarnCF("postgres", "Embedding failed, storing without vector", map[string]interface{}{"error": err.Error()})
			_, err = ms.db.Exec(`
				UPDATE PICO_DAILY_NOTES
				SET content = $1, updated_at = CURRENT_TIMESTAMP
				WHERE agent_id = $2 AND note_date = CURRENT_DATE`,
				newContent, ms.agentID,
			)
			return err
		}
		vecStr := float32SliceToString(emb)
		_, err = ms.db.Exec(`
			UPDATE PICO_DAILY_NOTES
			SET content = $1, embedding = $2::vector, updated_at = CURRENT_TIMESTAMP
			WHERE agent_id = $3 AND note_date = CURRENT_DATE`,
			newContent, vecStr, ms.agentID,
		)
		return err
	}

	_, err := ms.db.Exec(`
		UPDATE PICO_DAILY_NOTES
		SET content = $1, updated_at = CURRENT_TIMESTAMP
		WHERE agent_id = $2 AND note_date = CURRENT_DATE`,
		newContent, ms.agentID,
	)
	return err
}

// GetRecentDailyNotes returns daily notes from the last N days.
func (ms *MemoryStore) GetRecentDailyNotes(days int) string {
	rows, err := ms.db.Query(`
		SELECT content FROM PICO_DAILY_NOTES
		WHERE agent_id = $1 AND note_date >= CURRENT_DATE - INTERVAL '1 day' * $2
		ORDER BY note_date DESC`,
		ms.agentID, days,
	)
	if err != nil {
		logger.WarnCF("postgres", "Failed to read recent daily notes", map[string]interface{}{"error": err.Error()})
		return ""
	}
	defer rows.Close()

	var notes []string
	for rows.Next() {
		var content sql.NullString
		if err := rows.Scan(&content); err == nil && content.Valid {
			notes = append(notes, content.String)
		}
	}

	if len(notes) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, note := range notes {
		if i > 0 {
			sb.WriteString("\n\n---\n\n")
		}
		sb.WriteString(note)
	}
	return sb.String()
}

// GetMemoryContext returns formatted memory context for the agent prompt.
func (ms *MemoryStore) GetMemoryContext() string {
	var parts []string

	longTerm := ms.ReadLongTerm()
	if longTerm != "" {
		parts = append(parts, "## Long-term Memory\n\n"+longTerm)
	}

	recentNotes := ms.GetRecentDailyNotes(3)
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

// Remember stores a new memory with embedding for vector search.
func (ms *MemoryStore) Remember(text string, importance float64, category string) (string, error) {
	// Check for near-duplicate memories before inserting
	if existingID, updated := ms.deduplicateMemory(text, importance); updated {
		return existingID, nil
	}

	memoryID := uuid.New().String()[:8]

	if ms.embedding != nil && ms.embedding.Mode() == "api" {
		// API mode: compute embedding via external API
		emb, err := ms.embedding.EmbedText(text)
		if err != nil {
			logger.WarnCF("postgres", "Embedding failed, storing without vector", map[string]interface{}{"error": err.Error()})
			_, err = ms.db.Exec(`
				INSERT INTO PICO_MEMORIES (memory_id, agent_id, content, importance, category)
				VALUES ($1, $2, $3, $4, $5)`,
				memoryID, ms.agentID, text, importance, category,
			)
			if err != nil {
				return "", fmt.Errorf("failed to remember: %w", err)
			}
		} else {
			vecStr := float32SliceToString(emb)
			_, err = ms.db.Exec(`
				INSERT INTO PICO_MEMORIES (memory_id, agent_id, content, embedding, importance, category)
				VALUES ($1, $2, $3, $4::vector, $5, $6)`,
				memoryID, ms.agentID, text, vecStr, importance, category,
			)
			if err != nil {
				return "", fmt.Errorf("failed to remember: %w", err)
			}
		}
	} else {
		_, err := ms.db.Exec(`
			INSERT INTO PICO_MEMORIES (memory_id, agent_id, content, importance, category)
			VALUES ($1, $2, $3, $4, $5)`,
			memoryID, ms.agentID, text, importance, category,
		)
		if err != nil {
			return "", fmt.Errorf("failed to remember: %w", err)
		}
	}

	return memoryID, nil
}

// Recall searches for memories similar to the query using vector search.
func (ms *MemoryStore) Recall(query string, maxResults int) ([]agent.MemoryRecallResult, error) {
	if ms.embedding == nil || query == "" {
		return nil, nil
	}

	queryEmb, err := ms.embedding.EmbedText(query)
	if err != nil {
		logger.WarnCF("postgres", "Failed to embed recall query", map[string]interface{}{"error": err.Error()})
		return []agent.MemoryRecallResult{}, err
	}

	vecStr := float32SliceToString(queryEmb)

	rows, err := ms.db.Query(`
		SELECT memory_id, content, importance, category,
		       1 - (embedding <=> $3::vector) AS similarity
		FROM PICO_MEMORIES
		WHERE agent_id = $1 AND embedding IS NOT NULL
		ORDER BY embedding <=> $3::vector ASC
		LIMIT $2`,
		ms.agentID, maxResults, vecStr,
	)
	if err != nil {
		return []agent.MemoryRecallResult{}, fmt.Errorf("recall query failed: %w", err)
	}
	defer rows.Close()

	var results []agent.MemoryRecallResult
	for rows.Next() {
		var memID string
		var content string
		var importance float64
		var category string
		var score float64

		if err := rows.Scan(&memID, &content, &importance, &category, &score); err != nil {
			logger.WarnCF("postgres", "Failed to scan recall result", map[string]interface{}{"error": err.Error()})
			continue
		}

		results = append(results, agent.MemoryRecallResult{
			MemoryID:   memID,
			Text:       content,
			Importance: importance,
			Category:   category,
			Score:      score,
		})

		// Update access count and accessed_at
		ms.db.Exec(`
			UPDATE PICO_MEMORIES
			SET access_count = access_count + 1, accessed_at = CURRENT_TIMESTAMP
			WHERE memory_id = $1`,
			memID,
		)
	}

	return results, nil
}

// Forget removes a memory by ID.
func (ms *MemoryStore) Forget(memoryID string) error {
	_, err := ms.db.Exec(`
		DELETE FROM PICO_MEMORIES
		WHERE memory_id = $1 AND agent_id = $2`,
		memoryID, ms.agentID,
	)
	return err
}

// deduplicateMemory checks if a similar memory already exists and updates it.
func (ms *MemoryStore) deduplicateMemory(text string, importance float64) (string, bool) {
	// Simple deduplication: check for exact text match
	var existingID sql.NullString
	err := ms.db.QueryRow(`
		SELECT memory_id FROM PICO_MEMORIES
		WHERE agent_id = $1 AND content = $2
		LIMIT 1`,
		ms.agentID, text,
	).Scan(&existingID)

	if err == sql.ErrNoRows {
		return "", false
	}
	if err != nil {
		return "", false
	}

	// Update the existing memory
	if existingID.Valid {
		ms.db.Exec(`
			UPDATE PICO_MEMORIES
			SET importance = GREATEST(importance, $1), access_count = access_count + 1, updated_at = CURRENT_TIMESTAMP
			WHERE memory_id = $2`,
			importance, existingID.String,
		)
		return existingID.String, true
	}

	return "", false
}

// float32SliceToString converts a float32 slice to JSON string for storing in PostgreSQL.
func float32SliceToString(emb []float32) string {
	data, _ := json.Marshal(emb)
	return string(data)
}
