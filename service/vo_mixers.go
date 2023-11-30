package service

import (
	"strconv"
	"team.gg-server/models"
	"team.gg-server/third_party/riot"
)

// vo_mixers manage conversion of VAO -> VO

func SummonerSummaryMixer(d models.SummonerDAO) SummonerSummaryVO {
	return SummonerSummaryVO{
		ProfileIconId: d.ProfileIconId,
		GameName:      d.GameName,
		TagLine:       d.TagLine,
		Name:          d.Name,
		Puuid:         d.Puuid,
		SummonerLevel: d.SummonerLevel,
		LastUpdatedAt: d.LastUpdatedAt,
	}
}

func SummonerRankMixer(d models.LeagueDAO) SummonerRankVO {
	return SummonerRankVO{
		Tier:   d.Tier,
		Rank:   d.Rank,
		Lp:     d.LeaguePoints,
		Wins:   d.Wins,
		Losses: d.Losses,
	}
}

func SummonerMasteryMixer(d models.MasteryDAO) SummonerMasteryVO {
	var championName *string
	champion, ok := Champions[strconv.FormatInt(d.ChampionId, 10)]
	if ok {
		championName = &champion.Name
	}

	return SummonerMasteryVO{
		ChampionId:     d.ChampionId,
		ChampionName:   championName,
		ChampionLevel:  d.ChampionLevel,
		ChampionPoints: d.ChampionPoints,
	}
}

func SummonerMatchSummaryTeamMateMixer(d models.MatchParticipantDAO) TeammateVO {
	return TeammateVO{
		ChampionId:            d.ChampionId,
		SummonerName:          d.SummonerName,
		RiotIdName:            d.RiotIdName,
		RiotIdTagLine:         d.RiotIdTagLine,
		Puuid:                 d.Puuid,
		TotalDealtToChampions: d.TotalDamageDealtToChampions,
		Kills:                 d.Kills,
		IndividualPosition:    d.IndividualPosition,
		TeamPosition:          d.TeamPosition,
		ProfileIcon:           d.ProfileIcon,
	}
}

func SummonerMatchSummaryParticipantMixer(
	d SummonerMatchSummaryMXDAO, primaryPerkStyle int, subPerkStyle int) SummonerMatchParticipantVO {

	return SummonerMatchParticipantVO{
		MatchId:                     d.MatchId,
		ParticipantId:               d.ParticipantId,
		MatchParticipantId:          d.MatchParticipantId,
		Puuid:                       d.Puuid,
		Kills:                       d.Kills,
		Deaths:                      d.Deaths,
		Assists:                     d.Assists,
		ChampionId:                  d.ChampionId,
		ChampionLevel:               d.ChampionLevel,
		SummonerLevel:               d.SummonerLevel,
		SummonerName:                d.SummonerName,
		RiotIdName:                  d.RiotIdName,
		RiotIdTagLine:               d.RiotIdTagLine,
		ProfileIcon:                 d.ProfileIcon,
		Item0:                       d.Item0,
		Item1:                       d.Item1,
		Item2:                       d.Item2,
		Item3:                       d.Item3,
		Item4:                       d.Item4,
		Item5:                       d.Item5,
		Item6:                       d.Item6,
		Spell1Casts:                 d.Spell1Casts,
		Spell2Casts:                 d.Spell2Casts,
		Spell3Casts:                 d.Spell3Casts,
		Spell4Casts:                 d.Spell4Casts,
		Summoner1Casts:              d.Summoner1Casts,
		Summoner1Id:                 d.Summoner1Id,
		Summoner2Casts:              d.Summoner2Casts,
		Summoner2Id:                 d.Summoner2Id,
		PrimaryPerkStyle:            primaryPerkStyle,
		SubPerkStyle:                subPerkStyle,
		DoubleKills:                 d.DoubleKills,
		TripleKills:                 d.TripleKills,
		QuadraKills:                 d.QuadraKills,
		PentaKills:                  d.PentaKills,
		TotalMinionsKilled:          d.TotalMinionsKilled,
		TotalCCDealt:                d.TotalTimeCCDealt,
		TotalDamageDealtToChampions: d.TotalDamageDealtToChampions,
		GoldEarned:                  d.GoldEarned,
		Lane:                        d.Lane,
		Win:                         d.Win,
		IndividualPosition:          d.IndividualPosition,
		TeamPosition:                d.TeamPosition,
		GameEndedInEarlySurrender:   d.GameEndedInEarlySurrender,
		GameEndedInSurrender:        d.GameEndedInSurrender,
		TeamEarlySurrendered:        d.TeamEarlySurrendered,
	}
}

func SummonerMatchSummaryMixer(d SummonerMatchSummaryMXDAO, myStat SummonerMatchParticipantVO, team1 []TeammateVO, team2 []TeammateVO) MatchSummaryVO {
	return MatchSummaryVO{
		MatchId:            d.MatchId,
		GameStartTimestamp: d.GameStartTimestamp,
		GameEndTimestamp:   d.GameEndTimestamp,
		GameDuration:       d.GameDuration,
		QueueId:            d.QueueId,
		MyStat:             myStat,
		Team1:              team1,
		Team2:              team2,
	}
}

func IngameParticipantMixer(d riot.SpectatorParticipantDto) IngameParticipantVO {
	return IngameParticipantVO{
		ChampionId:    d.ChampionId,
		ProfileIconId: d.ProfileIconId,
		SummonerName:  d.SummonerName,
		SummonerId:    d.SummonerId,
	}
}
