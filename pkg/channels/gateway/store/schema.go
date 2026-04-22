package store

import "database/sql"

var gatewaySchema = []string{
	`CREATE TABLE IF NOT EXISTS pom_users (
		id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		username     VARCHAR(64)  NOT NULL UNIQUE,
		email        VARCHAR(255) NOT NULL UNIQUE,
		password     VARCHAR(255) NOT NULL,
		status       VARCHAR(16)  NOT NULL DEFAULT 'active',
		created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
		updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
	)`,
	`CREATE TABLE IF NOT EXISTS pom_agents (
		id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id        UUID         NOT NULL REFERENCES pom_users(id) ON DELETE CASCADE,
		name           VARCHAR(255) NOT NULL,
		description    TEXT         NOT NULL DEFAULT '',
		system_prompt  TEXT         NOT NULL DEFAULT '',
		model          VARCHAR(64)  NOT NULL,
		tools          JSONB        NOT NULL DEFAULT '[]',
		status         VARCHAR(16)  NOT NULL DEFAULT 'active',
		created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
		updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
	)`,
	`CREATE TABLE IF NOT EXISTS pom_gateway_sessions (
		id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id    UUID         NOT NULL REFERENCES pom_users(id)  ON DELETE CASCADE,
		agent_id   UUID         NOT NULL REFERENCES pom_agents(id) ON DELETE CASCADE,
		title      VARCHAR(255) NOT NULL DEFAULT '',
		created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
	)`,
}

// InitGatewaySchema creates the gateway tables if they do not exist.
func InitGatewaySchema(db *sql.DB) error {
	for _, stmt := range gatewaySchema {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
