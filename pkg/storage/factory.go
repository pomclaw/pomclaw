package storage

import (
	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/pkg/contracts"
	postgresdb "github.com/pomclaw/pomclaw/pkg/postgres"
)

// NewMemoryStore creates a MemoryStore based on config.StorageType.
func NewMemoryStore(memoriesModel model.MemoriesModel, dailyNotesModel model.DailyNotesModel) contracts.SqlMemoryStore {
	return postgresdb.NewMemoryStore(memoriesModel, dailyNotesModel)
}

// NewSessionStore creates a SessionStore based on config.StorageType.
func NewSessionStore(sessionsModel model.SessionsModel) contracts.SessionManagerInterface {
	return postgresdb.NewSessionStore(sessionsModel)
}

// NewPromptStore creates a PromptStore based on config.StorageType.
func NewPromptStore(promptsModel model.PromptsModel) contracts.PromptStoreInterface {
	return postgresdb.NewPromptStore(promptsModel)
}
