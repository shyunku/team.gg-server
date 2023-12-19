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
	"team.gg-server/third_party/riot"
	"team.gg-server/util"
	"time"
)

func UseCustomGameRouter(r *gin.RouterGroup) {
	g := r.Group("/custom-game")

	g.GET("/list", GetCustomGameConfigurationList)
	g.GET("/info", GetCustomGameConfiguration)
	g.POST("/create", CreateCustomGameConfiguration)

	g.PUT("/candidate", AddCandidateToCustomGameConfiguration)
	g.POST("/arrange", ArrangeCustomGameParticipant)
	g.POST("/favor-position", SetCustomGameParticipantFavorPosition)
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

func AddCandidateToCustomGameConfiguration(c *gin.Context) {
	var req AddCandidateToCustomGameRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid request")
		return
	}

	uid := c.GetString("uid")

	// check if user is creator of custom game
	customGameConfigurationDAO, exists, err := models.GetCustomGameDAO_byId(db.Root, req.CustomGameConfigId)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}
	if !exists {
		util.AbortWithStrJson(c, http.StatusNotFound, "custom game configuration not found")
		return
	}
	if customGameConfigurationDAO.CreatorUid != uid {
		util.AbortWithStrJson(c, http.StatusForbidden, "user is not creator of custom game")
		return
	}

	// get summoner
	summonerDAO, exists, err := models.GetSummonerDAO_byNameTag(db.Root, req.Name, req.TagLine)
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

		account, status, err := riot.GetAccountByRiotId(req.Name, req.TagLine)
		if err != nil {
			if status == http.StatusNotFound {
				util.AbortWithStrJson(c, http.StatusNotFound, "invalid game name")
				return
			}
			log.Error(err)
			util.AbortWithStrJson(c, http.StatusBadRequest, "internal server error")
			return
		}

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
		summonerDAO, exists, err = models.GetSummonerDAO_byNameTag(db.Root, req.Name, req.TagLine)
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

	// load summoner completed
	// check if candidate already exists
	candidateDAOs, err := models.GetCustomGameCandidateDAOs_byCustomGameConfigId(db.Root, req.CustomGameConfigId)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	for _, candidateDAO := range candidateDAOs {
		if candidateDAO.Puuid == summonerDAO.Puuid {
			util.AbortWithStrJson(c, http.StatusConflict, "candidate already exists")
			return
		}
	}

	// add candidate
	newCandidateDAO := models.CustomGameCandidateDAO{
		CustomGameConfigId: req.CustomGameConfigId,
		Puuid:              summonerDAO.Puuid,
		CustomTier:         nil,
		CustomRank:         nil,
		FlavorTop:          false,
		FlavorJungle:       false,
		FlavorMid:          false,
		FlavorAdc:          false,
		FlavorSupport:      false,
	}
	if err := newCandidateDAO.Upsert(db.Root); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	candidateVO, err := service.GetCustomGameCandidateVO(newCandidateDAO)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	c.JSON(http.StatusOK, candidateVO)
}

func ArrangeCustomGameParticipant(c *gin.Context) {
	var req ArrangeCustomGameParticipantRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid request")
		return
	}

	uid := c.GetString("uid")

	// check if user is creator of custom game
	customGameConfigurationDAO, exists, err := models.GetCustomGameDAO_byId(db.Root, req.CustomGameConfigId)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}
	if !exists {
		util.AbortWithStrJson(c, http.StatusNotFound, "custom game configuration not found")
		return
	}
	if customGameConfigurationDAO.CreatorUid != uid {
		util.AbortWithStrJson(c, http.StatusForbidden, "user is not creator of custom game")
		return
	}

	// validate request
	if req.Team != 1 && req.Team != 2 {
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid team")
		return
	}

	if req.TargetPosition != service.PositionTop &&
		req.TargetPosition != service.PositionJungle &&
		req.TargetPosition != service.PositionMid &&
		req.TargetPosition != service.PositionAdc &&
		req.TargetPosition != service.PositionSupport {
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid target position")
		return
	}

	tx, err := db.Root.BeginTxx(c, nil)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	// check if candidate exists in config
	_, exists, err = models.GetCustomGameCandidateDAO_byPuuid(tx, req.CustomGameConfigId, req.Puuid)
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

	// check if candidate exists in same place
	participantDAO, exists, err := models.GetCustomGameParticipantDAO_byPosition(tx, req.CustomGameConfigId, req.Team, req.TargetPosition)
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	if exists {
		// delete participant
		if err := models.DeleteCustomGameParticipantDAO_byCustomGameConfigId(tx, req.CustomGameConfigId, participantDAO.Puuid); err != nil {
			log.Error(err)
			_ = tx.Rollback()
			util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	// add participant
	newParticipantDAO := models.CustomGameParticipantDAO{
		CustomGameConfigId: req.CustomGameConfigId,
		Puuid:              req.Puuid,
		Team:               req.Team,
		Position:           req.TargetPosition,
	}
	if err := newParticipantDAO.Upsert(tx); err != nil {
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

	c.JSON(http.StatusOK, "ok")
}

func SetCustomGameParticipantFavorPosition(c *gin.Context) {
	var req SetCustomGameParticipantFavorPositionRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid request")
		return
	}

	uid := c.GetString("uid")

	// check if user is creator of custom game
	customGameConfigurationDAO, exists, err := models.GetCustomGameDAO_byId(db.Root, req.CustomGameConfigId)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}
	if !exists {
		util.AbortWithStrJson(c, http.StatusNotFound, "custom game configuration not found")
		return
	}
	if customGameConfigurationDAO.CreatorUid != uid {
		util.AbortWithStrJson(c, http.StatusForbidden, "user is not creator of custom game")
		return
	}

	tx, err := db.Root.BeginTxx(c, nil)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	// check if candidate exists in config
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

	if req.Enabled == nil {
		_ = tx.Rollback()
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid enabled")
		return
	}

	// update candidate
	if req.FavorPosition == service.PositionTop {
		candidateDAO.FlavorTop = *req.Enabled
	} else if req.FavorPosition == service.PositionJungle {
		candidateDAO.FlavorJungle = *req.Enabled
	} else if req.FavorPosition == service.PositionMid {
		candidateDAO.FlavorMid = *req.Enabled
	} else if req.FavorPosition == service.PositionAdc {
		candidateDAO.FlavorAdc = *req.Enabled
	} else if req.FavorPosition == service.PositionSupport {
		candidateDAO.FlavorSupport = *req.Enabled
	} else {
		_ = tx.Rollback()
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid target position")
		return
	}

	if err := candidateDAO.Upsert(tx); err != nil {
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
