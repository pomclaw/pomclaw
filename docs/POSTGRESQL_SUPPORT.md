# PostgreSQL Support in Pomclaw

## Overview

Pomclaw now supports **PostgreSQL** alongside Oracle AI Database through a flexible **factory pattern** architecture. This allows you to choose between Oracle and PostgreSQL as your storage backend while maintaining the same API and functionality.

## Architecture

### Design Pattern: Factory Pattern

The storage layer is abstracted through `pkg/storage/` which provides factory functions that route requests to the appropriate database implementation (Oracle or PostgreSQL) based on configuration.

```
┌─────────────────────────────────────┐
│     Application Layer               │
│  (agent, gateway, commands)         │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│   Storage Factory (pkg/storage/)    │
│  - ConnectionManager()              │
│  - MemoryStore()                    │
│  - SessionStore()                   │
│  - StateStore()                     │
│  - PromptStore()                    │
└──┬──────────────────────────┬───────┘
   │                          │
   ▼                          ▼
┌──────────────┐      ┌──────────────┐
│   Oracle     │      │ PostgreSQL   │
│  pkg/oracle/ │      │ pkg/postgres/│
└──────────────┘      └──────────────┘
```

## Configuration

### Environment Variables

```bash
# Storage Type Selection (default: oracle)
POM_STORAGE_TYPE=postgres

# PostgreSQL Connection Parameters
POM_POSTGRES_ENABLED=true
POM_POSTGRES_HOST=localhost
POM_POSTGRES_PORT=5432
POM_POSTGRES_DATABASE=pomclaw
POM_POSTGRES_USER=postgres
POM_POSTGRES_PASSWORD=your_password
POM_POSTGRES_SSL_MODE=disable

# PostgreSQL Agent ID (optional, defaults to 'default')
POM_POSTGRES_AGENT_ID=default

# Embedding Configuration (required)
POM_POSTGRES_EMBEDDING_PROVIDER=api
POM_POSTGRES_EMBEDDING_API_BASE=http://localhost:11434/api
POM_POSTGRES_EMBEDDING_API_KEY=optional
POM_POSTGRES_EMBEDDING_MODEL=nomic-embed-text:latest
```

### Config File Format (config.json)

```json
{
  "storageType": "postgres",
  "postgres": {
    "enabled": true,
    "host": "localhost",
    "port": 5432,
    "database": "pomclaw",
    "user": "postgres",
    "password": "password",
    "sslMode": "disable",
    "agentId": "default",
    "embeddingProvider": "api",
    "embeddingApiBase": "http://localhost:11434/api",
    "embeddingModel": "nomic-embed-text:latest"
  }
}
```

## Setup

### 1. Create PostgreSQL Database

```bash
# Using Docker
docker run --name pomclaw-postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=pomclaw \
  -p 5432:5432 \
  postgres:16-alpine

# Or connect to existing PostgreSQL instance
createdb -U postgres pomclaw
```

### 2. Install pgvector Extension (Optional)

For advanced vector search optimization:

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

### 3. Initialize Database Schema

```bash
./build/pomclaw setup-database
```

This command:
- Detects storage type from config (Oracle or PostgreSQL)
- Creates all necessary tables automatically
- Seeds workspace configuration
- Sets up vector search indexes

### 4. Run Agent

```bash
./build/pomclaw agent
```

## SQL Differences Handled

| Aspect | Oracle | PostgreSQL | How It's Handled |
|--------|--------|-----------|------------------|
| **Driver** | go-ora/v2 | lib/pq | Factory selects correct driver |
| **Vector Type** | `VECTOR` | `vector` (pgvector) | Schema adapter |
| **Vector Distance** | `VECTOR_DISTANCE(v1, v2)` | `v1 <=> v2` | Query abstraction |
| **UPSERT** | `MERGE INTO ... WHEN MATCHED` | `INSERT ... ON CONFLICT` | SQL layer abstraction |
| **Parameter Binding** | `:1, :2, :3` | `$1, $2, $3` | sql.DB handles both |
| **CLOB** | `CLOB` | `TEXT` | Direct mapping |
| **VARCHAR** | `VARCHAR2(n)` | `VARCHAR(n)` | Text fields adapted |
| **Auto-increment** | `GENERATED ALWAYS AS IDENTITY` | `SERIAL` | Table DDL adapted |
| **Timestamps** | `TIMESTAMP` | `TIMESTAMP WITH TIME ZONE` | Both work via sql.DB |

## Database Schema

### Core Tables (PostgreSQL)

```sql
-- Memory Store (Vector Search)
CREATE TABLE memory (
    id VARCHAR(255) PRIMARY KEY,
    agent_id VARCHAR(255) NOT NULL,
    namespace VARCHAR(255) NOT NULL,
    embedding vector(384),
    data TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_agent FOREIGN KEY (agent_id) REFERENCES agent(id)
);

-- Session Store
CREATE TABLE session (
    id VARCHAR(255) PRIMARY KEY,
    agent_id VARCHAR(255) NOT NULL,
    data TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- State Store
CREATE TABLE state (
    id VARCHAR(255) PRIMARY KEY,
    agent_id VARCHAR(255) NOT NULL,
    data TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Vector indexes
CREATE INDEX idx_memory_vector ON memory USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);
CREATE INDEX idx_memory_agent_namespace ON memory(agent_id, namespace);
```

## Features

### Memory Management
- **Vector Search**: Powered by pgvector extension
- **Similarity Search**: Find related memories using embedding distance
- **Namespace Isolation**: Organize memories by context
- **Full ACID Support**: All operations transactional

### Session Persistence
- **Session Storage**: Persist agent conversation sessions
- **State Management**: Store and retrieve agent state
- **Configuration**: Save workspace and agent configuration

### Embedding Service
- **API-based**: Uses external API for embeddings (no built-in like Oracle's ONNX)
- **Configurable**: Support for any OpenAI-compatible API
- **Compatible**: Works with ollama, Anthropic API, etc.

## Implementation Details

### Factory Functions (`pkg/storage/factory.go`)

```go
// Returns appropriate connection manager
NewConnectionManager(cfg) ConnectionManager

// Returns embedding service
NewEmbeddingService(cfg, db) EmbeddingService

// Returns storage implementations
NewMemoryStore() MemoryStore
NewSessionStore() SessionStore
NewStateStore() StateStore
NewPromptStore() PromptStore

// Initialize schema (database-agnostic)
InitSchema(ctx, db, cfg) error

// Get agent ID from config
GetAgentID(cfg) string
```

### PostgreSQL Packages

| Package | Purpose |
|---------|---------|
| `pkg/postgres/connection.go` | Connection pooling and lifecycle |
| `pkg/postgres/schema.go` | DDL statements for table creation |
| `pkg/postgres/memory_store.go` | Vector search implementation |
| `pkg/postgres/session_store.go` | Session persistence |
| `pkg/postgres/state_store.go` | Agent state storage |
| `pkg/postgres/config_store.go` | Configuration storage |
| `pkg/postgres/prompt_store.go` | Prompt templates and workspace |
| `pkg/postgres/embedding.go` | Vector embedding service |
| `pkg/postgres/vector_store.go` | Vector search utilities |

## Backward Compatibility

- **Default**: Oracle remains the default storage type (`StorageType = "oracle"`)
- **Commands**: Both `setup-oracle` and `setup-database` commands work
- **Config**: Existing Oracle configurations remain unchanged and functional
- **Opt-in**: PostgreSQL is entirely opt-in via configuration

## Migration Guide (Oracle → PostgreSQL)

### 1. Export Data from Oracle

```bash
# Export memory records
sqlplus user@oracle @export_memory.sql > memory_export.json

# Export sessions
sqlplus user@oracle @export_sessions.sql > sessions_export.json
```

### 2. Update Configuration

Edit `config.json`:
```json
{
  "storageType": "postgres",
  "postgres": { ... }
}
```

### 3. Initialize PostgreSQL Schema

```bash
./build/pomclaw setup-database
```

### 4. Import Data

```go
// Use storage factory to write to PostgreSQL backend
// Custom migration script needed based on your data structure
```

## Performance Considerations

### Vector Search Optimization

For optimal performance with pgvector:

1. **Index Strategy**
   ```sql
   -- IVFFLAT index (faster, approximate)
   CREATE INDEX idx_memory_vector ON memory
     USING ivfflat (embedding vector_cosine_ops)
     WITH (lists = 100);

   -- Or HNSW index (more accurate, slower)
   CREATE EXTENSION hnsw;
   CREATE INDEX idx_memory_vector ON memory
     USING hnsw (embedding vector_cosine_ops);
   ```

2. **Query Optimization**
   - Use appropriate distance operators: `<=>` (cosine), `<->` (L2), `<#>` (inner product)
   - Batch operations when possible
   - Monitor query execution plans

### Connection Pooling

```go
// Connection pool settings (configured in connection.go)
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

## Troubleshooting

### Issue: Connection refused
```
Error: dial tcp localhost:5432: connect: connection refused
```
**Solution**: Ensure PostgreSQL is running and accessible on the configured host/port.

### Issue: pgvector not available
```
Error: extension vector does not exist
```
**Solution**: Install pgvector extension:
```sql
CREATE EXTENSION vector;
```

### Issue: Embedding API timeout
```
Error: embedding service timeout
```
**Solution**:
- Check embedding API is running
- Increase timeout in config
- Verify `POM_POSTGRES_EMBEDDING_API_BASE` is correct

## Testing

### Unit Tests
```bash
go test ./pkg/postgres/...
go test ./pkg/storage/...
```

### Integration Tests
```bash
# Start test PostgreSQL container
docker-compose -f docker-compose.test.yml up

# Run tests
go test -tags=integration ./...
```

## Future Improvements

- [ ] MySQL support via same factory pattern
- [ ] SQLite support for embedded deployments
- [ ] Shared test matrix for both Oracle and PostgreSQL
- [ ] Query performance benchmarking tools
- [ ] Migration utilities for data transfer
- [ ] Read replica support for PostgreSQL

## References

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [pgvector Documentation](https://github.com/pgvector/pgvector)
- [Go sql/database Package](https://pkg.go.dev/database/sql)
- [lib/pq Driver](https://github.com/lib/pq)

## Support

For PostgreSQL-specific issues:
- Check PostgreSQL logs: `docker logs <container>`
- Verify network connectivity: `psql -h localhost -U postgres`
- Test embedding API: `curl http://localhost:11434/api/embeddings`

For general issues, refer to the main [README.md](../README.md).
