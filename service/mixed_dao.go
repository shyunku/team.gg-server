package service

type SummonerRankingMXDAO struct {
	Puuid        string  `db:"puuid" json:"puuid"`
	GameName     string  `db:"game_name" json:"gameName"`
	TagLine      string  `db:"tag_line" json:"tagLine"`
	Tier         *string `db:"tier" json:"tier"`
	LeagueRank   *string `db:"league_rank" json:"leagueRank"`
	LeaguePoints int     `db:"league_points" json:"leaguePoints"`

	RatingPoints float64 `db:"rating_points" json:"ratingPoints"`
	Ranking      int     `db:"ranking" json:"ranking"`
	Total        int     `db:"total" json:"total"`
}

type ChampionStatisticMXDAO struct {
	ChampionId int `db:"champion_id" json:"championId"`
	Win        int `db:"win" json:"win"`
	Total      int `db:"total" json:"total"`

	PickRate float64 `db:"pick_rate" json:"pickRate"`
	BanRate  float64 `db:"ban_rate" json:"banRate"`

	AvgMinionsKilled float64 `db:"avg_minions_killed" json:"avgMinionsKilled"`
	AvgKills         float64 `db:"avg_kills" json:"avgKills"`
	AvgDeaths        float64 `db:"avg_deaths" json:"avgDeaths"`
	AvgAssists       float64 `db:"avg_assists" json:"avgAssists"`
	AvgGoldEarned    float64 `db:"avg_gold_earned" json:"avgGoldEarned"`
}

type TierStatisticsTierCountMXDAO struct {
	QueueType  string `db:"queue_type" json:"queueType"`
	Tier       string `db:"tier" json:"tier"`
	LeagueRank string `db:"league_rank" json:"leagueRank"`
	Count      int    `db:"count" json:"count"`
}

type TierStatisticsTopRankersMXDAO struct {
	QueueType  string `db:"queue_type" json:"queueType"`
	Tier       string `db:"tier" json:"tier"`
	LeagueRank string `db:"league_rank" json:"leagueRank"`

	Puuid         string `db:"puuid" json:"puuid"`
	ProfileIconId int    `db:"profile_icon_id" json:"profileIconId"`
	GameName      string `db:"game_name" json:"gameName"`
	TagLine       string `db:"tag_line" json:"tagLine"`

	LeaguePoints int `db:"league_points" json:"leaguePoints"`
	Wins         int `db:"wins" json:"wins"`
	Losses       int `db:"losses" json:"losses"`
	Ranks        int `db:"ranks" json:"ranks"`
}

type MasteryStatisticsMXDAO struct {
	ChampionId    int     `db:"champion_id" json:"championId"`
	AvgMastery    float64 `db:"avg_mastery" json:"avgMastery"`
	MaxMastery    int     `db:"max_mastery" json:"maxMastery"`
	TotalMastery  int     `db:"total_mastery" json:"totalMastery"`
	MasteredCount int     `db:"mastered_count" json:"masteredCount"`
	Count         int     `db:"count" json:"count"`
}

type MasteryStatisticsTopRankersMXDAO struct {
	Puuid         string `db:"puuid" json:"puuid"`
	ProfileIconId int    `db:"profile_icon_id" json:"profileIconId"`
	GameName      string `db:"game_name" json:"gameName"`
	TagLine       string `db:"tag_line" json:"tagLine"`

	Ranks int `db:"ranks" json:"ranks"`

	ChampionId     int `db:"champion_id" json:"championId"`
	ChampionPoints int `db:"champion_points" json:"championPoints"`
}
