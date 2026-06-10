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
	UserID    string `json:"user_id,omitempty" db:"user_id"`
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
	UserID        string `json:"user_id,omitempty" db:"user_id"`
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
		FindByAgentIdAndUserId(ctx context.Context, agentId string, userId string) ([]*MemoryDocuments, error)
		DeleteByAgentIdAndPath(ctx context.Context, agentId string, path string) error

		// Document CRUD
		GetDocument(ctx context.Context, agentID, userID, path string) (string, error)
		PutDocument(ctx context.Context, agentID, userID, path, content string) error
		DeleteDocument(ctx context.Context, agentID, userID, path string) error
		ListDocuments(ctx context.Context, agentID, userID string) ([]DocumentInfo, error)

		// Admin queries
		ListAllDocumentsGlobal(ctx context.Context) ([]DocumentInfo, error)
		ListAllDocuments(ctx context.Context, agentID string) ([]DocumentInfo, error)
		GetDocumentDetail(ctx context.Context, agentID, userID, path string) (*DocumentDetail, error)
		ListChunks(ctx context.Context, agentID, userID, path string) ([]ChunkInfo, error)

		// Search
		Search(ctx context.Context, query string, agentID, userID string, opts MemorySearchOptions) ([]MemorySearchResult, error)

		// Indexing
		IndexDocument(ctx context.Context, agentID, userID, path string) error
		IndexAll(ctx context.Context, agentID, userID string) error
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
	query := `select id, agent_id, user_id, path, content, hash, custom_scope, created_at, updated_at from "public"."memory_documents" where agent_id = $1 order by updated_at desc`
	err := m.conn.QueryRowsCtx(ctx, &resp, query, agentId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customMemoryDocumentsModel) Finds(ctx context.Context) ([]*MemoryDocuments, error) {
	var resp []*MemoryDocuments
	query := `select id, agent_id, user_id, path, content, hash, custom_scope, created_at, updated_at from "public"."memory_documents" order by updated_at desc`
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customMemoryDocumentsModel) FindByAgentIdAndUserId(ctx context.Context, agentId string, userId string) ([]*MemoryDocuments, error) {
	var resp []*MemoryDocuments
	query := `select id, agent_id, user_id, path, content, hash, custom_scope, created_at, updated_at from "public"."memory_documents" where agent_id = $1 and user_id = $2 order by updated_at desc`
	err := m.conn.QueryRowsCtx(ctx, &resp, query, agentId, userId)
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
// userID can be empty string for shared/global documents.
// Falls back to shared document (empty userID) if per-user document not found.
func (m *customMemoryDocumentsModel) GetDocument(ctx context.Context, agentID, userID, path string) (string, error) {
	if agentID == "" || path == "" {
		return "", nil
	}

	// Try per-user first if userID is provided
	if userID != "" {
		docs, err := m.FindByAgentIdAndUserId(ctx, agentID, userID)
		if err == nil && len(docs) > 0 {
			for _, doc := range docs {
				if doc.Path == path {
					return doc.Content, nil
				}
			}
		}
	}

	// Fallback to shared document (empty userID)
	docs, err := m.FindByAgentIdAndUserId(ctx, agentID, "")
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
func (m *customMemoryDocumentsModel) Search(ctx context.Context, query string, agentID, userID string, opts MemorySearchOptions) ([]MemorySearchResult, error) {
	if agentID == "" {
		return nil, nil
	}

	var docs []*MemoryDocuments
	var err error

	// Query per-user documents if userID provided, otherwise query by agentID only
	if userID != "" {
		docs, err = m.FindByAgentIdAndUserId(ctx, agentID, userID)
	} else {
		docs, err = m.FindByAgentId(ctx, agentID)
	}

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
// userID can be empty string for shared/global documents.
func (m *customMemoryDocumentsModel) PutDocument(ctx context.Context, agentID, userID, path, content string) error {
	query := `INSERT INTO "public"."memory_documents" (agent_id, user_id, path, content, hash, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	          ON CONFLICT (agent_id, COALESCE(user_id, ''), path)
	          DO UPDATE SET content = EXCLUDED.content, hash = EXCLUDED.hash, updated_at = NOW()`

	hash := hashContent(content)
	_, err := m.conn.ExecCtx(ctx, query, agentID, userID, path, content, hash)
	return err
}

// DeleteDocument removes a memory document.
// userID can be empty string for shared documents.
func (m *customMemoryDocumentsModel) DeleteDocument(ctx context.Context, agentID, userID, path string) error {
	query := `DELETE FROM "public"."memory_documents" WHERE agent_id = $1 AND COALESCE(user_id, '') = $2 AND path = $3`
	res, err := m.conn.ExecCtx(ctx, query, agentID, userID, path)
	if err != nil {
		return err
	}
	// Check if anything was deleted
	if rowsAffected, err := res.RowsAffected(); err == nil && rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ListDocuments returns documents for an agent with optional userID filter.
// If userID is empty, returns only shared documents.
// If userID is provided, returns both shared and user-specific documents.
func (m *customMemoryDocumentsModel) ListDocuments(ctx context.Context, agentID, userID string) ([]DocumentInfo, error) {
	var query string
	var args []any

	if userID == "" {
		// Only shared documents
		query = `SELECT path, hash, agent_id, user_id, EXTRACT(EPOCH FROM updated_at)::bigint AS updated_at
		         FROM "public"."memory_documents"
		         WHERE agent_id = $1 AND COALESCE(user_id, '') = ''
		         ORDER BY updated_at DESC`
		args = []any{agentID}
	} else {
		// Both shared and user-specific documents
		query = `SELECT path, hash, agent_id, user_id, EXTRACT(EPOCH FROM updated_at)::bigint AS updated_at
		         FROM "public"."memory_documents"
		         WHERE agent_id = $1 AND (COALESCE(user_id, '') = '' OR user_id = $2)
		         ORDER BY updated_at DESC`
		args = []any{agentID, userID}
	}

	var docs []DocumentInfo
	err := m.conn.QueryRowsCtx(ctx, &docs, query, args...)
	if err != nil {
		return nil, err
	}
	return docs, nil
}

// ListAllDocumentsGlobal returns all documents across all agents (admin view).
func (m *customMemoryDocumentsModel) ListAllDocumentsGlobal(ctx context.Context) ([]DocumentInfo, error) {
	query := `SELECT agent_id, path, hash, user_id, EXTRACT(EPOCH FROM updated_at)::bigint AS updated_at
	          FROM "public"."memory_documents"
	          ORDER BY updated_at DESC`

	var docs []DocumentInfo
	err := m.conn.QueryRowsCtx(ctx, &docs, query)
	if err != nil {
		return nil, err
	}
	return docs, nil
}

// ListAllDocuments returns all documents for an agent (shared + all user instances).
func (m *customMemoryDocumentsModel) ListAllDocuments(ctx context.Context, agentID string) ([]DocumentInfo, error) {
	query := `SELECT agent_id, path, hash, user_id, EXTRACT(EPOCH FROM updated_at)::bigint AS updated_at
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
func (m *customMemoryDocumentsModel) GetDocumentDetail(ctx context.Context, agentID, userID, path string) (*DocumentDetail, error) {
	query := `SELECT path, content, hash, user_id,
	                 COALESCE((SELECT COUNT(*) FROM "public"."memory_chunks" WHERE path = $3), 0) AS chunk_count,
	                 COALESCE((SELECT COUNT(*) FROM "public"."memory_chunks" WHERE path = $3 AND embedding IS NOT NULL), 0) AS embedded_count,
	                 EXTRACT(EPOCH FROM created_at)::bigint AS created_at,
	                 EXTRACT(EPOCH FROM updated_at)::bigint AS updated_at
	          FROM "public"."memory_documents"
	          WHERE agent_id = $1 AND COALESCE(user_id, '') = $2 AND path = $3`

	var detail DocumentDetail
	err := m.conn.QueryRowCtx(ctx, &detail, query, agentID, userID, path)
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

// ListChunks returns chunks for a document.
func (m *customMemoryDocumentsModel) ListChunks(ctx context.Context, agentID, userID, path string) ([]ChunkInfo, error) {
	query := `SELECT id, start_line, end_line, text_preview,
	                 (embedding IS NOT NULL) AS has_embedding
	          FROM "public"."memory_chunks"
	          WHERE path = $1 AND agent_id = $2 AND COALESCE(user_id, '') = $3
	          ORDER BY start_line`

	var chunks []ChunkInfo
	err := m.conn.QueryRowsCtx(ctx, &chunks, query, path, agentID, userID)
	if err != nil {
		return nil, err
	}
	return chunks, nil
}

// IndexDocument prepares a document for search (placeholder for now).
// Full implementation would chunk text and generate embeddings.
func (m *customMemoryDocumentsModel) IndexDocument(ctx context.Context, agentID, userID, path string) error {
	// Placeholder: in a full implementation, would chunk and embed
	// For now, just mark as indexed by updating hash
	return nil
}

// IndexAll indexes all documents for an agent (placeholder).
func (m *customMemoryDocumentsModel) IndexAll(ctx context.Context, agentID, userID string) error {
	docs, err := m.ListDocuments(ctx, agentID, userID)
	if err != nil {
		return err
	}
	for _, doc := range docs {
		if err := m.IndexDocument(ctx, agentID, userID, doc.Path); err != nil {
			return err
		}
	}
	return nil
}

// hashContent generates SHA256 hash of content for deduplication.
func hashContent(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h)
}
