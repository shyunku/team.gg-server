package service

import (
	"database/sql"
	"errors"
	"team.gg-server/libs/db"
	"time"
)

func GetSummonerRecentMatchSummaryMXDAOs(puuid string, count int) ([]*SummonerMatchSummaryMXDAO, error) {
	var summaries []*SummonerMatchSummaryMXDAO
	if err := db.Root.Select(&summaries, `
		SELECT m.*, mp.*
		FROM summoner_matches sm
		LEFT JOIN matches m ON sm.match_id = m.match_id
		LEFT JOIN match_participants mp ON mp.match_id = m.match_id AND mp.puuid = sm.puuid
		WHERE sm.puuid = ? AND mp.match_id IS NOT NULL
		ORDER BY m.game_end_timestamp DESC
		LIMIT ?`, puuid, count,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]*SummonerMatchSummaryMXDAO, 0), nil
		}
		return nil, err
	}
	return summaries, nil
}

func GetSummonerMatchSummaryMXDAOS_before(puuid string, before int64, count int) ([]*SummonerMatchSummaryMXDAO, error) {
	var summaries []*SummonerMatchSummaryMXDAO
	if err := db.Root.Select(&summaries, `
		SELECT m.*, mp.*
		FROM summoner_matches sm
		LEFT JOIN matches m ON sm.match_id = m.match_id
		LEFT JOIN match_participants mp ON mp.match_id = m.match_id AND mp.puuid = sm.puuid
		WHERE sm.puuid = ?
		AND m.game_end_timestamp < ?
		ORDER BY m.game_end_timestamp DESC
		LIMIT ?`, puuid, before, count,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]*SummonerMatchSummaryMXDAO, 0), nil
		}
		return nil, err
	}
	return summaries, nil
}

func GetSummonerMatchSummariesBefore(puuid string, before time.Time) ([]*SummonerMatchSummaryMXDAO, error) {
	var summaries []*SummonerMatchSummaryMXDAO
	if err := db.Root.Select(&summaries, `
		SELECT m.*, mp.*
		FROM summoner_matches sm
		LEFT JOIN matches m ON sm.match_id = m.match_id
		LEFT JOIN match_participants mp ON mp.match_id = m.match_id AND mp.puuid = sm.puuid
		WHERE sm.puuid = ? AND m.game_end_timestamp < ?
		ORDER BY m.game_end_timestamp DESC
		LIMIT ?`, puuid, before, LoadInitialMatchCount,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]*SummonerMatchSummaryMXDAO, 0), nil
		}
		return nil, err
	}
	return summaries, nil
}

func GetChampionStatisticMXDAOs() ([]*ChampionStatisticMXDAO, error) {
	var statistics []*ChampionStatisticMXDAO
	if err := db.Root.Select(&statistics, `
		WITH ChampionStats AS (
			SELECT
				mp.champion_id AS champion_id,
				SUM(t.win) AS win,
				COUNT(*) AS total,
				AVG(mp.total_minions_killed) as avg_minions_killed,
				AVG(mp.kills) as avg_kills,
				AVG(mp.deaths) as avg_deaths,
				AVG(mp.assists) as avg_assists,
				AVG(mp.gold_earned) as avg_gold_earned
			FROM match_participants mp
			LEFT JOIN matches m ON mp.match_id = m.match_id
			LEFT JOIN match_teams t ON mp.team_id = t.team_id AND m.match_id = t.match_id
			LEFT JOIN match_team_bans b ON b.match_id = m.match_id AND b.champion_id = mp.champion_id
			GROUP BY mp.champion_id
		), BanStats AS (
			SELECT
				champion_id,
				COUNT(*) as total_bans
			FROM match_team_bans
			GROUP BY champion_id
		), MatchCount AS (
			SELECT
				COUNT(*) as matches
			FROM matches
		)
		SELECT
			cs.*,
			bs.total_bans / mc.matches as ban_rate,
			cs.total / mc.matches as pick_rate
		FROM ChampionStats cs
		LEFT JOIN BanStats bs ON cs.champion_id = bs.champion_id
		CROSS JOIN MatchCount mc;
	`); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]*ChampionStatisticMXDAO, 0), nil
		}
		return nil, err
	}
	return statistics, nil
}
