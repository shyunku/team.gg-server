package legacy_models

import (
	"team.gg-server/libs/db"
)

type LegacyMatchParticipantPerkStyleSelectionDAO struct {
	StyleId string `db:"style_id" json:"styleId"`

	Perk int `db:"perk" json:"perk"`
	Var1 int `db:"var1" json:"var1"`
	Var2 int `db:"var2" json:"var2"`
	Var3 int `db:"var3" json:"var3"`
}

func (m *LegacyMatchParticipantPerkStyleSelectionDAO) Insert(db db.Context) error {
	if _, err := db.Exec(`
		INSERT INTO legacy_match_participant_perk_style_selections
		    (style_id, perk, var1, var2, var3) 
		VALUES (?, ?, ?, ?, ?)`,
		m.StyleId, m.Perk, m.Var1, m.Var2, m.Var3,
	); err != nil {
		return err
	}
	return nil
}
