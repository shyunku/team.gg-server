package models

import (
	"database/sql"
	"errors"
	"team.gg-server/libs/database"
)

type LeagueEntity struct {
	Puuid        string `db:"puuid" json:"puuid"`
	LeagueId     string `db:"league_id" json:"leagueId"`
	QueueType    string `db:"queue_type" json:"queueType"`
	Tier         string `db:"tier" json:"tier"`
	Rank         string `db:"league_rank" json:"rank"`
	LeaguePoints int    `db:"league_points" json:"leaguePoints"`
	Wins         int    `db:"wins" json:"wins"`
	Losses       int    `db:"losses" json:"losses"`
	HotStreak    bool   `db:"hot_streak" json:"hotStreak"`
	Veteran      bool   `db:"veteran" json:"veteran"`
	FreshBlood   bool   `db:"fresh_blood" json:"freshBlood"`
	Inactive     bool   `db:"inactive" json:"inactive"`
	MsTarget     int    `db:"ms_target" json:"msTarget"`
	MsWins       int    `db:"ms_wins" json:"msWins"`
	MsLosses     int    `db:"ms_losses" json:"msLosses"`
	MsProgress   string `db:"ms_progress" json:"msProgress"`
}

func (l *LeagueEntity) UpsertTx(tx *sql.Tx) error {
	if _, err := tx.Exec(`
		INSERT INTO ranks
		    (puuid, league_id, queue_type, tier, league_rank, league_points, wins, losses, hot_streak, veteran, fresh_blood, inactive, ms_target, ms_wins, ms_losses, ms_progress) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		    			league_id = ?, queue_type = ?, tier = ?, league_rank = ?, league_points = ?, wins = ?, losses = ?, hot_streak = ?, veteran = ?, fresh_blood = ?, inactive = ?, ms_target = ?, ms_wins = ?, ms_losses = ?, ms_progress = ?`,
		l.Puuid, l.LeagueId, l.QueueType, l.Tier, l.Rank, l.LeaguePoints, l.Wins, l.Losses, l.HotStreak, l.Veteran, l.FreshBlood, l.Inactive, l.MsTarget, l.MsWins, l.MsLosses, l.MsProgress,
		l.LeagueId, l.QueueType, l.Tier, l.Rank, l.LeaguePoints, l.Wins, l.Losses, l.HotStreak, l.Veteran, l.FreshBlood, l.Inactive, l.MsTarget, l.MsWins, l.MsLosses, l.MsProgress,
	); err != nil {
		return err
	}
	return nil
}

func StrictGetLeagueByPuuid(puuid string) (*LeagueEntity, bool, error) {
	// check if league exists in db
	leagueEntity, err := GetLeagueByPuuid(puuid)
	if err != nil {
		return nil, false, err
	}
	if leagueEntity == nil {
		return nil, false, nil
	}
	return leagueEntity, true, nil
}

func StrictGetRankByPuuidAndQueueType(puuid string, queueType string) (*LeagueEntity, bool, error) {
	// check if league exists in db
	leagueEntity, err := GetRankByPuuidAndQueueType(puuid, queueType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if leagueEntity == nil {
		return nil, false, nil
	}
	return leagueEntity, true, nil
}

func GetLeagueByPuuid(puuid string) (*LeagueEntity, error) {
	var leagueEntity LeagueEntity
	if err := database.DB.Get(&leagueEntity, `
		SELECT *
		FROM ranks
		WHERE puuid = ?`, puuid); err != nil {
		return nil, err
	}
	return &leagueEntity, nil
}

func GetRankByPuuidAndQueueType(puuid string, queueType string) (*LeagueEntity, error) {
	var leagueEntity LeagueEntity
	if err := database.DB.Get(&leagueEntity, `
		SELECT *
		FROM ranks
		WHERE puuid = ? AND queue_type = ?`, puuid, queueType); err != nil {
		return nil, err
	}
	return &leagueEntity, nil
}
