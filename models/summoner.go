package models

import (
	"database/sql"
	"errors"
	"team.gg-server/libs/database"
	"team.gg-server/util"
	"time"
)

type SummonerEntity struct {
	AccountId     string    `db:"account_id" json:"accountId"`
	ProfileIconId int       `db:"profile_icon_id" json:"profileIconId"`
	RevisionDate  int64     `db:"revision_date" json:"revisionDate"`
	Name          string    `db:"name" json:"name"`
	Id            string    `db:"id" json:"id"`
	Puuid         string    `db:"puuid" json:"puuid"`
	SummonerLevel int64     `db:"summoner_level" json:"summonerLevel"`
	ShortenName   string    `db:"shorten_name" json:"shortenName"`
	LastUpdatedAt time.Time `db:"last_updated_at" json:"lastUpdatedAt"`
	Hits          int       `db:"hits" json:"hits"`
}

func (s *SummonerEntity) UpsertTx(tx *sql.Tx) error {
	if _, err := tx.Exec(`
		INSERT INTO summoners
		    (account_id, profile_icon_id, revision_date, name, id, puuid, summoner_level, shorten_name, last_updated_at, hits) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) 
		ON DUPLICATE KEY UPDATE 
			account_id = ?, profile_icon_id = ?, revision_date = ?, 
		    name = ?, id = ?, puuid = ?, summoner_level = ?, 
		    shorten_name = ?, last_updated_at = ?, hits = ?`,
		s.AccountId, s.ProfileIconId, s.RevisionDate, s.Name, s.Id, s.Puuid,
		s.SummonerLevel, s.ShortenName, s.LastUpdatedAt, s.Hits,
		s.AccountId, s.ProfileIconId, s.RevisionDate, s.Name, s.Id, s.Puuid,
		s.SummonerLevel, s.ShortenName, s.LastUpdatedAt, s.Hits,
	); err != nil {
		return err
	}
	return nil
}

func StrictGetSummonerByShortenName(name string) (*SummonerEntity, bool, error) {
	shortenName := util.ShortenSummonerName(name)
	// check if summoner exists in db
	summonerEntity, err := GetSummonerByShortenName(shortenName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if summonerEntity == nil {
		return nil, false, nil
	}
	return summonerEntity, true, nil
}

func StrictGetSummonerByPuuid(puuid string) (*SummonerEntity, bool, error) {
	// check if summoner exists in db
	summonerEntity, err := GetSummonerByPuuid(puuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if summonerEntity == nil {
		return nil, false, nil
	}
	return summonerEntity, true, nil
}

func GetSummonerByPuuid(puuid string) (*SummonerEntity, error) {
	var summoner SummonerEntity
	if err := database.DB.Get(&summoner, "SELECT * FROM summoners WHERE puuid = ?", puuid); err != nil {
		return nil, err
	}
	return &summoner, nil
}

func GetSummonerByShortenName(name string) (*SummonerEntity, error) {
	shortenName := util.ShortenSummonerName(name)
	var summoner SummonerEntity
	if err := database.DB.Get(&summoner, "SELECT * FROM summoners WHERE shorten_name = ?", shortenName); err != nil {
		return nil, err
	}
	return &summoner, nil
}
