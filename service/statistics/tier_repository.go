package statistics

import (
	"encoding/json"
	log "github.com/shyunku-libraries/go-logger"
	"os"
	"path"
	"team.gg-server/core"
	"team.gg-server/models/mixed/statistics_models"
	"team.gg-server/service"
	"team.gg-server/util"
	"time"
)

/* ----------------------- Tier statistics_models ----------------------- */

type TierStatisticsTopSummoners struct {
	Puuid         string `json:"puuid"`
	ProfileIconId int    `json:"profileIconId"`
	GameName      string `json:"gameName"`
	TagLine       string `json:"tagLine"`

	LeaguePoints int `json:"leaguePoints"`
	Wins         int `json:"wins"`
	Losses       int `json:"losses"`
	Ranks        int `json:"ranks"`
}

type TierStatisticsRankGroup struct {
	Rank      service.Rank                 `json:"rank"`
	Level     int                          `json:"level"`
	Summoners int                          `json:"summoners"`
	Rankers   []TierStatisticsTopSummoners `json:"rankers"`
}

type TierStatisticsTierGroup struct {
	Tier       service.Tier              `json:"tier"`
	Level      int                       `json:"level"`
	RankGroups []TierStatisticsRankGroup `json:"rankGroups"`
}

type TierStatisticsQueueGroup struct {
	QueueType  string                    `json:"queueType"`
	TierGroups []TierStatisticsTierGroup `json:"rankGroups"`
}

type TierStatistics struct {
	UpdatedAt   time.Time                  `json:"updatedAt"`
	QueueGroups []TierStatisticsQueueGroup `json:"queueGroups"`
}

type TierStatisticsRepository struct {
	Cache *TierStatistics
}

func NewTierStatisticsRepository() *TierStatisticsRepository {
	tsr := &TierStatisticsRepository{
		Cache: nil,
	}
	_, _ = tsr.Load()
	return tsr
}

func (tsr *TierStatisticsRepository) key() string {
	return "tier_statistics"
}

func (tsr *TierStatisticsRepository) Period() time.Duration {
	if core.DebugMode {
		return 1 * time.Hour
	}
	return 12 * time.Hour
}

func (tsr *TierStatisticsRepository) Loop() {
	for {
		if _, err := tsr.Collect(); err != nil {
			log.Error(err)
		}
		time.Sleep(tsr.Period())
	}
}

func (tsr *TierStatisticsRepository) Collect() (*TierStatistics, error) {
	log.Debugf("collecting %s...", tsr.key())
	timer := util.NewTimerWithName("TierStatisticsRepository")
	timer.Start()

	// collect data
	tierCountMXDAOs, err := statistics_models.GetTierStatisticsTierCountMXDAOs(StatisticsDB)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	topRankersMXDAOs, err := statistics_models.GetTierStatisticsTopRankersMXDAOs(StatisticsDB, 30)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	countMap := make(map[string]map[string]map[string]int)
	for _, tierCountMXDAO := range tierCountMXDAOs {
		if _, exists := countMap[tierCountMXDAO.QueueType]; !exists {
			countMap[tierCountMXDAO.QueueType] = make(map[string]map[string]int)
		}
		if _, exists := countMap[tierCountMXDAO.QueueType][tierCountMXDAO.Tier]; !exists {
			countMap[tierCountMXDAO.QueueType][tierCountMXDAO.Tier] = make(map[string]int)
		}
		countMap[tierCountMXDAO.QueueType][tierCountMXDAO.Tier][tierCountMXDAO.LeagueRank] = tierCountMXDAO.Count
	}

	statisticsMap := make(map[string]map[string]map[string][]TierStatisticsTopSummoners)
	for _, topRankerMXDAO := range topRankersMXDAOs {
		if _, exists := statisticsMap[topRankerMXDAO.QueueType]; !exists {
			statisticsMap[topRankerMXDAO.QueueType] = make(map[string]map[string][]TierStatisticsTopSummoners)
		}
		if _, exists := statisticsMap[topRankerMXDAO.QueueType][topRankerMXDAO.Tier]; !exists {
			statisticsMap[topRankerMXDAO.QueueType][topRankerMXDAO.Tier] = make(map[string][]TierStatisticsTopSummoners, 0)
		}
		if _, exists := statisticsMap[topRankerMXDAO.QueueType][topRankerMXDAO.Tier][topRankerMXDAO.LeagueRank]; !exists {
			statisticsMap[topRankerMXDAO.QueueType][topRankerMXDAO.Tier][topRankerMXDAO.LeagueRank] = make([]TierStatisticsTopSummoners, 0)
		}
		statisticsMap[topRankerMXDAO.QueueType][topRankerMXDAO.Tier][topRankerMXDAO.LeagueRank] = append(
			statisticsMap[topRankerMXDAO.QueueType][topRankerMXDAO.Tier][topRankerMXDAO.LeagueRank],
			TierStatisticsTopSummoners{
				ProfileIconId: topRankerMXDAO.ProfileIconId,
				GameName:      topRankerMXDAO.GameName,
				TagLine:       topRankerMXDAO.TagLine,
				Puuid:         topRankerMXDAO.Puuid,
				LeaguePoints:  topRankerMXDAO.LeaguePoints,
				Wins:          topRankerMXDAO.Wins,
				Losses:        topRankerMXDAO.Losses,
				Ranks:         topRankerMXDAO.Ranks,
			},
		)
	}

	queueGroups := make([]TierStatisticsQueueGroup, 0)
	for queueType, tierMap := range statisticsMap {
		tierGroups := make([]TierStatisticsTierGroup, 0)
		tierCountMap, exists := countMap[queueType]
		if !exists {
			log.Errorf("tier count map not found for queue type: %s", queueType)
			continue
		}

		for tier, rankMap := range tierMap {
			rankGroups := make([]TierStatisticsRankGroup, 0)
			rankCountMap, exists := tierCountMap[tier]
			if !exists {
				log.Errorf("tier count map not found for tier: %s", tier)
				continue
			}

			for rank, topSummoners := range rankMap {
				count, exists := rankCountMap[rank]
				if !exists {
					log.Errorf("tier count not found for rank: %s", rank)
					continue
				}

				rankLevel, err := service.GetRankLevel(service.Tier(tier), service.Rank(rank))
				if err != nil {
					log.Error(err)
					return nil, err
				}
				rankGroups = append(rankGroups, TierStatisticsRankGroup{
					Rank:      service.Rank(rank),
					Level:     rankLevel,
					Summoners: count,
					Rankers:   topSummoners,
				})
			}

			tierLevel, err := service.GetTierLevel(service.Tier(tier))
			if err != nil {
				log.Error(err)
				return nil, err
			}
			tierGroups = append(tierGroups, TierStatisticsTierGroup{
				Tier:       service.Tier(tier),
				Level:      tierLevel,
				RankGroups: rankGroups,
			})
		}
		queueGroups = append(queueGroups, TierStatisticsQueueGroup{
			QueueType:  queueType,
			TierGroups: tierGroups,
		})
	}

	tsr.Cache = &TierStatistics{
		UpdatedAt:   time.Now(),
		QueueGroups: queueGroups,
	}

	log.Debugf("%s data collected successfully in %s", tsr.key(), timer.GetDurationString())
	if err := tsr.Save(); err != nil {
		log.Warn(err)
	}

	return tsr.Cache, nil
}

func (tsr *TierStatisticsRepository) Save() error {
	if tsr.Cache == nil {
		log.Error("data is nil")
		return nil
	}

	// save data
	jsonData, err := json.Marshal(tsr.Cache)
	if err != nil {
		log.Error(err)
		return err
	}

	// create directory if not exists
	if err = os.MkdirAll(path.Join(util.GetProjectRootDirectory(), StatisticsDataPath), 0755); err != nil {
		log.Error(err)
		return err
	}

	filePath := keyPath(tsr.key())
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debugf("%s data saved to %s successfully", tsr.key(), filePath)
	return nil
}

func (tsr *TierStatisticsRepository) Load() (*TierStatistics, error) {
	if tsr.Cache != nil {
		return tsr.Cache, nil
	}

	// if there is no data, collect and save
	filePath := keyPath(tsr.key())
	_, err := os.Stat(filePath)
	if err != nil {
		log.Error(err)
		return nil, nil
	}

	// read file
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// unmarshal data
	err = json.Unmarshal(jsonData, &tsr.Cache)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return tsr.Cache, nil
}
