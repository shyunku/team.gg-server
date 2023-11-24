package v1

import (
	"github.com/gin-gonic/gin"
	log "github.com/shyunku-libraries/go-logger"
	"net/http"
	"team.gg-server/core"
	"team.gg-server/libs/database"
	"team.gg-server/models"
	"team.gg-server/service"
	"team.gg-server/util"
	"time"
)

func UseV1Router(r *gin.Engine) {
	g := r.Group("/v1")
	g.GET("/profileIcon", GetProfileIcon)
	g.GET("/summoner", GetSummonerInfo)
}

type GetProfileIconRequest struct {
	Id string `form:"id" binding:"required"`
}

func GetProfileIcon(c *gin.Context) {
	var req GetProfileIconRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid request")
	}

	profileIconUrl := "https://ddragon.leagueoflegends.com/cdn/" + core.DataDragonVersion + "/img/profileicon/" + req.Id + ".png"
	c.Redirect(http.StatusMovedPermanently, profileIconUrl)
}

type GetSummonerInfoRequest struct {
	SummonerName string `form:"summonerName" binding:"required"`
}

type SummonerSummary struct {
	ProfileIconId int       `json:"profileIconId"`
	Name          string    `json:"name"`
	Puuid         string    `json:"puuid"`
	SummonerLevel int64     `json:"summonerLevel"`
	LastUpdatedAt time.Time `json:"lastUpdatedAt"`
}

type SummonerRank struct {
	Tier   string `json:"tier"`
	Rank   string `json:"rank"`
	Lp     int    `json:"lp"`
	Wins   int    `json:"wins"`
	Losses int    `json:"losses"`
}

type SummonerMastery struct {
	ChampionId     int64 `json:"championId"`
	ChampionLevel  int   `json:"championLevel"`
	ChampionPoints int   `json:"championPoints"`
}

type GetSummonerInfoResponse struct {
	Summary  SummonerSummary   `json:"summary"`
	SoloRank *SummonerRank     `json:"soloRank"`
	FlexRank *SummonerRank     `json:"flexRank"`
	Mastery  []SummonerMastery `json:"mastery"`
}

func GetSummonerInfo(c *gin.Context) {
	var req GetSummonerInfoRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid request")
	}

	// check if summoner exists in db
	_, exists, err := models.StrictGetSummonerByShortenName(req.SummonerName)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	if !exists {
		tx, err := database.DB.BeginTx(c, nil)
		if err != nil {
			log.Error(err)
			util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
			return
		}

		// refresh summoner info from riot api
		if err := service.RefreshSummonerInfoByName(tx, req.SummonerName); err != nil {
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
	}

	// get summoner info from db
	summonerEntity, err := models.GetSummonerByShortenName(req.SummonerName)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	// TODO :: split by league id

	// get solo rank from db
	soloRankEntity, srExists, err := models.StrictGetRankByPuuidAndQueueType(summonerEntity.Puuid, service.QueueTypeSoloRank)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	// get flex rank from db
	flexRankEntity, frExists, err := models.StrictGetRankByPuuidAndQueueType(summonerEntity.Puuid, service.QueueTypeFlexRank)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	// get mastery from db
	masteryEntities, err := models.GetMasteriesByPuuidTx(summonerEntity.Puuid)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	// TODO :: get match history from db

	// make response
	res := GetSummonerInfoResponse{
		Summary: SummonerSummary{
			ProfileIconId: summonerEntity.ProfileIconId,
			Name:          summonerEntity.Name,
			Puuid:         summonerEntity.Puuid,
			SummonerLevel: summonerEntity.SummonerLevel,
			LastUpdatedAt: summonerEntity.LastUpdatedAt,
		},
		SoloRank: nil,
		FlexRank: nil,
	}
	if srExists {
		res.SoloRank = &SummonerRank{
			Tier:   soloRankEntity.Tier,
			Rank:   soloRankEntity.Rank,
			Lp:     soloRankEntity.LeaguePoints,
			Wins:   soloRankEntity.Wins,
			Losses: soloRankEntity.Losses,
		}
	}
	if frExists {
		res.FlexRank = &SummonerRank{
			Tier:   flexRankEntity.Tier,
			Rank:   flexRankEntity.Rank,
			Lp:     flexRankEntity.LeaguePoints,
			Wins:   flexRankEntity.Wins,
			Losses: flexRankEntity.Losses,
		}
	}
	for _, masteryEntity := range masteryEntities {
		res.Mastery = append(res.Mastery, SummonerMastery{
			ChampionId:     masteryEntity.ChampionId,
			ChampionLevel:  masteryEntity.ChampionLevel,
			ChampionPoints: masteryEntity.ChampionPoints,
		})
	}

	c.JSON(http.StatusOK, res)
}
