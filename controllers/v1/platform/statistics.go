package platform

import (
	"github.com/gin-gonic/gin"
	log "github.com/shyunku-libraries/go-logger"
	"net/http"
	"team.gg-server/service"
	"team.gg-server/util"
)

func UseStatisticsRouter(r *gin.RouterGroup) {
	g := r.Group("/statistics")

	g.GET("/champion", GetChampionStatistics)
	g.GET("/tier", GetTierStatistics)
	g.GET("/mastery", GetMasteryStatistics)
}

func GetChampionStatistics(c *gin.Context) {
	statistics, err := service.ChampionStatisticsRepo.Load()
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	c.JSON(http.StatusOK, statistics)
}

func GetTierStatistics(c *gin.Context) {
	statistics, err := service.TierStatisticsRepo.Load()
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	c.JSON(http.StatusOK, statistics)
}

func GetMasteryStatistics(c *gin.Context) {
	statistics, err := service.MasteryStatisticsRepo.Load()
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	c.JSON(http.StatusOK, statistics)
}
