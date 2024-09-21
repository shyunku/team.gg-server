package api

type GetSummonerPuuidRequestDto struct {
	GameName string  `form:"gameName" binding:"required"`
	TagLine  *string `form:"tagLine" binding:"required"`
}

type SetSummonerLineFavorRequestDto struct {
	CustomGameConfigId string `json:"customGameConfigId" binding:"required"`
	Puuid              string `json:"puuid" binding:"required"`
	FavorPosition      string `json:"favorPosition" binding:"required"`
	Strength           *int   `json:"strength" binding:"required"`
}
