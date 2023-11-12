package app

import (
	"context"
	"github.com/gogoclouds/project-layout/pkg/logger"
	"github.com/gogoclouds/project-layout/pkg/registry"
	"github.com/gogoclouds/project-layout/pkg/util"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type App struct {
	opts options

	mu sync.Mutex

	instance *registry.ServiceInstance
}

func New(opts ...Option) *App {
	o := options{
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
	return nil
}

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
		Name:      a.opts.name,
		Version:   "",
		Metadata:  nil,
		Endpoints: endpoints,
	}, nil
}
