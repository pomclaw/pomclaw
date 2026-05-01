package mysql

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pomclaw/pomclaw/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
)

type ConnectionManager struct {
	db  *sql.DB
	cfg *config.MySQLDBConfig
}

func NewConnectionManager(cfg *config.MySQLDBConfig) (*ConnectionManager, error) {
	dsn := BuildDSN(cfg)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.PoolMaxOpen)
	db.SetMaxIdleConns(cfg.PoolMaxIdle)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	cm := &ConnectionManager{db: db, cfg: cfg}

	if err := cm.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("MySQL ping failed: %w", err)
	}

	logx.Info("mysql", "Connection pool established")
	return cm, nil
}

func BuildDSN(cfg *config.MySQLDBConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
}

func (cm *ConnectionManager) DB() *sql.DB {
	return cm.db
}

func (cm *ConnectionManager) Ping() error {
	return cm.db.Ping()
}

func (cm *ConnectionManager) Close() error {
	logx.Info("mysql", "Closing connection pool")
	return cm.db.Close()
}
