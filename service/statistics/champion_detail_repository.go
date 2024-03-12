package statistics

import (
	"encoding/json"
	"fmt"
	uuid2 "github.com/google/uuid"
	log "github.com/shyunku-libraries/go-logger"
	"os"
	"path"
	"sort"
	"strconv"
	"team.gg-server/core"
	"team.gg-server/models/mixed/statistics_models"
	"team.gg-server/service"
	"team.gg-server/types"
	"team.gg-server/util"
	"time"
)

/* ----------------------- Champion Detail statistics_models ----------------------- */

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

func (m *MetaPick) toRealMeta() ChampionDetailStatisticsMeta {
	return ChampionDetailStatisticsMeta{
		MetaName:         fmt.Sprintf("meta-%s-%d", m.MajorTag, m.MetaRank),
		MajorTag:         m.MajorTag,
		MinorTag:         m.MinorTag,
		MajorPerkStyleId: m.PrimaryStyleId,
		MinorPerkStyleId: m.SubStyleId,
		ItemTree:         []int{m.Item0, m.Item1, m.Item2},
		Count:            m.Total,
		Win:              m.Wins,
		PickRate:         m.PickRate,
		WinRate:          m.WinRate,
	}
}

type MetaGroup []MetaPick

func (mg *MetaGroup) getTotalWinRate() float64 {
	totalWins := 0
	totalTotal := 0
	for _, metaPick := range *mg {
		totalWins += metaPick.Wins
		totalTotal += metaPick.Total
	}
	return float64(totalWins) / float64(totalTotal)
}

func (mg *MetaGroup) getTotalPickRate() float64 {
	totalPickRate := 0.0
	for _, metaPick := range *mg {
		totalPickRate += metaPick.PickRate
	}
	return totalPickRate
}

type ChampionDetailStatisticsMeta struct {
	MetaName string `json:"metaName"`

	MajorTag string  `json:"majorTag"`
	MinorTag *string `json:"minorTag"`

	MajorPerkStyleId int   `json:"majorPerkStyleId"`
	MinorPerkStyleId int   `json:"minorPerkStyleId"`
	ItemTree         []int `json:"itemTree"` // must be sorted as descending order with gold

	Count int `json:"count"`
	Win   int `json:"win"`

	WinRate  float64 `json:"winRate"`
	PickRate float64 `json:"pickRate"`
}

type ChampionDetailStatisticsMetaTree struct {
	MajorMetaPicks []ChampionDetailStatisticsMeta `json:"majorMetaPicks"`
	MinorMetaPicks []ChampionDetailStatisticsMeta `json:"minorMetaPick"`
	MetaPicks      []ChampionDetailStatisticsMeta `json:"metaPicks"`
}

type ChampionDetailStatisticsPositionMetaTree struct {
	Top     *ChampionDetailStatisticsMetaTree `json:"top"`
	Jungle  *ChampionDetailStatisticsMetaTree `json:"jungle"`
	Mid     *ChampionDetailStatisticsMetaTree `json:"mid"`
	ADC     *ChampionDetailStatisticsMetaTree `json:"adc"`
	Support *ChampionDetailStatisticsMetaTree `json:"support"`
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
	ChampionId    int    `json:"championId"`
	ChampionName  string `json:"championName"`
	ChampionTitle string `json:"championTitle"`
	ChampionStory string `json:"championStory"`

	Win         int     `json:"win"`
	Total       int     `json:"total"`
	AvgPickRate float64 `json:"avgPickRate"`
	AvgBanRate  float64 `json:"avgBanRate"`
	AvgWinRate  float64 `json:"avgWinRate"`

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
	championDetailStatisticMXDAOs, err := statistics_models.GetChampionDetailStatisticMXDAOs(StatisticsDB)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	championDetailStatisticsMXDAOmap := make(map[int]statistics_models.ChampionDetailStatisticMXDAO)
	for _, championDetailStatisticMXDAO := range championDetailStatisticMXDAOs {
		championDetailStatisticsMXDAOmap[championDetailStatisticMXDAO.ChampionId] = championDetailStatisticMXDAO
	}
	log.Debugf("championDetailStatisticMXDAOs: %d", len(championDetailStatisticMXDAOs))

	// collect meta
	championDetailStatisticsMetaMXDAOs, err := statistics_models.GetChampionDetailStatisticsMetaMXDAOs(StatisticsDB)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	championDetailStatisticsMetaMap := make(map[int][]statistics_models.ChampionDetailStatisticsMetaMXDAO)
	for _, meta := range championDetailStatisticsMetaMXDAOs {
		championId := meta.ChampionId
		if _, exists := championDetailStatisticsMetaMap[championId]; !exists {
			championDetailStatisticsMetaMap[championId] = make([]statistics_models.ChampionDetailStatisticsMetaMXDAO, 0)
		}
		championDetailStatisticsMetaMap[championId] = append(championDetailStatisticsMetaMap[championId], meta)
	}
	log.Debugf("championDetailStatisticsMetaMXDAOs: %d, size: %s",
		len(championDetailStatisticsMetaMXDAOs), util.MemorySizeOfArray(championDetailStatisticsMetaMXDAOs))

	stats := make(map[int]ChampionDetailStatisticsItem)
	for key, champion := range service.Champions {
		championId, err := strconv.Atoi(key)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		metas, ok := championDetailStatisticsMetaMap[championId]
		if !ok {
			log.Warnf("championId not found: %d", championId)
			continue
		}

		metaTree, err := cdsr.collectEachChampionMetas(metas)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		//// get champion detail meta statistics
		//metaTree, err := cdsr.collectEachChampionMetas(championId)
		//if err != nil {
		//	log.Error(err)
		//	return nil, err
		//}

		e := ChampionDetailStatisticsItem{
			ChampionId:    championId,
			ChampionName:  champion.Name,
			ChampionTitle: champion.Title,
			ChampionStory: champion.Blurb,
			Win:           0,
			Total:         0,
			AvgPickRate:   0,
			AvgBanRate:    0,
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
			MetaTree: *metaTree,
		}

		if championDetailStatisticMXDAO, exists := championDetailStatisticsMXDAOmap[championId]; exists {
			e.Win = championDetailStatisticMXDAO.Win
			e.Total = championDetailStatisticMXDAO.Total
			e.AvgPickRate = championDetailStatisticMXDAO.PickRate
			e.AvgBanRate = championDetailStatisticMXDAO.BanRate
			e.AvgWinRate = float64(championDetailStatisticMXDAO.Win) / float64(championDetailStatisticMXDAO.Total)
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

func (cdsr *ChampionDetailStatisticsRepository) collectEachChampionMetas(championDetailStatisticsMetaMXDAOs []statistics_models.ChampionDetailStatisticsMetaMXDAO) (*ChampionDetailStatisticsPositionMetaTree, error) {
	//championDetailStatisticsMetaMXDAOs, err := statistics_models.GetChampionDetailStatisticsMetaMXDAOs_byChampionId(StatisticsDB, championId)
	//if err != nil {
	//	log.Error(err)
	//	return nil, err
	//}

	metaTrees := &ChampionDetailStatisticsPositionMetaTree{
		Top:     nil,
		Jungle:  nil,
		Mid:     nil,
		ADC:     nil,
		Support: nil,
	}

	metaMXDAOsByPosition := map[string][]statistics_models.ChampionDetailStatisticsMetaMXDAO{
		types.TeamPositionTop:     make([]statistics_models.ChampionDetailStatisticsMetaMXDAO, 0),
		types.TeamPositionJungle:  make([]statistics_models.ChampionDetailStatisticsMetaMXDAO, 0),
		types.TeamPositionMid:     make([]statistics_models.ChampionDetailStatisticsMetaMXDAO, 0),
		types.TeamPositionAdc:     make([]statistics_models.ChampionDetailStatisticsMetaMXDAO, 0),
		types.TeamPositionSupport: make([]statistics_models.ChampionDetailStatisticsMetaMXDAO, 0),
	}
	for _, championDetailStatisticsMetaMXDAO := range championDetailStatisticsMetaMXDAOs {
		teamPosition := championDetailStatisticsMetaMXDAO.TeamPosition
		if _, exists := metaMXDAOsByPosition[teamPosition]; !exists {
			log.Warnf("team position not exists: %s", teamPosition)
			continue
		}
		metaMXDAOsByPosition[teamPosition] = append(metaMXDAOsByPosition[teamPosition], championDetailStatisticsMetaMXDAO)
	}

	for teamPosition, metaMXDAOs := range metaMXDAOsByPosition {
		metaGroup := make(MetaGroup, 0)
		for _, metaMXDAO := range metaMXDAOs {
			majorItems := make([]service.ItemDataVO, 0)
			for _, itemId := range []int{metaMXDAO.Item0Id, metaMXDAO.Item1Id, metaMXDAO.Item2Id} {
				item, exists := service.Items[itemId]
				if !exists {
					log.Errorf("item not found: %d", itemId)
					continue
				}
				majorItems = append(majorItems, item)
			}

			// get categories of tags from major items
			tagCategories := make(map[string]int)
			for _, item := range majorItems {
				for _, tag := range item.Tags {
					category := types.GetItemCategories(tag)
					if category != nil {
						if _, exists := tagCategories[*category]; !exists {
							tagCategories[*category] = 0
						}
						tagCategories[*category]++
					}
				}
			}

			// collect counts of tag categories
			type CategoryCount struct {
				category string
				count    int
			}
			categoryCounts := make([]CategoryCount, 0)
			for category, count := range tagCategories {
				categoryCounts = append(categoryCounts, CategoryCount{category: category, count: count})
			}

			// sort categories by count (desc)
			sort.SliceStable(categoryCounts, func(i, j int) bool {
				if categoryCounts[i].count == categoryCounts[j].count {
					return categoryCounts[i].category < categoryCounts[j].category
				}
				return categoryCounts[i].count > categoryCounts[j].count
			})

			// get major tag and minor tag
			var majorTag string
			var minorTag *string
			if len(categoryCounts) > 0 {
				majorTag = categoryCounts[0].category
				if len(categoryCounts) > 1 {
					minorTag = &categoryCounts[1].category
				}
			} else {
				continue
			}

			uuid := uuid2.New()
			metaPick := MetaPick{
				Id:             uuid.String(),
				PrimaryStyleId: metaMXDAO.PrimaryStyle,
				SubStyleId:     metaMXDAO.SubStyle,
				Item0:          metaMXDAO.Item0Id,
				Item1:          metaMXDAO.Item1Id,
				Item2:          metaMXDAO.Item2Id,
				Wins:           metaMXDAO.Wins,
				Total:          metaMXDAO.Total,
				WinRate:        metaMXDAO.WinRate,
				PickRate:       metaMXDAO.PickRate,
				MetaRank:       metaMXDAO.MetaRank,
				MajorTag:       majorTag,
				MinorTag:       minorTag,
			}

			metaGroup = append(metaGroup, metaPick)
		}

		// categorize meta picks by concept (concept = major tag + minor tag)
		conceptGroups := make(map[string]MetaGroup) // concept -> MetaGroup
		for _, metaPick := range metaGroup {
			concept := metaPick.MajorTag
			if metaPick.MinorTag != nil {
				concept += "-" + *metaPick.MinorTag
			}
			if _, exists := conceptGroups[concept]; !exists {
				conceptGroups[concept] = make(MetaGroup, 0)
			}
			conceptGroups[concept] = append(conceptGroups[concept], metaPick)
		}

		// pick popular 3 concept group
		concepts := make([]string, 0)
		for concept, _ := range conceptGroups {
			concepts = append(concepts, concept)
		}
		// sort concept groups (pickRate desc, winRate desc)
		sort.SliceStable(concepts, func(i, j int) bool {
			conceptI, conceptJ := concepts[i], concepts[j]
			groupI, groupJ := conceptGroups[conceptI], conceptGroups[conceptJ]
			pickRateI, pickRateJ := groupI.getTotalPickRate(), groupJ.getTotalPickRate()
			winRateI, winRateJ := groupI.getTotalWinRate(), groupJ.getTotalWinRate()
			if pickRateI != pickRateJ {
				return pickRateI > pickRateJ
			}
			return winRateI > winRateJ
		})

		popularConcepts := make([]string, 0)
		nonPopularConcepts := make([]string, 0)
		for ind, concept := range concepts {
			if ind < 5 {
				popularConcepts = append(popularConcepts, concept)
			} else {
				nonPopularConcepts = append(nonPopularConcepts, concept)
			}
		}

		var minorConcept *string // minor concept has low pick rate (with lower limit) but high win rate
		// sort non-popular concept groups (winRate desc, pickRate desc)
		sort.SliceStable(nonPopularConcepts, func(i, j int) bool {
			conceptI, conceptJ := concepts[i], concepts[j]
			groupI, groupJ := conceptGroups[conceptI], conceptGroups[conceptJ]
			pickRateI, pickRateJ := groupI.getTotalPickRate(), groupJ.getTotalPickRate()
			winRateI, winRateJ := groupI.getTotalWinRate(), groupJ.getTotalWinRate()
			if winRateI != winRateJ {
				return winRateI > winRateJ
			}
			return pickRateI > pickRateJ
		})
		if len(nonPopularConcepts) > 0 {
			minorConcept = &nonPopularConcepts[0]
		}

		pickMostPickRateMetas := func(metaGroup MetaGroup, count int) []MetaPick {
			metas := make([]MetaPick, 0)
			// sort meta picks (pickRate desc, winRate desc)
			sort.SliceStable(metaGroup, func(i, j int) bool {
				metaI, metaJ := metaGroup[i], metaGroup[j]
				if metaI.PickRate != metaJ.PickRate {
					return metaI.PickRate > metaJ.PickRate
				}
				return metaI.WinRate > metaJ.WinRate
			})
			// pick top {count} meta picks
			for ind, meta := range metaGroup {
				if ind < count {
					metas = append(metas, meta)
				} else {
					break
				}
			}
			return metas
		}

		majorMetaPicks := make([]MetaPick, 0)
		for _, concept := range popularConcepts {
			popularMetaGroup := conceptGroups[concept]
			popularMetas := pickMostPickRateMetas(popularMetaGroup, 3)
			if len(popularMetas) > 0 {
				majorMetaPicks = append(majorMetaPicks, popularMetas...)
			}
		}

		minorMetaPicks := make([]MetaPick, 0)
		if minorConcept != nil {
			minorMetaGroup := conceptGroups[*minorConcept]
			minorMetas := pickMostPickRateMetas(minorMetaGroup, 3)
			if len(minorMetas) > 0 {
				minorMetaPicks = append(minorMetaPicks, minorMetas...)
			}
		}

		metaTree := ChampionDetailStatisticsMetaTree{
			MajorMetaPicks: make([]ChampionDetailStatisticsMeta, 0),
			MinorMetaPicks: make([]ChampionDetailStatisticsMeta, 0),
			MetaPicks:      make([]ChampionDetailStatisticsMeta, 0),
		}
		for _, metaPick := range majorMetaPicks {
			metaTree.MajorMetaPicks = append(metaTree.MajorMetaPicks, metaPick.toRealMeta())
		}
		for _, metaPick := range minorMetaPicks {
			metaTree.MinorMetaPicks = append(metaTree.MinorMetaPicks, metaPick.toRealMeta())
		}
		for _, metaPick := range metaGroup {
			metaTree.MetaPicks = append(metaTree.MetaPicks, metaPick.toRealMeta())
		}

		if teamPosition == types.TeamPositionTop {
			metaTrees.Top = &metaTree
		} else if teamPosition == types.TeamPositionJungle {
			metaTrees.Jungle = &metaTree
		} else if teamPosition == types.TeamPositionMid {
			metaTrees.Mid = &metaTree
		} else if teamPosition == types.TeamPositionAdc {
			metaTrees.ADC = &metaTree
		} else if teamPosition == types.TeamPositionSupport {
			metaTrees.Support = &metaTree
		} else {
			log.Warnf("teamPosition not matched: %s", teamPosition)
		}
	}

	return metaTrees, nil
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
