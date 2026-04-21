package storage

import (
	"database/sql"
	"fmt"

	"github.com/pomclaw/pomclaw/pkg/agent"
	"github.com/pomclaw/pomclaw/pkg/config"
	oracledb "github.com/pomclaw/pomclaw/pkg/oracle"
	postgresdb "github.com/pomclaw/pomclaw/pkg/postgres"
)

// NewConnectionManager creates a ConnectionManager based on config.StorageType.
func NewConnectionManager(cfg *config.Config) (ConnectionManager, error) {
	storageType := cfg.StorageType

	switch storageType {
	case "postgres":
		return postgresdb.NewConnectionManager(&cfg.Postgres)
	case "oracle":
		return oracledb.NewConnectionManager(&cfg.Oracle)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", storageType)
	}
}

// InitSchema initializes the database schema based on config.StorageType.
func InitSchema(cfg *config.Config, db *sql.DB) error {
	storageType := cfg.StorageType

	switch storageType {
	case "postgres":
		return postgresdb.InitSchema(db)
	case "oracle":
		return oracledb.InitSchema(db)
	default:
		return fmt.Errorf("unknown storage type: %s", storageType)
	}
}

// NewEmbeddingService creates an EmbeddingService based on config.StorageType.
func NewEmbeddingService(cfg *config.Config, db *sql.DB) (EmbeddingService, error) {
	storageType := cfg.StorageType

	switch storageType {
	case "postgres":
		if cfg.Postgres.EmbeddingProvider == "api" && cfg.Postgres.EmbeddingAPIKey != "" {
			return postgresdb.NewAPIEmbeddingService(db, cfg.Postgres.EmbeddingAPIBase, cfg.Postgres.EmbeddingAPIKey, cfg.Postgres.EmbeddingModel), nil
		}
		return postgresdb.NewEmbeddingService(db), nil
	case "oracle":
		if cfg.Oracle.EmbeddingProvider == "api" && cfg.Oracle.EmbeddingAPIKey != "" {
			return oracledb.NewAPIEmbeddingService(db, cfg.Oracle.EmbeddingAPIBase, cfg.Oracle.EmbeddingAPIKey, cfg.Oracle.EmbeddingModel), nil
		}
		return oracledb.NewEmbeddingService(db, cfg.Oracle.ONNXModel)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", storageType)
	}
}

// NewMemoryStore creates a MemoryStore based on config.StorageType.
func NewMemoryStore(cfg *config.Config, db *sql.DB, embSvc interface{}) agent.OracleMemoryStore {
	storageType := cfg.StorageType

	var agentID string
	switch storageType {
	case "postgres":

		return postgresdb.NewMemoryStore(db, agentID, embSvc)
	case "oracle":
		agentID = cfg.Oracle.AgentID
		if agentID == "" {
			agentID = "default"
		}
		return oracledb.NewMemoryStore(db, agentID, embSvc)
	default:
		panic(fmt.Sprintf("unknown storage type: %s", storageType))
	}
}

// NewSessionStore creates a SessionStore based on config.StorageType.
func NewSessionStore(cfg *config.Config, db *sql.DB) agent.SessionManagerInterface {
	storageType := cfg.StorageType

	var agentID string
	switch storageType {
	case "postgres":

		return postgresdb.NewSessionStore(db, agentID)
	case "oracle":
		agentID = cfg.Oracle.AgentID
		if agentID == "" {
			agentID = "default"
		}
		return oracledb.NewSessionStore(db, agentID)
	default:
		panic(fmt.Sprintf("unknown storage type: %s", storageType))
	}
}

// NewStateStore creates a StateStore based on config.StorageType.
func NewStateStore(cfg *config.Config, db *sql.DB) agent.StateManagerInterface {
	storageType := cfg.StorageType

	var agentID string
	switch storageType {
	case "postgres":

		return postgresdb.NewStateStore(db, agentID)
	case "oracle":
		agentID = cfg.Oracle.AgentID
		if agentID == "" {
			agentID = "default"
		}
		return oracledb.NewStateStore(db, agentID)
	default:
		panic(fmt.Sprintf("unknown storage type: %s", storageType))
	}
}

// NewPromptStore creates a PromptStore based on config.StorageType.
func NewPromptStore(cfg *config.Config, db *sql.DB) agent.PromptStoreInterface {
	storageType := cfg.StorageType

	switch storageType {
	case "postgres":
		return postgresdb.NewPromptStore(db)

	default:
		panic(fmt.Sprintf("unknown storage type: %s", storageType))
	}
}

// GetAgentID returns the AgentID from config based on StorageType.
func GetAgentID(cfg *config.Config) string {
	storageType := cfg.StorageType

	switch storageType {
	case "oracle":
		if cfg.Oracle.AgentID != "" {
			return cfg.Oracle.AgentID
		}
	}
	return "default"
}
