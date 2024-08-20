package models

import (
	"database/sql"
	"errors"
	"team.gg-server/libs/db"
	legacy_models "team.gg-server/models/legacy"
)

type MatchTeamBanDAO struct {
	MatchId string `db:"match_id" json:"matchId"`
	TeamId  int    `db:"team_id" json:"teamId"`

	ChampionId int `db:"champion_id" json:"championId"`
	PickTurn   int `db:"pick_turn" json:"pickTurn"`
}

func (s *MatchTeamBanDAO) Insert(db db.Context) error {
	if _, err := db.Exec(`
		INSERT INTO match_team_bans
		    (match_id, team_id, champion_id, pick_turn) 
		VALUES (?, ?, ?, ?)`,
		s.MatchId, s.TeamId, s.ChampionId, s.PickTurn,
	); err != nil {
		return err
	}
	return nil
}

func (s *MatchTeamBanDAO) ToLegacy() legacy_models.LegacyMatchTeamBanDAO {
	return legacy_models.LegacyMatchTeamBanDAO{
		MatchId:    s.MatchId,
		TeamId:     s.TeamId,
		ChampionId: s.ChampionId,
		PickTurn:   s.PickTurn,
	}
}

func (s *MatchTeamBanDAO) Delete(db db.Context) error {
	if _, err := db.Exec("DELETE FROM match_team_bans WHERE match_id = ? AND team_id = ? AND champion_id = ?", s.MatchId, s.TeamId, s.ChampionId); err != nil {
		return err
	}
	return nil
}

func GetMatchTeamBanDAOs_byMatchId(db db.Context, matchId string) ([]MatchTeamBanDAO, error) {
	var bans []MatchTeamBanDAO
	if err := db.Select(&bans, `
		SELECT
			match_id,
			team_id,
			champion_id,
			pick_turn
		FROM match_team_bans
		WHERE match_id = ?;
	`, matchId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]MatchTeamBanDAO, 0), nil
		}
		return nil, err
	}
	return bans, nil
}
