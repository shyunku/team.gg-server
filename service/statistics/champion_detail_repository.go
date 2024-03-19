package statistics

import (
	"encoding/json"
	uuid2 "github.com/google/uuid"
	log "github.com/shyunku-libraries/go-logger"
	"os"
	"path"
	"sort"
	"strconv"
	"team.gg-server/core"
	"team.gg-server/libs/db"
	"team.gg-server/models/mixed"
	"team.gg-server/models/mixed/statistics_models"
	"team.gg-server/service"
	"team.gg-server/types"
	"team.gg-server/util"
	"time"
)

/* ----------------------- Champion Detail statistics_models ----------------------- */

type ChampionPositionStatistics struct {
	PickCount int `json:"pickCount"`
	WinCount  int `json:"winCount"`
}

type PerkSlot struct {
	Type      string `json:"type"`
	SlotLabel string `json:"slotLabel"`
	Perks     []int  `json:"perks"`
}

type PerkGroup struct {
	PerkStyleName string `json:"perkStyleName"`
	PerkStyleId   int    `json:"perkStyleId"`
	SubPerks      []int  `json:"subPerks"`
}

type PerkExtra struct {
	StatDefenseId int `json:"statDefenseId"`
	StatFlexId    int `json:"statFlexId"`
	StatOffenseId int `json:"statOffenseId"`
}

type ChampionDetailStatisticsMeta struct {
	MetaKey  string  `json:"metaKey"`
	MajorTag string  `json:"majorTag"`
	MinorTag *string `json:"minorTag"`

	Summoner1Id int `json:"summoner1Id"`
	Summoner2Id int `json:"summoner2Id"`

	MajorPerkGroup PerkGroup `json:"majorPerkGroup"` // 메인 룬 추천
	MinorPerkGroup PerkGroup `json:"minorPerkGroup"` // Sub 룬 추천
	PerkExtra      PerkExtra `json:"perkExtra"`      // 메인 룬 스탯 추천

	MainSlots []PerkSlot `json:"mainSlots"` // 메인 룬 Placeholders
	SubSlots  []PerkSlot `json:"subSlots"`  // Sub 룬 Placeholders
	StatSlots []PerkSlot `json:"statSlots"` // 메인 룬 스탯 Placeholders

	StartItemTree []int `json:"startItemTree"` // 시작 아이템 추천
	BasicItemTree []int `json:"basicItemTree"` // 기본 아이템 추천
	ItemTree      []int `json:"itemTree"`      // 메인 아이템 추천
	SubItemTree   []int `json:"subItemTree"`   // 부가 아이템 추천

	Count int `json:"count"` // 메타 픽 수
	Win   int `json:"win"`   // 해당 메타 승리 수

	WinRate  float64 `json:"winRate"`
	PickRate float64 `json:"pickRate"`
}

type ChampionCounterStatistics struct {
	CounterChampionId   int    `json:"counterChampionId"`
	CounterChampionName string `json:"counterChampionName"`

	AvgKills   float64 `json:"avgKills"`
	AvgDeaths  float64 `json:"avgDeaths"`
	AvgAssists float64 `json:"avgAssists"`

	CounterAvgKills   *float64 `json:"counterAvgKills"`
	CounterAvgDeaths  *float64 `json:"counterAvgDeaths"`
	CounterAvgAssists *float64 `json:"counterAvgAssists"`

	Summoner1Id int `json:"summoner1Id"`
	Summoner2Id int `json:"summoner2Id"`

	MajorPerkGroup PerkGroup `json:"majorPerkGroup"` // 메인 룬 추천
	MinorPerkGroup PerkGroup `json:"minorPerkGroup"` // Sub 룬 추천
	PerkExtra      PerkExtra `json:"perkExtra"`      // 메인 룬 스탯 추천

	MainSlots []PerkSlot `json:"mainSlots"` // 메인 룬 Placeholders
	SubSlots  []PerkSlot `json:"subSlots"`  // Sub 룬 Placeholders
	StatSlots []PerkSlot `json:"statSlots"` // 메인 룬 스탯 Placeholders

	StartItemTree []int `json:"startItemTree"` // 시작 아이템 추천
	BasicItemTree []int `json:"basicItemTree"` // 기본 아이템 추천
	ItemTree      []int `json:"itemTree"`      // 메인 아이템 추천
	SubItemTree   []int `json:"subItemTree"`   // 부가 아이템 추천

	Count int `json:"count"` // 메타 픽 수
	Win   int `json:"win"`   // 해당 메타 승리 수

	WinRate         float64  `json:"winRate"`
	ExpectedWinRate *float64 `json:"expectedWinRate"` // 해당 룬/아이템 조합의 기대 승률
}

type ChampionDetailStatisticsMetaTree struct {
	MajorMetaPicks []ChampionDetailStatisticsMeta    `json:"majorMetaPicks"`
	MinorMetaPicks []ChampionDetailStatisticsMeta    `json:"minorMetaPick"`
	MetaPicks      []ChampionDetailStatisticsMeta    `json:"metaPicks"`
	PickCount      int                               `json:"pickCount"`
	WinCount       int                               `json:"winCount"`
	CounterMap     map[int]ChampionCounterStatistics `json:"counterMap"`
}

type ChampionDetailStatisticsPositionMetaTree struct {
	Top     *ChampionDetailStatisticsMetaTree `json:"top"`
	Jungle  *ChampionDetailStatisticsMetaTree `json:"jungle"`
	Mid     *ChampionDetailStatisticsMetaTree `json:"mid"`
	Adc     *ChampionDetailStatisticsMetaTree `json:"adc"`
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

	ExtraStats      ChampionDetailStatisticsExtraStats       `json:"extraStats"`
	NormalizedStats ChampionDetailStatisticsNormalizedStats  `json:"normalizedStats"`
	MetaTree        ChampionDetailStatisticsPositionMetaTree `json:"metaTree"`
}

type ChampionDetailStatistics struct {
	UpdatedAt time.Time                            `json:"updatedAt"`
	Patches   []string                             `json:"patches"`
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

	// collect recent versions
	recentMatchGameVersions, recentMatchGameShortVersions, err := mixed.GetRecentMatchGameVersions_byDescendingShortVersion_withCount(StatisticsDB, 2)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Debugf("recentMatchGameVersions: %v", recentMatchGameVersions)

	// collect data
	championDetailStatisticsMXDAOmap := make(map[int]statistics_models.ChampionDetailStatisticMXDAO)
	championDetailStatisticMXDAOs, err := statistics_models.GetChampionDetailStatisticMXDAOs(StatisticsDB, recentMatchGameVersions)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	for _, championDetailStatisticMXDAO := range championDetailStatisticMXDAOs {
		championDetailStatisticsMXDAOmap[championDetailStatisticMXDAO.ChampionId] = championDetailStatisticMXDAO
	}
	log.Debugf("championDetailStatisticMXDAOs fetch complete: %d, size: %s",
		len(championDetailStatisticMXDAOs), util.MemorySizeOfArray(championDetailStatisticMXDAOs))

	// collect champion pick count by team position
	championPositionStatisticsMXDAOmap := make(map[int]map[string]ChampionPositionStatistics)
	championPositionStatisticsMXDAOs, err := statistics_models.GetChampionPositionStatisticsMXDAOs(StatisticsDB, recentMatchGameVersions)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	for _, championPositionStatisticsMXDAO := range championPositionStatisticsMXDAOs {
		championId := championPositionStatisticsMXDAO.ChampionId
		if _, exists := championPositionStatisticsMXDAOmap[championId]; !exists {
			championPositionStatisticsMXDAOmap[championId] = make(map[string]ChampionPositionStatistics)
		}
		teamPosition := championPositionStatisticsMXDAO.TeamPosition
		if teamPosition != types.TeamPositionTop &&
			teamPosition != types.TeamPositionJungle &&
			teamPosition != types.TeamPositionMid &&
			teamPosition != types.TeamPositionAdc &&
			teamPosition != types.TeamPositionSupport {
			log.Warnf("team position not matched: %s", teamPosition)
			continue
		}
		championPositionStatisticsMXDAOmap[championId][teamPosition] = ChampionPositionStatistics{
			PickCount: championPositionStatisticsMXDAO.Total,
			WinCount:  championPositionStatisticsMXDAO.Win,
		}
	}

	if err := statistics_models.CreateTemporaryTables(StatisticsDB, recentMatchGameVersions); err != nil {
		log.Error(err)
		return nil, err
	}
	defer func(db db.Context) {
		err := statistics_models.DropTemporaryTables(db)
		if err != nil {
			log.Error(err)
		}
	}(StatisticsDB)

	// collect meta
	championDetailStatisticsMetaMap := make(map[int][]statistics_models.ChampionDetailStatisticsMetaMXDAO)
	championDetailStatisticsMetaMXDAOs, err := statistics_models.GetChampionDetailStatisticsMetaMXDAOs(StatisticsDB)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	for _, meta := range championDetailStatisticsMetaMXDAOs {
		championId := meta.ChampionId
		if _, exists := championDetailStatisticsMetaMap[championId]; !exists {
			championDetailStatisticsMetaMap[championId] = make([]statistics_models.ChampionDetailStatisticsMetaMXDAO, 0)
		}
		championDetailStatisticsMetaMap[championId] = append(championDetailStatisticsMetaMap[championId], meta)
	}
	log.Debugf("championDetailStatisticsMetaMXDAOs fetch complete: %d, size: %s",
		len(championDetailStatisticsMetaMXDAOs), util.MemorySizeOfArray(championDetailStatisticsMetaMXDAOs))

	// collect counter data
	championCounterStatisticsMap := make(map[int][]statistics_models.ChampionCounterStatisticsMXDAO) // key: championId
	championCounterStatisticsMXDAOs, err := statistics_models.GetChampionCounterStatisticsMXDAOs(StatisticsDB)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	for _, counter := range championCounterStatisticsMXDAOs {
		championId := counter.ChampionId
		if _, exists := championCounterStatisticsMap[championId]; !exists {
			championCounterStatisticsMap[championId] = make([]statistics_models.ChampionCounterStatisticsMXDAO, 0)
		}
		championCounterStatisticsMap[championId] = append(championCounterStatisticsMap[championId], counter)
	}
	log.Debugf("championCounterStatisticsMXDAOs fetch complete: %d, size: %s",
		len(championCounterStatisticsMXDAOs), util.MemorySizeOfArray(championCounterStatisticsMXDAOs))

	stats := make(map[int]ChampionDetailStatisticsItem)
	for key, champion := range service.Champions {
		championId, err := strconv.Atoi(key)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		championPositionStatisticsMXDAO, exists := championPositionStatisticsMXDAOmap[championId]
		if !exists {
			log.Warnf("championId not found: %d", championId)
			continue
		}

		metas, ok := championDetailStatisticsMetaMap[championId]
		if !ok {
			log.Warnf("championId not found: %d", championId)
			continue
		}

		counters, exists := championCounterStatisticsMap[championId]
		if !exists {
			log.Warnf("championId not found: %d", championId)
			counters = make([]statistics_models.ChampionCounterStatisticsMXDAO, 0)
		}

		metaTree, err := cdsr.collectEachChampionMetas(championId, championPositionStatisticsMXDAO, metas, counters)
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
		Patches:   recentMatchGameShortVersions,
		Data:      stats,
	}

	log.Debugf("%s data collected successfully in %s", cdsr.key(), timer.GetDurationString())
	if err := cdsr.Save(); err != nil {
		log.Warn(err)
	}

	return cdsr.Cache, nil
}

func (cdsr *ChampionDetailStatisticsRepository) collectEachChampionMetas(
	championId int,
	countByPosition map[string]ChampionPositionStatistics,
	championDetailStatisticsMetaMXDAOs []statistics_models.ChampionDetailStatisticsMetaMXDAO,
	championCounterMXDAOs []statistics_models.ChampionCounterStatisticsMXDAO,
) (*ChampionDetailStatisticsPositionMetaTree, error) {

	metaTrees := &ChampionDetailStatisticsPositionMetaTree{
		Top:     nil,
		Jungle:  nil,
		Mid:     nil,
		Adc:     nil,
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

	// key: teamPosition -> counterChampionId -> ChampionCounterStatisticsMXDAO
	countersByPositionMap := make(map[string]map[int]statistics_models.ChampionCounterStatisticsMXDAO)
	for _, counter := range championCounterMXDAOs {
		teamPosition := counter.TeamPosition
		if _, exists := countersByPositionMap[teamPosition]; !exists {
			countersByPositionMap[teamPosition] = make(map[int]statistics_models.ChampionCounterStatisticsMXDAO)
		}
		countersByPositionMap[teamPosition][counter.EnemyChampionId] = counter
	}

	for teamPosition, metaMXDAOs := range metaMXDAOsByPosition {
		pickCount, winCount := 0, 0
		positionStatistics, ok := countByPosition[teamPosition]
		if ok {
			pickCount = positionStatistics.PickCount
			winCount = positionStatistics.WinCount
		} else {
			pickCount = 0
			winCount = 0
		}

		positionItems := make([]int, 0)
		for _, metaMXDAO := range metaMXDAOs {
			majorItems := getValidItems([]*int{
				&metaMXDAO.Item0Id,
				&metaMXDAO.Item1Id,
				&metaMXDAO.Item2Id,
				metaMXDAO.Item3Id,
				metaMXDAO.Item4Id,
				metaMXDAO.Item5Id,
			})
			for _, itemId := range majorItems {
				positionItems = append(positionItems, itemId)
			}
		}

		positionItemCounts := getDescSortedPositionItemCounts(positionItems)
		positionItemTags := getPositionItemTags(positionItems)
		lowDepthItemRecommends, err := getLowDepthItemRecommendations(championId, teamPosition, positionItemTags)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		metaGroup := make(MetaGroup, 0)
		for _, metaMXDAO := range metaMXDAOs {
			majorItems := getValidItems([]*int{
				&metaMXDAO.Item0Id,
				&metaMXDAO.Item1Id,
				&metaMXDAO.Item2Id,
				metaMXDAO.Item3Id,
				metaMXDAO.Item4Id,
				metaMXDAO.Item5Id,
			})

			// get categories of tags from major items
			tagCategories := make(map[string]int)
			for _, itemId := range majorItems {
				item := service.Items[itemId]
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
			maxCount := 0
			for category, count := range tagCategories {
				categoryCounts = append(categoryCounts, CategoryCount{category: category, count: count})
				if count > maxCount {
					maxCount = count
				}
			}

			var majorTag string
			var minorTag *string
			if maxCount > 1 {
				// sort categories by count (desc)
				sort.SliceStable(categoryCounts, func(i, j int) bool {
					if categoryCounts[i].count == categoryCounts[j].count {
						return categoryCounts[i].category < categoryCounts[j].category
					}
					return categoryCounts[i].count > categoryCounts[j].count
				})

				// get major tag and minor tag
				if len(categoryCounts) > 0 {
					majorTag = categoryCounts[0].category
					if len(categoryCounts) > 1 {
						minorTag = &categoryCounts[1].category
					}
				} else {
					continue
				}
			} else {
				// set major tag as first category & minor tag as nil
				if len(categoryCounts) > 0 {
					majorTag = categoryCounts[0].category
				} else {
					continue
				}
			}

			startItems, basicItems, subItems, err := getItemTrees(positionItemCounts, lowDepthItemRecommends, majorItems)
			if err != nil {
				log.Error(err)
				return nil, err
			}

			uuid := uuid2.New()
			pickRate := 0.0
			if pickCount > 0 {
				pickRate = float64(metaMXDAO.Total) / float64(pickCount)
			}
			item0Id := metaMXDAO.Item0Id
			item1Id := metaMXDAO.Item1Id
			item2Id := metaMXDAO.Item2Id
			metaPick := MetaPick{
				Id:                uuid.String(),
				Summoner1Id:       metaMXDAO.Summoner1Id,
				Summoner2Id:       metaMXDAO.Summoner2Id,
				PrimaryStyleId:    metaMXDAO.PrimaryStyle,
				PrimaryPerk0:      metaMXDAO.PrimaryPerk0,
				PrimaryPerk1:      metaMXDAO.PrimaryPerk1,
				PrimaryPerk2:      metaMXDAO.PrimaryPerk2,
				PrimaryPerk3:      metaMXDAO.PrimaryPerk3,
				SubStyleId:        metaMXDAO.SubStyle,
				SubPerk0:          metaMXDAO.SubPerk0,
				SubPerk1:          metaMXDAO.SubPerk1,
				StatPerkDefenseId: metaMXDAO.StatPerkDefense,
				StatPerkFlexId:    metaMXDAO.StatPerkFlex,
				StatPerkOffenseId: metaMXDAO.StatPerkOffense,
				Item0:             &item0Id,
				Item1:             &item1Id,
				Item2:             &item2Id,
				Item3:             metaMXDAO.Item3Id,
				Item4:             metaMXDAO.Item4Id,
				Item5:             metaMXDAO.Item5Id,
				Wins:              metaMXDAO.Wins,
				Total:             metaMXDAO.Total,
				WinRate:           metaMXDAO.WinRate,
				PickRate:          pickRate,
				MetaRank:          metaMXDAO.MetaRank,
				MajorTag:          majorTag,
				MinorTag:          minorTag,
				StartItems:        startItems,
				BasicItems:        basicItems,
				SubItems:          subItems,
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

		// pick popular concept groups
		concepts := make([]string, 0)
		for concept, _ := range conceptGroups {
			concepts = append(concepts, concept)
		}
		// sort concept groups (pickRate desc, winRate desc)
		sort.SliceStable(concepts, func(i, j int) bool {
			conceptI, conceptJ := concepts[i], concepts[j]
			groupI, groupJ := conceptGroups[conceptI], conceptGroups[conceptJ]
			pickCountI, pickCountJ := groupI.getTotalPickCount(), groupJ.getTotalPickCount()
			winRateI, winRateJ := groupI.getTotalWinRate(), groupJ.getTotalWinRate()
			if pickCountI != pickCountJ {
				return pickCountI > pickCountJ
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
			pickCountI, pickCountJ := groupI.getTotalPickCount(), groupJ.getTotalPickCount()
			winRateI, winRateJ := groupI.getTotalWinRate(), groupJ.getTotalWinRate()
			if winRateI != winRateJ {
				return winRateI > winRateJ
			}
			return pickCountI > pickCountJ
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

		counterMap := make(map[int]ChampionCounterStatistics)
		counters, exists := countersByPositionMap[teamPosition]
		if !exists {
			counters = make(map[int]statistics_models.ChampionCounterStatisticsMXDAO)
		}
		for counterChampionId, counterInfo := range counters {
			counterChampion, exists := service.Champions[strconv.Itoa(counterInfo.EnemyChampionId)]
			if !exists {
				log.Warnf("champion not found: %d", counterInfo.EnemyChampionId)
				continue
			}

			majorItems := getValidItems([]*int{
				counterInfo.Item0Id,
				counterInfo.Item1Id,
				counterInfo.Item2Id,
				counterInfo.Item3Id,
				counterInfo.Item4Id,
				counterInfo.Item5Id,
			})
			startItems, basicItems, subItems, err := getItemTrees(positionItemCounts, lowDepthItemRecommends, majorItems)
			if err != nil {
				log.Error(err)
				return nil, err
			}

			metaPick := MetaPick{
				Id:                uuid2.New().String(),
				Summoner1Id:       counterInfo.Summoner1Id,
				Summoner2Id:       counterInfo.Summoner2Id,
				PrimaryStyleId:    counterInfo.PrimaryStyle,
				PrimaryPerk0:      counterInfo.PrimaryPerk0,
				PrimaryPerk1:      counterInfo.PrimaryPerk1,
				PrimaryPerk2:      counterInfo.PrimaryPerk2,
				PrimaryPerk3:      counterInfo.PrimaryPerk3,
				SubStyleId:        counterInfo.SubStyle,
				SubPerk0:          counterInfo.SubPerk0,
				SubPerk1:          counterInfo.SubPerk1,
				StatPerkDefenseId: counterInfo.StatPerkDefense,
				StatPerkFlexId:    counterInfo.StatPerkFlex,
				StatPerkOffenseId: counterInfo.StatPerkOffense,
				Item0:             counterInfo.Item0Id,
				Item1:             counterInfo.Item1Id,
				Item2:             counterInfo.Item2Id,
				Item3:             counterInfo.Item3Id,
				Item4:             counterInfo.Item4Id,
				Item5:             counterInfo.Item5Id,
				Wins:              counterInfo.Wins,
				Total:             counterInfo.Total,
				WinRate:           counterInfo.WinRate,
				PickRate:          0,
				MetaRank:          0,
				MajorTag:          "",
				MinorTag:          nil,
				StartItems:        startItems,
				BasicItems:        basicItems,
				SubItems:          subItems,
			}
			realMeta, err := metaPick.toRealMeta()
			if err != nil {
				log.Error(err)
				return nil, err
			}

			counterMap[counterChampionId] = ChampionCounterStatistics{
				CounterChampionId:   counterInfo.EnemyChampionId,
				CounterChampionName: counterChampion.Name,
				AvgKills:            counterInfo.AvgKills,
				AvgDeaths:           counterInfo.AvgDeaths,
				AvgAssists:          counterInfo.AvgAssists,
				CounterAvgKills:     counterInfo.EnemyAvgKills,
				CounterAvgDeaths:    counterInfo.EnemyAvgDeaths,
				CounterAvgAssists:   counterInfo.EnemyAvgAssists,
				Summoner1Id:         counterInfo.Summoner1Id,
				Summoner2Id:         counterInfo.Summoner2Id,
				MajorPerkGroup:      realMeta.MajorPerkGroup,
				MinorPerkGroup:      realMeta.MinorPerkGroup,
				PerkExtra:           realMeta.PerkExtra,
				MainSlots:           realMeta.MainSlots,
				SubSlots:            realMeta.SubSlots,
				StatSlots:           realMeta.StatSlots,
				StartItemTree:       realMeta.StartItemTree,
				BasicItemTree:       realMeta.BasicItemTree,
				ItemTree:            realMeta.ItemTree,
				SubItemTree:         realMeta.SubItemTree,
				Count:               counterInfo.Total,
				Win:                 counterInfo.Wins,
				WinRate:             counterInfo.WinRate,
				ExpectedWinRate:     counterInfo.TotalWinRate,
			}
		}

		metaTree := ChampionDetailStatisticsMetaTree{
			MajorMetaPicks: make([]ChampionDetailStatisticsMeta, 0),
			MinorMetaPicks: make([]ChampionDetailStatisticsMeta, 0),
			MetaPicks:      make([]ChampionDetailStatisticsMeta, 0),
			PickCount:      pickCount,
			WinCount:       winCount,
			CounterMap:     counterMap,
		}
		for _, metaPick := range majorMetaPicks {
			meta, err := metaPick.toRealMeta()
			if err != nil {
				log.Error(err)
				return nil, err
			}
			metaTree.MajorMetaPicks = append(metaTree.MajorMetaPicks, *meta)
		}
		for _, metaPick := range minorMetaPicks {
			meta, err := metaPick.toRealMeta()
			if err != nil {
				log.Error(err)
				return nil, err
			}
			metaTree.MinorMetaPicks = append(metaTree.MinorMetaPicks, *meta)
		}
		for _, metaPick := range metaGroup {
			meta, err := metaPick.toRealMeta()
			if err != nil {
				log.Error(err)
				return nil, err
			}
			metaTree.MetaPicks = append(metaTree.MetaPicks, *meta)
		}

		if teamPosition == types.TeamPositionTop {
			metaTrees.Top = &metaTree
		} else if teamPosition == types.TeamPositionJungle {
			metaTrees.Jungle = &metaTree
		} else if teamPosition == types.TeamPositionMid {
			metaTrees.Mid = &metaTree
		} else if teamPosition == types.TeamPositionAdc {
			metaTrees.Adc = &metaTree
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
