package middlewares

import (
	"github.com/gin-gonic/gin"
	log "github.com/shyunku-libraries/go-logger"
	"net/http"
	"team.gg-server/libs/auth"
	"team.gg-server/util"
)

func AuthMiddleware(c *gin.Context) {
	accessToken, err := c.Cookie("accessToken")
	if err != nil {
		log.Warn(err)
		util.AbortWithErrJson(c, http.StatusUnauthorized, err)
		return
	}

	userId, err := auth.ValidateToken(accessToken)
	if err != nil {
		log.Warn(err)
		util.AbortWithErrJson(c, http.StatusUnauthorized, err)
		return
	}

	c.Set("uid", userId)
	c.Request.Header.Set("uid", userId)
	c.Next()
}
