package platform

type GetChampionStatisticsDetailRequestDto struct {
	ChampionId int `form:"championId" binding:"required"`
}
