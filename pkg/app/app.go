package app

import (
	"context"
	"github.com/gogoclouds/project-layout/pkg/host"
	"github.com/gogoclouds/project-layout/pkg/server"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gogoclouds/project-layout/pkg/logger"
	"github.com/gogoclouds/project-layout/pkg/registry"
	"github.com/gogoclouds/project-layout/pkg/util"
)

type App struct {
	opts options

	mu sync.Mutex

	instance *registry.ServiceInstance
}

func New(opts ...Option) *App {
	o := options{
		wg:              &sync.WaitGroup{},
		id:              util.UUID(),
		sigs:            []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT},
		registryTimeout: 10 * time.Second,
		exit:            make(chan struct{}),
	}

	for _, opt := range opts {
		opt(&o)
	}
	return &App{
		opts: o,
	}
}

// Run run server
// 1.注册服务
// 2.退出相关组件或服务
func (a *App) Run() error {
	instance, err := a.buildInstance()
	if err != nil {
		return err
	}
	a.mu.Lock()
	a.instance = instance
	a.mu.Unlock()

	opts := a.opts

	if opts.httpServer != nil {
		go server.RunHttpServer(opts.exit, opts.wg, opts.conf.Server.Http.Addr, opts.httpServer)
	}

	if opts.rpcServer != nil {
		go server.RunRpcServer(opts.exit, opts.wg, opts.conf.Server.Rpc.Addr, opts.rpcServer)
	}

	ctx := context.Background()

	for _, fn := range a.opts.beforeStart {
		if err = fn(ctx); err != nil {
			return err
		}
	}

	// 注册服务
	if opts.registrar != nil {
		ctx, cancel := context.WithTimeout(context.Background(), opts.registryTimeout)
		defer cancel()
		if err := opts.registrar.Registry(ctx, instance); err != nil {
			logger.Errorf("register service error: %v", err)
			return err
		}
	}

	for _, fn := range a.opts.afterStart {
		if err = fn(ctx); err != nil {
			return err
		}
	}

	// 监听退出信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, opts.sigs...)
	<-c

	if err = a.Stop(); err != nil {
		logger.Errorf("stop service error: %v", err)
	}
	for _, fn := range a.opts.beforeStop {
		err = fn(ctx)
	}

	close(opts.exit) // 通知http、rpc服务退出信号

	// 1.等待 Http 服务结束退出
	// 2.等待 RPC 服务结束退出
	a.opts.wg.Wait()

	for _, fn := range a.opts.afterStop {
		err = fn(ctx)
	}

	opts.wg.Wait()
	logger.Info("service has exited")
	return err
}

// Stop stop server
// 1.注销服务
func (a *App) Stop() error {
	a.mu.Lock()
	instance := a.instance
	a.mu.Unlock()
	if a.opts.registrar != nil && instance != nil {
		ctx, cancel := context.WithTimeout(context.Background(), a.opts.registryTimeout)
		defer cancel()
		if err := a.opts.registrar.Deregister(ctx, instance); err != nil {
			logger.Errorf("deregister service error: %w", err)
			return err
		}
	}
	return nil
}

func (a *App) buildInstance() (*registry.ServiceInstance, error) {
	endpoints := make([]string, 0)
	httpScheme, grpcScheme := false, false
	for _, e := range a.opts.endpoints {
		switch strings.ToLower(e.Scheme) {
		case "https", "http":
			httpScheme = true
		case "grpc":
			grpcScheme = true
		}
		endpoints = append(endpoints, e.String())
	}
	if !httpScheme {
		if rUrl, err := getRegistryUrl("http", a.opts.conf.Server.Http.Addr); err == nil {
			endpoints = append(endpoints, rUrl)
		} else {
			logger.Errorf("get http registry err:%v", err)
		}
	}
	if !grpcScheme {
		if rUrl, err := getRegistryUrl("grpc", a.opts.conf.Server.Rpc.Addr); err == nil {
			endpoints = append(endpoints, rUrl)
		} else {
			logger.Errorf("get grpc registry err:%v", err)
		}
	}
	return &registry.ServiceInstance{
		ID:        a.opts.id,
		Name:      a.opts.conf.Name,
		Version:   a.opts.conf.Version,
		Metadata:  nil,
		Endpoints: endpoints,
	}, nil
}

func getRegistryUrl(scheme, addr string) (string, error) {
	ip, err := host.OutBoundIP()
	if err != nil {
		return "", err
	}
	_, ports, err := net.SplitHostPort(addr)
	if err != nil {
		return "", err
	}
	return scheme + "://" + net.JoinHostPort(ip, ports), nil
}
