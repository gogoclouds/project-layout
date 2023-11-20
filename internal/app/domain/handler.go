package domain

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gogoclouds/project-layout/api/admin/v1/helloworld"
	"google.golang.org/grpc"
)

func LoadRouter(e *gin.Engine) {
	e.MaxMultipartMemory = 300 << 20 //MB

	//noAuthRouterGroup := e.Group("")
	//admin.NoAuthRouterRegister(noAuthRouterGroup, g.DB)

	//authRouterGroup := e.Group("")
	//authRouterGroup.Use(middleware.JWTAuth())

	//admin.RouterRegister(authRouterGroup, g.DB)
}

func RegisterServer(server *grpc.Server) {
	helloworld.RegisterGreeterServer(server, &GreeterService{})
}

type GreeterService struct {
	helloworld.UnimplementedGreeterServer
}

func (h *GreeterService) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	//TODO implement me
	panic("implement me")
}

func (h *GreeterService) mustEmbedUnimplementedGreeterServer() {
	//TODO implement me
	panic("implement me")
}
