package statistics

import (
	"path"
	"team.gg-server/libs/db"
	"team.gg-server/util"
	"time"
)

const StatisticsDataPath = "datafiles/statistics"

var (
	StatisticsDB                 *db.Database                        = nil
	ChampionDetailStatisticsRepo *ChampionDetailStatisticsRepository = nil
	TierStatisticsRepo           *TierStatisticsRepository           = nil
	MasteryStatisticsRepo        *MasteryStatisticsRepository        = nil
)

type Statistics[T any] interface {
	key() string
	Period() time.Duration
	Loop()
	Collect() (*T, error)
	Save() error
	Load() (*T, error)
}

func InitializeStatisticRepos() {
	ChampionDetailStatisticsRepo = NewChampionDetailStatisticsRepository()
	TierStatisticsRepo = NewTierStatisticsRepository()
	MasteryStatisticsRepo = NewMasteryStatisticsRepository()
}

func keyPath(key string) string {
	rootDir := util.GetProjectRootDirectory()
	return path.Join(rootDir, StatisticsDataPath, key+".json")
}
