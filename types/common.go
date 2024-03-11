package types

import "time"

const (
	RankTypeSolo = "RANKED_SOLO_5x5"
	RankTypeFlex = "RANKED_FLEX_SR"

	PerkStyleDescriptionTypePrimary = "primaryStyle"
	PerkStyleDescriptionTypeSub     = "subStyle"

	LoadInitialMatchCount    = 20
	LoadMoreMatchCount       = 10
	LoadInitialMatchCountDev = 10
	LoadMoreMatchCountDev    = 5

	DataExplorerLoopPeriod       = 1 * time.Second
	DataExplorerLoopPeriodDev    = 60 * time.Second
	DataExplorerLoadMatchesCount = 3

	SummonerRankingRevisionPeriod = 24 * time.Hour

	PositionTop     = "TOP"
	PositionJungle  = "JUNGLE"
	PositionMid     = "MID"
	PositionAdc     = "ADC"
	PositionSupport = "SUPPORT"

	TeamPositionTop     = "TOP"
	TeamPositionJungle  = "JUNGLE"
	TeamPositionMid     = "MIDDLE"
	TeamPositionAdc     = "BOTTOM"
	TeamPositionSupport = "UTILITY"

	WeightLineFairness     = 0.36
	WeightTierFairness     = 0.24
	WeightLineSatisfaction = 1 - WeightLineFairness - WeightTierFairness

	WeightTopInfluence     = 0.14
	WeightJungleInfluence  = 0.23
	WeightMidInfluence     = 0.25
	WeightAdcInfluence     = 0.21
	WeightSupportInfluence = 1 - WeightTopInfluence - WeightJungleInfluence - WeightMidInfluence - WeightAdcInfluence

	QueueTypeAll         = 0   // 전체
	QueueTypeNormalDraft = 400 // 일반 (드래프트)
	QueueTypeSolo        = 420 // 솔랭
	QueueTypeNormal      = 430 // 일반
	QueueTypeFlex        = 440 // 자유 5:5 랭크
	QueueTypeAram        = 450 // 칼바람
	QueueTypeClash       = 700 // 클래시
	QueueTypeUrf         = 900 // 우르프
	QueueTypePoro        = 920 // 포로왕?

	MatchDecoTypeFirstBloodKill     = "FIRST_BLOOD"
	MatchDecoTypeHighestDamage      = "HIGHEST_DAMAGE"
	MatchDecoTypeHighestDamageTaken = "HIGHEST_DAMAGE_TAKEN"
	MatchDecoTypeMostKill           = "MOST_KILL"
	MatchDecoTypeMostAssist         = "MOST_ASSIST"
	MatchDecoTypeMostMinionKill     = "MOST_MINION_KILL"
	MatchDecoTypeHighestKda         = "HIGHEST_KDA"
	MatchDecoTypeMostGold           = "MOST_GOLD"
	MatchDecoTypeMostVisionScore    = "MOST_VISION_SCORE"
	MatchDecoTypeMostWardPlaced     = "MOST_WARD_PLACED"
	MatchDecoTypeMostWardKilled     = "MOST_WARD_KILLED"
	MatchDecoTypeHighestVisionScore = "HIGHEST_VISION_SCORE"
)
