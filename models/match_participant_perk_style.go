package models

import (
	"database/sql"
	"team.gg-server/libs/database"
)

type MatchParticipantPerkStyleEntity struct {
	MatchParticipantId string `db:"match_participant_id" json:"matchParticipantId"`

	StyleId     string `db:"style_id" json:"styleId"`
	Description string `db:"description" json:"description"`
	Style       int    `db:"style" json:"style"`
}

func (m *MatchParticipantPerkStyleEntity) InsertTx(tx *sql.Tx) error {
	if _, err := tx.Exec(`
		INSERT INTO match_participant_perk_styles
		    (match_participant_id, style_id, description, style) 
		VALUES (?, ?, ?, ?)`,
		m.MatchParticipantId, m.StyleId, m.Description, m.Style,
	); err != nil {
		return err
	}
	return nil
}

func GetMatchParticipantPerkStylesByMatchParticipantId(matchParticipantId string) ([]*MatchParticipantPerkStyleEntity, error) {
	var matchParticipantPerkStyles []*MatchParticipantPerkStyleEntity
	if err := database.DB.Select(&matchParticipantPerkStyles, "SELECT * FROM match_participant_perk_styles WHERE match_participant_id = ?", matchParticipantId); err != nil {
		return nil, err
	}
	return matchParticipantPerkStyles, nil
}
