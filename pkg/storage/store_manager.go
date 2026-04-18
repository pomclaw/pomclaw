package storage

import (
	"database/sql"
	"sync"

	"github.com/pomclaw/pomclaw/pkg/agent"
	"github.com/pomclaw/pomclaw/pkg/config"
	"github.com/pomclaw/pomclaw/pkg/logger"
	oracledb "github.com/pomclaw/pomclaw/pkg/oracle"
	postgresdb "github.com/pomclaw/pomclaw/pkg/postgres"
)

// StoreManager manages store instances for multiple agents
type StoreManager struct {
	db               *sql.DB
	cfg              *config.Config
	embSvc           EmbeddingService
	agentConfigStore *AgentConfigStore
	storageType      string

	// Store instance caches (agentID -> Store)
	sessionStores sync.Map
	stateStores   sync.Map
	memoryStores  sync.Map
	promptStores  sync.Map
}

// NewStoreManager creates a new store manager
func NewStoreManager(cfg *config.Config, db *sql.DB, embSvc EmbeddingService) *StoreManager {
	storageType := cfg.StorageType
	if storageType == "" {
		storageType = "oracle"
	}

	return &StoreManager{
		db:               db,
		cfg:              cfg,
		embSvc:           embSvc,
		agentConfigStore: NewAgentConfigStore(db, storageType),
		storageType:      storageType,
	}
}

// GetSessionStore returns a session store for the given agentID (lazy-loaded and cached)
func (sm *StoreManager) GetSessionStore(agentID string) (agent.SessionManagerInterface, error) {
	// Try to load from cache
	if store, ok := sm.sessionStores.Load(agentID); ok {
		return store.(agent.SessionManagerInterface), nil
	}

	// Create new store instance
	var store agent.SessionManagerInterface
	switch sm.storageType {
	case "postgres":
		store = postgresdb.NewSessionStore(sm.db, agentID)
	case "oracle":
		store = oracledb.NewSessionStore(sm.db, agentID)
	default:
		store = postgresdb.NewSessionStore(sm.db, agentID)
	}

	// Cache it
	sm.sessionStores.Store(agentID, store)

	logger.DebugCF("storage", "Created session store", map[string]interface{}{
		"agent_id": agentID,
		"type":     sm.storageType,
	})

	return store, nil
}

// GetStateStore returns a state store for the given agentID (lazy-loaded and cached)
func (sm *StoreManager) GetStateStore(agentID string) (agent.StateManagerInterface, error) {
	// Try to load from cache
	if store, ok := sm.stateStores.Load(agentID); ok {
		return store.(agent.StateManagerInterface), nil
	}

	// Create new store instance
	var store agent.StateManagerInterface
	switch sm.storageType {
	case "postgres":
		store = postgresdb.NewStateStore(sm.db, agentID)
	case "oracle":
		store = oracledb.NewStateStore(sm.db, agentID)
	default:
		store = postgresdb.NewStateStore(sm.db, agentID)
	}

	// Cache it
	sm.stateStores.Store(agentID, store)

	logger.DebugCF("storage", "Created state store", map[string]interface{}{
		"agent_id": agentID,
		"type":     sm.storageType,
	})

	return store, nil
}

// GetMemoryStore returns a memory store for the given agentID (lazy-loaded and cached)
func (sm *StoreManager) GetMemoryStore(agentID string) (agent.MemoryStoreInterface, error) {
	// Try to load from cache
	if store, ok := sm.memoryStores.Load(agentID); ok {
		return store.(agent.MemoryStoreInterface), nil
	}

	// Create new store instance
	var store agent.MemoryStoreInterface
	switch sm.storageType {
	case "postgres":
		store = postgresdb.NewMemoryStore(sm.db, agentID, sm.embSvc)
	case "oracle":
		store = oracledb.NewMemoryStore(sm.db, agentID, sm.embSvc)
	default:
		store = postgresdb.NewMemoryStore(sm.db, agentID, sm.embSvc)
	}

	// Cache it
	sm.memoryStores.Store(agentID, store)

	logger.DebugCF("storage", "Created memory store", map[string]interface{}{
		"agent_id": agentID,
		"type":     sm.storageType,
	})

	return store, nil
}

// GetPromptStore returns a prompt store for the given agentID (lazy-loaded and cached)
func (sm *StoreManager) GetPromptStore(agentID string) (agent.PromptStoreInterface, error) {
	// Try to load from cache
	if store, ok := sm.promptStores.Load(agentID); ok {
		return store.(agent.PromptStoreInterface), nil
	}

	// Create new store instance
	var store interface{}
	switch sm.storageType {
	case "postgres":
		store = postgresdb.NewPromptStore(sm.db, agentID)
	case "oracle":
		store = oracledb.NewPromptStore(sm.db, agentID)
	default:
		store = postgresdb.NewPromptStore(sm.db, agentID)
	}

	// Cache it
	sm.promptStores.Store(agentID, store)

	logger.DebugCF("storage", "Created prompt store", map[string]interface{}{
		"agent_id": agentID,
		"type":     sm.storageType,
	})

	return store.(agent.PromptStoreInterface), nil
}

// GetAgentConfig retrieves an agent configuration from the database
func (sm *StoreManager) GetAgentConfig(agentID string) (*agent.AgentConfig, error) {
	cfg, err := sm.agentConfigStore.GetByAgentID(agentID)
	if err != nil {
		return nil, err
	}
	return &cfg.AgentConfig, nil
}

// GetAgentConfigsByUserID retrieves all agent configurations for a user
func (sm *StoreManager) GetAgentConfigsByUserID(userID string) ([]*AgentConfig, error) {
	return sm.agentConfigStore.GetByUserID(userID)
}

// CreateAgentConfig creates a new agent configuration
func (sm *StoreManager) CreateAgentConfig(cfg *AgentConfig) error {
	return sm.agentConfigStore.Create(cfg)
}

// UpdateAgentConfig updates an agent configuration
func (sm *StoreManager) UpdateAgentConfig(cfg *AgentConfig) error {
	return sm.agentConfigStore.Update(cfg)
}

// DeleteAgentConfig deletes an agent configuration
func (sm *StoreManager) DeleteAgentConfig(agentID string) error {
	// Also clear the caches for this agentID
	sm.sessionStores.Delete(agentID)
	sm.stateStores.Delete(agentID)
	sm.memoryStores.Delete(agentID)
	sm.promptStores.Delete(agentID)

	return sm.agentConfigStore.Delete(agentID)
}

// ClearCache clears all cached stores for the given agentID
func (sm *StoreManager) ClearCache(agentID string) {
	sm.sessionStores.Delete(agentID)
	sm.stateStores.Delete(agentID)
	sm.memoryStores.Delete(agentID)
	sm.promptStores.Delete(agentID)

	logger.InfoCF("storage", "Cleared store cache", map[string]interface{}{
		"agent_id": agentID,
	})
}
