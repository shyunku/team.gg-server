package legacy_models

import (
	"database/sql"
	"errors"
	"team.gg-server/core"
	"team.gg-server/libs/db"
	"team.gg-server/util"
)

type LegacyMatchDAO struct {
	MatchId string `db:"match_id" json:"matchId"`

	DataVersion        string `db:"data_version" json:"dataVersion"`
	GameCreation       int64  `db:"game_creation" json:"gameCreation"`
	GameDuration       int64  `db:"game_duration" json:"gameDuration"`
	GameEndTimestamp   int64  `db:"game_end_timestamp" json:"gameEndTimestamp"`
	GameId             int64  `db:"game_id" json:"gameId"`
	GameMode           string `db:"game_mode" json:"gameMode"`
	GameName           string `db:"game_name" json:"gameName"`
	GameStartTimestamp int64  `db:"game_start_timestamp" json:"gameStartTimestamp"`
	GameType           string `db:"game_type" json:"gameType"`
	GameVersion        string `db:"game_version" json:"gameVersion"`
	MapId              int    `db:"map_id" json:"mapId"`
	PlatformId         string `db:"platform_id" json:"platformId"`
	QueueId            int    `db:"queue_id" json:"queueId"`
	TournamentCode     string `db:"tournament_code" json:"tournamentCode"`
}

func (m *LegacyMatchDAO) Insert(db db.Context) error {
	if _, err := db.Exec(`
		INSERT INTO legacy_matches
		    (data_version, match_id, game_creation, game_duration, game_end_timestamp, game_id, game_mode, game_name, game_start_timestamp, game_type, game_version, map_id, platform_id, queue_id, tournament_code) 
		VALUE
		    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.DataVersion, m.MatchId, m.GameCreation, m.GameDuration, m.GameEndTimestamp, m.GameId, m.GameMode, m.GameName, m.GameStartTimestamp, m.GameType, m.GameVersion, m.MapId, m.PlatformId, m.QueueId, m.TournamentCode,
	); err != nil {
		return err
	}
	return nil
}

func GetLegacyMatchDAO(db db.Context, matchId string) (*LegacyMatchDAO, bool, error) {
	var matchEntity LegacyMatchDAO
	if err := db.Get(&matchEntity, "SELECT * FROM legacy_matches WHERE match_id = ?", matchId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &matchEntity, true, nil
}

// GetOldestLegacySummonerMatchDAO returns the oldest match for a summoner
func GetOldestLegacySummonerMatchDAO(db db.Context, puuid string) (*LegacyMatchDAO, bool, error) {
	var matchEntity LegacyMatchDAO
	if err := db.Get(&matchEntity, `
		SELECT m.*
		FROM legacy_summoner_matches sm
		LEFT JOIN legacy_matches m ON sm.match_id = m.match_id
		WHERE sm.puuid = ?
		ORDER BY m.game_end_timestamp DESC
		LIMIT 1;
	`, puuid); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &matchEntity, true, nil
}

func GetLegacyMatchDAOs_byPuuid(db db.Context, puuid string, count int) ([]LegacyMatchDAO, error) {
	if core.DebugOnProd {
		defer util.InspectFunctionExecutionTime()()
	}

	var matches []LegacyMatchDAO
	if err := db.Select(&matches, `
		SELECT m.*
		FROM legacy_summoner_matches sm
		LEFT JOIN legacy_matches m ON sm.match_id = m.match_id
		WHERE sm.puuid = ?
		ORDER BY m.game_end_timestamp DESC
		LIMIT ?;
	`, puuid, count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]LegacyMatchDAO, 0), nil
		}
		return nil, err
	}
	return matches, nil
}

func GetLegacyMatchDAOs_byPuuid_before(db db.Context, puuid string, before int64, count int) ([]LegacyMatchDAO, error) {
	if core.DebugOnProd {
		defer util.InspectFunctionExecutionTime()()
	}

	var matches []LegacyMatchDAO
	if err := db.Select(&matches, `
		SELECT m.*
		FROM legacy_summoner_matches sm
		LEFT JOIN legacy_matches m ON sm.match_id = m.match_id
		WHERE sm.puuid = ?
		AND m.game_end_timestamp < ?
		ORDER BY m.game_end_timestamp DESC
		LIMIT ?;
	`, puuid, before, count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]LegacyMatchDAO, 0), nil
		}
		return nil, err
	}
	return matches, nil
}

func GetLegacyMatchDAOs_byQueueId(db db.Context, puuid string, queueId, count int) ([]LegacyMatchDAO, error) {
	if core.DebugOnProd {
		defer util.InspectFunctionExecutionTime()()
	}

	var matches []LegacyMatchDAO
	if err := db.Select(&matches, `
		SELECT m.*
		FROM legacy_summoner_matches sm
		LEFT JOIN legacy_matches m ON sm.match_id = m.match_id
		WHERE sm.puuid = ?
		AND m.queue_id = ?
		ORDER BY m.game_end_timestamp DESC
		LIMIT ?;
	`, puuid, queueId, count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]LegacyMatchDAO, 0), nil
		}
		return nil, err
	}
	return matches, nil
}

func GetLegacyMatchDAOs_byQueueId_before(db db.Context, puuid string, queueId int, before int64, count int) ([]LegacyMatchDAO, error) {
	if core.DebugOnProd {
		defer util.InspectFunctionExecutionTime()()
	}

	var matches []LegacyMatchDAO
	if err := db.Select(&matches, `
		SELECT m.*
		FROM legacy_summoner_matches sm
		LEFT JOIN legacy_matches m ON sm.match_id = m.match_id
		WHERE sm.puuid = ?
		AND m.queue_id = ?
		AND m.game_end_timestamp < ?
		ORDER BY m.game_end_timestamp DESC
		LIMIT ?;
	`, puuid, queueId, before, count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]LegacyMatchDAO, 0), nil
		}
		return nil, err
	}
	return matches, nil
}
