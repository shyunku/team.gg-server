package controllers

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/shyunku-libraries/go-logger"
	"os"
	"team.gg-server/controllers/middlewares"
	"team.gg-server/controllers/socket"
	"team.gg-server/controllers/test"
	v1 "team.gg-server/controllers/v1"
	"team.gg-server/core"
)

var GlobalLogger = log.GetLogger()

func ping(c *gin.Context) {
	c.String(200, "pong")
}

func SetupRouter() *gin.Engine {
	gin.DefaultWriter = GlobalLogger
	gin.DefaultErrorWriter = GlobalLogger

	// setting cors
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:8080",
	}
	config.AllowCredentials = true
	//config.AllowHeaders = []string{
	//	"Origin",
	//	"Content-Length",
	//	"Content-Type",
	//	"Authorization",
	//}

	r := gin.Default()
	r.Use(cors.New(config))
	r.Use(middlewares.DefaultMiddleware)
	r.GET("/ping", ping)

	// platform routes
	v1.UseV1Router(r)
	if core.DebugMode {
		test.UseTestRouter(r)
	}
	socket.UseSocket(r)

	// 404
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": 404, "message": "Page not found"})
	})

	return r
}

func RunGin() {
	log.Infof("Starting server on port on %s...", core.AppServerPort)
	r := SetupRouter()
	if err := r.Run(fmt.Sprintf(":%s", core.AppServerPort)); err != nil {
		log.Fatal(err)
		os.Exit(-3)
	}
}
