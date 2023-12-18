package models

import (
	"team.gg-server/libs/db"
	"time"
)

type CustomGameConfigurationDAO struct {
	Id            string    `db:"id" json:"id"`
	Name          string    `db:"name" json:"name"`
	CreatorUid    string    `db:"creator_uid" json:"creatorUid"`
	CreatedAt     time.Time `db:"created_at" json:"createdAt"`
	LastUpdatedAt time.Time `db:"last_updated_at" json:"lastUpdatedAt"`
	Fairness      float64   `db:"fairness" json:"fairness"`
	LineFairness  float64   `db:"line_fairness" json:"lineFairness"`
	TierFairness  float64   `db:"tier_fairness" json:"tierFairness"`
}

func (c *CustomGameConfigurationDAO) Upsert(db db.Context) error {
	if _, err := db.Exec(`
	INSERT INTO custom_game_configurations (
		id, name, creator_uid, created_at, last_updated_at, fairness, line_fairness, tier_fairness
	) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?
	) ON DUPLICATE KEY UPDATE
	    name = ?,
		last_updated_at = ?,
		fairness = ?,
		line_fairness = ?,
		tier_fairness = ?`,
		c.Id, c.Name, c.CreatorUid, c.CreatedAt, c.LastUpdatedAt, c.Fairness, c.LineFairness, c.TierFairness,
		c.Name, c.LastUpdatedAt, c.Fairness, c.LineFairness, c.TierFairness,
	); err != nil {
		return err
	}
	return nil
}

func GetCustomGameDAOs_byCreatorUid(db db.Context, uid string) ([]CustomGameConfigurationDAO, error) {
	var customGameDAOs []CustomGameConfigurationDAO
	if err := db.Select(&customGameDAOs, `
		SELECT * FROM custom_game_configurations WHERE creator_uid = ?
	`, uid); err != nil {
		return nil, err
	}
	return customGameDAOs, nil
}

func GetCustomGameDAO_byId(db db.Context, id string) (*CustomGameConfigurationDAO, bool, error) {
	var customGameDAO CustomGameConfigurationDAO
	if err := db.Get(&customGameDAO, `
		SELECT * FROM custom_game_configurations WHERE id = ?
	`, id); err != nil {
		return nil, false, err
	}
	return &customGameDAO, true, nil
}
