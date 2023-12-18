package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/shyunku-libraries/go-logger"
	"net/http"
	"team.gg-server/libs/auth"
	"team.gg-server/libs/db"
	"team.gg-server/models"
	"team.gg-server/util"
)

func UseAuthRouter(r *gin.RouterGroup) {
	g := r.Group("/auth")

	g.POST("/login", Login)
	g.POST("/signup", Signup)
	g.POST("/logout", Logout)
	g.GET("/isAuthorized", IsAuthorized)
}

func Login(c *gin.Context) {
	var req LoginRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid request")
	}

	// check if user exists
	comparablePw := util.Sha256(req.UserId + req.EncryptedPassword)
	userDAO, exists, err := models.GetUserDAO_byUserId_withPassword(db.Root, req.UserId, comparablePw)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}
	if !exists {
		util.AbortWithStrJson(c, http.StatusUnauthorized, "user not found")
		return
	}

	// create auth token
	authTokenBundle, err := auth.CreateAuthToken(userDAO.Uid)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	// save refresh token to in-memory
	if err := auth.SaveRefreshToken(userDAO.Uid, authTokenBundle.RefreshToken); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "failed to save refresh token")
	}

	// save on cookie
	accessTokenExpireDuration, err := auth.GetAccessTokenExpireDuration()
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	c.SetCookie("accessToken", authTokenBundle.AccessToken.Token, int(accessTokenExpireDuration.Seconds()),
		"/", "", false, true)

	resp := LoginResponseDto{
		Uid:    userDAO.Uid,
		UserId: userDAO.UserId,
	}

	c.JSON(http.StatusOK, resp)
}

func Signup(c *gin.Context) {
	var req SignupRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid request")
	}

	if len(req.UserId) < 4 {
		util.AbortWithStrJson(c, http.StatusBadRequest, "user id length must be greater than 4")
		return
	}

	// check if user exists
	_, exists, err := models.GetUserDAO_byUserId(db.Root, req.UserId)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}
	if exists {
		util.AbortWithStrJson(c, http.StatusConflict, "user already exists")
		return
	}

	// create user
	uid := uuid.New().String()
	userDAO := models.UserDAO{
		Uid:               uid,
		UserId:            req.UserId,
		EncryptedPassword: util.Sha256(req.UserId + req.EncryptedPassword),
	}
	if err := userDAO.Upsert(db.Root); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	c.JSON(http.StatusOK, nil)
}

func Logout(c *gin.Context) { // delete cookie
	c.SetCookie("accessToken", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, nil)
}

func IsAuthorized(c *gin.Context) {
	// with cookie
	accessToken, err := c.Cookie("access_token")
	if err != nil {
		util.AbortWithStrJson(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// check if token is valid
	_, err = auth.ValidateToken(accessToken)
	if err != nil {
		util.AbortWithStrJson(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	c.JSON(http.StatusOK, nil)
}
