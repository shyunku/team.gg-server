package service

const (
	RankTypeSolo = "RANKED_SOLO_5x5"
	RankTypeFlex = "RANKED_FLEX_SR"

	PerkStyleDescriptionTypePrimary = "primaryStyle"
	PerkStyleDescriptionTypeSub     = "subStyle"

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

type ChampionInfo struct {
	Version string `json:"version"`
	Id      string `json:"id"`
	Key     string `json:"key"`
	Name    string `json:"name"`
	Title   string `json:"title"`
	Blurb   string `json:"blurb"`
	Info    struct {
		Attack     int `json:"attack"`
		Defense    int `json:"defense"`
		Magic      int `json:"magic"`
		Difficulty int `json:"difficulty"`
	} `json:"info"`
	Image struct {
		Full   string `json:"full"`
		Sprite string `json:"sprite"`
		Group  string `json:"group"`
		X      int    `json:"x"`
		Y      int    `json:"y"`
		W      int    `json:"w"`
		H      int    `json:"h"`
	} `json:"image"`
	Tags    []string `json:"tags"`
	Partype string   `json:"partype"`
	Stats   struct {
		Hp                   float64 `json:"hp"`
		Hpperlevel           float64 `json:"hpperlevel"`
		Mp                   float64 `json:"mp"`
		Mpperlevel           float64 `json:"mpperlevel"`
		Movespeed            float64 `json:"movespeed"`
		Armor                float64 `json:"armor"`
		Armorperlevel        float64 `json:"armorperlevel"`
		Spellblock           float64 `json:"spellblock"`
		Spellblockperlevel   float64 `json:"spellblockperlevel"`
		Attackrange          float64 `json:"attackrange"`
		Hpregen              float64 `json:"hpregen"`
		Hpregenperlevel      float64 `json:"hpregenperlevel"`
		Mpregen              float64 `json:"mpregen"`
		Mpregenperlevel      float64 `json:"mpregenperlevel"`
		Crit                 float64 `json:"crit"`
		Critperlevel         float64 `json:"critperlevel"`
		Attackdamage         float64 `json:"attackdamage"`
		Attackdamageperlevel float64 `json:"attackdamageperlevel"`
		Attackspeedperlevel  float64 `json:"attackspeedperlevel"`
		Attackspeed          float64 `json:"attackspeed"`
	} `json:"stats"`
}

type DDragonSummonerJsonDto struct {
	Type    string                       `json:"type"`
	Version string                       `json:"version"`
	Data    map[string]SummonerSpellInfo `json:"data"`
}

type SummonerSpellInfo struct {
	Id            string        `json:"id,omitempty"`
	Name          string        `json:"name,omitempty"`
	Description   string        `json:"description,omitempty"`
	Tooltip       string        `json:"tooltip,omitempty"`
	Maxrank       int           `json:"maxrank,omitempty"`
	Cooldown      []float64     `json:"cooldown,omitempty"`
	CooldownBurn  string        `json:"cooldownBurn,omitempty"`
	Cost          []int         `json:"cost,omitempty"`
	CostBurn      string        `json:"costBurn,omitempty"`
	Datavalues    interface{}   `json:"datavalues,omitempty"`
	Effect        []interface{} `json:"effect,omitempty"`
	EffectBurn    []*string     `json:"effectBurn,omitempty"`
	Vars          []interface{} `json:"vars,omitempty"`
	Key           string        `json:"key,omitempty"`
	SummonerLevel int           `json:"summonerLevel,omitempty"`
	Modes         []string      `json:"modes,omitempty"`
	CostType      string        `json:"costType,omitempty"`
	Maxammo       string        `json:"maxammo,omitempty"`
	Range         []int         `json:"range,omitempty"`
	RangeBurn     string        `json:"rangeBurn,omitempty"`
	Image         struct {
		Full   string `json:"full,omitempty"`
		Sprite string `json:"sprite,omitempty"`
		Group  string `json:"group,omitempty"`
		X      int    `json:"x,omitempty"`
		Y      int    `json:"y,omitempty"`
		W      int    `json:"w,omitempty"`
		H      int    `json:"h,omitempty"`
	} `json:"image"`
	Resource string `json:"resource,omitempty"`
}
