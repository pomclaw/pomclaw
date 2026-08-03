package model

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// DocumentInfo describes a memory document.
type DocumentInfo struct {
	Path      string `json:"path" db:"path"`
	Hash      string `json:"hash" db:"hash"`
	AgentID   string `json:"agent_id,omitempty" db:"agent_id"`
	UpdatedAt int64  `json:"updated_at" db:"updated_at"`
}

// MemorySearchResult is a single result from memory search.
type MemorySearchResult struct {
	Path      string  `json:"path" db:"-"`
	StartLine int     `json:"start_line" db:"-"`
	EndLine   int     `json:"end_line" db:"-"`
	Score     float64 `json:"score" db:"-"`
	Snippet   string  `json:"snippet" db:"-"`
	Source    string  `json:"source" db:"-"`
	Scope     string  `json:"scope,omitempty" db:"-"` // "global" or "personal"
}

// MemorySearchOptions configures a memory search query.
type MemorySearchOptions struct {
	MaxResults   int
	MinScore     float64
	Source       string // "memory", "sessions", ""
	PathPrefix   string
	VectorWeight float64 // per-agent override (0 = use store default)
	TextWeight   float64 // per-agent override (0 = use store default)
}

// DocumentDetail provides full document info including chunk/embedding stats.
type DocumentDetail struct {
	Path          string `json:"path" db:"path"`
	Content       string `json:"content" db:"content"`
	Hash          string `json:"hash" db:"hash"`
	ChunkCount    int    `json:"chunk_count" db:"chunk_count"`
	EmbeddedCount int    `json:"embedded_count" db:"embedded_count"`
	CreatedAt     int64  `json:"created_at" db:"created_at"`
	UpdatedAt     int64  `json:"updated_at" db:"updated_at"`
}

// ChunkInfo describes a single memory chunk.
type ChunkInfo struct {
	ID           string `json:"id" db:"id"`
	StartLine    int    `json:"start_line" db:"start_line"`
	EndLine      int    `json:"end_line" db:"end_line"`
	TextPreview  string `json:"text_preview" db:"text_preview"`
	HasEmbedding bool   `json:"has_embedding" db:"has_embedding"`
}

var _ MemoryDocumentsModel = (*customMemoryDocumentsModel)(nil)

type (
	// MemoryDocumentsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMemoryDocumentsModel.
	MemoryDocumentsModel interface {
		memoryDocumentsModel
		withSession(session sqlx.Session) MemoryDocumentsModel
		FindByAgentId(ctx context.Context, agentId string) ([]*MemoryDocuments, error)
		Finds(ctx context.Context) ([]*MemoryDocuments, error)
		DeleteByAgentIdAndPath(ctx context.Context, agentId string, path string) error

		// Document CRUD
		GetDocument(ctx context.Context, agentID, path string) (string, error)
		PutDocument(ctx context.Context, agentID, path, content string) error
		DeleteDocument(ctx context.Context, agentID, path string) error
		ListDocuments(ctx context.Context, agentID string) ([]DocumentInfo, error)

		// Admin queries
		ListAllDocumentsGlobal(ctx context.Context) ([]DocumentInfo, error)
		ListAllDocuments(ctx context.Context, agentID string) ([]DocumentInfo, error)
		GetDocumentDetail(ctx context.Context, agentID, path string) (*DocumentDetail, error)
		ListChunks(ctx context.Context, agentID, path string) ([]ChunkInfo, error)

		// Search
		Search(ctx context.Context, query string, agentID string, opts MemorySearchOptions) ([]MemorySearchResult, error)

		// Indexing
		IndexDocument(ctx context.Context, agentID, path string) error
		IndexAll(ctx context.Context, agentID string) error

		// Long-term memory helpers (using memory_documents table)
		ReadLongTerm(ctx context.Context, agentID string) ([]string, error)
		Recall(ctx context.Context, agentID string, limit int) ([]*MemoryDocuments, error)
	}

	customMemoryDocumentsModel struct {
		*defaultMemoryDocumentsModel
	}
)

// NewMemoryDocumentsModel returns a model for the database table.
func NewMemoryDocumentsModel(conn sqlx.SqlConn) MemoryDocumentsModel {
	return &customMemoryDocumentsModel{
		defaultMemoryDocumentsModel: newMemoryDocumentsModel(conn),
	}
}

func (m *customMemoryDocumentsModel) withSession(session sqlx.Session) MemoryDocumentsModel {
	return NewMemoryDocumentsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customMemoryDocumentsModel) FindByAgentId(ctx context.Context, agentId string) ([]*MemoryDocuments, error) {
	var resp []*MemoryDocuments
	query := `select id, agent_id, path, content, hash, custom_scope, created_at, updated_at from "public"."memory_documents" where agent_id = $1 order by updated_at desc`
	err := m.conn.QueryRowsCtx(ctx, &resp, query, agentId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customMemoryDocumentsModel) Finds(ctx context.Context) ([]*MemoryDocuments, error) {
	var resp []*MemoryDocuments
	query := `select id, agent_id, path, content, hash, custom_scope, created_at, updated_at from "public"."memory_documents" order by updated_at desc`
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customMemoryDocumentsModel) DeleteByAgentIdAndPath(ctx context.Context, agentId string, path string) error {
	query := `delete from "public"."memory_documents" where agent_id = $1 and path = $2`
	_, err := m.conn.ExecCtx(ctx, query, agentId, path)
	return err
}

// GetDocument retrieves the full content of a memory document.
func (m *customMemoryDocumentsModel) GetDocument(ctx context.Context, agentID, path string) (string, error) {
	if agentID == "" || path == "" {
		return "", nil
	}

	docs, err := m.FindByAgentId(ctx, agentID)
	if err != nil {
		return "", err
	}
	for _, doc := range docs {
		if doc.Path == path {
			return doc.Content, nil
		}
	}

	return "", nil
}

// Search performs a keyword-based search on memory documents.
// Returns documents matching the query string (case-insensitive).
func (m *customMemoryDocumentsModel) Search(ctx context.Context, query string, agentID string, opts MemorySearchOptions) ([]MemorySearchResult, error) {
	if agentID == "" {
		return nil, nil
	}

	docs, err := m.FindByAgentId(ctx, agentID)
	if err != nil {
		return nil, err
	}

	// Keyword search (case-insensitive substring match)
	var results []MemorySearchResult
	queryLower := strings.ToLower(query)

	for _, doc := range docs {
		contentLower := strings.ToLower(doc.Content)
		if strings.Contains(contentLower, queryLower) {
			// Extract snippet around the match
			idx := strings.Index(contentLower, queryLower)
			start := idx - 50
			if start < 0 {
				start = 0
			}
			end := idx + len(query) + 50
			if end > len(doc.Content) {
				end = len(doc.Content)
			}
			snippet := doc.Content[start:end]
			if start > 0 {
				snippet = "..." + snippet
			}
			if end < len(doc.Content) {
				snippet = snippet + "..."
			}

			results = append(results, MemorySearchResult{
				Path:    doc.Path,
				Score:   0.5, // Placeholder: keyword match score
				Snippet: snippet,
				Source:  "memory",
			})
		}
	}

	// Limit results
	if opts.MaxResults > 0 && len(results) > opts.MaxResults {
		results = results[:opts.MaxResults]
	}

	return results, nil
}

// PutDocument inserts or updates a memory document (upsert).
func (m *customMemoryDocumentsModel) PutDocument(ctx context.Context, agentID, path, content string) error {
	query := `INSERT INTO "public"."memory_documents" (agent_id, path, content, hash, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, NOW(), NOW())
	          ON CONFLICT (agent_id, path)
	          DO UPDATE SET content = EXCLUDED.content, hash = EXCLUDED.hash, updated_at = NOW()`

	hash := hashContent(content)
	_, err := m.conn.ExecCtx(ctx, query, agentID, path, content, hash)
	return err
}

// DeleteDocument removes a memory document.
func (m *customMemoryDocumentsModel) DeleteDocument(ctx context.Context, agentID, path string) error {
	query := `DELETE FROM "public"."memory_documents" WHERE agent_id = $1 AND path = $2`
	res, err := m.conn.ExecCtx(ctx, query, agentID, path)
	if err != nil {
		return err
	}
	// Check if anything was deleted
	if rowsAffected, err := res.RowsAffected(); err == nil && rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ListDocuments returns documents for an agent.
func (m *customMemoryDocumentsModel) ListDocuments(ctx context.Context, agentID string) ([]DocumentInfo, error) {
	query := `SELECT path, hash, agent_id, EXTRACT(EPOCH FROM updated_at)::bigint AS updated_at
	         FROM "public"."memory_documents"
	         WHERE agent_id = $1
	         ORDER BY updated_at DESC`

	var docs []DocumentInfo
	err := m.conn.QueryRowsCtx(ctx, &docs, query, agentID)
	if err != nil {
		return nil, err
	}
	return docs, nil
}

// ListAllDocumentsGlobal returns all documents across all agents (admin view).
func (m *customMemoryDocumentsModel) ListAllDocumentsGlobal(ctx context.Context) ([]DocumentInfo, error) {
	query := `SELECT agent_id, path, hash, EXTRACT(EPOCH FROM updated_at)::bigint AS updated_at
	          FROM "public"."memory_documents"
	          ORDER BY updated_at DESC`

	var docs []DocumentInfo
	err := m.conn.QueryRowsCtx(ctx, &docs, query)
	if err != nil {
		return nil, err
	}
	return docs, nil
}

// ListAllDocuments returns all documents for an agent.
func (m *customMemoryDocumentsModel) ListAllDocuments(ctx context.Context, agentID string) ([]DocumentInfo, error) {
	query := `SELECT agent_id, path, hash, EXTRACT(EPOCH FROM updated_at)::bigint AS updated_at
	          FROM "public"."memory_documents"
	          WHERE agent_id = $1
	          ORDER BY updated_at DESC`

	var docs []DocumentInfo
	err := m.conn.QueryRowsCtx(ctx, &docs, query, agentID)
	if err != nil {
		return nil, err
	}
	return docs, nil
}

// GetDocumentDetail returns full document info including stats.
func (m *customMemoryDocumentsModel) GetDocumentDetail(ctx context.Context, agentID, path string) (*DocumentDetail, error) {
	query := `SELECT path, content, hash,
	                 COALESCE((SELECT COUNT(*) FROM "public"."memory_chunks" WHERE path = $2), 0) AS chunk_count,
	                 COALESCE((SELECT COUNT(*) FROM "public"."memory_chunks" WHERE path = $2 AND embedding IS NOT NULL), 0) AS embedded_count,
	                 EXTRACT(EPOCH FROM created_at)::bigint AS created_at,
	                 EXTRACT(EPOCH FROM updated_at)::bigint AS updated_at
	          FROM "public"."memory_documents"
	          WHERE agent_id = $1 AND path = $2`

	var detail DocumentDetail
	err := m.conn.QueryRowCtx(ctx, &detail, query, agentID, path)
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

// ListChunks returns chunks for a document.
func (m *customMemoryDocumentsModel) ListChunks(ctx context.Context, agentID, path string) ([]ChunkInfo, error) {
	query := `SELECT id, start_line, end_line, text_preview,
	                 (embedding IS NOT NULL) AS has_embedding
	          FROM "public"."memory_chunks"
	          WHERE path = $1 AND agent_id = $2
	          ORDER BY start_line`

	var chunks []ChunkInfo
	err := m.conn.QueryRowsCtx(ctx, &chunks, query, path, agentID)
	if err != nil {
		return nil, err
	}
	return chunks, nil
}

// IndexDocument prepares a document for search (placeholder for now).
// Full implementation would chunk text and generate embeddings.
func (m *customMemoryDocumentsModel) IndexDocument(ctx context.Context, agentID, path string) error {
	// Placeholder: in a full implementation, would chunk and embed
	// For now, just mark as indexed by updating hash
	return nil
}

// IndexAll indexes all documents for an agent (placeholder).
func (m *customMemoryDocumentsModel) IndexAll(ctx context.Context, agentID string) error {
	docs, err := m.ListDocuments(ctx, agentID)
	if err != nil {
		return err
	}
	for _, doc := range docs {
		if err := m.IndexDocument(ctx, agentID, doc.Path); err != nil {
			return err
		}
	}
	return nil
}

// ReadLongTerm reads recent document contents for an agent, ordered by recency.
func (m *customMemoryDocumentsModel) ReadLongTerm(ctx context.Context, agentID string) ([]string, error) {
	query := `SELECT content FROM "public"."memory_documents"
	WHERE agent_id = $1
	ORDER BY updated_at DESC
	LIMIT 50`
	var results []string
	err := m.conn.QueryRowsCtx(ctx, &results, query, agentID)
	return results, err
}

// Recall returns recent documents for an agent from memory_documents.
func (m *customMemoryDocumentsModel) Recall(ctx context.Context, agentID string, limit int) ([]*MemoryDocuments, error) {
	query := `SELECT id, agent_id, path, content, hash, custom_scope, created_at, updated_at
	          FROM "public"."memory_documents"
	          WHERE agent_id = $1
	          ORDER BY updated_at DESC
	          LIMIT $2`
	var results []*MemoryDocuments
	err := m.conn.QueryRowsCtx(ctx, &results, query, agentID, limit)
	return results, err
}

// hashContent generates SHA256 hash of content for deduplication.
func hashContent(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h)
}
