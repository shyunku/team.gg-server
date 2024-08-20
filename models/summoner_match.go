package models

import (
	"database/sql"
	"errors"
	"team.gg-server/libs/db"
	legacy_models "team.gg-server/models/legacy"
)

type SummonerMatchDAO struct {
	Puuid   string `db:"puuid" json:"puuid"`
	MatchId string `db:"match_id" json:"matchId"`
}

func (s *SummonerMatchDAO) Upsert(db db.Context) error {
	if _, err := db.Exec(`
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

func (s *SummonerMatchDAO) ToLegacy() legacy_models.LegacySummonerMatchDAO {
	return legacy_models.LegacySummonerMatchDAO{
		Puuid:   s.Puuid,
		MatchId: s.MatchId,
	}
}

func (s *SummonerMatchDAO) Delete(db db.Context) error {
	if _, err := db.Exec("DELETE FROM summoner_matches WHERE puuid = ? AND match_id = ?", s.Puuid, s.MatchId); err != nil {
		return err
	}
	return nil
}

func GetSummonerMatchDAO(db db.Context, puuid string, matchId string) (*SummonerMatchDAO, bool, error) {
	// check if summoner exists in db
	var summonerMatch SummonerMatchDAO
	if err := db.Get(&summonerMatch, "SELECT * FROM summoner_matches WHERE puuid = ? AND match_id = ?", puuid, matchId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &summonerMatch, true, nil
}

func GetSummonerMatchDAOs_byMatchId(db db.Context, matchId string) ([]SummonerMatchDAO, error) {
	var summonerMatches []SummonerMatchDAO
	if err := db.Select(&summonerMatches, "SELECT * FROM summoner_matches WHERE match_id = ?", matchId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]SummonerMatchDAO, 0), nil
		}
		return nil, err
	}
	return summonerMatches, nil
}
