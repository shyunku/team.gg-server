package v1

import "team.gg-server/service"

type LoginRequestDto struct {
	UserId            string `json:"userId" binding:"required"`
	EncryptedPassword string `json:"encryptedPassword" binding:"required"`
}

type LoginResponseDto struct {
	Uid    string `json:"uid"`
	UserId string `json:"userId"`
}

type SignupRequestDto struct {
	UserId            string `json:"userId" binding:"required"`
	EncryptedPassword string `json:"encryptedPassword" binding:"required"`
}

type GetSummonerInfoRequestDto struct {
	GameName string  `form:"gameName" binding:"required"`
	TagLine  *string `form:"tagLine" binding:"required"`
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

type GetIngameInfoRequestDto struct {
	Puuid string `form:"puuid" binding:"required"`
}

type GetIngameInfoResponseDto struct {
	GameType          string                        `json:"gameType"`
	MapId             int64                         `json:"mapId"`
	GameStartTime     int64                         `json:"gameStartTime"`
	GameMode          string                        `json:"gameMode"`
	GameQueueConfigId int64                         `json:"gameQueueConfigId"`
	Team1             []service.IngameParticipantVO `json:"team1"`
	Team2             []service.IngameParticipantVO `json:"team2"`
}
