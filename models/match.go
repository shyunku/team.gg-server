package models

import (
	"database/sql"
	"errors"
	"team.gg-server/libs/db"
)

type MatchDAO struct {
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

func (m *MatchDAO) Insert(db db.Context) error {
	if _, err := db.Exec(`
		INSERT INTO matches
		    (data_version, match_id, game_creation, game_duration, game_end_timestamp, game_id, game_mode, game_name, game_start_timestamp, game_type, game_version, map_id, platform_id, queue_id, tournament_code) 
		VALUE
		    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.DataVersion, m.MatchId, m.GameCreation, m.GameDuration, m.GameEndTimestamp, m.GameId, m.GameMode, m.GameName, m.GameStartTimestamp, m.GameType, m.GameVersion, m.MapId, m.PlatformId, m.QueueId, m.TournamentCode,
	); err != nil {
		return err
	}
	return nil
}

func GetMatchDAO(db db.Context, matchId string) (*MatchDAO, bool, error) {
	var matchEntity MatchDAO
	if err := db.Get(&matchEntity, "SELECT * FROM matches WHERE match_id = ?", matchId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &matchEntity, true, nil
}

func GetSummonerMatchDAOs_before(db db.Context, puuid string, before int64, limit int64) ([]*MatchDAO, error) {
	var matches []*MatchDAO
	if err := db.Select(&matches, `
		SELECT m.*
		FROM summoner_matches sm
		LEFT JOIN matches m ON sm.match_id = m.match_id
		WHERE sm.puuid = ?
		AND m.game_end_timestamp < ?
		ORDER BY m.game_end_timestamp DESC 
		LIMIT ?;
	`, puuid, before, limit); err != nil {
		return nil, err
	}
	return matches, nil
}

// GetOldestSummonerMatchDAO returns the oldest match for a summoner
func GetOldestSummonerMatchDAO(db db.Context, puuid string) (*MatchDAO, bool, error) {
	var matchEntity MatchDAO
	if err := db.Get(&matchEntity, `
		SELECT m.*
		FROM summoner_matches sm
		LEFT JOIN matches m ON sm.match_id = m.match_id
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
