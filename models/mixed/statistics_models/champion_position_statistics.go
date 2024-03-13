package statistics_models

import (
	"database/sql"
	"errors"
	"team.gg-server/libs/db"
)

type ChampionPositionStatisticsMXDAO struct {
	ChampionId   int    `db:"champion_id" json:"championId"`
	TeamPosition string `db:"team_position" json:"teamPosition"`
	Win          int    `db:"win" json:"win"`
	Total        int    `db:"total" json:"total"`
}

func GetChampionPositionStatisticsMXDAOs(db db.Context) ([]ChampionPositionStatisticsMXDAO, error) {
	var statistics []ChampionPositionStatisticsMXDAO
	if err := db.Select(&statistics, `
		SELECT
			champion_id,
			team_position,
			SUM(win) AS win,
			COUNT(*) AS total
		FROM match_participants
		WHERE team_position != ''
		GROUP BY champion_id, team_position
	`); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]ChampionPositionStatisticsMXDAO, 0), nil
		}
		return nil, err
	}
	return statistics, nil
}
