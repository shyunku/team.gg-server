package models

import (
	"database/sql"
	"errors"
	"team.gg-server/libs/db"
	"team.gg-server/util"
	"time"
)

type SummonerDAO struct {
	AccountId       string    `db:"account_id" json:"accountId"`
	ProfileIconId   int       `db:"profile_icon_id" json:"profileIconId"`
	RevisionDate    int64     `db:"revision_date" json:"revisionDate"`
	GameName        string    `db:"game_name" json:"gameName"`
	TagLine         string    `db:"tag_line" json:"tagLine"`
	Name            string    `db:"name" json:"name"`
	Id              string    `db:"id" json:"id"`
	Puuid           string    `db:"puuid" json:"puuid"`
	SummonerLevel   int64     `db:"summoner_level" json:"summonerLevel"`
	ShortenGameName string    `db:"shorten_game_name" json:"shortenGameName"`
	ShortenName     string    `db:"shorten_name" json:"shortenName"`
	LastUpdatedAt   time.Time `db:"last_updated_at" json:"lastUpdatedAt"`
}

func (s *SummonerDAO) Upsert(db db.Context) error {
	if _, err := db.Exec(`
		INSERT INTO summoners
		    (account_id, profile_icon_id, revision_date, game_name, tag_line, name, id, puuid, summoner_level, shorten_game_name, shorten_name, last_updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) 
		ON DUPLICATE KEY UPDATE 
			account_id = ?, profile_icon_id = ?, revision_date = ?, 
		    game_name = ?, tag_line = ?,
		    name = ?, id = ?, puuid = ?, summoner_level = ?, 
		    shorten_game_name = ?, shorten_name = ?, last_updated_at = ?`,
		s.AccountId, s.ProfileIconId, s.RevisionDate,
		s.GameName, s.TagLine, s.Name, s.Id, s.Puuid,
		s.SummonerLevel, s.ShortenGameName, s.ShortenName, s.LastUpdatedAt,
		s.AccountId, s.ProfileIconId, s.RevisionDate,
		s.GameName, s.TagLine, s.Name, s.Id, s.Puuid,
		s.SummonerLevel, s.ShortenGameName, s.ShortenName, s.LastUpdatedAt,
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

func GetSummonerDAO_byNameTag(db db.Context, gameName string, tagLine string) (*SummonerDAO, bool, error) {
	shortenName := util.ShortenSummonerName(gameName)
	// check if summoner exists in db
	var summonerEntity SummonerDAO
	if err := db.Get(&summonerEntity,
		"SELECT * FROM summoners WHERE shorten_game_name = ? AND tag_line = ?",
		shortenName, tagLine); err != nil {
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
