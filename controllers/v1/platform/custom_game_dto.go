package platform

import "team.gg-server/service"

type GetCustomGameConfigurationsResponseDto []service.CustomGameConfigurationSummaryVO

type GetCustomGameConfigurationRequestDto struct {
	Id string `form:"id" binding:"required"`
}

type GetCustomGameConfigurationResponseDto service.CustomGameConfigurationVO

type GetTierRankRequestDto struct {
	RatingPoint *float64 `form:"ratingPoint" binding:"required"`
}

type GetTierRankResponseDto struct {
	Tier string `json:"tier"`
	Rank string `json:"rank"`
	Lp   int64  `json:"lp"`
}

type GetCustomConfigurationBalanceRequestDto struct {
	Id string `form:"id" binding:"required"`
}

type GetCustomConfigurationSummaryResponseDto service.CustomGameConfigurationSummaryVO

type AddCandidateToCustomGameRequestDto struct {
	CustomGameConfigId string `json:"customGameConfigId" binding:"required"`
	Name               string `json:"name" binding:"required"`
	TagLine            string `json:"tagLine" binding:"required"`
}

type AddCandidateToCustomGameResponseDto service.CustomGameCandidateVO

type ArrangeCustomGameParticipantRequestDto struct {
	CustomGameConfigId string `json:"customGameConfigId" binding:"required"`
	Puuid              string `json:"puuid" binding:"required"`
	Team               int    `json:"team" binding:"required"`
	TargetPosition     string `json:"targetPosition" binding:"required"`
}

type UnarrangeCustomGameParticipantRequestDto struct {
	CustomGameConfigId string `json:"customGameConfigId" binding:"required"`
	Puuid              string `json:"puuid" binding:"required"`
}

type SetCustomGameParticipantFavorPositionRequestDto struct {
	CustomGameConfigId string `json:"customGameConfigId" binding:"required"`
	Puuid              string `json:"puuid" binding:"required"`
	FavorPosition      string `json:"favorPosition" binding:"required"`
	Strength           *int   `json:"strength" binding:"required"`
}

type SetCustomGameCandidateCustomTierRankRequestDto struct {
	CustomGameConfigId string  `json:"customGameConfigId" binding:"required"`
	Puuid              string  `json:"puuid" binding:"required"`
	Tier               *string `json:"tier"`
	Rank               *string `json:"rank"`
}

type OptimizeCustomGameConfigurationRequestDto struct {
	Id string `json:"id" binding:"required"`

	LineFairnessWeight *float64 `json:"lineFairnessWeight" binding:"required"`
	TierFairnessWeight *float64 `json:"tierFairnessWeight" binding:"required"`

	TopInfluenceWeight     *float64 `json:"topInfluenceWeight" binding:"required"`
	JungleInfluenceWeight  *float64 `json:"jungleInfluenceWeight" binding:"required"`
	MidInfluenceWeight     *float64 `json:"midInfluenceWeight" binding:"required"`
	AdcInfluenceWeight     *float64 `json:"adcInfluenceWeight" binding:"required"`
	SupportInfluenceWeight *float64 `json:"supportInfluenceWeight" binding:"required"`
}

type UtilityRequestDto struct {
	Id string `json:"id" binding:"required"`
}
