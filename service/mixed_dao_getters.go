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
