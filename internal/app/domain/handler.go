package domain

import (
	"github.com/gin-gonic/gin"
)

func LoadRouter(e *gin.Engine) {
	e.MaxMultipartMemory = 300 << 20 //MB

	//noAuthRouterGroup := e.Group("")
	//admin.NoAuthRouterRegister(noAuthRouterGroup, g.DB)

	//authRouterGroup := e.Group("")
	//authRouterGroup.Use(middleware.JWTAuth())

	//admin.RouterRegister(authRouterGroup, g.DB)
}
