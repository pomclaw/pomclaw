package store

import (
	"database/sql"
	"time"
)

// User represents a row in pom_users.
type User struct {
	ID        string
	Username  string
	Email     string
	Password  string // bcrypt hash
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateUser inserts a new user and returns the created record.
func CreateUser(db *sql.DB, username, email, passwordHash string) (*User, error) {
	u := &User{}
	err := db.QueryRow(
		`INSERT INTO pom_users (username, email, password)
		 VALUES ($1, $2, $3)
		 RETURNING id, username, email, status, created_at, updated_at`,
		username, email, passwordHash,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetUserByUsername returns the user with the given username, including the password hash.
func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	u := &User{}
	err := db.QueryRow(
		`SELECT id, username, email, password, status, created_at, updated_at
		 FROM pom_users WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}
