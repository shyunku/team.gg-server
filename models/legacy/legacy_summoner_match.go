package legacy_models

import (
	"database/sql"
	"errors"
	"team.gg-server/libs/db"
)

type LegacySummonerMatchDAO struct {
	Puuid   string `db:"puuid" json:"puuid"`
	MatchId string `db:"match_id" json:"matchId"`
}

func (s *LegacySummonerMatchDAO) Upsert(db db.Context) error {
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

func GetLegacySummonerMatchDAO(db db.Context, puuid string, matchId string) (*LegacySummonerMatchDAO, bool, error) {
	// check if summoner exists in db
	var summonerMatch LegacySummonerMatchDAO
	if err := db.Get(&summonerMatch, "SELECT * FROM summoner_matches WHERE puuid = ? AND match_id = ?", puuid, matchId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &summonerMatch, true, nil
}
