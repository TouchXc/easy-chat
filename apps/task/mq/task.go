package main

import (
	"easy-chat/apps/task/mq/internal/config"
	"easy-chat/apps/task/mq/internal/handler"
	"easy-chat/apps/task/mq/internal/svc"
	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/dev/task.yaml", "the config file")

func main() {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)
	if err := c.SetUp(); err != nil {
		panic(err)
	}
	// 创建服务上下文。
	ctx := svc.NewServiceContext(c)
	// 初始化监听器，用于处理服务请求
	listen := handler.NewListen(ctx)
	// 创建服务组，用于统一管理和启动服务。
	serviceGroup := service.NewServiceGroup()
	// 将所有服务添加到服务组中。
	for _, s := range listen.Services() {
		serviceGroup.Add(s)
	}
	// 启动服务组，开始监听和处理请求。
	fmt.Println("Staring mqueue at...")
	serviceGroup.Start()

}
