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
	leagueVo, err := SummonerRankMixer(*leagueDAO)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return leagueVo, nil
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

func GetSummonerRecentMatchSummaryVOs(puuid string, queueId, count int) ([]MatchSummaryVO, error) {
	matchSummaryMXDAOs, err := GetSummonerRecentMatchSummaryMXDAOs_byQueueId(puuid, queueId, count)
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

func GetCustomGameConfigurationVOs(uid string) ([]CustomGameConfigurationSummaryVO, error) {
	customGameConfigurationDAOs, err := models.GetCustomGameDAOs_byCreatorUid(db.Root, uid)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	customGameConfigurationVOs := make([]CustomGameConfigurationSummaryVO, 0)
	for _, customGameConfigurationDAO := range customGameConfigurationDAOs {
		customGameConfigurationVOs = append(customGameConfigurationVOs, CustomGameConfigurationSummaryMixer(customGameConfigurationDAO))
	}

	return customGameConfigurationVOs, nil
}

func GetCustomGameCandidateVO(candidateDAO models.CustomGameCandidateDAO) (*CustomGameCandidateVO, error) {
	summonerDao, exists, err := models.GetSummonerDAO_byPuuid(db.Root, candidateDAO.Puuid)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("summoner dao not found with puuid (%s)", candidateDAO.Puuid)
	}
	summonerVO := SummonerSummaryMixer(*summonerDao)

	var soloLeagueVO *SummonerRankVO
	soloLeagueDAO, exists, err := models.GetLeagueDAO(db.Root, candidateDAO.Puuid, RankTypeSolo)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if exists {
		soloLeagueVO, err = SummonerRankMixer(*soloLeagueDAO)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}

	var flexLeagueVO *SummonerRankVO
	flexLeagueDAO, exists, err := models.GetLeagueDAO(db.Root, candidateDAO.Puuid, RankTypeFlex)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if exists {
		flexLeagueVO, err = SummonerRankMixer(*flexLeagueDAO)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}

	var customRankVO *SummonerRankVO
	if candidateDAO.CustomTier != nil && candidateDAO.CustomRank != nil {
		customRankVO = &SummonerRankVO{
			Tier:   *candidateDAO.CustomTier,
			Rank:   *candidateDAO.CustomRank,
			Lp:     0,
			Wins:   0,
			Losses: 0,
		}
		ratingPoint, err := CalculateRatingPoint(*candidateDAO.CustomTier, *candidateDAO.CustomRank, 0)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		customRankVO.RatingPoint = ratingPoint
	}

	positionFavorVO := CustomGameCandidatePositionFavorVO{
		Top:     candidateDAO.FlavorTop,
		Jungle:  candidateDAO.FlavorJungle,
		Mid:     candidateDAO.FlavorMid,
		Adc:     candidateDAO.FlavorAdc,
		Support: candidateDAO.FlavorSupport,
	}

	masteryDAO, err := models.GetMasteryDAOs(db.Root, candidateDAO.Puuid)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	masteryVOs := make([]SummonerMasteryVO, 0)
	for _, mastery := range masteryDAO {
		masteryVOs = append(masteryVOs, SummonerMasteryMixer(*mastery))
	}

	return &CustomGameCandidateVO{
		Summary:       summonerVO,
		SoloRank:      soloLeagueVO,
		FlexRank:      flexLeagueVO,
		CustomRank:    customRankVO,
		PositionFavor: positionFavorVO,
		Mastery:       masteryVOs,
	}, nil
}

func GetCustomGameConfigurationVO(configurationId string) (*CustomGameConfigurationVO, error) {
	customGameConfigurationDAO, exists, err := models.GetCustomGameDAO_byId(db.Root, configurationId)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("custom game configuration dao not found with id (%s)", configurationId)
	}

	candidateVOs := make([]CustomGameCandidateVO, 0)
	team1ParticipantsVOs := make([]CustomGameParticipantVO, 0)
	team2ParticipantsVOs := make([]CustomGameParticipantVO, 0)

	// get candidates
	candidateDAOs, err := models.GetCustomGameCandidateDAOs_byCustomGameConfigId(db.Root, configurationId)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	for _, candidateDAO := range candidateDAOs {
		summonerVO, err := GetCustomGameCandidateVO(candidateDAO)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		candidateVOs = append(candidateVOs, *summonerVO)
	}

	// get participants
	participantDAOs, err := models.GetCustomGameParticipantDAOs_byCustomGameConfigId(db.Root, configurationId)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	for _, participantDAO := range participantDAOs {
		participantVO := CustomGameConfigurationParticipantMixer(*participantDAO)
		if participantDAO.Team == 1 {
			team1ParticipantsVOs = append(team1ParticipantsVOs, participantVO)
		} else {
			team2ParticipantsVOs = append(team2ParticipantsVOs, participantVO)
		}
	}

	customGameConfigurationVO := CustomGameConfigurationMixer(
		*customGameConfigurationDAO,
		candidateVOs,
		team1ParticipantsVOs,
		team2ParticipantsVOs,
	)
	return &customGameConfigurationVO, nil
}

func GetCustomGameConfigurationBalanceVO(customGameConfigId string) (*CustomGameConfigurationBalanceVO, error) {
	customGameConfigDAO, exists, err := models.GetCustomGameDAO_byId(db.Root, customGameConfigId)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("custom game config dao not found with id (%s)", customGameConfigId)
	}

	fairnessVO := CustomGameConfigurationFairnessMixer(*customGameConfigDAO)
	return &fairnessVO, nil
}
