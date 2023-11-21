package app

import (
	"context"
	"os"
	"os/signal"
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

	if a.opts.rpcServer != nil {
		//a.opts.rpcServer.Serve()
	}

	// 注册服务
	if a.opts.registrar != nil {
		ctx, cancel := context.WithTimeout(context.Background(), a.opts.registryTimeout)
		defer cancel()
		if err := a.opts.registrar.Registry(ctx, instance); err != nil {
			logger.Errorf("register service error: %w", err)
			return err
		}
	}
	// 监听退出信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, a.opts.sigs...)
	<-c

	if err = a.Stop(); err != nil {
		logger.Errorf("stop service error: %w", err)
	}

	close(a.opts.exit) // 通知http、rpc服务退出信号

	// 1.等待 Http 服务结束退出
	// 2.等待 RPC 服务结束退出
	a.opts.wg.Wait()
	logger.Info("service has exited")
	return nil
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
	for _, e := range a.opts.endpoints {
		endpoints = append(endpoints, e.String())
	}
	return &registry.ServiceInstance{
		ID:        a.opts.id,
		Name:      a.opts.conf.Name,
		Version:   a.opts.conf.Version,
		Metadata:  nil,
		Endpoints: endpoints,
	}, nil
}
