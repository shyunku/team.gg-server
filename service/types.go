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

	PositionTopEffectiveness     = 0.14
	PositionJungleEffectiveness  = 0.23
	PositionMidEffectiveness     = 0.25
	PositionAdcEffectiveness     = 0.21
	PositionSupportEffectiveness = 1 - PositionTopEffectiveness - PositionJungleEffectiveness - PositionMidEffectiveness - PositionAdcEffectiveness

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
	SupportedPositions = []string{PositionTop, PositionJungle, PositionMid, PositionAdc, PositionSupport}
)
