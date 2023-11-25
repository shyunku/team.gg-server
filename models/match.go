package models

import (
	"database/sql"
	"errors"
	"team.gg-server/libs/database"
)

type MatchEntity struct {
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

func (m *MatchEntity) InsertTx(tx *sql.Tx) error {
	if _, err := tx.Exec(`
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

func StrictGetMatchByMatchId(matchId string) (*MatchEntity, bool, error) {
	// check if match exists in db
	matchEntity, err := GetMatchEntityByMatchId(matchId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if matchEntity == nil {
		return nil, false, nil
	}
	return matchEntity, true, nil
}

func GetMatchEntityByMatchId(matchId string) (*MatchEntity, error) {
	var match MatchEntity
	if err := database.DB.Get(&match, "SELECT * FROM matches WHERE match_id = ?", matchId); err != nil {
		return nil, err
	}
	return &match, nil
}
