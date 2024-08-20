package models

import (
	"team.gg-server/libs/db"
	legacy_models "team.gg-server/models/legacy"
)

type MatchParticipantPerkDAO struct {
	MatchParticipantId string `db:"match_participant_id" json:"matchParticipantId"`

	StatPerkDefense int `db:"stat_perk_defense" json:"statPerkDefense"`
	StatPerkFlex    int `db:"stat_perk_flex" json:"statPerkFlex"`
	StatPerkOffense int `db:"stat_perk_offense" json:"statPerkOffense"`
}

func (m *MatchParticipantPerkDAO) InsertTx(db db.Context) error {
	if _, err := db.Exec(`
		INSERT INTO match_participant_perks
		    (match_participant_id, stat_perk_defense, stat_perk_flex, stat_perk_offense) 
		VALUES (?, ?, ?, ?)`,
		m.MatchParticipantId, m.StatPerkDefense, m.StatPerkFlex, m.StatPerkOffense,
	); err != nil {
		return err
	}
	return nil
}

func (m *MatchParticipantPerkDAO) ToLegacy() legacy_models.LegacyMatchParticipantPerkDAO {
	return legacy_models.LegacyMatchParticipantPerkDAO{
		MatchParticipantId: m.MatchParticipantId,
		StatPerkDefense:    m.StatPerkDefense,
		StatPerkFlex:       m.StatPerkFlex,
		StatPerkOffense:    m.StatPerkOffense,
	}
}

func (m *MatchParticipantPerkDAO) Delete(db db.Context) error {
	if _, err := db.Exec("DELETE FROM match_participant_perks WHERE match_participant_id = ?", m.MatchParticipantId); err != nil {
		return err
	}
	return nil
}

func GetMatchParticipantPerkDAOs_byMatchParticipantId(db db.Context, matchParticipantId string) ([]MatchParticipantPerkDAO, error) {
	var perks []MatchParticipantPerkDAO
	if err := db.Select(&perks, `
		SELECT
			match_participant_id,
			stat_perk_defense,
			stat_perk_flex,
			stat_perk_offense
		FROM match_participant_perks
		WHERE match_participant_id = ?;
	`, matchParticipantId); err != nil {
		return nil, err
	}
	return perks, nil
}
