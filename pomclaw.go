// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package main

import (
	"flag"
	"fmt"
	"github.com/pomclaw/pomclaw/internal/config"
	"github.com/pomclaw/pomclaw/internal/handler"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/pkg/agent"
	"github.com/pomclaw/pomclaw/pkg/channels"
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

	server := rest.MustNewServer(c.RestConf)

	// Create agent loop powered by Eino framework
	agentLoop, err := agent.NewAgentLoop(&c, ctx.Conn.DB(), ctx.MsgBus)
	if err != nil {
		panic(err)
	}

	channelManager, err := channels.NewManager(&c, ctx.MsgBus)
	if err != nil {
		panic(err)
	}
	sg := service.NewServiceGroup()
	sg.Add(server)
	sg.Add(agentLoop)
	sg.Add(channelManager)

	defer sg.Stop()

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	sg.Start()
}
