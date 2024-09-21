package api

import (
	"github.com/gin-gonic/gin"
	log "github.com/shyunku-libraries/go-logger"
	"net/http"
	"team.gg-server/controllers/socket"
	"team.gg-server/libs/db"
	"team.gg-server/models"
	"team.gg-server/service"
	"team.gg-server/third_party/riot/api"
	"team.gg-server/util"
)

func UseApiRouter(r *gin.RouterGroup) {
	g := r.Group("/api")

	g.GET("/summonerPuuid", getSummonerPuuid)
	g.POST("/summonerLineFavor", setSummonerLineFavor)
}

func getSummonerPuuid(c *gin.Context) {
	var req GetSummonerPuuidRequestDto
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid request")
		return
	}

	tagLine := "KR1"
	if req.TagLine != nil {
		tagLine = *req.TagLine
	}

	if req.GameName == "" {
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid game name")
		return
	}

	var puuid string
	summonerDAO, exists, err := models.GetSummonerDAO_byNameTag(db.Root, req.GameName, tagLine)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}
	if !exists {
		account, status, err := api.GetAccountByRiotId(req.GameName, tagLine)
		if err != nil {
			if status == http.StatusNotFound {
				util.AbortWithStrJson(c, http.StatusNotFound, "summoner not found")
			} else {
				log.Error(err)
				util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
			}
			return
		}
		puuid = account.Puuid
	} else {
		puuid = summonerDAO.Puuid
	}

	c.JSON(http.StatusOK, puuid)
}

func setSummonerLineFavor(c *gin.Context) {
	var req SetSummonerLineFavorRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid request")
		return
	}

	tx, err := db.Root.BeginTxx(c, nil)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	candidateDAO, exists, err := models.GetCustomGameCandidateDAO_byPuuid(tx, req.CustomGameConfigId, req.Puuid)
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}
	if !exists {
		_ = tx.Rollback()
		util.AbortWithStrJson(c, http.StatusBadRequest, "candidate not found")
		return
	}

	if len(req.Strengths) != 5 {
		_ = tx.Rollback()
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid line favor length")
		return
	}

	for _, strength := range req.Strengths {
		if strength < -1 || strength > 2 {
			_ = tx.Rollback()
			util.AbortWithStrJson(c, http.StatusBadRequest, "invalid strength")
			return
		}
	}

	// update candidate
	candidateDAO.FlavorTop = req.Strengths[0]
	candidateDAO.FlavorJungle = req.Strengths[1]
	candidateDAO.FlavorMid = req.Strengths[2]
	candidateDAO.FlavorAdc = req.Strengths[3]
	candidateDAO.FlavorSupport = req.Strengths[4]

	if err := candidateDAO.Upsert(tx); err != nil {
		log.Error(err)
		_ = tx.Rollback()
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := service.RecalculateCustomGameBalance(tx, req.CustomGameConfigId); err != nil {
		log.Error(err)
		_ = tx.Rollback()
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := tx.Commit(); err != nil {
		log.Error(err)
		_ = tx.Rollback()
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	socket.SocketIO.BroadcastToCustomConfigRoom(req.CustomGameConfigId, socket.EventCustomConfigUpdated, nil)
	c.JSON(http.StatusOK, nil)
}
