package v1

import (
	"github.com/gin-gonic/gin"
	log "github.com/shyunku-libraries/go-logger"
	"net/http"
	"regexp"
	"team.gg-server/service"
	"team.gg-server/util"
)

func UseIconRouter(r *gin.RouterGroup) {
	g := r.Group("/icon")

	g.GET("/champion", GetChampionIcon)
	g.GET("/profile", GetProfileIcon)
	g.GET("/summonerSpell", GetSummonerSpellIcon)
	g.GET("/item", GetItemIcon)
	g.GET("/perkStyle", GetPerkStyleIcon)
}

type GetChampionIconRequest struct {
	Key string `form:"key" binding:"required"`
}

func GetChampionIcon(c *gin.Context) {
	var req GetChampionIconRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid request")
	}

	champion, ok := service.Champions[req.Key]
	if !ok {
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid champion key")
		return
	}

	championId := champion.Id
	championIconUrl := "https://ddragon.leagueoflegends.com/cdn/" + service.DataDragonVersion + "/img/champion/" + championId + ".png"
	c.Redirect(http.StatusMovedPermanently, championIconUrl)
}

type GetProfileIconRequest struct {
	Id string `form:"id" binding:"required"`
}

func GetProfileIcon(c *gin.Context) {
	var req GetProfileIconRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid request")
		return
	}

	profileIconUrl := "https://ddragon.leagueoflegends.com/cdn/" + service.DataDragonVersion + "/img/profileicon/" + req.Id + ".png"
	c.Redirect(http.StatusMovedPermanently, profileIconUrl)
}

type GetSummonerSpellIconRequest struct {
	Id string `form:"id" binding:"required"`
}

func GetSummonerSpellIcon(c *gin.Context) {
	var req GetSummonerSpellIconRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid request")
		return
	}

	spellInfo, ok := service.SummonerSpells[req.Id]
	if !ok {
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid spell id")
		return
	}

	spellImgName := spellInfo.Image.Full
	imgBytes, err := service.LoadDDragonImageFile("/spell/" + spellImgName)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	c.Header("Cache-Control", "public, max-age=31536000")
	c.Data(http.StatusOK, "image/png", imgBytes)
}

type GetItemIconRequest struct {
	Id string `form:"id" binding:"required"`
}

func GetItemIcon(c *gin.Context) {
	var req GetItemIconRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid request")
		return
	}
	imgBytes, err := service.LoadDDragonImageFile("/item/" + req.Id + ".png")
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	c.Header("Cache-Control", "public, max-age=31536000")
	c.Data(http.StatusOK, "image/png", imgBytes)
}

type GetPerkIconRequest struct {
	Id int `form:"id" binding:"required"`
}

func GetPerkStyleIcon(c *gin.Context) {
	var req GetPerkIconRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid request")
		return
	}

	perkStyle, ok1 := service.PerkStyles[req.Id]
	perk, ok2 := service.Perks[req.Id]
	if !ok1 && !ok2 {
		log.Debug(service.Perks)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid perk id")
		return
	}

	var perkImgPathRaw string
	if ok1 {
		perkImgPathRaw = perkStyle.IconPath
	} else {
		perkImgPathRaw = perk.IconPath
	}

	re := regexp.MustCompile(`(?m)/perk-images/(.*)`)
	perkImgPath := re.FindStringSubmatch(perkImgPathRaw)[1]
	path := "https://ddragon.leagueoflegends.com/cdn/img/perk-images/" + perkImgPath
	c.Redirect(http.StatusMovedPermanently, path)
}
