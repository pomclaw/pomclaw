package storage

import (
	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/pkg/contracts"
	postgresdb "github.com/pomclaw/pomclaw/pkg/postgres"
)

// NewSessionStore creates a SessionStore based on config.StorageType.
func NewSessionStore(sessionsModel model.SessionsModel) contracts.SessionManagerInterface {
	return postgresdb.NewSessionStore(sessionsModel)
}

// NewPromptStore creates a PromptStore based on config.StorageType.
func NewPromptStore(promptsModel model.PromptsModel) contracts.PromptStoreInterface {
	return postgresdb.NewPromptStore(promptsModel)
}
