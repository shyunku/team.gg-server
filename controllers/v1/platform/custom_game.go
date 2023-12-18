package platform

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/shyunku-libraries/go-logger"
	"net/http"
	"team.gg-server/libs/db"
	"team.gg-server/models"
	"team.gg-server/service"
	"team.gg-server/util"
	"time"
)

func UseCustomGameRouter(r *gin.RouterGroup) {
	g := r.Group("/custom-game")

	g.GET("/list", GetCustomGameConfigurationList)
	g.GET("/info", GetCustomGameConfiguration)
	g.POST("/create", CreateCustomGameConfiguration)
}

func GetCustomGameConfigurationList(c *gin.Context) {
	uid := c.GetString("uid")

	// get all custom games from db
	resp, err := service.GetCustomGameConfigurationVOs(uid)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}
	c.JSON(http.StatusOK, resp)
}

func GetCustomGameConfiguration(c *gin.Context) {
	var req GetCustomGameConfigurationRequestDto
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid request")
		return
	}

	resp, err := service.GetCustomGameConfigurationVO(req.Id)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	c.JSON(http.StatusOK, resp)
}

func CreateCustomGameConfiguration(c *gin.Context) {
	uid := c.GetString("uid")

	// get all custom games from db
	customGameConfigurationDAOs, err := models.GetCustomGameDAOs_byCreatorUid(db.Root, uid)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	namePrefix := "내전 팀 구성"
	nameSuffix := 1
	name := fmt.Sprintf("%s %d", namePrefix, nameSuffix)
	for _, customGameConfigurationDAO := range customGameConfigurationDAOs {
		if customGameConfigurationDAO.Name == name {
			nameSuffix++
			name = fmt.Sprintf("%s %d", namePrefix, nameSuffix)
		}
	}

	// create custom game configuration
	newId := uuid.New().String()
	now := time.Now()
	newCustomGameConfigurationDAO := models.CustomGameConfigurationDAO{
		Id:            newId,
		Name:          name,
		CreatorUid:    uid,
		CreatedAt:     now,
		LastUpdatedAt: now,
		Fairness:      0,
		LineFairness:  0,
		TierFairness:  0,
	}
	if err := newCustomGameConfigurationDAO.Upsert(db.Root); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	c.JSON(http.StatusOK, newId)
}
