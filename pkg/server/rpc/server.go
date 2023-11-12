package rpc

import (
	"github.com/gogoclouds/project-layout/pkg/host"
	apimd "github.com/gogoclouds/project-layout/pkg/metadata"
	"github.com/gogoclouds/project-layout/pkg/server/rpc/serverinterceptors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"net"
	"net/url"
	"time"
)

type ServerOption func(s *Server)

type Server struct {
	*grpc.Server

	address            string
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
	grpcOptions        []grpc.ServerOption

	timeout  time.Duration
	listen   net.Listener
	health   *health.Server
	endpoint *url.URL
}

func WithAddress(address string) ServerOption {
	return func(s *Server) {
		s.address = address
	}
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		address: ":0",
		timeout: 1 * time.Second,
		health:  health.NewServer(),
	}
	for _, o := range opts {
		o(srv)
	}
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		serverinterceptors.UnaryRecoverInterceptor,
	}
	if srv.timeout > 0 {
		unaryInterceptors = append(unaryInterceptors, serverinterceptors.UnaryTimeoutInterceptor(srv.timeout))
	}
	if len(srv.unaryInterceptors) > 0 {
		unaryInterceptors = append(unaryInterceptors, srv.unaryInterceptors...)
	}
	grpcOpts := []grpc.ServerOption{grpc.ChainUnaryInterceptor(unaryInterceptors...)}
	if len(srv.grpcOptions) > 0 {
		grpcOpts = append(grpcOpts, srv.grpcOptions...)
	}
	srv.Server = grpc.NewServer(grpcOpts...)
	// 解析address
	if err := srv.listenAndEndpoint(); err != nil {
		return nil
	}
	// 注册 health
	grpc_health_v1.RegisterHealthServer(srv.Server, srv.health)
	// 可以支持用户通过grpc的一个接口查看当前支持的所有rpc服务
	apimd.RegisterMetadataServer(srv.Server, apimd.NewServer(srv.Server))
	reflection.Register(srv.Server)
	return srv
}

func (s *Server) listenAndEndpoint() error {
	if s.listen == nil {
		liste, err := net.Listen("tcp", s.address)
		if err != nil {
			return err
		}
		s.listen = liste
	}
	addr, err := host.Extract(s.address, s.listen)
	if err != nil {
		_ = s.listen.Close()
		return err
	}
	s.endpoint = &url.URL{Scheme: "grpc", Host: addr}
	return nil
}
