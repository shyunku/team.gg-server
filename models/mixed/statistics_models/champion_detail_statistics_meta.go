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

	Summoner1Id int `db:"summoner1_id" json:"summoner1Id"`
	Summoner2Id int `db:"summoner2_id" json:"summoner2Id"`

	Item0Id int  `db:"item0_id" json:"item0Id"`
	Item1Id int  `db:"item1_id" json:"item1Id"`
	Item2Id int  `db:"item2_id" json:"item2Id"`
	Item3Id *int `db:"item3_id" json:"item3Id"`
	Item4Id *int `db:"item4_id" json:"item4Id"`
	Item5Id *int `db:"item5_id" json:"item5Id"`

	Item0Name string  `db:"item0_name" json:"item0Name"`
	Item1Name string  `db:"item1_name" json:"item1Name"`
	Item2Name string  `db:"item2_name" json:"item2Name"`
	Item3Name *string `db:"item3_name" json:"item3Name"`
	Item4Name *string `db:"item4_name" json:"item4Name"`
	Item5Name *string `db:"item5_name" json:"item5Name"`

	ItemCount int     `db:"item_count" json:"itemCount"`
	Wins      int     `db:"wins" json:"wins"`
	Total     int     `db:"total" json:"total"`
	WinRate   float64 `db:"win_rate" json:"winRate"`

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
				mp.summoner1_id,
				mp.summoner2_id,
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
				AND si.id != 0
				AND si.required_ally IS NULL
				AND si.gold_purchasable IS TRUE
				AND si.gold_total > 0
				AND si.depth >= 3
				AND mp.team_position != ''
		),
		ItemTreeGroups AS (
			SELECT match_participant_id,
				champion_id,
				champion_name,
				team_position,
				summoner1_id,
				summoner2_id,
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
			GROUP BY match_participant_id, champion_id, champion_name, team_position,
					 primary_style, sub_style, summoner1_id, summoner2_id
		),
		SummonerSpellGroups AS (
			SELECT champion_id,
				   team_position,
				   primary_style,
				   sub_style,
				   summoner1_id,
				   summoner2_id,
				   COUNT(*) AS count
			FROM ItemTreeGroups
			GROUP BY champion_id, team_position, primary_style, sub_style,
					 summoner1_id, summoner2_id
		),
		MaxSummonerSpells AS (
			SELECT *,
				   ROW_NUMBER() OVER (
					   PARTITION BY champion_id, team_position, primary_style, sub_style ORDER BY count DESC
				   ) AS spell_rank
			FROM SummonerSpellGroups
		),
		FullItemTreeGroups AS (
			SELECT champion_id,
				champion_name,
				team_position,
				primary_style,
				sub_style,
				item0_id, item1_id, item2_id,
				item3_id, item4_id, item5_id,
				item0_name, item1_name, item2_name,
				item3_name, item4_name, item5_name,
				IF(item0_id IS NOT NULL, 1, 0)
				+ IF(item1_id IS NOT NULL, 1, 0)
				+ IF(item2_id IS NOT NULL, 1, 0)
				+ IF(item3_id IS NOT NULL, 1, 0)
				+ IF(item4_id IS NOT NULL, 1, 0)
				+ IF(item5_id IS NOT NULL, 1, 0)
				AS item_count,
				COUNT(*) AS full_item_tree_count
			FROM ItemTreeGroups
			WHERE item0_id IS NOT NULL
				AND item1_id IS NOT NULL
				AND item2_id IS NOT NULL
			GROUP BY champion_id, champion_name, team_position, primary_style, sub_style,
				item0_id, item1_id, item2_id, item3_id, item4_id, item5_id,
				item0_name, item1_name, item2_name, item3_name, item4_name, item5_name
		),
		MaxItemCombination AS (
			SELECT *,
				ROW_NUMBER() OVER (PARTITION BY champion_id, champion_name, team_position, primary_style, sub_style,
					item0_id, item1_id, item2_id ORDER BY item_count DESC, full_item_tree_count DESC) AS item_combo_rank
			FROM FullItemTreeGroups
		),
		RefinedMetaGroups AS (
			SELECT
				champion_id,
				champion_name,
				team_position,
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
				SUM(wins) / SUM(count) AS win_rate
			FROM ItemTreeGroups img
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
		),
		FinalRankedMetas AS (
			SELECT rm.*,
					mic.item3_id,
					mic.item4_id,
					mic.item5_id,
					mic.item3_name,
					mic.item4_name,
					mic.item5_name,
					mss.summoner1_id,
					mss.summoner2_id
			FROM RankedMetas rm
			LEFT JOIN MaxItemCombination mic ON rm.champion_id = mic.champion_id
				AND rm.champion_name = mic.champion_name
				AND rm.team_position = mic.team_position
				AND rm.primary_style = mic.primary_style
				AND rm.sub_style = mic.sub_style
				AND rm.item0_id = mic.item0_id
				AND rm.item1_id = mic.item1_id
				AND rm.item2_id = mic.item2_id
				AND mic.item_combo_rank = 1
			LEFT JOIN MaxSummonerSpells mss ON rm.champion_id = mss.champion_id
				AND rm.team_position = mss.team_position
				AND rm.primary_style = mss.primary_style
				AND rm.sub_style = mss.sub_style
				AND mss.spell_rank = 1
		)
		SELECT *
		FROM FinalRankedMetas
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

// GetChampionDetailStatisticsMetaMXDAOs_previous deprecated
func GetChampionDetailStatisticsMetaMXDAOs_previous(db db.Context) ([]ChampionDetailStatisticsMetaMXDAO, error) {
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
		ItemTreeGroups AS (
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
				SUM(wins) / SUM(count) AS win_rate
			FROM ItemTreeGroups img
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
