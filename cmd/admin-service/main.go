package main

import (
	"flag"
	"github.com/gogoclouds/project-layout/pkg/app"
	"github.com/gogoclouds/project-layout/pkg/conf"
	"github.com/gogoclouds/project-layout/pkg/logger"
)

var filepath = flag.String("config", "config/config.yaml", "config file path")

func init() {
	flag.String("name", "gogo-service", "service name")
	flag.String("env", "dev", "run config context")
	flag.String("logger.level", "info", "logger level")
	flag.Int("port", 8080, "http port 8080, rpc port 9080")
	conf.BindPFlags()
}

func main() {
	newApp := app.New(
		app.WithConfig(*filepath),
		app.WithLogger(),
		app.WithDB(),
		app.WithRedis(),
		//app.WithGinServer(),
		//app.WithGrpcServer(),
	)
	_ = newApp
	//if err := newApp.Run(); err != nil {
	//	logger.Panic(err.Error())
	//}
	logger.Info("service run ...")
	select {}
}
