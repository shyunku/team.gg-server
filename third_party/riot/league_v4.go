package riot

import (
	"encoding/json"
	"team.gg-server/libs/http"
)

type LeagueItemDto struct {
	LeagueId     string `json:"leagueId"`
	SummonerId   string `json:"summonerId"`
	SummonerName string `json:"summonerName"`
	QueueType    string `json:"queueType"`
	Tier         string `json:"tier"`
	Rank         string `json:"rank"`
	LeaguePoints int    `json:"leaguePoints"`
	Wins         int    `json:"wins"`
	Losses       int    `json:"losses"`
	HotStreak    bool   `json:"hotStreak"`
	Veteran      bool   `json:"veteran"`
	FreshBlood   bool   `json:"freshBlood"`
	Inactive     bool   `json:"inactive"`
	MiniSeries   struct {
		Target   int    `json:"target"`
		Wins     int    `json:"wins"`
		Losses   int    `json:"losses"`
		Progress string `json:"progress"`
	}
}

type LeagueDto []LeagueItemDto

func GetLeaguesBySummonerId(summonerId string) (*LeagueDto, error) {
	incrementApiCalls()
	resp, err := http.Get(http.GetRequest{
		Url: CreateUrl(RegionKr, "/lol/league/v4/entries/by-summoner/"+summonerId),
	})
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, resp.Err
	}

	var league LeagueDto
	if err := json.Unmarshal(resp.Body, &league); err != nil {
		return nil, err
	}

	return &league, nil
}
