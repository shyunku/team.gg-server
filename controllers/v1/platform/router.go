package platform

import (
	"github.com/gin-gonic/gin"
	"team.gg-server/controllers/middlewares"
)

func UsePlatformRouter(r *gin.RouterGroup) {
	g := r.Group("/platform")
	g.Use(middlewares.AuthMiddleware)
	UseCustomGameRouter(g)
}
