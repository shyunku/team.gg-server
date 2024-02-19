package platform

import (
	"team.gg-server/service"
	"time"
)

type GetChampionStatisticsResponseDto struct {
	UpdatedAt *time.Time                    `json:"updatedAt"`
	Stats     []service.ChampionStatisticVO `json:"stats"`
}
