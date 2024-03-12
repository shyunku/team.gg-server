package statistics

import (
	"encoding/json"
	log "github.com/shyunku-libraries/go-logger"
	"os"
	"path"
	"strconv"
	"team.gg-server/core"
	"team.gg-server/models/mixed/statistics_models"
	"team.gg-server/service"
	"team.gg-server/util"
	"time"
)

/* ----------------------- Champion statistics_models ----------------------- */

type ChampionStatisticsItem struct {
	ChampionId   int    `json:"championId"`
	ChampionName string `json:"championName"`

	Win   int `json:"win"`
	Total int `json:"total"`

	AvgPickRate float64 `json:"avgPickRate"`
	AvgBanRate  float64 `json:"avgBanRate"`

	AvgMinionsKilled float64 `json:"avgMinionsKilled"`
	AvgKills         float64 `json:"avgKills"`
	AvgDeaths        float64 `json:"avgDeaths"`
	AvgAssists       float64 `json:"avgAssists"`
	AvgGoldEarned    float64 `json:"avgGoldEarned"`
}

type ChampionStatistics struct {
	UpdatedAt time.Time                `json:"updatedAt"`
	Data      []ChampionStatisticsItem `json:"data"`
}

type ChampionStatisticsRepository struct {
	Cache *ChampionStatistics
}

func NewChampionStatisticsRepository() *ChampionStatisticsRepository {
	csr := &ChampionStatisticsRepository{
		Cache: nil,
	}
	_, _ = csr.Load()
	return csr
}

func (csr *ChampionStatisticsRepository) key() string {
	return "champion_statistics"
}

func (csr *ChampionStatisticsRepository) Period() time.Duration {
	if core.DebugMode {
		return 1 * time.Hour
	}
	return 6 * time.Hour
}

func (csr *ChampionStatisticsRepository) Loop() {
	// must be run in a goroutine
	for {
		if _, err := csr.Collect(); err != nil {
			log.Error(err)
		}
		time.Sleep(csr.Period())
	}
}

func (csr *ChampionStatisticsRepository) Collect() (*ChampionStatistics, error) {
	log.Debugf("collecting %s...", csr.key())
	timer := util.NewTimerWithName("ChampionStatisticsRepository")
	timer.Start()

	// collect data
	championStatisticMXDAOs, err := statistics_models.GetChampionStatisticMXDAOs(StatisticsDB)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	championStatisticsMXDAOmap := make(map[int]*statistics_models.ChampionStatisticMXDAO)
	for _, championStatisticMXDAO := range championStatisticMXDAOs {
		championStatisticsMXDAOmap[championStatisticMXDAO.ChampionId] = championStatisticMXDAO
	}

	stats := make([]ChampionStatisticsItem, 0)
	for key, champion := range service.Champions {
		championId, err := strconv.Atoi(key)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		if championStatisticMXDAO, exists := championStatisticsMXDAOmap[championId]; exists {
			stats = append(stats, ChampionStatisticsItem{
				ChampionId:       championId,
				ChampionName:     champion.Name,
				Win:              championStatisticMXDAO.Win,
				Total:            championStatisticMXDAO.Total,
				AvgPickRate:      championStatisticMXDAO.PickRate,
				AvgBanRate:       championStatisticMXDAO.BanRate,
				AvgMinionsKilled: championStatisticMXDAO.AvgMinionsKilled,
				AvgKills:         championStatisticMXDAO.AvgKills,
				AvgDeaths:        championStatisticMXDAO.AvgDeaths,
				AvgAssists:       championStatisticMXDAO.AvgAssists,
				AvgGoldEarned:    championStatisticMXDAO.AvgGoldEarned,
			})
		} else {
			stats = append(stats, ChampionStatisticsItem{
				ChampionId:       championId,
				ChampionName:     champion.Name,
				Win:              0,
				Total:            0,
				AvgPickRate:      0,
				AvgBanRate:       0,
				AvgMinionsKilled: 0,
				AvgKills:         0,
				AvgDeaths:        0,
				AvgAssists:       0,
				AvgGoldEarned:    0,
			})
		}
	}

	csr.Cache = &ChampionStatistics{
		UpdatedAt: time.Now(),
		Data:      stats,
	}

	log.Debugf("%s data collected successfully in %s", csr.key(), timer.GetDurationString())
	if err := csr.Save(); err != nil {
		log.Warn(err)
	}

	return csr.Cache, nil
}

func (csr *ChampionStatisticsRepository) Save() error {
	if csr.Cache == nil {
		log.Error("data is nil")
		return nil
	}

	// save data
	jsonData, err := json.Marshal(csr.Cache)
	if err != nil {
		log.Error(err)
		return err
	}

	// create directory if not exists
	if err = os.MkdirAll(path.Join(util.GetProjectRootDirectory(), StatisticsDataPath), 0755); err != nil {
		log.Error(err)
		return err
	}

	filePath := keyPath(csr.key())
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debugf("%s data saved to %s successfully", csr.key(), filePath)
	return nil
}

func (csr *ChampionStatisticsRepository) Load() (*ChampionStatistics, error) {
	if csr.Cache != nil {
		return csr.Cache, nil
	}

	// if there is no data, collect and save
	filePath := keyPath(csr.key())
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Debugf("file not found: %s", filePath)
			return csr.Collect()
		}
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
	err = json.Unmarshal(jsonData, &csr.Cache)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return csr.Cache, nil
}
