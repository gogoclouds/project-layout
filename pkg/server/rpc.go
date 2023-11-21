package server

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"net"
	"sync"
)

func RunRpcServer(exit <-chan struct{}, wg *sync.WaitGroup, addr string, register func(server *grpc.Server)) {
	defer wg.Done()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()

	// 注册健康检查服务
	healthgrpc.RegisterHealthServer(s, health.NewServer())
	register(s)

	go func() {
		if err = s.Serve(lis); err != nil {
			panic(err)
		}
	}()
	<-exit
	s.GracefulStop() // 优雅停止
}

// RPC Dial

var rpcClientMap = make(map[string]*grpc.ClientConn)

func RpcDial(serverName string) (*grpc.ClientConn, error) {
	if cc, ok := rpcClientMap[serverName]; ok {
		state := cc.GetState()
		if state == connectivity.Ready {
			return cc, nil
		}
	}

	// conn, err := grpc.Dial(serverName, grpc.WithInsecure())
	conn, err := grpc.Dial(serverName, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	rpcClientMap[serverName] = conn
	return conn, nil
}