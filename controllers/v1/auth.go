package v1

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/shyunku-libraries/go-logger"
	"net/http"
	"net/url"
	"strings"
	util2 "team.gg-server/controllers/util"
	"team.gg-server/core"
	"team.gg-server/libs/auth"
	"team.gg-server/libs/db"
	"team.gg-server/models"
	"team.gg-server/service"
	"team.gg-server/util"
)

func UseAuthRouter(r *gin.RouterGroup) {
	g := r.Group("/auth")

	g.POST("/login", Login)
	g.POST("/signup", Signup)
	g.POST("/logout", Logout)
	g.GET("/rsoLogin", RsoLogin)
	g.GET("/rsoLogout", RsoLogout)
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
	refreshTokenExpireDuration, err := auth.GetRefreshTokenExpireDuration()
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}
	util2.SetAccessTokenCookie(c, authTokenBundle.AccessToken.Token, int(refreshTokenExpireDuration.Seconds()))

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

func Logout(c *gin.Context) {
	// delete cookie
	util2.DeleteAccessTokenCookie(c)
	c.JSON(http.StatusOK, nil)
}

func RsoLogin(c *gin.Context) {
	codeRaw, codeExists := c.GetQuery("code")
	if !codeExists {
		util.AbortWithStrJson(c, http.StatusBadRequest, "code not found")
		return
	}
	code, err := url.QueryUnescape(codeRaw)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	issRaw, issExists := c.GetQuery("iss")
	if !issExists {
		util.AbortWithStrJson(c, http.StatusBadRequest, "iss not found")
		return
	}
	iss, err := url.QueryUnescape(issRaw)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	stateRaw, stateExists := c.GetQuery("state")
	if !stateExists {
		util.AbortWithStrJson(c, http.StatusBadRequest, "state not found")
		return
	}
	state, err := url.QueryUnescape(stateRaw)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}
	splitedStates := strings.Split(state, "|")
	if len(splitedStates) != 2 {
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid state")
		return
	}
	platform := splitedStates[0]
	token := splitedStates[1]

	//sessionState, sessionStateExists := c.GetQuery("session_state")
	//if !sessionStateExists {
	//	util.AbortWithStrJson(c, http.StatusBadRequest, "session_state not found")
	//	return
	//}

	if iss != "https://auth.riotgames.com" {
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid iss")
		return
	}

	// send request
	tokenAuthUrl := "https://auth.riotgames.com/token"
	callbackUri := core.RsoClientCallbackUri

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", callbackUri)

	req, err := http.NewRequest("POST", tokenAuthUrl, strings.NewReader(data.Encode()))
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	username := core.RsoClientId
	password := core.RsoClientSecret
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	errorRaw, errorExists := result["error"]
	if errorExists {
		errorDesc, errorDescExists := result["error_description"]
		if errorDescExists {
			log.Error(errorDesc.(string))
		}
		util.AbortWithStrJson(c, http.StatusInternalServerError, errorRaw.(string))
		return
	}

	accessTokenRaw, accessTokenExists := result["access_token"]
	if !accessTokenExists {
		util.AbortWithStrJson(c, http.StatusInternalServerError, "access_token not found")
		return
	}

	accessToken := accessTokenRaw.(string)

	// identify user
	req, err = http.NewRequest("GET", "https://asia.api.riotgames.com/riot/account/v1/accounts/me", nil)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	puuidRaw, puuidExists := userInfo["puuid"]
	if !puuidExists {
		util.AbortWithStrJson(c, http.StatusInternalServerError, "puuid not found")
		return
	}
	puuid := puuidRaw.(string)

	// find summoner
	_, exists, err := models.GetSummonerDAO_byPuuid(db.Root, puuid)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}
	if !exists {
		// fetch summoner info
		if _, _, err = service.RenewSummonerInfoByPuuid(db.Root, puuid); err != nil {
			log.Error(err)
			util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	thirdPartyIntegrationDAO := models.ThirdPartyIntegrationDAO{
		Puuid:    puuid,
		Platform: platform,
		Token:    token,
	}
	if err := thirdPartyIntegrationDAO.Upsert(db.Root); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	c.Redirect(http.StatusMovedPermanently, "https://team-gg.net/#/oauth_complete")
}

func RsoLogout(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}
