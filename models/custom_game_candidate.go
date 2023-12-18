package models

import "team.gg-server/libs/db"

type CustomGameCandidateDAO struct {
	CustomGameConfigId string `db:"custom_game_config_id" json:"customGameConfigId"`
	Puuid              string `db:"puuid" json:"puuid"`
}

func (c *CustomGameCandidateDAO) Upsert(db db.Context) error {
	if _, err := db.Exec(`
	INSERT INTO custom_game_candidates (
		custom_game_config_id, puuid
	) VALUES (
		?, ?
	) ON DUPLICATE KEY UPDATE
	    custom_game_config_id = ?,
		puuid = ?`,
		c.CustomGameConfigId, c.Puuid,
		c.CustomGameConfigId, c.Puuid,
	); err != nil {
		return err
	}
	return nil
}

func GetCustomGameCandidateDAOs_byCustomGameConfigId(db db.Context, customGameConfigId string) ([]CustomGameCandidateDAO, error) {
	var customGameCandidatesDAO []CustomGameCandidateDAO
	if err := db.Select(&customGameCandidatesDAO, `
		SELECT * FROM custom_game_candidates WHERE custom_game_config_id = ?
	`, customGameConfigId); err != nil {
		return nil, err
	}
	return customGameCandidatesDAO, nil
}
