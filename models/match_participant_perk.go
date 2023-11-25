package models

import "database/sql"

type MatchParticipantPerkEntity struct {
	MatchParticipantId string `db:"match_participant_id" json:"matchParticipantId"`

	StatPerkDefense int `db:"stat_perk_defense" json:"statPerkDefense"`
	StatPerkFlex    int `db:"stat_perk_flex" json:"statPerkFlex"`
	StatPerkOffense int `db:"stat_perk_offense" json:"statPerkOffense"`
}

func (m *MatchParticipantPerkEntity) InsertTx(tx *sql.Tx) error {
	if _, err := tx.Exec(`
		INSERT INTO match_participant_perks
		    (match_participant_id, stat_perk_defense, stat_perk_flex, stat_perk_offense) 
		VALUES (?, ?, ?, ?)`,
		m.MatchParticipantId, m.StatPerkDefense, m.StatPerkFlex, m.StatPerkOffense,
	); err != nil {
		return err
	}
	return nil
}
