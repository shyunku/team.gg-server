package v1

import (
	"github.com/gin-gonic/gin"
	log "github.com/shyunku-libraries/go-logger"
	"net/http"
	"strconv"
	"team.gg-server/libs/database"
	"team.gg-server/models"
	"team.gg-server/service"
	"team.gg-server/util"
	"time"
)

func UseV1Router(r *gin.Engine) {
	g := r.Group("/v1")
	UseIconRouter(g)

	g.GET("/summoner", GetSummonerInfo)
	g.POST("/renewSummoner", RenewSummonerInfo)
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
	ChampionId     int64   `json:"championId"`
	ChampionName   *string `json:"championName"`
	ChampionLevel  int     `json:"championLevel"`
	ChampionPoints int     `json:"championPoints"`
}

type SummonerMatchParticipant struct {
	MatchId            string `json:"matchId"`
	ParticipantId      int    `json:"participantId"`
	MatchParticipantId string `json:"matchParticipantId"`
	Puuid              string `json:"puuid"`
	Kills              int    `json:"kills"`
	Deaths             int    `json:"deaths"`
	Assists            int    `json:"assists"`
	ChampionId         int    `json:"championId"`
	ChampionLevel      int    `json:"championLevel"`
	SummonerLevel      int    `json:"summonerLevel"`
	SummonerName       string `json:"summonerName"`
	RiotIdName         string `json:"riotIdName"`
	RiotIdTagLine      string `json:"riotIdTagLine"`
	ProfileIcon        int    `json:"profileIcon"`

	Item0 int `json:"item0"`
	Item1 int `json:"item1"`
	Item2 int `json:"item2"`
	Item3 int `json:"item3"`
	Item4 int `json:"item4"`
	Item5 int `json:"item5"`
	Item6 int `json:"item6"`

	Spell1Casts    int `json:"spell1Casts"`
	Spell2Casts    int `json:"spell2Casts"`
	Spell3Casts    int `json:"spell3Casts"`
	Spell4Casts    int `json:"spell4Casts"`
	Summoner1Casts int `json:"summoner1Casts"`
	Summoner1Id    int `json:"summoner1Id"`
	Summoner2Casts int `json:"summoner2Casts"`
	Summoner2Id    int `json:"summoner2Id"`

	PrimaryPerkStyle int `json:"primaryPerkStyle"`
	SubPerkStyle     int `json:"subPerkStyle"`

	DoubleKills int `json:"doubleKills"`
	TripleKills int `json:"tripleKills"`
	QuadraKills int `json:"quadraKills"`
	PentaKills  int `json:"pentaKills"`

	TotalMinionsKilled          int `json:"totalMinionsKilled"`
	TotalCCDealt                int `json:"totalCCDealt"`
	TotalDamageDealtToChampions int `json:"totalDamageDealtToChampions"`

	GoldEarned int    `json:"goldEarned"`
	Lane       string `json:"lane"`
	Win        bool   `json:"win"`

	IndividualPosition string `json:"individualPosition"`
	TeamPosition       string `json:"teamPosition"`

	GameEndedInEarlySurrender bool `json:"gameEndedInEarlySurrender"`
	GameEndedInSurrender      bool `json:"gameEndedInSurrender"`
	TeamEarlySurrendered      bool `json:"teamEarlySurrendered"`
}

type TeamMate struct {
	ChampionId            int    `json:"championId"`
	SummonerName          string `json:"summonerName"`
	Puuid                 string `json:"puuid"`
	TotalDealtToChampions int    `json:"totalDealtToChampions"`
	Kills                 int    `json:"kills"`
	IndividualPosition    string `json:"individualPosition"`
	TeamPosition          string `json:"teamPosition"`
	ProfileIcon           int    `json:"profileIcon"`
}

type MatchSummary struct {
	MatchId            string                   `json:"matchId"`
	GameStartTimestamp int64                    `json:"gameStartTimestamp"`
	GameEndTimestamp   int64                    `json:"gameEndTimestamp"`
	GameDuration       int64                    `json:"gameDuration"`
	QueueId            int                      `json:"queueId"`
	MyStat             SummonerMatchParticipant `json:"myStat"`
	TeamChampionKills  int                      `json:"teamChampionKills"`
	Team1              []TeamMate               `json:"team1"`
	Team2              []TeamMate               `json:"team2"`
}

type GetSummonerInfoResponse struct {
	Summary  SummonerSummary   `json:"summary"`
	SoloRank *SummonerRank     `json:"soloRank"`
	FlexRank *SummonerRank     `json:"flexRank"`
	Mastery  []SummonerMastery `json:"mastery"`
	Matches  []MatchSummary    `json:"matches"`
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
	soloRankEntity, srExists, err := models.StrictGetRankByPuuidAndQueueType(summonerEntity.Puuid, service.RankTypeSolo)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	// get flex rank from db
	flexRankEntity, frExists, err := models.StrictGetRankByPuuidAndQueueType(summonerEntity.Puuid, service.RankTypeFlex)
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

	// get matches from db
	matchSummaries, err := service.GetSummonerRecentMatchSummaries(summonerEntity.Puuid)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	// get match participants from db
	matchParticipantsMap := make(map[string][]models.MatchParticipantEntity)
	for _, matchSummary := range matchSummaries {
		matchParticipants, err := models.GetMatchParticipantsByMatchId(matchSummary.MatchId)
		if err != nil {
			log.Error(err)
			util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
			return
		}
		if _, ok := matchParticipantsMap[matchSummary.MatchId]; !ok {
			matchParticipantsMap[matchSummary.MatchId] = make([]models.MatchParticipantEntity, 0)
		}

		for _, matchParticipant := range matchParticipants {
			matchParticipantsMap[matchSummary.MatchId] = append(matchParticipantsMap[matchSummary.MatchId], matchParticipant)
		}
	}

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
		var championName *string
		champion, ok := service.Champions[strconv.FormatInt(masteryEntity.ChampionId, 10)]
		if ok {
			championName = &champion.Name
		}
		res.Mastery = append(res.Mastery, SummonerMastery{
			ChampionId:     masteryEntity.ChampionId,
			ChampionName:   championName,
			ChampionLevel:  masteryEntity.ChampionLevel,
			ChampionPoints: masteryEntity.ChampionPoints,
		})
	}
	for _, matchSummary := range matchSummaries {
		// get match participant perk styles from db
		perks, err := models.GetMatchParticipantPerkStylesByMatchParticipantId(matchSummary.MatchParticipantId)
		if err != nil {
			log.Warn(err)
		}
		primaryPerkStyle := 0
		subPerkStyle := 0
		for _, perk := range perks {
			if perk.Description == service.PerkStyleDescriptionTypePrimary {
				primaryPerkStyle = perk.Style
			} else if perk.Description == service.PerkStyleDescriptionTypeSub {
				subPerkStyle = perk.Style
			}
		}

		summary := MatchSummary{
			MatchId:            matchSummary.MatchId,
			GameStartTimestamp: matchSummary.GameStartTimestamp,
			GameEndTimestamp:   matchSummary.GameEndTimestamp,
			GameDuration:       matchSummary.GameDuration,
			QueueId:            matchSummary.QueueId,
			MyStat: SummonerMatchParticipant{
				MatchId:                     matchSummary.MatchId,
				ParticipantId:               matchSummary.ParticipantId,
				MatchParticipantId:          matchSummary.MatchParticipantId,
				Puuid:                       matchSummary.Puuid,
				Kills:                       matchSummary.Kills,
				Deaths:                      matchSummary.Deaths,
				Assists:                     matchSummary.Assists,
				ChampionId:                  matchSummary.ChampionId,
				ChampionLevel:               matchSummary.ChampionLevel,
				SummonerLevel:               matchSummary.SummonerLevel,
				SummonerName:                matchSummary.SummonerName,
				RiotIdName:                  matchSummary.RiotIdName,
				RiotIdTagLine:               matchSummary.RiotIdTagLine,
				ProfileIcon:                 matchSummary.ProfileIcon,
				Item0:                       matchSummary.Item0,
				Item1:                       matchSummary.Item1,
				Item2:                       matchSummary.Item2,
				Item3:                       matchSummary.Item3,
				Item4:                       matchSummary.Item4,
				Item5:                       matchSummary.Item5,
				Item6:                       matchSummary.Item6,
				Spell1Casts:                 matchSummary.Spell1Casts,
				Spell2Casts:                 matchSummary.Spell2Casts,
				Spell3Casts:                 matchSummary.Spell3Casts,
				Spell4Casts:                 matchSummary.Spell4Casts,
				Summoner1Casts:              matchSummary.Summoner1Casts,
				Summoner1Id:                 matchSummary.Summoner1Id,
				Summoner2Casts:              matchSummary.Summoner2Casts,
				Summoner2Id:                 matchSummary.Summoner2Id,
				PrimaryPerkStyle:            primaryPerkStyle,
				SubPerkStyle:                subPerkStyle,
				DoubleKills:                 matchSummary.DoubleKills,
				TripleKills:                 matchSummary.TripleKills,
				QuadraKills:                 matchSummary.QuadraKills,
				PentaKills:                  matchSummary.PentaKills,
				TotalMinionsKilled:          matchSummary.TotalMinionsKilled,
				TotalCCDealt:                matchSummary.TotalTimeCCDealt,
				TotalDamageDealtToChampions: matchSummary.TotalDamageDealtToChampions,
				GoldEarned:                  matchSummary.GoldEarned,
				Lane:                        matchSummary.Lane,
				Win:                         matchSummary.Win,
				IndividualPosition:          matchSummary.IndividualPosition,
				TeamPosition:                matchSummary.TeamPosition,
				GameEndedInEarlySurrender:   matchSummary.GameEndedInEarlySurrender,
				GameEndedInSurrender:        matchSummary.GameEndedInSurrender,
				TeamEarlySurrendered:        matchSummary.TeamEarlySurrendered,
			},
		}

		matchTeamParticipants, ok := matchParticipantsMap[matchSummary.MatchId]
		if !ok {
			continue
		}
		team1Participants := make([]TeamMate, 0)
		team2Participants := make([]TeamMate, 0)
		for _, matchTeamParticipant := range matchTeamParticipants {
			teamMate := TeamMate{
				ChampionId:            matchTeamParticipant.ChampionId,
				SummonerName:          matchTeamParticipant.SummonerName,
				Puuid:                 matchTeamParticipant.Puuid,
				TotalDealtToChampions: matchTeamParticipant.TotalDamageDealtToChampions,
				Kills:                 matchTeamParticipant.Kills,
				IndividualPosition:    matchTeamParticipant.IndividualPosition,
				TeamPosition:          matchTeamParticipant.TeamPosition,
				ProfileIcon:           matchTeamParticipant.ProfileIcon,
			}
			if matchTeamParticipant.TeamId == 100 {
				team1Participants = append(team1Participants, teamMate)
			} else {
				team2Participants = append(team2Participants, teamMate)
			}
		}

		summary.Team1 = team1Participants
		summary.Team2 = team2Participants

		res.Matches = append(res.Matches, summary)
	}

	c.JSON(http.StatusOK, res)
}

type RenewSummonerInfoRequest struct {
	Puuid string `json:"puuid" binding:"required"`
}

func RenewSummonerInfo(c *gin.Context) {
	var req RenewSummonerInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid request")
	}

	// check if summoner exists in db
	_, exists, err := models.StrictGetSummonerByPuuid(req.Puuid)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}
	if !exists {
		util.AbortWithStrJson(c, http.StatusBadRequest, "invalid puuid")
		return
	}

	tx, err := database.DB.BeginTx(c, nil)
	if err != nil {
		log.Error(err)
		util.AbortWithStrJson(c, http.StatusInternalServerError, "internal server error")
		return
	}

	// refresh summoner info from riot api
	if err := service.RefreshSummonerInfoByPuuid(tx, req.Puuid); err != nil {
		log.Error("failed to refresh summoner info by puuid: ", err)
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
