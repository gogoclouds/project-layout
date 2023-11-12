package app

import (
	"github.com/fsnotify/fsnotify"
	"github.com/gogoclouds/project-layout/config"
	"github.com/gogoclouds/project-layout/pkg/cache"
	"github.com/gogoclouds/project-layout/pkg/conf"
	"github.com/gogoclouds/project-layout/pkg/db"
	"github.com/gogoclouds/project-layout/pkg/logger"
	"github.com/gogoclouds/project-layout/pkg/registry"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/url"
	"os"
	"time"
)

type Option func(o *options)

type options struct {
	conf *config.Service

	id        string
	name      string
	endpoints []*url.URL

	sigs            []os.Signal
	registrar       registry.ServiceRegistrar
	registryTimeout time.Duration
	rpcServer       *grpc.Server

	db    *gorm.DB
	redis redis.UniversalClient
}

func WithId(id string) Option {
	return func(o *options) {
		o.id = id
	}
}

func WithName(name string) Option {
	return func(o *options) {
		o.name = name
	}
}

func WithEndpoints(endpoints []*url.URL) Option {
	return func(o *options) {
		o.endpoints = endpoints
	}
}

func WithSignal(sigs []os.Signal) Option {
	return func(o *options) {
		o.sigs = sigs
	}
}

func WithRegistrar(registrar registry.ServiceRegistrar) Option {
	return func(o *options) {
		o.registrar = registrar
	}
}

func WithRegistrarTimeout(rt time.Duration) Option {
	return func(o *options) {
		o.registryTimeout = rt
	}
}

func WithConfig(filename string) Option {
	return func(o *options) {
		var err error
		o.conf, err = conf.Load[config.Service](filename, func(e fsnotify.Event) {
			//logger.S(config.Conf.Logger.Level)
		})
		if err != nil {
			panic(err)
		}
	}
}

func WithLogger() Option {
	return func(o *options) {
		o.conf.Logger.Filename = o.conf.Name
		o.conf.Logger.TimeFormat = o.conf.TimeFormat
		logger.InitZapLogger(o.conf.Logger)
	}
}

func WithDB() Option {
	return func(o *options) {
		db, err := db.NewDB(mysql.Open(o.conf.DB.Source), o.conf.DB)
		if err != nil {
			logger.Panic(err.Error())
		}
		o.db = db
	}
}

func WithRedis() Option {
	return func(o *options) {
		newRedis, err := cache.NewRedis(o.conf.Redis)
		if err != nil {
			logger.Panic(err.Error())
		}
		o.redis = newRedis
	}
}

func WithGinServer() Option {
	return func(o *options) {
		// TODO
	}
}

func WithGrpcServer() Option {
	return func(o *options) {
		// TODO
	}
}