package service

import (
	"fmt"
	log "github.com/shyunku-libraries/go-logger"
	"team.gg-server/libs/db"
	"team.gg-server/models"
)

// vo_getters configure vo with VAO and mixed-VAO (null-safe)

func GetSummonerSummaryVO_byName(summonerName string) (*SummonerSummaryVO, error) {
	// find summoner by name on db
	summonerDao, exists, err := models.GetSummonerDAO_byName(db.Root, summonerName)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("summoner dao not found with name (%s)", summonerName)
	}
	summonerVo := SummonerSummaryMixer(*summonerDao)
	return &summonerVo, nil
}

func GetSummonerSummaryVO_byPuuid(puuid string) (*SummonerSummaryVO, error) {
	// find summoner by name on db
	summonerDao, exists, err := models.GetSummonerDAO_byPuuid(db.Root, puuid)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("summoner dao not found with puuid (%s)", puuid)
	}
	summonerVo := SummonerSummaryMixer(*summonerDao)
	return &summonerVo, nil
}

// GetSummonerRankVO returns SummonerRankVO by puuid and rankType
// this function assumes that summoner info & rank info has consistency
func GetSummonerRankVO(puuid string, rankType string) (*SummonerRankVO, error) {
	leagueDAO, exists, err := models.GetLeagueDAO(db.Root, puuid, rankType)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	leagueVo := SummonerRankMixer(*leagueDAO)
	return &leagueVo, nil
}

func GetSummonerMasteryVOs(puuid string) ([]SummonerMasteryVO, error) {
	masteries, err := models.GetMasteryDAOs(db.Root, puuid)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var masteryVos []SummonerMasteryVO
	for _, mastery := range masteries {
		masteryVos = append(masteryVos, SummonerMasteryMixer(*mastery))
	}
	return masteryVos, nil
}

func GetSummonerRecentMatchSummaryVOs(puuid string, count int) ([]MatchSummaryVO, error) {
	matchSummaryMXDAOs, err := GetSummonerRecentMatchSummaryMXDAOs(puuid, count)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return getSummonerMatchSummaryVOs(puuid, matchSummaryMXDAOs)
}

func GetSummonerMatchSummaryVOs_before(puuid string, before int64, count int) ([]MatchSummaryVO, error) {
	matchSummaryMXDAOs, err := GetSummonerMatchSummaryMXDAOS_before(puuid, before, count)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return getSummonerMatchSummaryVOs(puuid, matchSummaryMXDAOs)
}

func getSummonerMatchSummaryVOs(puuid string, matchSummaryMXDAOs []*SummonerMatchSummaryMXDAO) ([]MatchSummaryVO, error) {
	matchSummaryVOs := make([]MatchSummaryVO, 0)
	for _, summonerRecentMatchSummaryMXDAO := range matchSummaryMXDAOs {
		matchParticipants, err := models.GetMatchParticipantDAOs(db.Root, summonerRecentMatchSummaryMXDAO.MatchId)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		var myStat *SummonerMatchParticipantVO

		team1Participants := make([]TeammateVO, 0)
		team2Participants := make([]TeammateVO, 0)
		for _, matchParticipant := range matchParticipants {
			teamMate := SummonerMatchSummaryTeamMateMixer(matchParticipant)
			if matchParticipant.TeamId == 100 {
				team1Participants = append(team1Participants, teamMate)
			} else {
				team2Participants = append(team2Participants, teamMate)
			}
		}

		for _, matchParticipant := range matchParticipants {
			perks, err := models.GetMatchParticipantPerkStyleDAOs(db.Root, matchParticipant.MatchParticipantId)
			if err != nil {
				log.Warn(err)
			}

			primaryPerkStyle := 0
			subPerkStyle := 0
			for _, perk := range perks {
				if perk.Description == PerkStyleDescriptionTypePrimary {
					primaryPerkStyle = perk.Style
				} else if perk.Description == PerkStyleDescriptionTypeSub {
					subPerkStyle = perk.Style
				}
			}

			if matchParticipant.Puuid == puuid {
				me := SummonerMatchSummaryParticipantMixer(*summonerRecentMatchSummaryMXDAO, primaryPerkStyle, subPerkStyle)
				myStat = &me
			}
		}

		if myStat == nil {
			return nil, fmt.Errorf("myStat is nil")
		}

		matchSummaryVOs = append(matchSummaryVOs, SummonerMatchSummaryMixer(
			*summonerRecentMatchSummaryMXDAO,
			*myStat,
			team1Participants,
			team2Participants,
		))
	}

	return matchSummaryVOs, nil
}
