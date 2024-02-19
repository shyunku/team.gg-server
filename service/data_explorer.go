package service

import (
	"context"
	log "github.com/shyunku-libraries/go-logger"
	"team.gg-server/libs/db"
	"team.gg-server/models"
	"team.gg-server/third_party/riot"
	"time"
)

type DataExplorer struct {
	lastExploredTime *time.Time
	exploreCaches    int
	cacheHit         int
}

func NewDataExplorer() *DataExplorer {
	return &DataExplorer{
		lastExploredTime: nil,
		exploreCaches:    0,
	}
}

func (de *DataExplorer) Loop() {
	for {
		if time.Since(riot.LastApiCallTime) > riot.ApiCallIdleThreshold {
			// do something
			de.Explore()
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}

func (de *DataExplorer) finalizeIteration(success bool) {
	now := time.Now()
	de.lastExploredTime = &now
	de.exploreCaches++
	if success {
		de.cacheHit++
	}
}

func (de *DataExplorer) Explore() {
	de.exploreCaches++
	log.Debugf("DataExplorer: explored %d/%d", de.cacheHit, de.exploreCaches)

	var err error
	// update something new
	err = de.fetchNewSummoner()

	de.finalizeIteration(err == nil)
}

func (de *DataExplorer) GetExploreCaches() int {
	return de.exploreCaches
}

func (de *DataExplorer) fetchNewSummoner() error {
	ctx := context.Background()
	tx, err := db.Root.BeginTxx(ctx, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	// get random summoner match
	participant, err := models.GetRandomMatchParticipantDAO(tx)
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return err
	}

	// search for original summoner
	_, exists, err := models.GetSummonerDAO_byPuuid(tx, participant.Puuid)
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return err
	}
	if exists {
		_ = tx.Rollback()
		return nil
	}

	// get summoner info
	summonerDAO, err := RenewSummonerInfoByPuuid(tx, participant.Puuid)
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return err
	}

	// get summoner rank
	if err := RenewSummonerLeague(tx, summonerDAO.Id, summonerDAO.Puuid); err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return err
	}

	// get summoner recent matches
	if err := RenewSummonerRecentMatchesWithCount(tx, summonerDAO.Puuid, 5); err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return err
	}

	log.Debugf("DataExplorer: fetched new summoner %s#%s", summonerDAO.GameName, summonerDAO.TagLine)
	return nil
}
