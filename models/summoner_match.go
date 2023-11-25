package models

import (
	"database/sql"
	"errors"
)

type SummonerMatchEntity struct {
	Puuid   string `db:"puuid" json:"puuid"`
	MatchId string `db:"match_id" json:"matchId"`
}

func (s *SummonerMatchEntity) UpsertTx(tx *sql.Tx) error {
	if _, err := tx.Exec(`
		INSERT INTO summoner_matches
		    (puuid, match_id) 
		VALUES (?, ?) 
		ON DUPLICATE KEY UPDATE 
			puuid = ?, match_id = ?`,
		s.Puuid, s.MatchId, s.Puuid, s.MatchId,
	); err != nil {
		return err
	}
	return nil
}

func StrictGetSummonerMatch(tx *sql.Tx, puuid string, matchId string) (*SummonerMatchEntity, bool, error) {
	// check if summoner exists in db
	summonerMatch, err := GetSummonerMatchTx(tx, puuid, matchId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if summonerMatch == nil {
		return nil, false, nil
	}
	return summonerMatch, true, nil
}

func GetSummonerMatchTx(tx *sql.Tx, puuid string, matchId string) (*SummonerMatchEntity, error) {
	var summonerMatch SummonerMatchEntity
	if err := tx.QueryRow("SELECT * FROM summoner_matches WHERE puuid = ? AND match_id = ?", puuid, matchId).Scan(&summonerMatch.Puuid, &summonerMatch.MatchId); err != nil {
		return nil, err
	}
	return &summonerMatch, nil
}
