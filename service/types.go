package service

const (
	RankTypeSolo = "RANKED_SOLO_5x5"
	RankTypeFlex = "RANKED_FLEX_SR"

	PerkStyleDescriptionTypePrimary = "primaryStyle"
	PerkStyleDescriptionTypeSub     = "subStyle"

	LoadInitialMatchCount = 10
	LoadMoreMatchCount    = 5

	PositionTop     = "TOP"
	PositionJungle  = "JUNGLE"
	PositionMid     = "MID"
	PositionAdc     = "ADC"
	PositionSupport = "SUPPORT"

	WeightLineFairness     = 0.36
	WeightTierFairness     = 0.24
	WeightLineSatisfaction = 1 - WeightLineFairness - WeightTierFairness

	WeightTopInfluence     = 0.14
	WeightJungleInfluence  = 0.23
	WeightMidInfluence     = 0.25
	WeightAdcInfluence     = 0.21
	WeightSupportInfluence = 1 - WeightTopInfluence - WeightJungleInfluence - WeightMidInfluence - WeightAdcInfluence

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

var (
	GetSupportedPositions    = [...]string{PositionTop, PositionJungle, PositionMid, PositionAdc, PositionSupport}
	GetPossibleTeamPositions = func() []CustomGameTeamPositionVO {
		return []CustomGameTeamPositionVO{
			{Team: 1, Position: PositionTop},
			{Team: 1, Position: PositionJungle},
			{Team: 1, Position: PositionMid},
			{Team: 1, Position: PositionAdc},
			{Team: 1, Position: PositionSupport},
			{Team: 2, Position: PositionTop},
			{Team: 2, Position: PositionJungle},
			{Team: 2, Position: PositionMid},
			{Team: 2, Position: PositionAdc},
			{Team: 2, Position: PositionSupport},
		}
	}
)
