// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"github.com/pomclaw/pomclaw/internal/config"
	"github.com/pomclaw/pomclaw/pkg/agent"
	"github.com/pomclaw/pomclaw/pkg/contracts"
	"github.com/pomclaw/pomclaw/pkg/storage"
	"github.com/zeromicro/go-zero/core/logx"
)

type ServiceContext struct {
	Config config.Config

	Conn           storage.ConnectionManager
	SessionManager contracts.SessionManagerInterface
	Agent          *agent.AgentLoop
}

func NewServiceContext(c config.Config) *ServiceContext {

	//// 初始化 PostgreSQL 连接
	//psqlConn := postgres.New(c.PSQLConfig.DataSource)
	//_ = psqlConn

	//mysqlConn, err := mysql.NewConnectionManager(&c.MySQL)
	//if err != nil {
	//	panic(err)
	//}
	//

	// Connect using factory
	conn, err := storage.NewConnectionManager(&c)
	if err != nil {
		panic(err)
	}

	embSvc, err := storage.NewEmbeddingService(&c, conn.DB())
	if err != nil {
		logx.Error("agent", "Failed to create embedding service", map[string]interface{}{"error": err.Error()})
		panic(err)
	}
	logx.Info("agent", "Using embedding service", map[string]interface{}{"type": c.StorageType})

	stateStore := storage.NewStateStore(&c, conn.DB())
	memoryStore := storage.NewMemoryStore(&c, conn.DB(), embSvc)
	promptStoreRaw := storage.NewPromptStore(&c, conn.DB())
	sessionManager := storage.NewSessionStore(&c, conn.DB())

	a, err := agent.NewAgentLoop(&c, stateStore, memoryStore, promptStoreRaw, sessionManager)
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config: c,

		Conn:           conn,
		SessionManager: sessionManager,
		Agent:          a,
	}
}
