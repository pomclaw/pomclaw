package storage

import (
	"database/sql"
	"github.com/pomclaw/pomclaw/internal/config"
	"github.com/pomclaw/pomclaw/pkg/contracts"
	postgresdb "github.com/pomclaw/pomclaw/pkg/postgres"
)

// NewConnectionManager creates a ConnectionManager based on config.StorageType.
func NewConnectionManager(cfg *config.Config) (ConnectionManager, error) {
	return postgresdb.NewConnectionManager(&cfg.Postgres)
}

// NewEmbeddingService creates an EmbeddingService based on config.StorageType.
func NewEmbeddingService(cfg *config.Config, db *sql.DB) (EmbeddingService, error) {
	if cfg.Postgres.EmbeddingProvider == "api" && cfg.Postgres.EmbeddingAPIKey != "" {
		return postgresdb.NewAPIEmbeddingService(db, cfg.Postgres.EmbeddingAPIBase, cfg.Postgres.EmbeddingAPIKey, cfg.Postgres.EmbeddingModel), nil
	}
	return postgresdb.NewEmbeddingService(db), nil
}

// NewMemoryStore creates a MemoryStore based on config.StorageType.
func NewMemoryStore(cfg *config.Config, db *sql.DB, embSvc interface{}) contracts.SqlMemoryStore {
	return postgresdb.NewMemoryStore(db, "default", embSvc)
}

// NewSessionStore creates a SessionStore based on config.StorageType.
func NewSessionStore(cfg *config.Config, db *sql.DB) contracts.SessionManagerInterface {
	return postgresdb.NewSessionStore(db)
}

// NewPromptStore creates a PromptStore based on config.StorageType.
func NewPromptStore(cfg *config.Config, db *sql.DB) contracts.PromptStoreInterface {
	return postgresdb.NewPromptStore(db)
}
