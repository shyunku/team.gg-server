package statistics_models

import (
	"database/sql"
	"errors"
	"team.gg-server/libs/db"
)

type ChampionDetailStatisticsMetaMXDAO struct {
	ChampionId   int    `db:"champion_id" json:"championId"`
	ChampionName string `db:"champion_name" json:"championName"`
	TeamPosition string `db:"team_position" json:"teamPosition"`

	PrimaryStyle int `db:"primary_style" json:"primaryStyle"`
	SubStyle     int `db:"sub_style" json:"subStyle"`

	Item0Id int `db:"item0_id" json:"item0Id"`
	Item1Id int `db:"item1_id" json:"item1Id"`
	Item2Id int `db:"item2_id" json:"item2Id"`

	Item0Name string `db:"item0_name" json:"item0Name"`
	Item1Name string `db:"item1_name" json:"item1Name"`
	Item2Name string `db:"item2_name" json:"item2Name"`

	ItemCount int     `db:"item_count" json:"itemCount"`
	Wins      int     `db:"wins" json:"wins"`
	Total     int     `db:"total" json:"total"`
	WinRate   float64 `db:"win_rate" json:"winRate"`
	PickRate  float64 `db:"pick_rate" json:"pickRate"`

	MetaRank int `db:"meta_rank" json:"metaRank"`
}

// TODO :: resolve memory issue (can be overflowed by memory limit)

func GetChampionDetailStatisticsMetaMXDAOs(db db.Context) ([]ChampionDetailStatisticsMetaMXDAO, error) {
	var championDetailStatisticsMetaMXDAOs []ChampionDetailStatisticsMetaMXDAO
	if err := db.Select(&championDetailStatisticsMetaMXDAOs, `
		WITH PreCalculated AS (
			SELECT
				match_participant_id,
				MAX(CASE WHEN description = 'primaryStyle' THEN style END) AS primary_style,
				MAX(CASE WHEN description = 'subStyle' THEN style END) AS sub_style
			FROM match_participant_perk_styles
			GROUP BY match_participant_id
		),
		ItemDetails AS (
			SELECT
				mp.match_id,
				mp.match_participant_id,
				mp.champion_id,
				mp.champion_name,
				mp.team_position,
				si.id AS item_id,
				si.name AS item_name,
				si.gold_total AS gold_value,
				pc.primary_style,
				pc.sub_style,
				mt.win,
				ROW_NUMBER() OVER (PARTITION BY mp.match_id, mp.match_participant_id ORDER BY si.depth DESC, si.gold_total DESC) AS item_rank
			FROM match_participants mp
			JOIN static_items si ON si.id IN (mp.item0, mp.item1, mp.item2, mp.item3, mp.item4, mp.item5, mp.item6)
			LEFT JOIN match_teams mt ON mp.match_id = mt.match_id AND mp.team_id = mt.team_id
			LEFT JOIN PreCalculated pc ON mp.match_participant_id = pc.match_participant_id
			WHERE si.id IS NOT NULL
				AND si.required_ally IS NULL
				AND si.id != 0
				AND si.gold_total > 0
				AND si.depth >= 3
				AND mp.team_position != ''
		),
		ItemMetaGroups AS (
			SELECT match_participant_id, 
				champion_id,
				champion_name,
				team_position,
				primary_style,
				sub_style,
				MAX(CASE WHEN item_rank = 1 THEN item_id END)   AS item0_id,
				MAX(CASE WHEN item_rank = 1 THEN item_name END) AS item0_name,
				MAX(CASE WHEN item_rank = 2 THEN item_id END)   AS item1_id,
				MAX(CASE WHEN item_rank = 2 THEN item_name END) AS item1_name,
				MAX(CASE WHEN item_rank = 3 THEN item_id END)   AS item2_id,
				MAX(CASE WHEN item_rank = 3 THEN item_name END) AS item2_name,
				MAX(CASE WHEN item_rank = 4 THEN item_id END)   AS item3_id,
				MAX(CASE WHEN item_rank = 4 THEN item_name END) AS item3_name,
				MAX(CASE WHEN item_rank = 5 THEN item_id END)   AS item4_id,
				MAX(CASE WHEN item_rank = 5 THEN item_name END) AS item4_name,
				MAX(CASE WHEN item_rank = 6 THEN item_id END)   AS item5_id,
				MAX(CASE WHEN item_rank = 6 THEN item_name END) AS item5_name,
				MAX(CASE WHEN item_rank = 7 THEN item_id END)   AS item6_id,
				MAX(CASE WHEN item_rank = 7 THEN item_name END) AS item6_name,
				SUM(win)                                        AS wins,
				COUNT(*)                                        AS count
			FROM ItemDetails
			GROUP BY match_participant_id, champion_id, champion_name, team_position, primary_style, sub_style
		),
		PickCounts AS (
			SELECT
				champion_id,
				team_position,
				COUNT(*) AS total
			FROM ItemDetails
			GROUP BY champion_id, team_position
		),
		RefinedMetaGroups AS (
			SELECT
				img.champion_id,
				champion_name,
				img.team_position,
				primary_style,
				sub_style,
				item0_id,
				item1_id,
				item2_id,
				item0_name,
				item1_name,
				item2_name,
				SUM(wins) AS wins,
				SUM(count) AS total,
				SUM(wins) / SUM(count) AS win_rate,
				SUM(count) / SUM(pc.total) AS pick_rate
			FROM ItemMetaGroups img
			JOIN PickCounts pc ON pc.champion_id = img.champion_id
				AND pc.team_position = img.team_position
			WHERE item0_id IS NOT NULL
				AND item1_id IS NOT NULL
				AND item2_id IS NOT NULL
			GROUP BY champion_id, champion_name, team_position,
					 primary_style, sub_style,
					 item0_id, item1_id, item2_id,
					 item0_name, item1_name, item2_name
		),
		RankedMetas AS (
			SELECT *,
				ROW_NUMBER() OVER (
					PARTITION BY champion_id, champion_name, team_position
					ORDER BY total DESC, win_rate DESC
				) AS meta_rank
			FROM RefinedMetaGroups
		)
		SELECT *
		FROM RankedMetas
		WHERE meta_rank <= 20 OR (win_rate >= 0.5 AND total >= 50)
		ORDER BY champion_name ASC, team_position ASC, meta_rank ASC;
	`); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]ChampionDetailStatisticsMetaMXDAO, 0), nil
		}
		return nil, err
	}

	return championDetailStatisticsMetaMXDAOs, nil
}

func GetChampionDetailStatisticsMetaMXDAOs_byChampionId(db db.Context, championId int) ([]ChampionDetailStatisticsMetaMXDAO, error) {
	var championDetailStatisticsMetaMXDAOs []ChampionDetailStatisticsMetaMXDAO
	if err := db.Select(&championDetailStatisticsMetaMXDAOs, `
		WITH PreCalculated AS (
			SELECT
				match_participant_id,
				MAX(CASE WHEN description = 'primaryStyle' THEN style END) AS primary_style,
				MAX(CASE WHEN description = 'subStyle' THEN style END) AS sub_style
			FROM match_participant_perk_styles
			GROUP BY match_participant_id
		),
		ItemDetails AS (
			SELECT
				mp.match_id,
				mp.match_participant_id,
				mp.champion_id,
				mp.champion_name,
				mp.team_position,
				si.id AS item_id,
				si.name AS item_name,
				si.gold_total AS gold_value,
				pc.primary_style,
				pc.sub_style,
				mt.win,
				ROW_NUMBER() OVER (PARTITION BY mp.match_id, mp.match_participant_id ORDER BY si.depth DESC, si.gold_total DESC) AS item_rank
			FROM match_participants mp
			JOIN static_items si ON si.id IN (mp.item0, mp.item1, mp.item2, mp.item3, mp.item4, mp.item5, mp.item6)
			LEFT JOIN match_teams mt ON mp.match_id = mt.match_id AND mp.team_id = mt.team_id
			LEFT JOIN PreCalculated pc ON mp.match_participant_id = pc.match_participant_id
			WHERE si.id IS NOT NULL
				AND mp.champion_id = ?
				AND mp.team_position != ''
				AND si.id != 0
				AND si.depth >= 3
				AND si.gold_total > 0
				AND si.required_ally IS NULL
		),
		ItemMetaGroups AS (
			SELECT match_participant_id, 
				champion_id,
				champion_name,
				team_position,
				primary_style,
				sub_style,
				MAX(CASE WHEN item_rank = 1 THEN item_id END)   AS item0_id,
				MAX(CASE WHEN item_rank = 1 THEN item_name END) AS item0_name,
				MAX(CASE WHEN item_rank = 2 THEN item_id END)   AS item1_id,
				MAX(CASE WHEN item_rank = 2 THEN item_name END) AS item1_name,
				MAX(CASE WHEN item_rank = 3 THEN item_id END)   AS item2_id,
				MAX(CASE WHEN item_rank = 3 THEN item_name END) AS item2_name,
				SUM(win)                                        AS wins,
				COUNT(*)                                        AS count
			FROM ItemDetails
			GROUP BY match_participant_id, champion_id, champion_name, team_position, primary_style, sub_style
		),
		PickCounts AS (
			SELECT
				champion_id,
				team_position,
				COUNT(*) AS total
			FROM ItemDetails
			GROUP BY champion_id, team_position
		),
		RefinedMetaGroups AS (
			SELECT
				img.champion_id,
				champion_name,
				img.team_position,
				primary_style,
				sub_style,
				item0_id,
				item1_id,
				item2_id,
				item0_name,
				item1_name,
				item2_name,
				SUM(wins) AS wins,
				SUM(count) AS total,
				SUM(wins) / SUM(count) AS win_rate,
				SUM(count) / SUM(pc.total) AS pick_rate
			FROM ItemMetaGroups img
			JOIN PickCounts pc ON pc.champion_id = img.champion_id
				AND pc.team_position = img.team_position
			WHERE item0_id IS NOT NULL
				AND item1_id IS NOT NULL
				AND item2_id IS NOT NULL
			GROUP BY champion_id, champion_name, team_position,
					 primary_style, sub_style,
					 item0_id, item1_id, item2_id,
					 item0_name, item1_name, item2_name
		),
		RankedMetas AS (
			SELECT *,
				ROW_NUMBER() OVER (
					PARTITION BY champion_id, champion_name, team_position
					ORDER BY total DESC, win_rate DESC
				) AS meta_rank
			FROM RefinedMetaGroups
		)
		SELECT *
		FROM RankedMetas
		WHERE meta_rank <= 20 OR (win_rate >= 0.5 AND total >= 50)
		ORDER BY champion_name ASC, team_position ASC, meta_rank ASC;
	`, championId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]ChampionDetailStatisticsMetaMXDAO, 0), nil
		}
		return nil, err
	}

	return championDetailStatisticsMetaMXDAOs, nil
}
