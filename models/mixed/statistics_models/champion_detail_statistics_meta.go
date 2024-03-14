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

	PrimaryStyle    int `db:"primary_style" json:"primaryStyle"`
	PrimaryPerk0    int `db:"primary_perk0" json:"primaryPerk0"`
	PrimaryPerk1    int `db:"primary_perk1" json:"primaryPerk1"`
	PrimaryPerk2    int `db:"primary_perk2" json:"primaryPerk2"`
	PrimaryPerk3    int `db:"primary_perk3" json:"primaryPerk3"`
	SubStyle        int `db:"sub_style" json:"subStyle"`
	SubPerk0        int `db:"sub_perk0" json:"subPerk0"`
	SubPerk1        int `db:"sub_perk1" json:"subPerk1"`
	StatPerkDefense int `db:"stat_perk_defense" json:"statPerkDefense"`
	StatPerkFlex    int `db:"stat_perk_flex" json:"statPerkFlex"`
	StatPerkOffense int `db:"stat_perk_offense" json:"statPerkOffense"`

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
		WITH SortedPerks AS ( # Perk 정렬
			SELECT style_id,
				   perk,
				   ROW_NUMBER() OVER (PARTITION BY style_id ORDER BY perk DESC) AS perk_rank
			FROM match_participant_perk_style_selections
			ORDER BY style_id
		),
		PerkGroups AS ( # 정렬된 Perk 한 줄로 그룹화
			SELECT match_participant_id,
				   mpps.style_id,
				   MAX(CASE WHEN sp.perk_rank = 1 THEN sp.perk END) AS perk0,
				   MAX(CASE WHEN sp.perk_rank = 2 THEN sp.perk END) AS perk1,
				   MAX(CASE WHEN sp.perk_rank = 3 THEN sp.perk END) AS perk2,
				   MAX(CASE WHEN sp.perk_rank = 4 THEN sp.perk END) AS perk3
			FROM match_participant_perk_styles mpps
			LEFT JOIN SortedPerks sp ON mpps.style_id = sp.style_id
			GROUP BY mpps.match_participant_id, mpps.style_id
		),
		PerkStyleGroups AS ( # PerkStyle 그룹화
			SELECT
				mpps.match_participant_id,
				MAX(CASE WHEN description = 'primaryStyle' THEN style END) AS primary_style,
				MAX(CASE WHEN description = 'primaryStyle' THEN perk0 END) AS primary_perk0,
				MAX(CASE WHEN description = 'primaryStyle' THEN perk1 END) AS primary_perk1,
				MAX(CASE WHEN description = 'primaryStyle' THEN perk2 END) AS primary_perk2,
				MAX(CASE WHEN description = 'primaryStyle' THEN perk3 END) AS primary_perk3,
				MAX(CASE WHEN description = 'subStyle' THEN style END) AS sub_style,
				MAX(CASE WHEN description = 'subStyle' THEN perk0 END) AS sub_perk0,
				MAX(CASE WHEN description = 'subStyle' THEN perk1 END) AS sub_perk1,
				mpp.stat_perk_defense,
				mpp.stat_perk_flex,
				mpp.stat_perk_offense
			FROM match_participant_perk_styles mpps
			LEFT JOIN match_participant_perks mpp ON mpps.match_participant_id = mpp.match_participant_id
			LEFT JOIN PerkGroups pg ON mpps.match_participant_id = pg.match_participant_id
								   AND mpps.style_id = pg.style_id
			GROUP BY match_participant_id
		),
		ItemDetails AS ( # 아이템 0~6 정보 flatten by row
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
				pc.primary_perk0,
				pc.primary_perk1,
				pc.primary_perk2,
				pc.primary_perk3,
				pc.sub_style,
				pc.sub_perk0,
				pc.sub_perk1,
				pc.stat_perk_defense,
				pc.stat_perk_flex,
				pc.stat_perk_offense,
				mt.win,
				ROW_NUMBER() OVER (PARTITION BY mp.match_id, mp.match_participant_id ORDER BY si.depth DESC, si.gold_total DESC) AS item_rank
			FROM match_participants mp
			JOIN static_items si ON si.id IN (mp.item0, mp.item1, mp.item2, mp.item3, mp.item4, mp.item5, mp.item6)
			LEFT JOIN match_teams mt ON mp.match_id = mt.match_id AND mp.team_id = mt.team_id
			LEFT JOIN PerkStyleGroups pc ON mp.match_participant_id = pc.match_participant_id
			WHERE si.id IS NOT NULL
				AND si.id != 0
				AND si.required_ally IS NULL
				AND si.gold_purchasable IS TRUE
				AND si.gold_total > 0
				AND si.depth >= 3
				AND mp.team_position != ''
				AND pc.primary_style IS NOT NULL
				AND pc.sub_style IS NOT NULL
				AND pc.primary_style != 0
				AND pc.sub_style != 0
		),
		MainGroup AS ( # 아이템 0~6 정보 그룹화 (participant row)
			SELECT match_participant_id,
				champion_id,
				champion_name,
				team_position,
				summoner1_id,
				summoner2_id,
				primary_style,
				primary_perk0,
				primary_perk1,
				primary_perk2,
				primary_perk3,
				sub_style,
				sub_perk0,
				sub_perk1,
				stat_perk_defense,
				stat_perk_flex,
				stat_perk_offense,
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
		SummonerSpellCounts AS ( # 스펠 그룹 카운트
			SELECT champion_id,
				   team_position,
				   primary_style,
				   sub_style,
				   summoner1_id,
				   summoner2_id,
				   COUNT(*) AS count
			FROM MainGroup
			GROUP BY champion_id, team_position, primary_style, sub_style,
					 summoner1_id, summoner2_id
		),
		SummonerSpellRanks AS ( # 스펠 그룹 정렬
			SELECT *,
				   ROW_NUMBER() OVER (
					   PARTITION BY champion_id, team_position, primary_style, sub_style ORDER BY count DESC
				   ) AS spell_rank
			FROM SummonerSpellCounts
		),
		PerkCounts AS ( # Perk 그룹 카운트
			SELECT champion_id,
				   team_position,
				   primary_style,
				   sub_style,
				   primary_perk0,
				   primary_perk1,
				   primary_perk2,
				   primary_perk3,
				   sub_perk0,
				   sub_perk1,
				   stat_perk_defense,
				   stat_perk_flex,
				   stat_perk_offense,
				   COUNT(*) AS count
			FROM MainGroup
			GROUP BY champion_id, team_position, primary_style, sub_style,
					 primary_perk0, primary_perk1, primary_perk2, primary_perk3,
					 sub_perk0, sub_perk1, stat_perk_defense, stat_perk_flex, stat_perk_offense
		),
		PerkRanks AS ( # Perk 그룹 정렬
			SELECT *,
			   ROW_NUMBER() OVER (
				   PARTITION BY champion_id, team_position, primary_style, sub_style ORDER BY count DESC
			   ) AS perk_rank
			FROM PerkCounts
		),
		FullItemTreeGroups AS ( # 템트리 그룹화
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
			FROM MainGroup
			WHERE item0_id IS NOT NULL
				AND item1_id IS NOT NULL
				AND item2_id IS NOT NULL
			GROUP BY champion_id, champion_name, team_position, primary_style, sub_style,
				item0_id, item1_id, item2_id, item3_id, item4_id, item5_id,
				item0_name, item1_name, item2_name, item3_name, item4_name, item5_name
		),
		FullItemTreeRanks AS ( # 템트리 그룹 정렬
			SELECT *,
				ROW_NUMBER() OVER (PARTITION BY champion_id, champion_name, team_position, primary_style, sub_style,
					item0_id, item1_id, item2_id ORDER BY item_count DESC, full_item_tree_count DESC) AS item_combo_rank
			FROM FullItemTreeGroups
		),
		RefinedMetaGroups AS ( # 메타 그룹화
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
			FROM MainGroup img
			WHERE item0_id IS NOT NULL
				AND item1_id IS NOT NULL
				AND item2_id IS NOT NULL
			GROUP BY champion_id, champion_name, team_position,
					 primary_style, sub_style,
					 item0_id, item1_id, item2_id,
					 item0_name, item1_name, item2_name
		),
		RankedMetas AS ( # 메타 그룹 정렬
			SELECT *,
				ROW_NUMBER() OVER (
					PARTITION BY champion_id, champion_name, team_position
					ORDER BY total DESC, win_rate DESC
				) AS meta_rank
			FROM RefinedMetaGroups
		),
		FinalRankedMetas AS ( # 최종 메타 그룹 정렬
			SELECT rm.*,
				   fitr.item3_id,
				   fitr.item4_id,
				   fitr.item5_id,
				   fitr.item3_name,
				   fitr.item4_name,
				   fitr.item5_name,
				   mss.summoner1_id,
				   mss.summoner2_id,
				   pr.primary_perk0,
				   pr.primary_perk1,
				   pr.primary_perk2,
				   pr.primary_perk3,
				   pr.sub_perk0,
				   pr.sub_perk1,
				   pr.stat_perk_defense,
				   pr.stat_perk_flex,
				   pr.stat_perk_offense
			FROM RankedMetas rm
			LEFT JOIN FullItemTreeRanks fitr ON rm.champion_id = fitr.champion_id
				AND rm.champion_name = fitr.champion_name
				AND rm.team_position = fitr.team_position
				AND rm.primary_style = fitr.primary_style
				AND rm.sub_style = fitr.sub_style
				AND rm.item0_id = fitr.item0_id
				AND rm.item1_id = fitr.item1_id
				AND rm.item2_id = fitr.item2_id
				AND fitr.item_combo_rank = 1
			LEFT JOIN SummonerSpellRanks mss ON rm.champion_id = mss.champion_id
				AND rm.team_position = mss.team_position
				AND rm.primary_style = mss.primary_style
				AND rm.sub_style = mss.sub_style
				AND mss.spell_rank = 1
			LEFT JOIN PerkRanks pr ON rm.champion_id = pr.champion_id
				AND rm.team_position = pr.team_position
				AND rm.primary_style = pr.primary_style
				AND rm.sub_style = pr.sub_style
				AND pr.perk_rank = 1
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
