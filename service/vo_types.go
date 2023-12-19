package service

import "time"

/* ------------------------ Service VOs ------------------------ */

type SummonerSummaryVO struct {
	ProfileIconId int       `json:"profileIconId"`
	GameName      string    `json:"gameName"`
	TagLine       string    `json:"tagLine"`
	Name          string    `json:"name"`
	Puuid         string    `json:"puuid"`
	SummonerLevel int64     `json:"summonerLevel"`
	LastUpdatedAt time.Time `json:"lastUpdatedAt"`
}

type SummonerRankVO struct {
	Tier        string `json:"tier"`
	Rank        string `json:"rank"`
	Lp          int    `json:"lp"`
	Wins        int    `json:"wins"`
	Losses      int    `json:"losses"`
	RatingPoint int64  `json:"ratingPoint"`
}

type SummonerMasteryVO struct {
	ChampionId     int64   `json:"championId"`
	ChampionName   *string `json:"championName"`
	ChampionLevel  int     `json:"championLevel"`
	ChampionPoints int     `json:"championPoints"`
}

type SummonerMatchParticipantVO struct {
	MatchId            string `json:"matchId"`
	ParticipantId      int    `json:"participantId"`
	MatchParticipantId string `json:"matchParticipantId"`
	Puuid              string `json:"puuid"`
	Kills              int    `json:"kills"`
	Deaths             int    `json:"deaths"`
	Assists            int    `json:"assists"`
	ChampionId         int    `json:"championId"`
	ChampionLevel      int    `json:"championLevel"`
	SummonerLevel      int    `json:"summonerLevel"`
	SummonerName       string `json:"summonerName"`
	RiotIdName         string `json:"riotIdName"`
	RiotIdTagLine      string `json:"riotIdTagLine"`
	ProfileIcon        int    `json:"profileIcon"`

	Item0 int `json:"item0"`
	Item1 int `json:"item1"`
	Item2 int `json:"item2"`
	Item3 int `json:"item3"`
	Item4 int `json:"item4"`
	Item5 int `json:"item5"`
	Item6 int `json:"item6"`

	Spell1Casts    int `json:"spell1Casts"`
	Spell2Casts    int `json:"spell2Casts"`
	Spell3Casts    int `json:"spell3Casts"`
	Spell4Casts    int `json:"spell4Casts"`
	Summoner1Casts int `json:"summoner1Casts"`
	Summoner1Id    int `json:"summoner1Id"`
	Summoner2Casts int `json:"summoner2Casts"`
	Summoner2Id    int `json:"summoner2Id"`

	PrimaryPerkStyle int `json:"primaryPerkStyle"`
	SubPerkStyle     int `json:"subPerkStyle"`

	DoubleKills int `json:"doubleKills"`
	TripleKills int `json:"tripleKills"`
	QuadraKills int `json:"quadraKills"`
	PentaKills  int `json:"pentaKills"`

	TotalMinionsKilled          int `json:"totalMinionsKilled"`
	TotalCCDealt                int `json:"totalCCDealt"`
	TotalDamageDealtToChampions int `json:"totalDamageDealtToChampions"`

	GoldEarned int    `json:"goldEarned"`
	Lane       string `json:"lane"`
	Win        bool   `json:"win"`

	IndividualPosition string `json:"individualPosition"`
	TeamPosition       string `json:"teamPosition"`

	GameEndedInEarlySurrender bool `json:"gameEndedInEarlySurrender"`
	GameEndedInSurrender      bool `json:"gameEndedInSurrender"`
	TeamEarlySurrendered      bool `json:"teamEarlySurrendered"`
}

type TeammateVO struct {
	ChampionId            int    `json:"championId"`
	SummonerName          string `json:"summonerName"`
	RiotIdName            string `json:"riotIdName"`
	RiotIdTagLine         string `json:"riotIdTagLine"`
	Puuid                 string `json:"puuid"`
	TotalDealtToChampions int    `json:"totalDealtToChampions"`
	Kills                 int    `json:"kills"`
	IndividualPosition    string `json:"individualPosition"`
	TeamPosition          string `json:"teamPosition"`
	ProfileIcon           int    `json:"profileIcon"`
}

type MatchSummaryVO struct {
	MatchId            string                     `json:"matchId"`
	GameStartTimestamp int64                      `json:"gameStartTimestamp"`
	GameEndTimestamp   int64                      `json:"gameEndTimestamp"`
	GameDuration       int64                      `json:"gameDuration"`
	QueueId            int                        `json:"queueId"`
	MyStat             SummonerMatchParticipantVO `json:"myStat"`
	Team1              []TeammateVO               `json:"team1"`
	Team2              []TeammateVO               `json:"team2"`
}

type IngameParticipantVO struct {
	ChampionId    int64  `json:"championId"`
	ProfileIconId int64  `json:"profileIconId"`
	SummonerName  string `json:"summonerName"`
	SummonerId    string `json:"summonerId"`
}

/* ------------------------ Preload VOs ------------------------ */

type ChampionDataVO struct {
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

type SummonerSpellDataVO struct {
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

type PerkInfoVO struct {
	Id                                  int         `json:"id"`
	Name                                string      `json:"name"`
	MajorChangePatchVersion             string      `json:"majorChangePatchVersion"`
	ToolTip                             string      `json:"tooltip"`
	ShortDesc                           string      `json:"shortDesc"`
	LongDesc                            string      `json:"longDesc"`
	RecommendationDescriptor            string      `json:"recommendationDescriptor"`
	IconPath                            string      `json:"iconPath"`
	EndOfGameStatDescs                  []string    `json:"endOfGameStatDescs"`
	RecommendationDescriptionAttributes interface{} `json:"recommendationDescriptionAttributes"`
}

type PerkStyleInfoVO struct {
	Id               int         `json:"id"`
	Name             string      `json:"name"`
	ToolTip          string      `json:"tooltip"`
	IconPath         string      `json:"iconPath"`
	AssetMap         interface{} `json:"assetMap"`
	IsAdvanced       bool        `json:"isAdvanced"`
	AllowedSubStyles []int       `json:"allowedSubStyles"`
	SubStyleBonus    []struct {
		StyleId int `json:"styleId"`
		PerkId  int `json:"perkId"`
	} `json:"subStyleBonus"`
	Slots []struct {
		Type      string `json:"type"`
		SlotLabel string `json:"slotLabel"`
		Perks     []int  `json:"perks"`
	} `json:"slots"`
	DefaultPageName            string `json:"defaultPageName"`
	DefaultSubStyle            int    `json:"defaultSubStyle"`
	DefaultPerks               []int  `json:"defaultPerks"`
	DefaultPerksWhenSplashed   []int  `json:"defaultPerksWhenSplashed"`
	DefaultStatModsPerSubStyle []struct {
		Id    string `json:"id"`
		Perks []int  `json:"perks"`
	} `json:"defaultStatModsPerSubStyle"`
}

type CustomGameConfigurationSummaryVO struct {
	Id            string    `json:"id"`
	Name          string    `json:"name"`
	LastUpdatedAt time.Time `json:"lastUpdatedAt"`
	Fairness      float64   `json:"fairness"`
}

type CustomGameCandidatePositionFavorVO struct {
	Top     bool `json:"top"`
	Jungle  bool `json:"jungle"`
	Mid     bool `json:"mid"`
	Adc     bool `json:"adc"`
	Support bool `json:"support"`
}

type CustomGameCandidateVO struct {
	Summary       SummonerSummaryVO                  `json:"summary"`
	SoloRank      *SummonerRankVO                    `json:"soloRank"`
	FlexRank      *SummonerRankVO                    `json:"flexRank"`
	CustomRank    *SummonerRankVO                    `json:"customRank"`
	PositionFavor CustomGameCandidatePositionFavorVO `json:"positionFavor"`
	Mastery       []SummonerMasteryVO                `json:"mastery"`
}

type CustomGameParticipantVO struct {
	Position string `json:"position"`
	Puuid    string `json:"puuid"`
}

type CustomGameConfigurationVO struct {
	Id            string    `json:"id"`
	Name          string    `json:"name"`
	CreatorUid    string    `json:"creatorUid"`
	CreatedAt     time.Time `json:"createdAt"`
	LastUpdatedAt time.Time `json:"lastUpdatedAt"`
	Fairness      float64   `json:"fairness"`
	LineFairness  float64   `json:"lineFairness"`
	TierFairness  float64   `json:"tierFairness"`

	Candidates []CustomGameCandidateVO `json:"candidates"`

	Team1 []CustomGameParticipantVO `json:"team1"`
	Team2 []CustomGameParticipantVO `json:"team2"`
}
