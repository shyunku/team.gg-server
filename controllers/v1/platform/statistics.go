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
}

func GetChampionStatistics(c *gin.Context) {
	championStatVOs, err := service.GetChampionStatisticsVOs()
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := GetChampionStatisticsResponseDto{
		UpdatedAt: nil,
		Stats:     championStatVOs,
	}
	c.JSON(http.StatusOK, resp)
}
