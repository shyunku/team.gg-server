package models

import (
	"database/sql"
	"errors"
	"team.gg-server/libs/db"
	"team.gg-server/util"
	"time"
)

type SummonerDAO struct {
	AccountId     string    `db:"account_id" json:"accountId"`
	ProfileIconId int       `db:"profile_icon_id" json:"profileIconId"`
	RevisionDate  int64     `db:"revision_date" json:"revisionDate"`
	Name          string    `db:"name" json:"name"`
	Id            string    `db:"id" json:"id"`
	Puuid         string    `db:"puuid" json:"puuid"`
	SummonerLevel int64     `db:"summoner_level" json:"summonerLevel"`
	ShortenName   string    `db:"shorten_name" json:"shortenName"`
	LastUpdatedAt time.Time `db:"last_updated_at" json:"lastUpdatedAt"`
}

func (s *SummonerDAO) Upsert(db db.Context) error {
	if _, err := db.Exec(`
		INSERT INTO summoners
		    (account_id, profile_icon_id, revision_date, name, id, puuid, summoner_level, shorten_name, last_updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) 
		ON DUPLICATE KEY UPDATE 
			account_id = ?, profile_icon_id = ?, revision_date = ?, 
		    name = ?, id = ?, puuid = ?, summoner_level = ?, 
		    shorten_name = ?, last_updated_at = ?`,
		s.AccountId, s.ProfileIconId, s.RevisionDate, s.Name, s.Id, s.Puuid,
		s.SummonerLevel, s.ShortenName, s.LastUpdatedAt,
		s.AccountId, s.ProfileIconId, s.RevisionDate, s.Name, s.Id, s.Puuid,
		s.SummonerLevel, s.ShortenName, s.LastUpdatedAt,
	); err != nil {
		return err
	}
	return nil
}

func GetSummonerDAO_byName(db db.Context, name string) (*SummonerDAO, bool, error) {
	shortenName := util.ShortenSummonerName(name)
	// check if summoner exists in db
	var summonerEntity SummonerDAO
	if err := db.Get(&summonerEntity, "SELECT * FROM summoners WHERE shorten_name = ?", shortenName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &summonerEntity, true, nil
}

func GetSummonerDAO_byPuuid(db db.Context, puuid string) (*SummonerDAO, bool, error) {
	// check if summoner exists in db
	var summonerEntity SummonerDAO
	if err := db.Get(&summonerEntity, "SELECT * FROM summoners WHERE puuid = ?", puuid); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &summonerEntity, true, nil
}
