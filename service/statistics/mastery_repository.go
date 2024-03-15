package statistics

import (
	"encoding/json"
	"fmt"
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

/* ----------------------- Mastery statistics_models ----------------------- */

type MasteryStatisticsTopSummoners struct {
	Puuid         string `json:"puuid"`
	ProfileIconId int    `json:"profileIconId"`
	GameName      string `json:"gameName"`
	TagLine       string `json:"tagLine"`

	ChampionPoints int `json:"championPoints"`
	Ranks          int `json:"ranks"`
}

type MasteryStatisticsItem struct {
	ChampionId   int    `json:"championId"`
	ChampionName string `json:"championName"`

	AvgMastery   float64 `json:"avgMastery"`
	MaxMastery   int64   `json:"maxMastery"`
	TotalMastery int64   `json:"totalMastery"`

	MasteredCount int                             `json:"masteredCount"`
	Summoners     int                             `json:"summoners"`
	Rankers       []MasteryStatisticsTopSummoners `json:"rankers"`
}

type MasteryStatistics struct {
	UpdatedAt     time.Time               `json:"updatedAt"`
	MasteryGroups []MasteryStatisticsItem `json:"masteryGroups"`
}

type MasteryStatisticsRepository struct {
	Cache *MasteryStatistics
}

func NewMasteryStatisticsRepository() *MasteryStatisticsRepository {
	msr := &MasteryStatisticsRepository{
		Cache: nil,
	}
	return msr
}

func (msr *MasteryStatisticsRepository) key() string {
	return "mastery_statistics"
}

func (msr *MasteryStatisticsRepository) Period() time.Duration {
	if core.DebugMode {
		return 1 * time.Hour
	}
	return 12 * time.Hour
}

func (msr *MasteryStatisticsRepository) Loop() {
	for {
		if _, err := msr.Collect(); err != nil {
			log.Error(err)
		}
		time.Sleep(msr.Period())
	}
}

func (msr *MasteryStatisticsRepository) Collect() (*MasteryStatistics, error) {
	log.Debugf("collecting %s...", msr.key())
	timer := util.NewTimerWithName("MasteryStatisticsRepository")
	timer.Start()

	// collect data
	masteryMXDAOs, err := statistics_models.GetMasteryStatisticsMXDAOs(StatisticsDB)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	masteryTopRankersMXDAO, err := statistics_models.GetMasteryStatisticsTopRankersMXDAOs(StatisticsDB, 30)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	masteryMap := make(map[int]MasteryStatisticsItem)
	for _, masteryMXDAO := range masteryMXDAOs {
		if _, exists := masteryMap[masteryMXDAO.ChampionId]; !exists {
			champion, exists := service.Champions[strconv.Itoa(masteryMXDAO.ChampionId)]
			if !exists {
				log.Errorf("champion not found: %d", masteryMXDAO.ChampionId)
				return nil, fmt.Errorf("champion not found: %d", masteryMXDAO.ChampionId)
			}

			masteryMap[masteryMXDAO.ChampionId] = MasteryStatisticsItem{
				ChampionId:    masteryMXDAO.ChampionId,
				ChampionName:  champion.Name,
				AvgMastery:    masteryMXDAO.AvgMastery,
				MaxMastery:    int64(masteryMXDAO.MaxMastery),
				TotalMastery:  int64(masteryMXDAO.TotalMastery),
				MasteredCount: masteryMXDAO.MasteredCount,
				Summoners:     masteryMXDAO.Count,
				Rankers:       make([]MasteryStatisticsTopSummoners, 0),
			}
		}
	}

	for _, masteryTopRankerMXDAO := range masteryTopRankersMXDAO {
		mastery, exists := masteryMap[masteryTopRankerMXDAO.ChampionId]
		if !exists {
			log.Errorf("mastery not found: %d", masteryTopRankerMXDAO.ChampionId)
			return nil, fmt.Errorf("mastery not found: %d", masteryTopRankerMXDAO.ChampionId)
		}
		mastery.Rankers = append(mastery.Rankers, MasteryStatisticsTopSummoners{
			Puuid:          masteryTopRankerMXDAO.Puuid,
			ProfileIconId:  masteryTopRankerMXDAO.ProfileIconId,
			GameName:       masteryTopRankerMXDAO.GameName,
			TagLine:        masteryTopRankerMXDAO.TagLine,
			ChampionPoints: masteryTopRankerMXDAO.ChampionPoints,
			Ranks:          masteryTopRankerMXDAO.Ranks,
		})
		masteryMap[masteryTopRankerMXDAO.ChampionId] = mastery
	}

	msr.Cache = &MasteryStatistics{
		UpdatedAt:     time.Now(),
		MasteryGroups: make([]MasteryStatisticsItem, 0),
	}
	for _, mastery := range masteryMap {
		msr.Cache.MasteryGroups = append(msr.Cache.MasteryGroups, mastery)
	}

	log.Debugf("%s data collected successfully in %s", msr.key(), timer.GetDurationString())
	if err := msr.Save(); err != nil {
		log.Warn(err)
	}

	return msr.Cache, nil
}

func (msr *MasteryStatisticsRepository) Save() error {
	if msr.Cache == nil {
		log.Error("data is nil")
		return nil
	}

	// save data
	jsonData, err := json.Marshal(msr.Cache)
	if err != nil {
		log.Error(err)
		return err
	}

	// create directory if not exists
	if err = os.MkdirAll(path.Join(util.GetProjectRootDirectory(), StatisticsDataPath), 0755); err != nil {
		log.Error(err)
		return err
	}

	filePath := keyPath(msr.key())
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debugf("%s data saved to %s successfully", msr.key(), filePath)
	return nil
}

func (msr *MasteryStatisticsRepository) Load() (*MasteryStatistics, error) {
	if msr.Cache != nil {
		return msr.Cache, nil
	}

	// if there is no data, collect and save
	filePath := keyPath(msr.key())
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Debugf("file not found: %s", filePath)
			return msr.Collect()
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
	err = json.Unmarshal(jsonData, &msr.Cache)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return msr.Cache, nil
}
