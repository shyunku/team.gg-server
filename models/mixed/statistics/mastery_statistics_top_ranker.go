package statistics

import (
	"database/sql"
	"errors"
	"team.gg-server/libs/db"
)

type MasteryStatisticsTopRankersMXDAO struct {
	Puuid         string `db:"puuid" json:"puuid"`
	ProfileIconId int    `db:"profile_icon_id" json:"profileIconId"`
	GameName      string `db:"game_name" json:"gameName"`
	TagLine       string `db:"tag_line" json:"tagLine"`

	Ranks int `db:"ranks" json:"ranks"`

	ChampionId     int `db:"champion_id" json:"championId"`
	ChampionPoints int `db:"champion_points" json:"championPoints"`
}

func GetMasteryStatisticsTopRankersMXDAOs(db db.Context, topRanks int) ([]*MasteryStatisticsTopRankersMXDAO, error) {
	var topRankers []*MasteryStatisticsTopRankersMXDAO
	if err := db.Select(&topRankers, `
		WITH RankedMasteries AS (
			SELECT
			    puuid,
				champion_id,
				champion_points,
				ROW_NUMBER() OVER (PARTITION BY champion_id ORDER BY champion_points DESC) AS ranks
			FROM masteries
		)
		SELECT s.puuid, s.game_name, s.tag_line, s.profile_icon_id, rm.champion_id, rm.champion_points, rm.ranks
		FROM RankedMasteries rm
		LEFT JOIN summoners s ON rm.puuid = s.puuid
		WHERE ranks <= ?;
	`, topRanks); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]*MasteryStatisticsTopRankersMXDAO, 0), nil
		}
		return nil, err
	}
	return topRankers, nil
}
