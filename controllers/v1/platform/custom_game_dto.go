package platform

import "team.gg-server/service"

type GetCustomGameConfigurationsResponseDto []service.CustomGameConfigurationSummaryVO

type GetCustomGameConfigurationRequestDto struct {
	Id string `form:"id" binding:"required"`
}

type GetCustomGameConfigurationResponseDto service.CustomGameConfigurationVO

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

type SetCustomGameParticipantFavorPositionRequestDto struct {
	CustomGameConfigId string `json:"customGameConfigId" binding:"required"`
	Puuid              string `json:"puuid" binding:"required"`
	FavorPosition      string `json:"favorPosition" binding:"required"`
	Enabled            *bool  `json:"enabled" binding:"required"`
}
