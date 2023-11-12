package config

import (
	"github.com/gogoclouds/project-layout/pkg/cache"
	"github.com/gogoclouds/project-layout/pkg/db"
	"github.com/gogoclouds/project-layout/pkg/enum"
	"github.com/gogoclouds/project-layout/pkg/logger"
)

var Conf *Service

type Service struct {
	Name       string       `yaml:"name"`    // 服务名
	Version    string       `yaml:"version"` // 版本号
	Env        enum.EnvType `yaml:"env"`
	TimeFormat string       `yaml:"timeFormat"`
	Server     struct {
		Http Transport `yaml:"http"`
		Rpc  Transport `yaml:"rpc"`
	}
	KV     KV              `yaml:"kv"`
	Logger logger.Config   `yaml:"logger"`
	DB     db.Config       `yaml:"db"`
	Redis  cache.RedisConf `yaml:"redis"`
}

// Transport 传输协议
type Transport struct {
	Addr    string `yaml:"addr"`    // 0.0.0.0:8000
	Timeout string `yaml:"timeout"` // 1s
}
