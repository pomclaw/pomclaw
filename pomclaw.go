// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package main

import (
	"flag"
	"fmt"

	"github.com/pomclaw/pomclaw/internal/config"
	"github.com/pomclaw/pomclaw/internal/handler"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/pomclaw.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)

	server := rest.MustNewServer(c.RestConf, rest.WithCors())

	// Create Protocol v3 WebSocket gateway for real-time communication
	wsServer := handler.NewWSServer(ctx)

	sg := service.NewServiceGroup()
	sg.Add(server)
	sg.Add(wsServer)

	defer sg.Stop()

	// Register all HTTP routes including WebSocket
	handler.RegisterHandlers(server, ctx, wsServer)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	fmt.Printf("  - REST API: http://%s:%d/api\n", c.Host, c.Port)
	fmt.Printf("  - WebSocket: ws://%s:%d/ws\n", c.Host, c.Port)
	sg.Start()
}
