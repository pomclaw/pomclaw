package oracle

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/pomclaw/pomclaw/pkg/logger"
)

// Table DDL statements with POM_ prefix
var tableDDL = map[string]string{
	"POM_META": `CREATE TABLE POM_META (
        meta_key   VARCHAR2(255) PRIMARY KEY,
        meta_value VARCHAR2(4000),
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`,

	"POM_MEMORIES": `CREATE TABLE POM_MEMORIES (
        memory_id    VARCHAR2(64) PRIMARY KEY,
        agent_id     VARCHAR2(64) NOT NULL,
        content      CLOB,
        embedding    VECTOR,
        importance   NUMBER(3,2) DEFAULT 0.5,
        category     VARCHAR2(255),
        access_count NUMBER DEFAULT 0,
        created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        accessed_at  TIMESTAMP,
        updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`,

	"POM_DAILY_NOTES": `CREATE TABLE POM_DAILY_NOTES (
        note_id    VARCHAR2(64) PRIMARY KEY,
        agent_id   VARCHAR2(64) NOT NULL,
        note_date  DATE NOT NULL,
        content    CLOB,
        embedding  VECTOR,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`,

	"POM_SESSIONS": `CREATE TABLE POM_SESSIONS (
        session_key VARCHAR2(255) PRIMARY KEY,
        agent_id    VARCHAR2(64) NOT NULL,
        messages    CLOB,
        summary     CLOB,
        created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`,

	"POM_STATE": `CREATE TABLE POM_STATE (
        state_key   VARCHAR2(255) NOT NULL,
        agent_id    VARCHAR2(64) NOT NULL,
        state_value VARCHAR2(4000),
        updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (state_key, agent_id)
    )`,

	"POM_CONFIG": `CREATE TABLE POM_CONFIG (
        config_key   VARCHAR2(255) NOT NULL,
        agent_id     VARCHAR2(64) NOT NULL,
        config_value CLOB,
        updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (config_key, agent_id)
    )`,

	"POM_PROMPTS": `CREATE TABLE POM_PROMPTS (
        prompt_name VARCHAR2(255) NOT NULL,
        agent_id    VARCHAR2(64) NOT NULL,
        content     CLOB,
        updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (prompt_name, agent_id)
    )`,

	"POM_TRANSCRIPTS": `CREATE TABLE POM_TRANSCRIPTS (
        id           NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        session_key  VARCHAR2(255),
        agent_id     VARCHAR2(64),
        sequence_num NUMBER,
        role         VARCHAR2(32),
        content      CLOB,
        created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`,
}

// Regular index DDL
var indexDDL = []string{
	"CREATE INDEX IDX_POM_MEMORIES_AGENT ON POM_MEMORIES(agent_id)",
	"CREATE INDEX IDX_POM_DAILY_AGENT_DATE ON POM_DAILY_NOTES(agent_id, note_date)",
	"CREATE INDEX IDX_POM_SESSIONS_AGENT ON POM_SESSIONS(agent_id)",
	"CREATE INDEX IDX_POM_TRANSCRIPTS_SESSION ON POM_TRANSCRIPTS(session_key)",
	"CREATE INDEX IDX_POM_STATE_AGENT ON POM_STATE(agent_id)",
	"CREATE INDEX IDX_POM_MEMORIES_AGENT_CAT ON POM_MEMORIES(agent_id, category)",
}

// Vector index DDL
var vectorIndexDDL = []string{
	`CREATE VECTOR INDEX IDX_POM_MEMORIES_VEC ON POM_MEMORIES(embedding)
     ORGANIZATION NEIGHBOR PARTITIONS
     DISTANCE COSINE
     WITH TARGET ACCURACY 95`,
	`CREATE VECTOR INDEX IDX_POM_DAILY_NOTES_VEC ON POM_DAILY_NOTES(embedding)
     ORGANIZATION NEIGHBOR PARTITIONS
     DISTANCE COSINE
     WITH TARGET ACCURACY 95`,
}

// InitSchema creates all tables and indexes idempotently.
func InitSchema(db *sql.DB) error {
	logger.InfoC("oracle", "Initializing Oracle schema...")

	// Create tables
	tableOrder := []string{
		"POM_META", "POM_MEMORIES", "POM_DAILY_NOTES", "POM_SESSIONS",
		"POM_STATE", "POM_CONFIG", "POM_PROMPTS", "POM_TRANSCRIPTS",
	}

	for _, tableName := range tableOrder {
		ddl := tableDDL[tableName]
		if _, err := db.Exec(ddl); err != nil {
			if isORA00955(err) {
				logger.DebugCF("oracle", "Table already exists", map[string]interface{}{"table": tableName})
			} else {
				return fmt.Errorf("failed to create table %s: %w", tableName, err)
			}
		} else {
			logger.InfoCF("oracle", "Created table", map[string]interface{}{"table": tableName})
		}
	}

	// Create regular indexes
	for _, ddl := range indexDDL {
		if _, err := db.Exec(ddl); err != nil {
			if isORA00955(err) || isORA01408(err) {
				// Index already exists
			} else {
				logger.WarnCF("oracle", "Index creation warning", map[string]interface{}{"error": err.Error()})
			}
		}
	}

	// Create vector indexes
	for _, ddl := range vectorIndexDDL {
		if _, err := db.Exec(ddl); err != nil {
			if isORA00955(err) || isORA01408(err) {
				// Index already exists
			} else {
				logger.WarnCF("oracle", "Vector index creation warning", map[string]interface{}{"error": err.Error()})
			}
		}
	}

	// Set schema version
	setSchemaVersion(db, "1.0.0")

	logger.InfoC("oracle", "Schema initialization complete")
	return nil
}

// setSchemaVersion updates or inserts the schema version in POM_META.
func setSchemaVersion(db *sql.DB, version string) {
	_, err := db.Exec(`
        MERGE INTO POM_META m
        USING (SELECT 'schema_version' AS meta_key FROM DUAL) s
        ON (m.meta_key = s.meta_key)
        WHEN MATCHED THEN
            UPDATE SET meta_value = :1, updated_at = CURRENT_TIMESTAMP
        WHEN NOT MATCHED THEN
            INSERT (meta_key, meta_value) VALUES ('schema_version', :2)
    `, version, version)
	if err != nil {
		logger.WarnCF("oracle", "Failed to set schema version", map[string]interface{}{"error": err.Error()})
	}
}

// isORA00955 checks if the error is ORA-00955 (name already used by existing object).
func isORA00955(err error) bool {
	return strings.Contains(err.Error(), "ORA-00955")
}

// isORA01408 checks if the error is ORA-01408 (such column list already indexed).
func isORA01408(err error) bool {
	return strings.Contains(err.Error(), "ORA-01408")
}
