# Storage Architecture

## Overview

Pomclaw uses a **factory pattern** to abstract database operations, supporting multiple storage backends while maintaining a consistent API. This document explains the storage layer architecture and how to extend it.

## Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
│  (cmd/pomclaw/main.go, agent, gateway)                 │
└────────────────────┬────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│            Storage Abstraction (pkg/storage/)               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ Types (Interfaces)                                    │  │
│  │  - ConnectionManager                                  │  │
│  │  - MemoryStore                                        │  │
│  │  - SessionStore                                       │  │
│  │  - StateStore                                         │  │
│  │  - PromptStore                                        │  │
│  │  - EmbeddingService                                   │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ Factory Functions                                     │  │
│  │  - NewConnectionManager(cfg)                          │  │
│  │  - NewMemoryStore()                                   │  │
│  │  - NewSessionStore()                                  │  │
│  │  - NewStateStore()                                    │  │
│  │  - NewPromptStore()                                   │  │
│  │  - NewEmbeddingService(cfg, db)                       │  │
│  │  - InitSchema(ctx, db, cfg)                           │  │
│  └───────────────────────────────────────────────────────┘  │
└────────┬──────────────────────────────────────────┬──────────┘
         │                                          │
         ▼                                          ▼
    ┌─────────────┐                         ┌──────────────┐
    │ Oracle      │                         │ PostgreSQL   │
    │ Backend     │                         │ Backend      │
    │ pkg/oracle/ │                         │ pkg/postgres/│
    └─────────────┘                         └──────────────┘
         │                                          │
         ▼                                          ▼
    ┌─────────────┐                         ┌──────────────┐
    │ Oracle      │                         │ PostgreSQL   │
    │ Database    │                         │ Database     │
    │ (ONNX       │                         │ (pgvector    │
    │  embeddings)│                         │  embeddings) │
    └─────────────┘                         └──────────────┘
```

## Core Interfaces

All storage backends must implement these interfaces:

### ConnectionManager

```go
type ConnectionManager interface {
    // Initialize schema/tables
    InitSchema(ctx context.Context) error

    // Get database connection
    GetDB() *sql.DB

    // Close connection
    Close() error
}
```

### MemoryStore

```go
type MemoryStore interface {
    // Create memory record with embedding
    Create(ctx context.Context, record *MemoryRecord) error

    // Search by similarity
    Search(ctx context.Context, embedding []float32, limit int) ([]MemoryRecord, error)

    // Get by ID
    Get(ctx context.Context, id string) (*MemoryRecord, error)

    // Update record
    Update(ctx context.Context, record *MemoryRecord) error

    // Delete record
    Delete(ctx context.Context, id string) error
}
```

### SessionStore

```go
type SessionStore interface {
    // Create session
    Create(ctx context.Context, session *Session) error

    // Get session
    Get(ctx context.Context, id string) (*Session, error)

    // Update session
    Update(ctx context.Context, session *Session) error

    // Delete session
    Delete(ctx context.Context, id string) error
}
```

### StateStore

```go
type StateStore interface {
    // Set state value
    Set(ctx context.Context, id string, state interface{}) error

    // Get state value
    Get(ctx context.Context, id string) (interface{}, error)

    // Delete state
    Delete(ctx context.Context, id string) error
}
```

### PromptStore

```go
type PromptStore interface {
    // Get prompt template
    GetPrompt(ctx context.Context, key string) (string, error)

    // List all prompts
    ListPrompts(ctx context.Context) ([]Prompt, error)
}
```

### EmbeddingService

```go
type EmbeddingService interface {
    // Generate embeddings for text
    Embed(ctx context.Context, text string) ([]float32, error)

    // Generate embeddings for multiple texts
    EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
}
```

## Configuration-Driven Selection

### How Storage Type is Selected

1. **Default**: Oracle (backward compatibility)
2. **Environment Variable**: `PICO_STORAGE_TYPE=postgres`
3. **Config File**: `"storageType": "postgres"` in config.json
4. **Command-line**: Could be extended for CLI flags

### Flow Diagram

```
┌──────────────────────┐
│ Load Configuration   │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────────────────┐
│ Check StorageType in Config      │
│ (default: "oracle")              │
└──────────┬───────────────────────┘
           │
      ┌────┴─────┐
      ▼          ▼
  oracle     postgres
      │          │
      ▼          ▼
Check          Check
Oracle.        Postgres.
Enabled        Enabled
      │          │
      ▼          ▼
   Init       Init
   Oracle    PostgreSQL
```

## Implementation Examples

### Adding a New Storage Backend (MySQL)

1. **Create package** `pkg/mysql/`

2. **Implement interfaces**:
   ```go
   // pkg/mysql/connection.go
   type ConnectionManager struct {
       db *sql.DB
   }

   func (m *ConnectionManager) InitSchema(ctx context.Context) error {
       // MySQL DDL statements
       ...
   }

   func (m *ConnectionManager) GetDB() *sql.DB {
       return m.db
   }

   func (m *ConnectionManager) Close() error {
       return m.db.Close()
   }
   ```

3. **Add factory cases** in `pkg/storage/factory.go`:
   ```go
   case "mysql":
       return mysql.NewConnectionManager(cfg)
   ```

4. **Add config struct** in `pkg/config/config.go`:
   ```go
   type MySQLConfig struct {
       Enabled          bool
       Host             string
       Port             int
       Database         string
       User             string
       Password         string
   }
   ```

### Using the Storage Layer in Code

```go
// Initialize storage with factory (automatically selects backend)
connMgr := storage.NewConnectionManager(cfg)
defer connMgr.Close()

// Get database connection
db := connMgr.GetDB()

// Create stores
memStore := storage.NewMemoryStore(db, cfg)
sessionStore := storage.NewSessionStore(db, cfg)

// Use stores (same API for any backend)
record := &storage.MemoryRecord{
    ID:        "mem-1",
    AgentID:   "agent-1",
    Embedding: embedding,
    Data:      jsonData,
}
err := memStore.Create(ctx, record)

// Vector search (works the same for Oracle or PostgreSQL)
results, err := memStore.Search(ctx, queryEmbedding, 10)
```

## Database-Agnostic Patterns

### 1. Parameter Binding

```go
// Don't write: "SELECT * FROM table WHERE id = ?" (MySQL-only)
// Instead:    "SELECT * FROM table WHERE id = $1" (both work with sql.DB)

// sql.Rows.Scan() and db.QueryRow() handle both :N and $N placeholders
err := db.QueryRowContext(ctx,
    "SELECT * FROM table WHERE id = $1",
    id).Scan(&result)
```

### 2. Schema Initialization

```go
// pkg/storage/factory.go: Database-agnostic schema setup
func InitSchema(ctx context.Context, db *sql.DB, cfg *config.Config) error {
    storageType := cfg.StorageType
    if storageType == "" {
        storageType = "oracle"
    }

    switch storageType {
    case "oracle":
        return oracle.InitSchema(ctx, db)
    case "postgres":
        return postgres.InitSchema(ctx, db)
    default:
        return fmt.Errorf("unknown storage type: %s", storageType)
    }
}
```

### 3. UPSERT Operations

**Oracle**:
```sql
MERGE INTO table t
USING (SELECT :id AS id, :data AS data FROM dual) s
ON (t.id = s.id)
WHEN MATCHED THEN UPDATE SET data = s.data
WHEN NOT MATCHED THEN INSERT (id, data) VALUES (s.id, s.data)
```

**PostgreSQL**:
```sql
INSERT INTO table (id, data) VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE SET data = $2
```

**Abstraction**:
```go
// Each backend implements Insert+Update logic differently,
// but both achieve the same result - upsert semantics
type MemoryStore interface {
    CreateOrUpdate(ctx context.Context, record *MemoryRecord) error
}
```

## Testing Strategy

### Unit Tests (Storage Factory)

```go
// pkg/storage/factory_test.go
func TestNewConnectionManager(t *testing.T) {
    tests := []struct {
        name        string
        storageType string
        expectedType interface{}
    }{
        {
            name:        "Oracle backend",
            storageType: "oracle",
            expectedType: (*oracle.ConnectionManager)(nil),
        },
        {
            name:        "PostgreSQL backend",
            storageType: "postgres",
            expectedType: (*postgres.ConnectionManager)(nil),
        },
    }
    // Test implementation...
}
```

### Integration Tests (Both Backends)

```go
// tests/integration_test.go
func TestMemoryStore(t *testing.T) {
    // Test with both Oracle and PostgreSQL containers
    backends := []string{"oracle", "postgres"}

    for _, backend := range backends {
        t.Run(backend, func(t *testing.T) {
            cfg := setupTestDB(backend)
            memStore := storage.NewMemoryStore(cfg)

            // Run same tests for both backends
            testCreateMemory(t, memStore)
            testSearchMemory(t, memStore)
            testUpdateMemory(t, memStore)
        })
    }
}
```

## Deployment Considerations

### Docker Compose (Multi-Backend Support)

```yaml
version: '3'
services:
  # Oracle Backend
  oracle-db:
    image: container-registry.oracle.com/database/free:latest
    environment:
      ORACLE_PWD: password
    ports:
      - "1521:1521"

  # PostgreSQL Backend
  postgres-db:
    image: postgres:16-alpine
    environment:
      POSTGRES_PASSWORD: password
      POSTGRES_DB: pomclaw
    ports:
      - "5432:5432"
    extensions:
      - image: pgvector/pgvector:latest

  # Pomclaw (Oracle)
  app-oracle:
    build: .
    environment:
      PICO_STORAGE_TYPE: oracle
      PICO_ORACLE_ENABLED: "true"
    depends_on:
      - oracle-db

  # Pomclaw (PostgreSQL)
  app-postgres:
    build: .
    environment:
      PICO_STORAGE_TYPE: postgres
      PICO_POSTGRES_ENABLED: "true"
    depends_on:
      - postgres-db
```

## Performance Characteristics

### Oracle vs PostgreSQL

| Operation | Oracle | PostgreSQL | Notes |
|-----------|--------|-----------|-------|
| **Vector Insert** | O(1) | O(log n) | Index insertion cost |
| **Vector Search** | O(log n) | O(log n) | pgvector IVFFLAT |
| **Session Create** | O(1) | O(1) | Direct insert |
| **Memory Search** | O(log n) | O(log n) | Both indexed |
| **Connection Pool** | 25 max | 25 max | Configurable |

### Vector Search Distance Operators

| Database | Operator | Type | Notes |
|----------|----------|------|-------|
| Oracle | `VECTOR_DISTANCE()` | Function | Euclidean distance |
| PostgreSQL | `<=>` | Operator | Cosine distance |
| PostgreSQL | `<->` | Operator | L2 distance |
| PostgreSQL | `<#>` | Operator | Inner product |

## Monitoring & Debugging

### Connection Pool Stats

```go
// Get connection pool statistics
stats := db.Stats()
fmt.Printf("Open connections: %d\n", stats.OpenConnections)
fmt.Printf("In-use connections: %d\n", stats.InUse)
fmt.Printf("Idle connections: %d\n", stats.Idle)
fmt.Printf("Wait count: %d\n", stats.WaitCount)
fmt.Printf("Wait duration: %v\n", stats.WaitDuration)
fmt.Printf("Max idle closed: %d\n", stats.MaxIdleClosed)
fmt.Printf("Max lifetime closed: %d\n", stats.MaxLifetimeClosed)
```

### Query Logging

```go
// Enable query logging in test/debug
import _ "github.com/golang-sql/sqlc/cmd/log"

// Or use custom hooks:
db.QueryRow = loggedQueryRow
```

## References

- [Go database/sql Package](https://pkg.go.dev/database/sql)
- [lib/pq Documentation](https://github.com/lib/pq)
- [go-ora/v2 Documentation](https://github.com/sijms/go-ora)
- [Factory Pattern Design](https://refactoring.guru/design-patterns/factory-method)

