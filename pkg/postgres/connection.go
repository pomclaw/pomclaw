package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pomclaw/pomclaw/pkg/config"
	"github.com/pomclaw/pomclaw/pkg/logger"
	_ "github.com/lib/pq"
)

// ConnectionManager wraps *sql.DB for PostgreSQL connectivity.
type ConnectionManager struct {
	db  *sql.DB
	cfg *config.PostgresDBConfig
}

// NewConnectionManager creates a new ConnectionManager and establishes the pool.
func NewConnectionManager(cfg *config.PostgresDBConfig) (*ConnectionManager, error) {
	connStr := BuildConnStr(cfg)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.PoolMaxOpen)
	db.SetMaxIdleConns(cfg.PoolMaxIdle)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	cm := &ConnectionManager{
		db:  db,
		cfg: cfg,
	}

	if err := cm.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("PostgreSQL ping failed: %w", err)
	}

	logger.InfoC("postgres", "Connection pool established")
	return cm, nil
}

// BuildConnStr constructs the PostgreSQL connection string.
func BuildConnStr(cfg *config.PostgresDBConfig) string {
	return fmt.Sprintf(
		"host=%s port=%d database=%s user=%s password=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.User,
		cfg.Password,
		cfg.SSLMode,
	)
}

// DB returns the underlying *sql.DB.
func (cm *ConnectionManager) DB() *sql.DB {
	return cm.db
}

// Ping checks the PostgreSQL connection.
func (cm *ConnectionManager) Ping() error {
	return cm.db.Ping()
}

// Close closes the connection pool.
func (cm *ConnectionManager) Close() error {
	logger.InfoC("postgres", "Closing connection pool")
	return cm.db.Close()
}

// WithTx executes a function within a transaction.
func (cm *ConnectionManager) WithTx(fn func(tx *sql.Tx) error) error {
	tx, err := cm.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	return tx.Commit()
}
