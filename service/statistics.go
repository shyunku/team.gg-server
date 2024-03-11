package service

import (
	"encoding/json"
	"fmt"
	uuid2 "github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	log "github.com/shyunku-libraries/go-logger"
	"os"
	"path"
	"sort"
	"strconv"
	"team.gg-server/core"
	"team.gg-server/models/mixed/statistics"
	"team.gg-server/types"
	"team.gg-server/util"
	"time"
)

const StatisticsDataPath = "datafiles/statistics"

var (
	StatisticsDB                 *sqlx.DB                            = nil
	ChampionStatisticsRepo       *ChampionStatisticsRepository       = nil
	ChampionDetailStatisticsRepo *ChampionDetailStatisticsRepository = nil
	TierStatisticsRepo           *TierStatisticsRepository           = nil
	MasteryStatisticsRepo        *MasteryStatisticsRepository        = nil
)

func InitializeStatisticRepos() {
	ChampionStatisticsRepo = NewChampionStatisticsRepository()
	ChampionDetailStatisticsRepo = NewChampionDetailStatisticsRepository()
	TierStatisticsRepo = NewTierStatisticsRepository()
	MasteryStatisticsRepo = NewMasteryStatisticsRepository()
}

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
	championStatisticMXDAOs, err := statistics.GetChampionStatisticMXDAOs(StatisticsDB)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	championStatisticsMXDAOmap := make(map[int]*statistics.ChampionStatisticMXDAO)
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

/* ----------------------- Champion Detail statistics ----------------------- */

type ChampionDetailStatisticsMeta struct {
	MetaName string `json:"metaName"`

	MajorTag string  `json:"majorTag"`
	MinorTag *string `json:"minorTag"`

	MajorPerkStyleId int   `json:"majorPerkStyleId"`
	MinorPerkStyleId int   `json:"minorPerkStyleId"`
	ItemTree         []int `json:"itemTree"` // must be sorted as descending order with gold

	Count    int     `json:"count"`
	Win      int     `json:"win"`
	PickRate float64 `json:"pickRate"`
}

type ChampionDetailStatisticsMetaTree struct {
	MajorMetaPicks []ChampionDetailStatisticsMeta `json:"majorMetaPicks"`
	MinorMetaPick  *ChampionDetailStatisticsMeta  `json:"minorMetaPick"`
}

type ChampionDetailStatisticsPositionMetaTree struct {
	Top     ChampionDetailStatisticsMetaTree `json:"top"`
	Jungle  ChampionDetailStatisticsMetaTree `json:"jungle"`
	Mid     ChampionDetailStatisticsMetaTree `json:"mid"`
	ADC     ChampionDetailStatisticsMetaTree `json:"adc"`
	Support ChampionDetailStatisticsMetaTree `json:"support"`
}

type ChampionDetailStatisticsExtraStats struct {
	AvgMinionsKilled float64 `json:"avgMinionsKilled"`
	AvgKills         float64 `json:"avgKills"`
	AvgDeaths        float64 `json:"avgDeaths"`
	AvgAssists       float64 `json:"avgAssists"`
	AvgGoldEarned    float64 `json:"avgGoldEarned"`
}

type ChampionDetailStatisticsNormalizedStats struct {
	AvgTotalHealPerSec           float64 `json:"avgTotalHealPerSec"`
	AvgVisionScorePerSec         float64 `json:"avgVisionScorePerSec"`
	AvgTotalDamageTakenPerSec    float64 `json:"avgTotalDamageTakenPerSec"`
	AvgTotalTimeCCDealtPerSec    float64 `json:"avgTotalTimeCCDealtPerSec"`
	AvgDamageSelfMitigatedPerSec float64 `json:"avgDamageSelfMitigatedPerSec"`
}

type ChampionDetailStatisticsItem struct {
	ChampionId   int    `json:"championId"`
	ChampionName string `json:"championName"`

	Win         int     `json:"win"`
	Total       int     `json:"total"`
	AvgPickRate float64 `json:"avgPickRate"`
	AvgBanRate  float64 `json:"avgBanRate"`

	ExtraStats      ChampionDetailStatisticsExtraStats      `json:"extraStats"`
	NormalizedStats ChampionDetailStatisticsNormalizedStats `json:"normalizedStats"`

	TeamPosition string                                   `json:"teamPosition"`
	MetaTree     ChampionDetailStatisticsPositionMetaTree `json:"metaTree"`
}

type ChampionDetailStatistics struct {
	UpdatedAt time.Time                            `json:"updatedAt"`
	Data      map[int]ChampionDetailStatisticsItem `json:"data"`
}

type ChampionDetailStatisticsRepository struct {
	Cache *ChampionDetailStatistics
}

func NewChampionDetailStatisticsRepository() *ChampionDetailStatisticsRepository {
	cdsr := &ChampionDetailStatisticsRepository{
		Cache: nil,
	}
	_, _ = cdsr.Load()
	return cdsr
}

func (cdsr *ChampionDetailStatisticsRepository) key() string {
	return "champion_detail_statistics"
}

func (cdsr *ChampionDetailStatisticsRepository) Period() time.Duration {
	if core.DebugMode {
		return 1 * time.Hour
	}
	return 24 * time.Hour
}

func (cdsr *ChampionDetailStatisticsRepository) Loop() {
	// must be run in a goroutine
	for {
		if _, err := cdsr.Collect(); err != nil {
			log.Error(err)
		}
		time.Sleep(cdsr.Period())
	}
}

func (cdsr *ChampionDetailStatisticsRepository) Collect() (*ChampionDetailStatistics, error) {
	log.Debugf("collecting %s...", cdsr.key())
	timer := util.NewTimerWithName("ChampionDetailStatisticsRepository")
	timer.Start()

	// collect data
	championDetailStatisticMXDAOs, err := statistics.GetChampionDetailStatisticMXDAOs(StatisticsDB)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Debugf("championDetailStatisticMXDAOs: %d", len(championDetailStatisticMXDAOs))

	// TODO :: load by champion id, not in bulk (important!! -> memory issue > 3GB)
	championDetailStatisticsMXDAOmap := make(map[int]statistics.ChampionDetailStatisticMXDAO)
	for _, championDetailStatisticMXDAO := range championDetailStatisticMXDAOs {
		championDetailStatisticsMXDAOmap[championDetailStatisticMXDAO.ChampionId] = championDetailStatisticMXDAO
	}
	log.Debugf("championDetailStatisticsMXDAOmap: %d", len(championDetailStatisticsMXDAOmap))

	// TODO :: calculate asynchronously
	type MetaPick struct {
		Id             string
		PrimaryStyleId int
		SubStyleId     int
		Item0          int
		Item1          int
		Item2          int
		Wins           int
		Total          int
		WinRate        float64
		PickRate       float64
		MetaRank       int
		MajorTag       string
		MinorTag       *string
	}
	metaPickToRealMeta := func(metaPick MetaPick) ChampionDetailStatisticsMeta {
		return ChampionDetailStatisticsMeta{
			MetaName:         fmt.Sprintf("meta-%s-%d", metaPick.MajorTag, metaPick.MetaRank),
			MajorTag:         metaPick.MajorTag,
			MinorTag:         metaPick.MinorTag,
			MajorPerkStyleId: metaPick.PrimaryStyleId,
			MinorPerkStyleId: metaPick.SubStyleId,
			ItemTree:         []int{metaPick.Item0, metaPick.Item1, metaPick.Item2},
			Count:            metaPick.Total,
			Win:              metaPick.Wins,
			PickRate:         metaPick.PickRate,
		}
	}
	metaPicksToRealMetaTree := func(metaPicks []MetaPick, minorMeta *MetaPick) ChampionDetailStatisticsMetaTree {
		majorMetaPicks := make([]ChampionDetailStatisticsMeta, 0)
		for _, metaPick := range metaPicks {
			majorMetaPicks = append(majorMetaPicks, metaPickToRealMeta(metaPick))
		}
		var minorMetaPick *ChampionDetailStatisticsMeta
		if minorMeta != nil {
			p := metaPickToRealMeta(*minorMeta)
			minorMetaPick = &p
		}
		return ChampionDetailStatisticsMetaTree{
			MajorMetaPicks: majorMetaPicks,
			MinorMetaPick:  minorMetaPick,
		}
	}

	metaPicks := make(map[int]map[string][]MetaPick) // championId -> teamPosition
	championDetailStatisticsMetaMXDAOs, err := statistics.GetChampionDetailStatisticsMetaMXDAOs(StatisticsDB)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	for _, championDetailStatisticsMetaMXDAO := range championDetailStatisticsMetaMXDAOs {
		item0, exists := Items[championDetailStatisticsMetaMXDAO.Item0Id]
		if !exists {
			log.Errorf("item not found: %d", championDetailStatisticsMetaMXDAO.Item0Id)
			continue
		}
		item1, exists := Items[championDetailStatisticsMetaMXDAO.Item1Id]
		if !exists {
			log.Errorf("item not found: %d", championDetailStatisticsMetaMXDAO.Item1Id)
			continue
		}
		item2, exists := Items[championDetailStatisticsMetaMXDAO.Item2Id]
		if !exists {
			log.Errorf("item not found: %d", championDetailStatisticsMetaMXDAO.Item2Id)
			continue
		}

		// TODO :: exclude some tags (meaningless tags)
		tags := make(map[string]int)
		items := []ItemDataVO{item0, item1, item2}
		for _, item := range items {
			for _, tag := range item.Tags {
				if _, exists := tags[tag]; !exists {
					tags[tag] = 0
				}
				tags[tag]++
			}
		}

		type tagCount struct {
			tag   string
			count int
		}
		tagCounts := make([]tagCount, 0)
		for tag, count := range tags {
			tagCounts = append(tagCounts, tagCount{tag: tag, count: count})
		}

		// sort by count (desc)
		sort.SliceStable(tagCounts, func(i, j int) bool {
			if tagCounts[i].count == tagCounts[j].count {
				return tagCounts[i].tag < tagCounts[j].tag
			}
			return tagCounts[i].count > tagCounts[j].count
		})

		var majorTag string
		var minorTag *string
		if len(tagCounts) > 0 {
			majorTag = tagCounts[0].tag
			if len(tagCounts) > 1 {
				minorTag = &tagCounts[1].tag
			}
		} else {
			continue
		}

		if _, exists := metaPicks[championDetailStatisticsMetaMXDAO.ChampionId]; !exists {
			metaPicks[championDetailStatisticsMetaMXDAO.ChampionId] = make(map[string][]MetaPick)
		}
		if _, exists := metaPicks[championDetailStatisticsMetaMXDAO.ChampionId][championDetailStatisticsMetaMXDAO.TeamPosition]; !exists {
			metaPicks[championDetailStatisticsMetaMXDAO.ChampionId][championDetailStatisticsMetaMXDAO.TeamPosition] = make([]MetaPick, 0)
		}

		uuid := uuid2.New()
		metaPick := MetaPick{
			Id:             uuid.String(),
			PrimaryStyleId: championDetailStatisticsMetaMXDAO.PrimaryStyle,
			SubStyleId:     championDetailStatisticsMetaMXDAO.SubStyle,
			Item0:          championDetailStatisticsMetaMXDAO.Item0Id,
			Item1:          championDetailStatisticsMetaMXDAO.Item1Id,
			Item2:          championDetailStatisticsMetaMXDAO.Item2Id,
			Wins:           championDetailStatisticsMetaMXDAO.Wins,
			Total:          championDetailStatisticsMetaMXDAO.Total,
			WinRate:        championDetailStatisticsMetaMXDAO.WinRate,
			PickRate:       championDetailStatisticsMetaMXDAO.PickRate,
			MetaRank:       championDetailStatisticsMetaMXDAO.MetaRank,
			MajorTag:       majorTag,
			MinorTag:       minorTag,
		}

		metaPicks[championDetailStatisticsMetaMXDAO.ChampionId][championDetailStatisticsMetaMXDAO.TeamPosition] = append(
			metaPicks[championDetailStatisticsMetaMXDAO.ChampionId][championDetailStatisticsMetaMXDAO.TeamPosition],
			metaPick,
		)
	}

	type metaGroup []MetaPick
	type MetaCluster struct {
		Top3Meta  []MetaPick
		MinorMeta *MetaPick
	}

	majorMetaPicks := make(map[int]map[string]MetaCluster) // championId -> teamPosition
	for championId, teamPositionMap := range metaPicks {
		if _, exists := majorMetaPicks[championId]; !exists {
			majorMetaPicks[championId] = make(map[string]MetaCluster)
		}
		for teamPosition, metaPicks := range teamPositionMap {
			// group meta picks by concept (concept = major tag + minor tag)
			metaGroups := make(map[string]metaGroup)
			for _, metaPick := range metaPicks {
				concept := fmt.Sprintf("%s-%s", metaPick.MajorTag, *metaPick.MinorTag)
				if _, exists := metaGroups[concept]; !exists {
					metaGroups[concept] = make(metaGroup, 0)
				}
				metaGroups[concept] = append(metaGroups[concept], metaPick)
			}

			// pick best 3 meta group by average pick rate & best win rate meta group
			metaGroupPickRate := make(map[string]float64)
			metaGroupWinRate := make(map[string]float64)
			for concept, metaGroup := range metaGroups {
				totalPickRate := 0.0
				win := 0
				total := 0
				for _, metaPick := range metaGroup {
					totalPickRate += metaPick.PickRate
					win += metaPick.Wins
					total += metaPick.Total
				}
				if _, exists := metaGroupPickRate[concept]; !exists {
					metaGroupPickRate[concept] = 0
				}
				metaGroupPickRate[concept] += totalPickRate
				metaGroupWinRate[concept] = float64(win) / float64(total)
			}

			// pick best 3 meta group
			popularMetaConcepts := make([]string, 0)
			popularConcepts := make([]string, 0)
			popularMetas := make([]MetaPick, 0)
			for concept, _ := range metaGroups {
				popularMetaConcepts = append(popularMetaConcepts, concept)
			}
			sort.SliceStable(popularMetaConcepts, func(i, j int) bool {
				pickRateI, pickRateJ := metaGroupPickRate[popularMetaConcepts[i]], metaGroupPickRate[popularMetaConcepts[j]]
				winRateI, winRateJ := metaGroupWinRate[popularMetaConcepts[i]], metaGroupWinRate[popularMetaConcepts[j]]
				if pickRateI != pickRateJ {
					return pickRateI > pickRateJ
				}
				return winRateI > winRateJ
			})
			for ind, concept := range popularMetaConcepts {
				if ind < 3 {
					popularConcepts = append(popularConcepts, concept)
				}
			}
			for _, concept := range popularConcepts {
				mostMetaPicks := metaGroups[concept]
				// pick the highest pick rate & win rate meta pick from the concept group
				var mostWinRateMeta *MetaPick
				maxWinRate := 0.0
				maxPickRate := 0.0
				for _, metaPick := range mostMetaPicks {
					if metaPick.PickRate > maxPickRate {
						maxPickRate = metaPick.PickRate
						mostWinRateMeta = &metaPick
					} else if metaPick.PickRate == maxPickRate && metaPick.WinRate > maxWinRate {
						maxWinRate = metaPick.WinRate
						mostWinRateMeta = &metaPick
					}
				}
				if mostWinRateMeta != nil {
					popularMetas = append(popularMetas, *mostWinRateMeta)
				}
			}

			// pick best win rate meta group except mostMetaPicks
			var mostWinRateConcept *string
			maxWinRate := 0.0
			maxPickRate := 0.0
			for concept, _ := range metaGroups {
				conceptUsed := false
				for _, popularMeta := range popularMetaConcepts {
					if concept == popularMeta {
						conceptUsed = true
						break
					}
				}
				if conceptUsed {
					continue
				}
				winRate := metaGroupWinRate[concept]
				pickRate := metaGroupPickRate[concept]
				if pickRate > 0.05 {
					continue
				}
				if mostWinRateConcept == nil || pickRate > maxPickRate || (pickRate == maxPickRate && winRate > maxWinRate) {
					mostWinRateConcept = &concept
					maxWinRate = winRate
					maxPickRate = pickRate
				}
			}

			var mostWinRateMeta *MetaPick
			if mostWinRateConcept != nil {
				mostMetaPicks, exists := metaGroups[*mostWinRateConcept]
				if !exists {
					continue
				}

				// pick the highest pick rate & win rate meta pick from the concept group
				maxWinRate = 0.0
				maxPickRate = 0.0
				for _, metaPick := range mostMetaPicks {
					if metaPick.PickRate > maxPickRate {
						maxPickRate = metaPick.PickRate
						mostWinRateMeta = &metaPick
					} else if metaPick.PickRate == maxPickRate && metaPick.WinRate > maxWinRate {
						maxWinRate = metaPick.WinRate
						mostWinRateMeta = &metaPick
					}
				}
			}

			majorMetaPicks[championId][teamPosition] = MetaCluster{
				Top3Meta:  popularMetas,
				MinorMeta: mostWinRateMeta,
			}
		}
	}

	stats := make(map[int]ChampionDetailStatisticsItem)
	for key, champion := range Champions {
		championId, err := strconv.Atoi(key)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		e := ChampionDetailStatisticsItem{
			ChampionId:   championId,
			ChampionName: champion.Name,
			Win:          0,
			Total:        0,
			AvgPickRate:  0,
			AvgBanRate:   0,
			ExtraStats: ChampionDetailStatisticsExtraStats{
				AvgMinionsKilled: 0,
				AvgKills:         0,
				AvgDeaths:        0,
				AvgAssists:       0,
				AvgGoldEarned:    0,
			},
			NormalizedStats: ChampionDetailStatisticsNormalizedStats{
				AvgTotalHealPerSec:           0,
				AvgVisionScorePerSec:         0,
				AvgTotalDamageTakenPerSec:    0,
				AvgTotalTimeCCDealtPerSec:    0,
				AvgDamageSelfMitigatedPerSec: 0,
			},
			MetaTree: ChampionDetailStatisticsPositionMetaTree{
				Top: ChampionDetailStatisticsMetaTree{
					MajorMetaPicks: make([]ChampionDetailStatisticsMeta, 0),
					MinorMetaPick:  nil,
				},
				Jungle: ChampionDetailStatisticsMetaTree{
					MajorMetaPicks: make([]ChampionDetailStatisticsMeta, 0),
					MinorMetaPick:  nil,
				},
				Mid: ChampionDetailStatisticsMetaTree{
					MajorMetaPicks: make([]ChampionDetailStatisticsMeta, 0),
					MinorMetaPick:  nil,
				},
				ADC: ChampionDetailStatisticsMetaTree{
					MajorMetaPicks: make([]ChampionDetailStatisticsMeta, 0),
					MinorMetaPick:  nil,
				},
				Support: ChampionDetailStatisticsMetaTree{
					MajorMetaPicks: make([]ChampionDetailStatisticsMeta, 0),
					MinorMetaPick:  nil,
				},
			},
		}

		championDetailStatisticMXDAO, exists := championDetailStatisticsMXDAOmap[championId]
		topMetaTree, l1Exists := majorMetaPicks[championId][types.TeamPositionTop]
		jungleMetaTree, l2Exists := majorMetaPicks[championId][types.TeamPositionJungle]
		midMetaTree, l3Exists := majorMetaPicks[championId][types.TeamPositionMid]
		adcMetaTree, l4Exists := majorMetaPicks[championId][types.TeamPositionAdc]
		supportMetaTree, l5Exists := majorMetaPicks[championId][types.TeamPositionSupport]

		if exists {
			e.Win = championDetailStatisticMXDAO.Win
			e.Total = championDetailStatisticMXDAO.Total
			e.AvgPickRate = championDetailStatisticMXDAO.PickRate
			e.AvgBanRate = championDetailStatisticMXDAO.BanRate
			e.ExtraStats = ChampionDetailStatisticsExtraStats{
				AvgMinionsKilled: championDetailStatisticMXDAO.AvgMinionsKilled,
				AvgKills:         championDetailStatisticMXDAO.AvgKills,
				AvgDeaths:        championDetailStatisticMXDAO.AvgDeaths,
				AvgAssists:       championDetailStatisticMXDAO.AvgAssists,
				AvgGoldEarned:    championDetailStatisticMXDAO.AvgGoldEarned,
			}
			e.NormalizedStats = ChampionDetailStatisticsNormalizedStats{
				AvgTotalHealPerSec:           championDetailStatisticMXDAO.AvgHealPerSec,
				AvgVisionScorePerSec:         championDetailStatisticMXDAO.AvgVisionScorePerSec,
				AvgTotalDamageTakenPerSec:    championDetailStatisticMXDAO.AvgDamageTakenPerSec,
				AvgTotalTimeCCDealtPerSec:    championDetailStatisticMXDAO.AvgTimeCCDealtPerSec,
				AvgDamageSelfMitigatedPerSec: championDetailStatisticMXDAO.AvgDamageSelfMitigatedPerSec,
			}
		}

		if l1Exists {
			e.MetaTree.Top = metaPicksToRealMetaTree(topMetaTree.Top3Meta, topMetaTree.MinorMeta)
		}
		if l2Exists {
			e.MetaTree.Jungle = metaPicksToRealMetaTree(jungleMetaTree.Top3Meta, jungleMetaTree.MinorMeta)
		}
		if l3Exists {
			e.MetaTree.Mid = metaPicksToRealMetaTree(midMetaTree.Top3Meta, midMetaTree.MinorMeta)
		}
		if l4Exists {
			e.MetaTree.ADC = metaPicksToRealMetaTree(adcMetaTree.Top3Meta, adcMetaTree.MinorMeta)
		}
		if l5Exists {
			e.MetaTree.Support = metaPicksToRealMetaTree(supportMetaTree.Top3Meta, supportMetaTree.MinorMeta)
		}

		stats[championId] = e
	}

	cdsr.Cache = &ChampionDetailStatistics{
		UpdatedAt: time.Now(),
		Data:      stats,
	}

	log.Debugf("%s data collected successfully in %s", cdsr.key(), timer.GetDurationString())
	if err := cdsr.Save(); err != nil {
		log.Warn(err)
	}

	return cdsr.Cache, nil
}

func (cdsr *ChampionDetailStatisticsRepository) Save() error {
	if cdsr.Cache == nil {
		log.Error("data is nil")
		return nil
	}

	// save data
	jsonData, err := json.Marshal(cdsr.Cache)
	if err != nil {
		log.Error(err)
		return err
	}

	// create directory if not exists
	if err = os.MkdirAll(path.Join(util.GetProjectRootDirectory(), StatisticsDataPath), 0755); err != nil {
		log.Error(err)
		return err
	}

	filePath := keyPath(cdsr.key())
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debugf("%s data saved to %s successfully", cdsr.key(), filePath)
	return nil
}

func (cdsr *ChampionDetailStatisticsRepository) Load() (*ChampionDetailStatistics, error) {
	if cdsr.Cache != nil {
		return cdsr.Cache, nil
	}

	// if there is no data, collect and save
	filePath := keyPath(cdsr.key())
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Debugf("file not found: %s", filePath)
			return cdsr.Collect()
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
	err = json.Unmarshal(jsonData, &cdsr.Cache)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return cdsr.Cache, nil
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
	timer := util.NewTimerWithName("TierStatisticsRepository")
	timer.Start()

	// collect data
	tierCountMXDAOs, err := statistics.GetTierStatisticsTierCountMXDAOs(StatisticsDB)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	topRankersMXDAOs, err := statistics.GetTierStatisticsTopRankersMXDAOs(StatisticsDB, 30)
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
		if os.IsNotExist(err) {
			log.Debugf("file not found: %s", filePath)
			return tsr.Collect()
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
	err = json.Unmarshal(jsonData, &tsr.Cache)
	if err != nil {
		log.Error(err)
		return nil, err
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
	msr := &MasteryStatisticsRepository{
		Cache: nil,
	}
	_, _ = msr.Load()
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
	timer := util.NewTimerWithName("MasteryStatisticsRepository")
	timer.Start()

	// collect data
	masteryMXDAOs, err := statistics.GetMasteryStatisticsMXDAOs(StatisticsDB)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	masteryTopRankersMXDAO, err := statistics.GetMasteryStatisticsTopRankersMXDAOs(StatisticsDB, 30)
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
