package models

import (
	"database/sql"
	"errors"
	"team.gg-server/libs/db"
	legacy_models "team.gg-server/models/legacy"
)

type MatchTeamDAO struct {
	MatchId string `db:"match_id" json:"matchId"`
	TeamId  int    `db:"team_id" json:"teamId"`
	Win     bool   `db:"win" json:"win"`

	BaronFirst      bool `db:"baron_first" json:"baronFirst"`
	BaronKills      int  `db:"baron_kills" json:"baronKills"`
	ChampionFirst   bool `db:"champion_first" json:"championFirst"`
	ChampionKills   int  `db:"champion_kills" json:"championKills"`
	DragonFirst     bool `db:"dragon_first" json:"dragonFirst"`
	DragonKills     int  `db:"dragon_kills" json:"dragonKills"`
	InhibitorFirst  bool `db:"inhibitor_first" json:"inhibitorFirst"`
	InhibitorKills  int  `db:"inhibitor_kills" json:"inhibitorKills"`
	RiftHeraldFirst bool `db:"rift_herald_first" json:"riftHeraldFirst"`
	RiftHeraldKills int  `db:"rift_herald_kills" json:"riftHeraldKills"`
	TowerFirst      bool `db:"tower_first" json:"towerFirst"`
	TowerKills      int  `db:"tower_kills" json:"towerKills"`
}

func (s *MatchTeamDAO) Insert(db db.Context) error {
	if _, err := db.Exec(`
		INSERT INTO match_teams
		    (match_id, team_id, win, baron_first, baron_kills, champion_first, 
		     champion_kills, dragon_first, dragon_kills, inhibitor_first, 
		     inhibitor_kills, rift_herald_first, rift_herald_kills, 
		     tower_first, tower_kills) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.MatchId, s.TeamId, s.Win, s.BaronFirst, s.BaronKills, s.ChampionFirst,
		s.ChampionKills, s.DragonFirst, s.DragonKills, s.InhibitorFirst,
		s.InhibitorKills, s.RiftHeraldFirst, s.RiftHeraldKills,
		s.TowerFirst, s.TowerKills,
	); err != nil {
		return err
	}
	return nil
}

func (s *MatchTeamDAO) ToLegacy() legacy_models.LegacyMatchTeamDAO {
	return legacy_models.LegacyMatchTeamDAO{
		MatchId:         s.MatchId,
		TeamId:          s.TeamId,
		Win:             s.Win,
		BaronFirst:      s.BaronFirst,
		BaronKills:      s.BaronKills,
		ChampionFirst:   s.ChampionFirst,
		ChampionKills:   s.ChampionKills,
		DragonFirst:     s.DragonFirst,
		DragonKills:     s.DragonKills,
		InhibitorFirst:  s.InhibitorFirst,
		InhibitorKills:  s.InhibitorKills,
		RiftHeraldFirst: s.RiftHeraldFirst,
		RiftHeraldKills: s.RiftHeraldKills,
		TowerFirst:      s.TowerFirst,
		TowerKills:      s.TowerKills,
	}
}

func (s *MatchTeamDAO) Delete(db db.Context) error {
	if _, err := db.Exec("DELETE FROM match_teams WHERE match_id = ? AND team_id = ?", s.MatchId, s.TeamId); err != nil {
		return err
	}
	return nil
}

func GetMatchTeamDAOs_byMatchId(db db.Context, matchId string) ([]MatchTeamDAO, error) {
	var matchTeams []MatchTeamDAO
	if err := db.Select(&matchTeams, `
		SELECT match_id, team_id, win, baron_first, baron_kills, champion_first, 
		       champion_kills, dragon_first, dragon_kills, inhibitor_first, 
		       inhibitor_kills, rift_herald_first, rift_herald_kills, 
		       tower_first, tower_kills
		FROM match_teams
		WHERE match_id = ?;
	`, matchId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]MatchTeamDAO, 0), nil
		}
		return nil, err
	}
	return matchTeams, nil
}
