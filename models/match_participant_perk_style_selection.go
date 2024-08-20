package models

import (
	"team.gg-server/libs/db"
	legacy_models "team.gg-server/models/legacy"
)

type MatchParticipantPerkStyleSelectionDAO struct {
	StyleId string `db:"style_id" json:"styleId"`

	Perk int `db:"perk" json:"perk"`
	Var1 int `db:"var1" json:"var1"`
	Var2 int `db:"var2" json:"var2"`
	Var3 int `db:"var3" json:"var3"`
}

func (m *MatchParticipantPerkStyleSelectionDAO) Insert(db db.Context) error {
	if _, err := db.Exec(`
		INSERT INTO match_participant_perk_style_selections
		    (style_id, perk, var1, var2, var3) 
		VALUES (?, ?, ?, ?, ?)`,
		m.StyleId, m.Perk, m.Var1, m.Var2, m.Var3,
	); err != nil {
		return err
	}
	return nil
}

func (m *MatchParticipantPerkStyleSelectionDAO) ToLegacy() legacy_models.LegacyMatchParticipantPerkStyleSelectionDAO {
	return legacy_models.LegacyMatchParticipantPerkStyleSelectionDAO{
		StyleId: m.StyleId,
		Perk:    m.Perk,
		Var1:    m.Var1,
		Var2:    m.Var2,
		Var3:    m.Var3,
	}
}

func (m *MatchParticipantPerkStyleSelectionDAO) Delete(db db.Context) error {
	if _, err := db.Exec("DELETE FROM match_participant_perk_style_selections WHERE style_id = ?", m.StyleId); err != nil {
		return err
	}
	return nil
}

func GetMatchParticipantPerkStyleSelectionDAOs(db db.Context, styleId string) ([]MatchParticipantPerkStyleSelectionDAO, error) {
	//if core.DebugOnProd {
	//	defer util.InspectFunctionExecutionTime()()
	//}
	var matchParticipantPerkStyleSelections []MatchParticipantPerkStyleSelectionDAO
	if err := db.Select(&matchParticipantPerkStyleSelections, `
		SELECT * FROM match_participant_perk_style_selections WHERE style_id = ?
	`, styleId); err != nil {
		return nil, err
	}
	return matchParticipantPerkStyleSelections, nil
}
