// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"github.com/pomclaw/pomclaw/internal/config"
	"github.com/pomclaw/pomclaw/pkg/bus"
	"github.com/pomclaw/pomclaw/pkg/storage"
)

type ServiceContext struct {
	Config config.Config

	MsgBus *bus.MessageBus
	Conn   storage.ConnectionManager
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

	msgBus := bus.NewMessageBus()

	return &ServiceContext{
		Config: c,

		Conn:   conn,
		MsgBus: msgBus,
	}
}
