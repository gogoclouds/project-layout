package app

import (
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
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
	"sync"
	"time"
)

type Option func(o *options)

type options struct {
	conf *config.Service

	wg   *sync.WaitGroup
	exit chan struct{}

	id        string
	endpoints []*url.URL

	sigs            []os.Signal
	registrar       registry.ServiceRegistrar
	registryTimeout time.Duration
	httpServer      func(e *gin.Engine)
	rpcServer       func(s *grpc.Server)

	db    *gorm.DB
	redis redis.UniversalClient
}

func WithId(id string) Option {
	return func(o *options) {
		o.id = id
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

func WithDB(tables ...[]string) Option {
	// TODO gorm.AutoMerge
	return func(o *options) {
		newDB, err := db.NewDB(mysql.Open(o.conf.DB.Source), o.conf.DB)
		if err != nil {
			logger.Panic(err.Error())
		}
		o.db = newDB
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

func WithGinServer(router func(e *gin.Engine)) Option {
	return func(o *options) {
		o.httpServer = router
	}
}

func WithGrpcServer(svr func(rpcServer *grpc.Server)) Option {
	return func(o *options) {
		o.rpcServer = svr
	}
}
