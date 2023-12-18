package v1

import (
	"github.com/gin-gonic/gin"
	log "github.com/shyunku-libraries/go-logger"
	"net/http"
	"team.gg-server/controllers/v1/platform"
	"team.gg-server/libs/db"
	"team.gg-server/models"
	"team.gg-server/service"
	"team.gg-server/third_party/riot"
	"team.gg-server/util"
	"time"
)

func UseV1Router(r *gin.Engine) {
	g := r.Group("/v1")
	UseIconRouter(g)
	UseAuthRouter(g)
	platform.UsePlatformRouter(g)

	g.GET("/summoner", GetSummonerInfo)
	g.POST("/renewSummoner", RenewSummonerInfo)
	g.POST("/loadMatches", LoadMatches)
	g.GET("/ingame", GetIngameInfo)
}

func GetSummonerInfo(c *gin.Context) {
	var req GetSummonerInfoRequestDto
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid request")
	}

	tagLine := "KR1"
	if req.TagLine != nil {
		tagLine = *req.TagLine
	}

	if req.GameName == "" {
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid game name")
		return
	}

	summonerDAO, exists, err := models.GetSummonerDAO_byNameTag(db.Root, req.GameName, tagLine)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}
	if !exists {
		// need to renew summoner
		tx, err := db.Root.BeginTxx(c, nil)
		if err != nil {
			log.Error(err)
			util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
			return
		}

		account, status, err := riot.GetAccountByRiotId(req.GameName, tagLine)
		if err != nil {
			if status == http.StatusNotFound {
				util.AbortWithStrJson(c, http.StatusNotFound, "invalid game name")
				return
			}
			log.Error(err)
			util.AbortWithStrJson(c, http.StatusBadRequest, "internal server error")
			return
		}

		//summoner, status, err := riot.GetSummonerByName(req.SummonerName)
		//if err != nil {
		//	if status == http.StatusNotFound {
		//		util.AbortWithStrJson(c, http.StatusNotFound, "invalid summoner name")
		//		return
		//	}
		//	log.Error(err)
		//	util.AbortWithStrJson(c, http.StatusBadRequest, "internal server error")
		//	return
		//}

		if err := service.RenewSummonerTotal(tx, account.Puuid); err != nil {
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

		// retry
		summonerDAO, exists, err = models.GetSummonerDAO_byNameTag(db.Root, req.GameName, tagLine)
		if err != nil {
			log.Error(err)
			util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
			return
		}
		if !exists {
			util.AbortWithStrJson(c, http.StatusBadRequest, "invalid summoner name")
			return
		}
	}

	// configure VOs
	summaryVO, err := service.GetSummonerSummaryVO_byPuuid(summonerDAO.Puuid)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	soloRankVO, err := service.GetSummonerRankVO(summonerDAO.Puuid, service.RankTypeSolo)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	flexRankVO, err := service.GetSummonerRankVO(summonerDAO.Puuid, service.RankTypeFlex)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	masteryVOs, err := service.GetSummonerMasteryVOs(summonerDAO.Puuid)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	matchesVOs, err := service.GetSummonerRecentMatchSummaryVOs(summonerDAO.Puuid, service.LoadInitialMatchCount)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := GetSummonerInfoResponseDto{
		Summary:  *summaryVO,
		SoloRank: soloRankVO,
		FlexRank: flexRankVO,
		Mastery:  masteryVOs,
		Matches:  matchesVOs,
	}

	c.JSON(http.StatusOK, resp)
}

func RenewSummonerInfo(c *gin.Context) {
	var req RenewSummonerInfoRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid request")
	}

	tx, err := db.Root.BeginTxx(c, nil)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := service.RenewSummonerTotal(tx, req.Puuid); err != nil {
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

	c.JSON(http.StatusOK, nil)
}

func LoadMatches(c *gin.Context) {
	var req LoadMatchesRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid request")
	}

	_, exists, err := models.GetOldestSummonerMatchDAO(db.Root, req.Puuid)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}
	if !exists {
		// just renew matches (recent)
		if err := service.RenewSummonerRecentMatches(db.Root, req.Puuid); err != nil {
			log.Error(err)
			util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
			return
		}
	} else {
		// renew matches (before requested time)
		beforeTime := time.UnixMilli(*req.Before)
		if err := service.RenewSummonerMatchesBefore(db.Root, req.Puuid, beforeTime); err != nil {
			log.Error(err)
			util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	if req.Before == nil {
		now := time.Now().UnixMilli()
		req.Before = &now
	}

	resp, err := service.GetSummonerMatchSummaryVOs_before(req.Puuid, *req.Before, service.LoadMoreMatchCount)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	c.JSON(http.StatusOK, resp)
}

func GetIngameInfo(c *gin.Context) {
	var req GetIngameInfoRequestDto
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid request")
	}

	summonerDAO, exists, err := models.GetSummonerDAO_byPuuid(db.Root, req.Puuid)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}
	if !exists {
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid puuid")
		return
	}

	spectatorInfo, _, err := riot.GetSpectatorInfo(summonerDAO.Id)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	team1 := make([]service.IngameParticipantVO, 0)
	team2 := make([]service.IngameParticipantVO, 0)
	for _, participant := range spectatorInfo.Participants {
		if participant.TeamId == 100 {
			team1 = append(team1, service.IngameParticipantMixer(participant))
		} else {
			team2 = append(team2, service.IngameParticipantMixer(participant))
		}
	}

	resp := GetIngameInfoResponseDto{
		GameType:          spectatorInfo.GameType,
		MapId:             spectatorInfo.MapId,
		GameStartTime:     spectatorInfo.GameStartTime,
		GameMode:          spectatorInfo.GameMode,
		GameQueueConfigId: spectatorInfo.GameQueueConfigId,
		Team1:             team1,
		Team2:             team2,
	}

	c.JSON(http.StatusOK, resp)
}
