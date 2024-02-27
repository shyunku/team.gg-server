package service

import (
	"encoding/json"
	"fmt"
	log "github.com/shyunku-libraries/go-logger"
	"os"
	"path"
	"strconv"
	"team.gg-server/core"
	"team.gg-server/util"
	"time"
)

const StatisticsDataPath = "datafiles/statistics"

var (
	ChampionStatisticsRepo = NewChampionStatisticsRepository()
	TierStatisticsRepo     = NewTierStatisticsRepository()
	MasteryStatisticsRepo  = NewMasteryStatisticsRepository()
)

func keyPath(key string) string {
	rootDir := util.GetProjectRootDirectory()
	return path.Join(rootDir, StatisticsDataPath, key+".json")
}

type Statistics[T any] interface {
	key() string
	Period() time.Duration
	Loop()
	Collect() (*T, error)
	Save() error
	Load() (*T, error)
}

/* ----------------------- Champion statistics ----------------------- */

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
	return &ChampionStatisticsRepository{
		Cache: nil,
	}
}

func (csr *ChampionStatisticsRepository) key() string {
	return "champion_statistics"
}

func (csr *ChampionStatisticsRepository) Period() time.Duration {
	if core.DebugMode {
		return 10 * time.Second
	}
	return 1 * time.Hour
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

	// collect data
	championStatisticMXDAOs, err := GetChampionStatisticMXDAOs()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	championStatisticsMXDAOmap := make(map[int]*ChampionStatisticMXDAO)
	for _, championStatisticMXDAO := range championStatisticMXDAOs {
		championStatisticsMXDAOmap[championStatisticMXDAO.ChampionId] = championStatisticMXDAO
	}

	stats := make([]ChampionStatisticsItem, 0)
	for key, champion := range Champions {
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
	// if there is no data, collect and save
	filePath := keyPath(csr.key())
	_, err := os.Stat(filePath)
	if err != nil {
		log.Error(err)
		return nil, nil
	}
	if os.IsNotExist(err) {
		log.Debugf("file not found: %s", filePath)
		return csr.Collect()
	}

	return csr.Cache, nil
}

/* ----------------------- Tier statistics ----------------------- */

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
	Rank      Rank                         `json:"rank"`
	Level     int                          `json:"level"`
	Summoners int                          `json:"summoners"`
	Rankers   []TierStatisticsTopSummoners `json:"rankers"`
}

type TierStatisticsTierGroup struct {
	Tier       Tier                      `json:"tier"`
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
	return &TierStatisticsRepository{
		Cache: nil,
	}
}

func (tsr *TierStatisticsRepository) key() string {
	return "tier_statistics"
}

func (tsr *TierStatisticsRepository) Period() time.Duration {
	if core.DebugMode {
		return 60 * time.Second
	}
	return 6 * time.Hour
}

func (tsr *TierStatisticsRepository) Loop() {
	// must be run in a goroutine
	for {
		if _, err := tsr.Collect(); err != nil {
			log.Error(err)
		}
		time.Sleep(tsr.Period())
	}
}

func (tsr *TierStatisticsRepository) Collect() (*TierStatistics, error) {
	log.Debugf("collecting %s...", tsr.key())

	// collect data
	tierCountMXDAOs, err := GetTierStatisticsTierCountMXDAOs()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	topRankersMXDAOs, err := GetTierStatisticsTopRankersMXDAOs(30)
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

				rankLevel, err := GetRankLevel(Tier(tier), Rank(rank))
				if err != nil {
					log.Error(err)
					return nil, err
				}
				rankGroups = append(rankGroups, TierStatisticsRankGroup{
					Rank:      Rank(rank),
					Level:     rankLevel,
					Summoners: count,
					Rankers:   topSummoners,
				})
			}

			tierLevel, err := GetTierLevel(Tier(tier))
			if err != nil {
				log.Error(err)
				return nil, err
			}
			tierGroups = append(tierGroups, TierStatisticsTierGroup{
				Tier:       Tier(tier),
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
	// if there is no data, collect and save
	filePath := keyPath(tsr.key())
	_, err := os.Stat(filePath)
	if err != nil {
		log.Error(err)
		return nil, nil
	}
	if os.IsNotExist(err) {
		log.Debugf("file not found: %s", filePath)
		return tsr.Collect()
	}

	return tsr.Cache, nil
}

/* ----------------------- Mastery statistics ----------------------- */

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
	return &MasteryStatisticsRepository{
		Cache: nil,
	}
}

func (msr *MasteryStatisticsRepository) key() string {
	return "mastery_statistics"
}

func (msr *MasteryStatisticsRepository) Period() time.Duration {
	if core.DebugMode {
		return 60 * time.Second
	}
	return 6 * time.Hour
}

func (msr *MasteryStatisticsRepository) Loop() {
	// must be run in a goroutine
	for {
		if _, err := msr.Collect(); err != nil {
			log.Error(err)
		}
		time.Sleep(msr.Period())
	}
}

func (msr *MasteryStatisticsRepository) Collect() (*MasteryStatistics, error) {
	log.Debugf("collecting %s...", msr.key())

	// collect data
	masteryMXDAOs, err := GetMasteryStatisticsMXDAOs()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	masteryTopRankersMXDAO, err := GetMasteryStatisticsTopRankersMXDAOs(30)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	masteryMap := make(map[int]MasteryStatisticsItem)
	for _, masteryMXDAO := range masteryMXDAOs {
		if _, exists := masteryMap[masteryMXDAO.ChampionId]; !exists {
			champion, exists := Champions[strconv.Itoa(masteryMXDAO.ChampionId)]
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
	// if there is no data, collect and save
	filePath := keyPath(msr.key())
	_, err := os.Stat(filePath)
	if err != nil {
		log.Error(err)
		return nil, nil
	}

	if os.IsNotExist(err) {
		log.Debugf("file not found: %s", filePath)
		return msr.Collect()
	}

	return msr.Cache, nil
}
