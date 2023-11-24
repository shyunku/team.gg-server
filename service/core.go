package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	log "github.com/shyunku-libraries/go-logger"
	"team.gg-server/libs/http"
	"team.gg-server/models"
	"team.gg-server/third_party/riot"
	"team.gg-server/util"
	"time"
)

func GetLatestDataDragonVersion() (string, error) {
	resp, err := http.Get(http.GetRequest{
		Url: "https://ddragon.leagueoflegends.com/api/versions.json",
	})
	if err != nil {
		return "", err
	}

	var versions []string
	if err := json.Unmarshal(resp.Body, &versions); err != nil {
		return "", err
	}

	return versions[0], nil
}

func RefreshSummonerInfoByName(tx *sql.Tx, name string) error {
	// get summoner by name
	summoner, err := riot.GetSummonerByName(name)
	if err != nil {
		log.Warnf("failed to get summoner by name (%s)", name)
		return err
	}

	// TODO :: update summoner?

	return RefreshSummonerInfoByPuuid(tx, summoner.Puuid)
}

func RefreshSummonerInfoByPuuid(tx *sql.Tx, puuid string) error {
	if puuid == "" {
		return errors.New("puuid is required")
	}

	/* -------------------- refresh summoner -------------------- */
	summoner, err := riot.GetSummonerByPuuid(puuid)
	if err != nil {
		log.Warnf("failed to get summoner by puuid (%s)", puuid)
		return err
	}
	if err := refreshSummoner(tx, summoner); err != nil {
		return err
	}

	/* -------------------- refresh summoner rank -------------------- */
	leagues, err := riot.GetLeaguesBySummonerId(summoner.Id)
	if err != nil {
		log.Warnf("failed to get league by summoner id (%s)", summoner.Id)
		return err
	}
	if err := refreshRank(tx, summoner, leagues); err != nil {
		return err
	}

	/* -------------------- refresh mastery -------------------- */
	mastery, err := riot.GetMasteryBySummonerId(summoner.Id)
	if err != nil {
		log.Warnf("failed to get mastery by summoner id (%s)", summoner.Id)
		return err
	}
	if err := refreshMastery(tx, summoner, mastery); err != nil {
		return err
	}

	return nil
}

func refreshSummoner(tx *sql.Tx, summoner *riot.SummonerDto) error {
	// check if summoner exists in db
	summonerEntity, exists, err := models.StrictGetSummonerByPuuid(summoner.Puuid)
	if err != nil {
		return err
	}

	if !exists {
		// create new summoner
		summonerEntity = &models.SummonerEntity{
			AccountId:     summoner.AccountId,
			ProfileIconId: summoner.ProfileIconId,
			RevisionDate:  summoner.RevisionDate,
			Name:          summoner.Name,
			Id:            summoner.Id,
			Puuid:         summoner.Puuid,
			SummonerLevel: summoner.SummonerLevel,
			ShortenName:   util.ShortenSummonerName(summoner.Name),
			LastUpdatedAt: time.Now(),
			Hits:          0,
		}
	} else {
		// update summoner
		summonerEntity.AccountId = summoner.AccountId
		summonerEntity.ProfileIconId = summoner.ProfileIconId
		summonerEntity.RevisionDate = summoner.RevisionDate
		summonerEntity.Name = summoner.Name
		summonerEntity.Id = summoner.Id
		summonerEntity.Puuid = summoner.Puuid
		summonerEntity.SummonerLevel = summoner.SummonerLevel
		summonerEntity.ShortenName = util.ShortenSummonerName(summoner.Name)
		summonerEntity.LastUpdatedAt = time.Now()
		summonerEntity.Hits += 1
	}
	if err := summonerEntity.UpsertTx(tx); err != nil {
		return err
	}

	return nil
}

func refreshRank(tx *sql.Tx, summoner *riot.SummonerDto, leagues *riot.LeagueDto) error {
	for _, league := range *leagues {
		if league.SummonerId != summoner.Id {
			log.Errorf("league summoner id (%s) != summoner id (%s)", league.SummonerId, summoner.Id)
			return errors.New("league summoner id is not equal to summoner id")
		}

		// create new league
		leagueEntity := &models.LeagueEntity{
			Puuid:      summoner.Puuid,
			LeagueId:   league.LeagueId,
			QueueType:  league.QueueType,
			Tier:       league.Tier,
			Rank:       league.Rank,
			Wins:       league.Wins,
			Losses:     league.Losses,
			HotStreak:  league.HotStreak,
			Veteran:    league.Veteran,
			FreshBlood: league.FreshBlood,
			Inactive:   league.Inactive,
			MsTarget:   league.MiniSeries.Target,
			MsWins:     league.MiniSeries.Wins,
			MsLosses:   league.MiniSeries.Losses,
			MsProgress: league.MiniSeries.Progress,
		}

		if err := leagueEntity.UpsertTx(tx); err != nil {
			return err
		}
	}

	return nil
}

func refreshMastery(tx *sql.Tx, summoner *riot.SummonerDto, masteries *riot.MasteryDto) error {
	for _, mastery := range *masteries {
		if mastery.Puuid != summoner.Puuid {
			log.Errorf("mastery puuid (%s) != summoner puuid (%s)", mastery.Puuid, summoner.Puuid)
			return errors.New("mastery puuid is not equal to summoner puuid")
		}

		// upsert mastery
		masteryEntity := &models.MasteryEntity{
			Puuid:                        summoner.Puuid,
			ChampionId:                   mastery.ChampionId,
			ChampionLevel:                mastery.ChampionLevel,
			ChampionPoints:               mastery.ChampionPoints,
			ChampionPointsSinceLastLevel: mastery.ChampionPointsSinceLastLevel,
			ChampionPointsUntilNextLevel: mastery.ChampionPointsUntilNextLevel,
			ChestGranted:                 mastery.ChestGranted,
			LastPlayTime:                 time.UnixMilli(mastery.LastPlayTime),
			TokensEarned:                 mastery.TokensEarned,
		}

		if err := masteryEntity.Upsert(tx); err != nil {
			return err
		}
	}

	return nil
}
