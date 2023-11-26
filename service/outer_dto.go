package service

// Third party API response DTOs

type ChampionsInfoDto struct {
	Type    string                    `json:"type"`
	Format  string                    `json:"format"`
	Version string                    `json:"version"`
	Data    map[string]ChampionDataVO `json:"data"`
}

type DDragonSummonerJsonDto struct {
	Type    string                         `json:"type"`
	Version string                         `json:"version"`
	Data    map[string]SummonerSpellDataVO `json:"data"`
}

type PerksInfoDto []PerkInfoVO

type PerkStylesInfoDto struct {
	SchemeVersion string            `json:"schemeVersion"`
	Styles        []PerkStyleInfoVO `json:"styles"`
}
