package util

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"team.gg-server/core"
)

func SetAccessTokenCookie(c *gin.Context, token string, refreshTokenExpireDuration int) {
	secureMode := !core.DebugMode
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie("accessToken", token, refreshTokenExpireDuration, "/", "", secureMode, true)
}

func DeleteAccessTokenCookie(c *gin.Context) {
	secureMode := !core.DebugMode
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie("accessToken", "", -1, "/", "", secureMode, true)
}
