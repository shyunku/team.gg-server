package models

import (
	"team.gg-server/libs/db"
	legacy_models "team.gg-server/models/legacy"
)

type MatchParticipantPerkStyleDAO struct {
	MatchParticipantId string `db:"match_participant_id" json:"matchParticipantId"`

	StyleId     string `db:"style_id" json:"styleId"`
	Description string `db:"description" json:"description"`
	Style       int    `db:"style" json:"style"`
}

func (m *MatchParticipantPerkStyleDAO) Insert(db db.Context) error {
	if _, err := db.Exec(`
		INSERT INTO match_participant_perk_styles
		    (match_participant_id, style_id, description, style) 
		VALUES (?, ?, ?, ?)`,
		m.MatchParticipantId, m.StyleId, m.Description, m.Style,
	); err != nil {
		return err
	}
	return nil
}

func (m *MatchParticipantPerkStyleDAO) ToLegacy() legacy_models.LegacyMatchParticipantPerkStyleDAO {
	return legacy_models.LegacyMatchParticipantPerkStyleDAO{
		MatchParticipantId: m.MatchParticipantId,
		StyleId:            m.StyleId,
		Description:        m.Description,
		Style:              m.Style,
	}
}

func (m *MatchParticipantPerkStyleDAO) Delete(db db.Context) error {
	if _, err := db.Exec("DELETE FROM match_participant_perk_styles WHERE match_participant_id = ? AND style_id = ?",
		m.MatchParticipantId, m.StyleId); err != nil {
		return err
	}
	return nil
}

func GetMatchParticipantPerkStyleDAOs(db db.Context, matchParticipantId string) ([]MatchParticipantPerkStyleDAO, error) {
	//if core.DebugOnProd {
	//	defer util.InspectFunctionExecutionTime()()
	//}
	var matchParticipantPerkStyles []MatchParticipantPerkStyleDAO
	if err := db.Select(&matchParticipantPerkStyles, `
		SELECT * FROM match_participant_perk_styles WHERE match_participant_id = ?
	`, matchParticipantId); err != nil {
		return nil, err
	}
	return matchParticipantPerkStyles, nil
}
