package models

import "team.gg-server/libs/db"

type CustomGameParticipantDAO struct {
	CustomGameConfigId string `db:"custom_game_config_id" json:"customGameConfigId"`
	Puuid              string `db:"puuid" json:"puuid"`
	Team               int    `db:"team" json:"team"`
	Position           string `db:"position" json:"position"`
}

func (s *CustomGameParticipantDAO) Upsert(db db.Context) error {
	if _, err := db.Exec(`
		INSERT INTO custom_game_participants
		    (custom_game_config_id, puuid, team, position) 
		VALUES (?, ?, ?, ?) 
		ON DUPLICATE KEY UPDATE 
			custom_game_config_id = ?, puuid = ?, team = ?, position = ?`,
		s.CustomGameConfigId, s.Puuid, s.Team, s.Position,
		s.CustomGameConfigId, s.Puuid, s.Team, s.Position,
	); err != nil {
		return err
	}
	return nil
}

func GetCustomGameParticipantsDAOs_byCustomGameConfigId(db db.Context, customGameConfigId string) ([]*CustomGameParticipantDAO, error) {
	var customGameParticipants []*CustomGameParticipantDAO
	if err := db.Select(&customGameParticipants, "SELECT * FROM custom_game_participants WHERE custom_game_config_id = ?", customGameConfigId); err != nil {
		return nil, err
	}
	return customGameParticipants, nil
}
