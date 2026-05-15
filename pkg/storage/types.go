package storage

import (
	"database/sql"

	"github.com/pomclaw/pomclaw/internal/config"
	"github.com/pomclaw/pomclaw/pkg/contracts"
)

// ConnectionManager is the generic interface for database connections.
type ConnectionManager interface {
	DB() *sql.DB
	Close() error
	Ping() error
}

// SchemaManager handles schema initialization.
type SchemaManager interface {
	InitSchema(db *sql.DB) error
}

// EmbeddingService handles vector embeddings (both API and in-database).
// This is used by both Oracle and PostgreSQL implementations.
type EmbeddingService interface {
	Embed(text string) ([]float32, error)
	Mode() string
}

// StorageFactory provides methods to create storage implementations based on config.
type StorageFactory interface {
	CreateConnectionManager(cfg *config.Config) (ConnectionManager, error)
	CreateMemoryStore(db *sql.DB, agentID string, embSvc EmbeddingService) contracts.SqlMemoryStore
	CreateSessionStore(db *sql.DB, agentID string) contracts.SessionManagerInterface
	CreatePromptStore(db *sql.DB, agentID string) interface{}
	InitSchema(db *sql.DB) error
}
