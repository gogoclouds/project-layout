package main

import (
	"flag"
	"github.com/gogoclouds/project-layout/internal/app/domain"
	"github.com/gogoclouds/project-layout/pkg/app"
	"github.com/gogoclouds/project-layout/pkg/conf"
	"github.com/gogoclouds/project-layout/pkg/logger"
	"github.com/gogoclouds/project-layout/pkg/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"time"
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
	etcdClient := getEtcdClient()
	defer etcdClient.Close()

	newApp := app.New(
		app.WithConfig(*filepath),
		app.WithLogger(),
		app.WithDB(),
		app.WithRedis(),
		app.WithGinServer(domain.LoadRouter),
		app.WithGrpcServer(domain.RegisterServer),
		app.WithRegistrar(etcd.New(etcdClient)),
	)
	if err := newApp.Run(); err != nil {
		logger.Panic(err.Error())
	}
}

func getEtcdClient() *clientv3.Client {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second, DialOptions: []grpc.DialOption{grpc.WithBlock()},
	})
	if err != nil {
		panic(err)
	}
	return client
}
