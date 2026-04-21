package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/pomclaw/pomclaw/pkg/logger"
)

// Table DDL statements with POM_ prefix - PostgreSQL syntax
var tableDDL = map[string]string{
	"POM_META": `CREATE TABLE IF NOT EXISTS POM_META (
        meta_key   VARCHAR(255) PRIMARY KEY,
        meta_value VARCHAR(4000),
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    )`,

	"POM_MEMORIES": `CREATE TABLE IF NOT EXISTS POM_MEMORIES (
        memory_id    VARCHAR(64) PRIMARY KEY,
        agent_id     VARCHAR(64) NOT NULL,
        content      TEXT,
        embedding    vector(1536),
        importance   NUMERIC(3,2) DEFAULT 0.5,
        category     VARCHAR(255),
        access_count INTEGER DEFAULT 0,
        created_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        accessed_at  TIMESTAMP WITH TIME ZONE,
        updated_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    )`,

	"POM_DAILY_NOTES": `CREATE TABLE IF NOT EXISTS POM_DAILY_NOTES (
        note_id    VARCHAR(64) PRIMARY KEY,
        agent_id   VARCHAR(64) NOT NULL,
        note_date  DATE NOT NULL,
        content    TEXT,
        embedding  vector(1536),
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    )`,

	"POM_SESSIONS": `CREATE TABLE IF NOT EXISTS POM_SESSIONS (
        session_key VARCHAR(255) PRIMARY KEY,
        agent_id    VARCHAR(64) NOT NULL,
        messages    TEXT,
        summary     TEXT,
        created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    )`,

	"POM_STATE": `CREATE TABLE IF NOT EXISTS POM_STATE (
        state_key   VARCHAR(255) NOT NULL,
        agent_id    VARCHAR(64) NOT NULL,
        state_value VARCHAR(4000),
        updated_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (state_key, agent_id)
    )`,

	"POM_CONFIG": `CREATE TABLE IF NOT EXISTS POM_CONFIG (
        config_key   VARCHAR(255) NOT NULL,
        agent_id     VARCHAR(64) NOT NULL,
        config_value TEXT,
        updated_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (config_key, agent_id)
    )`,

	"POM_PROMPTS": `CREATE TABLE IF NOT EXISTS POM_PROMPTS (
        prompt_name VARCHAR(255) NOT NULL,
        agent_id    VARCHAR(64) NOT NULL,
        content     TEXT,
        updated_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (prompt_name, agent_id)
    )`,

	"POM_TRANSCRIPTS": `CREATE TABLE IF NOT EXISTS POM_TRANSCRIPTS (
        id           SERIAL PRIMARY KEY,
        session_key  VARCHAR(255),
        agent_id     VARCHAR(64),
        sequence_num INTEGER,
        role         VARCHAR(32),
        content      TEXT,
        created_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    )`,
}

// Regular index DDL
var indexDDL = []string{
	"CREATE INDEX IF NOT EXISTS IDX_POM_MEMORIES_AGENT ON POM_MEMORIES(agent_id)",
	"CREATE INDEX IF NOT EXISTS IDX_POM_DAILY_AGENT_DATE ON POM_DAILY_NOTES(agent_id, note_date)",
	"CREATE INDEX IF NOT EXISTS IDX_POM_SESSIONS_AGENT ON POM_SESSIONS(agent_id)",
	"CREATE INDEX IF NOT EXISTS IDX_POM_TRANSCRIPTS_SESSION ON POM_TRANSCRIPTS(session_key)",
	"CREATE INDEX IF NOT EXISTS IDX_POM_STATE_AGENT ON POM_STATE(agent_id)",
	"CREATE INDEX IF NOT EXISTS IDX_POM_MEMORIES_AGENT_CAT ON POM_MEMORIES(agent_id, category)",
}

// Vector index DDL - PostgreSQL with pgvector extension
var vectorIndexDDL = []string{
	"CREATE INDEX IF NOT EXISTS IDX_POM_MEMORIES_VEC ON POM_MEMORIES USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)",
	"CREATE INDEX IF NOT EXISTS IDX_POM_DAILY_NOTES_VEC ON POM_DAILY_NOTES USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)",
}

// InitSchema creates all tables and indexes idempotently.
func InitSchema(db *sql.DB) error {
	logger.InfoC("postgres", "Initializing PostgreSQL schema...")

	// Enable pgvector extension
	if _, err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector"); err != nil {
		logger.WarnCF("postgres", "Failed to create vector extension", map[string]interface{}{"error": err.Error()})
		// Continue - might already exist
	}

	// Create tables
	tableOrder := []string{
		"POM_META", "POM_MEMORIES", "POM_DAILY_NOTES", "POM_SESSIONS",
		"POM_STATE", "POM_CONFIG", "POM_PROMPTS", "POM_TRANSCRIPTS",
	}

	for _, tableName := range tableOrder {
		ddl := tableDDL[tableName]
		if _, err := db.Exec(ddl); err != nil {
			// Only fail on non-exists errors
			if !strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("failed to create table %s: %w", tableName, err)
			}
			logger.DebugCF("postgres", "Table already exists", map[string]interface{}{"table": tableName})
		} else {
			logger.InfoCF("postgres", "Created table", map[string]interface{}{"table": tableName})
		}
	}

	// Create regular indexes
	for _, ddl := range indexDDL {
		if _, err := db.Exec(ddl); err != nil {
			logger.WarnCF("postgres", "Index creation warning", map[string]interface{}{"error": err.Error()})
		}
	}

	// Create vector indexes
	for _, ddl := range vectorIndexDDL {
		if _, err := db.Exec(ddl); err != nil {
			logger.WarnCF("postgres", "Vector index creation warning", map[string]interface{}{"error": err.Error()})
		}
	}

	// Set schema version
	setSchemaVersion(db, "1.0.0")

	logger.InfoC("postgres", "Schema initialization complete")
	return nil
}

// setSchemaVersion updates or inserts the schema version in POM_META.
func setSchemaVersion(db *sql.DB, version string) {
	_, err := db.Exec(`
        INSERT INTO POM_META (meta_key, meta_value)
        VALUES ('schema_version', $1)
        ON CONFLICT (meta_key) DO UPDATE
        SET meta_value = $1, updated_at = CURRENT_TIMESTAMP
    `, version)
	if err != nil {
		logger.WarnCF("postgres", "Failed to set schema version", map[string]interface{}{"error": err.Error()})
	}
}
