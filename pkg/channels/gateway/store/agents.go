package store

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Agent represents a row in pom_agents.
type Agent struct {
	ID           string
	UserID       string
	Name         string
	Description  string
	SystemPrompt string
	Model        string
	Tools        json.RawMessage
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

const agentCols = `id, user_id, name, description, system_prompt, model, tools, status, created_at, updated_at`

func scanAgent(row interface {
	Scan(...any) error
}) (*Agent, error) {
	a := &Agent{}
	err := row.Scan(&a.ID, &a.UserID, &a.Name, &a.Description, &a.SystemPrompt,
		&a.Model, &a.Tools, &a.Status, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}

// ListAgents returns all agents owned by the given user.
func ListAgents(db *sql.DB, userID string) ([]*Agent, error) {
	rows, err := db.Query(
		`SELECT `+agentCols+` FROM pom_agents WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []*Agent
	for rows.Next() {
		a, err := scanAgent(rows)
		if err != nil {
			return nil, err
		}
		agents = append(agents, a)
	}
	return agents, rows.Err()
}

// CreateAgent inserts a new agent and returns the created record.
func CreateAgent(db *sql.DB, userID, name, description, systemPrompt, model string, tools json.RawMessage) (*Agent, error) {
	agentID := GenerateID()
	return scanAgent(db.QueryRow(
		`INSERT INTO pom_agents (id, user_id, name, description, system_prompt, model, tools)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING `+agentCols,
		agentID, userID, name, description, systemPrompt, model, tools,
	))
}

// GetAgent returns the agent with the given id owned by the given user.
func GetAgent(db *sql.DB, id, userID string) (*Agent, error) {
	return scanAgent(db.QueryRow(
		`SELECT `+agentCols+` FROM pom_agents WHERE id = $1 AND user_id = $2`,
		id, userID,
	))
}

// UpdateAgent updates mutable fields and returns the updated record.
func UpdateAgent(db *sql.DB, id, userID, name, description, systemPrompt, model string, tools json.RawMessage) (*Agent, error) {
	return scanAgent(db.QueryRow(
		`UPDATE pom_agents
		 SET name=$3, description=$4, system_prompt=$5, model=$6, tools=$7, updated_at=NOW()
		 WHERE id=$1 AND user_id=$2
		 RETURNING `+agentCols,
		id, userID, name, description, systemPrompt, model, tools,
	))
}

// DeleteAgent removes an agent owned by the given user.
func DeleteAgent(db *sql.DB, id, userID string) error {
	result, err := db.Exec(
		`DELETE FROM pom_agents WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
