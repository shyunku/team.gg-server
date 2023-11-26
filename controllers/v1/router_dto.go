package v1

import "team.gg-server/service"

type GetSummonerInfoRequestDto struct {
	SummonerName string `form:"summonerName" binding:"required"`
}

type GetSummonerInfoResponseDto struct {
	Summary  service.SummonerSummaryVO   `json:"summary"`
	SoloRank *service.SummonerRankVO     `json:"soloRank"`
	FlexRank *service.SummonerRankVO     `json:"flexRank"`
	Mastery  []service.SummonerMasteryVO `json:"mastery"`
	Matches  []service.MatchSummaryVO    `json:"matches"`
}

type RenewSummonerInfoRequestDto struct {
	Puuid string `json:"puuid" binding:"required"`
}

type LoadMatchesRequestDto struct {
	Puuid  string `json:"puuid" binding:"required"`
	Before *int64 `json:"before" binding:"required"`
}

type LoadMatchesResponseDto []service.MatchSummaryVO
