package service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	log "github.com/shyunku-libraries/go-logger"
	"team.gg-server/libs/db"
	"team.gg-server/models"
	"team.gg-server/third_party/riot"
	"team.gg-server/util"
	"time"
)

// RenewSummonerTotal updates summoner info, league, mastery, matches
// you should use db context with transaction (to prevent inconsistency)
func RenewSummonerTotal(tx *sqlx.Tx, puuid string) error {
	// update summoner info
	if err := RenewSummonerInfoByPuuid(tx, puuid); err != nil {
		log.Error(err)
		return err
	}

	// get summoner by puuid in db
	summonerDao, exists, err := models.GetSummonerDAO_byPuuid(tx, puuid)
	if err != nil {
		log.Error(err)
		return err
	}
	if !exists {
		return fmt.Errorf("summoner (%s) doesn't exist", puuid)
	}

	// update summoner league
	if err := RenewSummonerLeague(tx, summonerDao.Id, summonerDao.Puuid); err != nil {
		log.Error(err)
		return err
	}

	// update summoner mastery
	if err := RenewSummonerMastery(tx, summonerDao.Id, summonerDao.Puuid); err != nil {
		log.Error(err)
		return err
	}

	// update summoner recent matches
	if err := RenewSummonerRecentMatches(tx, summonerDao.Puuid); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func RenewSummonerInfoByPuuid(db db.Context, puuid string) error {
	summoner, _, err := riot.GetSummonerByPuuid(puuid)
	if err != nil {
		log.Error(err)
		return err
	}

	account, _, err := riot.GetAccountByPuuid(puuid)
	if err != nil {
		log.Error(err)
		return err
	}

	if err := renewSummonerInfo(db, summoner, account); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func renewSummonerInfo(db db.Context, summoner *riot.SummonerDto, account *riot.AccountByRiotIdDto) error {
	// make new summoner DAO
	summonerDao := &models.SummonerDAO{
		AccountId:       summoner.AccountId,
		ProfileIconId:   summoner.ProfileIconId,
		RevisionDate:    summoner.RevisionDate,
		GameName:        account.GameName,
		TagLine:         account.TagLine,
		Name:            summoner.Name,
		Id:              summoner.Id,
		Puuid:           summoner.Puuid,
		SummonerLevel:   summoner.SummonerLevel,
		ShortenGameName: util.ShortenSummonerName(account.GameName),
		ShortenName:     util.ShortenSummonerName(summoner.Name),
		LastUpdatedAt:   time.Now(),
	}

	// insert summoner DAO to db
	if err := summonerDao.Upsert(db); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// RenewSummonerLeague updates summoner league info
// this assumes that summoner info is already stored in this context.
func RenewSummonerLeague(db db.Context, summonerId string, puuid string) error {
	leagues, err := riot.GetLeaguesBySummonerId(summonerId)
	if err != nil {
		log.Warnf("failed to get league by summoner id (%s)", summonerId, puuid)
		return err
	}

	for _, league := range *leagues {
		if league.SummonerId != summonerId {
			log.Errorf("league summoner id (%s) != summoner id (%s)", league.SummonerId, summonerId)
			return errors.New("league summoner id is not equal to summoner id")
		}

		// create new league
		leagueEntity := &models.LeagueDAO{
			Puuid:        puuid,
			LeagueId:     league.LeagueId,
			QueueType:    league.QueueType,
			Tier:         league.Tier,
			Rank:         league.Rank,
			Wins:         league.Wins,
			Losses:       league.Losses,
			LeaguePoints: league.LeaguePoints,
			HotStreak:    league.HotStreak,
			Veteran:      league.Veteran,
			FreshBlood:   league.FreshBlood,
			Inactive:     league.Inactive,
			MsTarget:     league.MiniSeries.Target,
			MsWins:       league.MiniSeries.Wins,
			MsLosses:     league.MiniSeries.Losses,
			MsProgress:   league.MiniSeries.Progress,
		}

		if err := leagueEntity.Upsert(db); err != nil {
			return err
		}
	}

	return nil
}

func RenewSummonerMastery(db db.Context, summonerId string, puuid string) error {
	masteries, err := riot.GetMasteryByPuuid(puuid)
	if err != nil {
		log.Warnf("failed to get mastery by summoner id (%s)", summonerId)
		return err
	}

	for _, mastery := range *masteries {
		if mastery.Puuid != puuid {
			log.Errorf("mastery puuid (%s) != summoner puuid (%s)", mastery.Puuid, puuid)
			return errors.New("mastery puuid is not equal to summoner puuid")
		}

		// upsert mastery
		masteryEntity := &models.MasteryDAO{
			Puuid:                        puuid,
			ChampionId:                   mastery.ChampionId,
			ChampionLevel:                mastery.ChampionLevel,
			ChampionPoints:               mastery.ChampionPoints,
			ChampionPointsSinceLastLevel: mastery.ChampionPointsSinceLastLevel,
			ChampionPointsUntilNextLevel: mastery.ChampionPointsUntilNextLevel,
			ChestGranted:                 mastery.ChestGranted,
			LastPlayTime:                 time.UnixMilli(mastery.LastPlayTime),
			TokensEarned:                 mastery.TokensEarned,
		}

		if err := masteryEntity.Upsert(db); err != nil {
			return err
		}
	}

	return nil
}

func RenewSummonerRecentMatches(db db.Context, puuid string) error {
	matches, err := riot.GetMatchIdsInterval(puuid, nil, nil, LoadInitialMatchCount)
	if err != nil {
		log.Warnf("failed to get match ids by puuid (%s)", puuid)
		return err
	}

	for _, matchId := range *matches {
		if err := RenewSummonerMatchIfNecessary(db, puuid, matchId); err != nil {
			return err
		}
	}

	return nil
}

func RenewSummonerMatchesBefore(db db.Context, puuid string, before time.Time) error {
	matches, err := riot.GetMatchIdsInterval(puuid, nil, &before, LoadMoreMatchCount)
	if err != nil {
		log.Warnf("failed to get match ids by puuid (%s)", puuid)
		return err
	}

	for _, matchId := range *matches {
		if err := RenewSummonerMatchIfNecessary(db, puuid, matchId); err != nil {
			return err
		}
	}

	return nil
}

func RenewSummonerMatchIfNecessary(db db.Context, puuid string, matchId string) error {
	matchDAO, exists, err := models.GetMatchDAO(db, matchId)
	if err != nil {
		log.Error(err)
		return err
	}

	if !exists {
		// no match data in local db
		// get match from riot api
		match, err := riot.GetMatchByMatchId(matchId)
		if err != nil {
			log.Error(err)
			return err
		}

		// insert match into db
		matchDAO = &models.MatchDAO{
			MatchId:            matchId,
			DataVersion:        match.Metadata.DataVersion,
			GameCreation:       match.Info.GameCreation,
			GameDuration:       match.Info.GameDuration,
			GameEndTimestamp:   match.Info.GameEndTimestamp,
			GameId:             match.Info.GameId,
			GameMode:           match.Info.GameMode,
			GameName:           match.Info.GameName,
			GameStartTimestamp: match.Info.GameStartTimestamp,
			GameType:           match.Info.GameType,
			GameVersion:        match.Info.GameVersion,
			MapId:              match.Info.MapId,
			PlatformId:         match.Info.PlatformId,
			QueueId:            match.Info.QueueId,
			TournamentCode:     match.Info.TournamentCode,
		}
		if err := matchDAO.Insert(db); err != nil {
			log.Error(err)
			return err
		}

		// insert new match participants
		for _, p := range match.Info.Participants {
			matchParticipantId := uuid.New().String()

			if p.Puuid == puuid {
				// insert new summoner match (upsert)
				summonerMatchEntity := models.SummonerMatchDAO{
					Puuid:   p.Puuid,
					MatchId: matchId,
				}
				if err := summonerMatchEntity.Upsert(db); err != nil {
					log.Error(err)
					return err
				}
			}

			// insert new match participant
			matchParticipantEntity := models.MatchParticipantDAO{
				MatchId:                        matchId,
				ParticipantId:                  p.ParticipantId,
				MatchParticipantId:             matchParticipantId,
				Puuid:                          p.Puuid,
				Kills:                          p.Kills,
				Deaths:                         p.Deaths,
				Assists:                        p.Assists,
				ChampionId:                     p.ChampionId,
				ChampionLevel:                  p.ChampLevel,
				ChampionName:                   p.ChampionName,
				ChampExperience:                p.ChampExperience,
				SummonerLevel:                  p.SummonerLevel,
				SummonerName:                   p.SummonerName,
				RiotIdName:                     p.RiotIdGameName,
				RiotIdTagLine:                  p.RiotIdTagline,
				ProfileIcon:                    p.ProfileIcon,
				MagicDamageDealtToChampions:    p.MagicDamageDealtToChampions,
				PhysicalDamageDealtToChampions: p.PhysicalDamageDealtToChampions,
				TrueDamageDealtToChampions:     p.TrueDamageDealtToChampions,
				TotalDamageDealtToChampions:    p.TotalDamageDealtToChampions,
				MagicDamageTaken:               p.MagicDamageTaken,
				PhysicalDamageTaken:            p.PhysicalDamageTaken,
				TrueDamageTaken:                p.TrueDamageTaken,
				TotalDamageTaken:               p.TotalDamageTaken,
				TotalHeal:                      p.TotalHeal,
				TotalHealsOnTeammates:          p.TotalHealsOnTeammates,
				Item0:                          p.Item0,
				Item1:                          p.Item1,
				Item2:                          p.Item2,
				Item3:                          p.Item3,
				Item4:                          p.Item4,
				Item5:                          p.Item5,
				Item6:                          p.Item6,
				Spell1Casts:                    p.Spell1Casts,
				Spell2Casts:                    p.Spell2Casts,
				Spell3Casts:                    p.Spell3Casts,
				Spell4Casts:                    p.Spell4Casts,
				Summoner1Casts:                 p.Summoner1Casts,
				Summoner1Id:                    p.Summoner1Id,
				Summoner2Casts:                 p.Summoner2Casts,
				Summoner2Id:                    p.Summoner2Id,
				FirstBloodAssist:               p.FirstBloodAssist,
				FirstBloodKill:                 p.FirstBloodKill,
				DoubleKills:                    p.DoubleKills,
				TripleKills:                    p.TripleKills,
				QuadraKills:                    p.QuadraKills,
				PentaKills:                     p.PentaKills,
				TotalMinionsKilled:             p.TotalMinionsKilled,
				TotalTimeCCDealt:               p.TotalTimeCCDealt,
				NeutralMinionsKilled:           p.NeutralMinionsKilled,
				GoldSpent:                      p.GoldSpent,
				GoldEarned:                     p.GoldEarned,
				IndividualPosition:             p.IndividualPosition,
				TeamPosition:                   p.TeamPosition,
				Lane:                           p.Lane,
				Role:                           p.Role,
				TeamId:                         p.TeamId,
				VisionScore:                    p.VisionScore,
				Win:                            p.Win,
				GameEndedInEarlySurrender:      p.GameEndedInEarlySurrender,
				GameEndedInSurrender:           p.GameEndedInSurrender,
				TeamEarlySurrendered:           p.TeamEarlySurrendered,
			}
			if err := matchParticipantEntity.Insert(db); err != nil {
				log.Error(err)
				return err
			}

			// insert new match participant detail
			matchParticipantDetailEntity := models.MatchParticipantDetailDAO{
				MatchParticipantId:             matchParticipantId,
				MatchId:                        matchId,
				BaronKills:                     p.BaronKills,
				BountyLevel:                    p.BountyLevel,
				ChampionTransform:              p.ChampionTransform,
				ConsumablesPurchased:           p.ConsumablesPurchased,
				DamageDealtToBuildings:         p.DamageDealtToBuildings,
				DamageDealtToObjectives:        p.DamageDealtToObjectives,
				DamageDealtToTurrets:           p.DamageDealtToTurrets,
				DamageSelfMitigated:            p.DamageSelfMitigated,
				DetectorWardsPlaced:            p.DetectorWardsPlaced,
				DragonKills:                    p.DragonKills,
				PhysicalDamageDealt:            p.PhysicalDamageDealt,
				MagicDamageDealt:               p.MagicDamageDealt,
				TotalDamageDealt:               p.TotalDamageDealt,
				LargestCriticalStrike:          p.LargestCriticalStrike,
				LargestKillingSpree:            p.LargestKillingSpree,
				LargestMultiKill:               p.LargestMultiKill,
				FirstTowerAssist:               p.FirstTowerAssist,
				FirstTowerKill:                 p.FirstTowerKill,
				InhibitorKills:                 p.InhibitorKills,
				InhibitorTakedowns:             p.InhibitorTakedowns,
				InhibitorsLost:                 p.InhibitorsLost,
				ItemsPurchased:                 p.ItemsPurchased,
				KillingSprees:                  p.KillingSprees,
				NexusKills:                     p.NexusKills,
				NexusTakedowns:                 p.NexusTakedowns,
				NexusLost:                      p.NexusLost,
				LongestTimeSpentLiving:         p.LongestTimeSpentLiving,
				ObjectiveStolen:                p.ObjectiveStolen,
				ObjectiveStolenAssists:         p.ObjectiveStolenAssists,
				SightWardsBoughtInGame:         p.SightWardsBoughtInGame,
				VisionWardsBoughtInGame:        p.VisionWardsBoughtInGame,
				SummonerId:                     p.SummonerId,
				TimeCCingOthers:                p.TimeCCingOthers,
				TimePlayed:                     p.TimePlayed,
				TotalDamageShieldedOnTeammates: p.TotalDamageShieldedOnTeammates,
				TotalTimeSpentDead:             p.TotalTimeSpentDead,
				TotalUnitsHealed:               p.TotalUnitsHealed,
				TrueDamageDealt:                p.TrueDamageDealt,
				TurretKills:                    p.TurretKills,
				TurretTakedowns:                p.TurretTakedowns,
				TurretsLost:                    p.TurretsLost,
				UnrealKills:                    p.UnrealKills,
				WardsKilled:                    p.WardsKilled,
				WardsPlaced:                    p.WardsPlaced,
			}
			if err := matchParticipantDetailEntity.Insert(db); err != nil {
				log.Error(err)
				return err
			}

			// insert new match participant perk
			matchParticipantPerkEntity := models.MatchParticipantPerkDAO{
				MatchParticipantId: matchParticipantId,
				StatPerkDefense:    p.Perks.StatPerks.Defense,
				StatPerkFlex:       p.Perks.StatPerks.Flex,
				StatPerkOffense:    p.Perks.StatPerks.Offense,
			}
			if err := matchParticipantPerkEntity.InsertTx(db); err != nil {
				log.Error(err)
				return err
			}

			// insert new match participant perk style
			for _, style := range p.Perks.Styles {
				styleId := uuid.New().String()
				matchParticipantPerkStyleEntity := models.MatchParticipantPerkStyleDAO{
					MatchParticipantId: matchParticipantId,
					StyleId:            styleId,
					Description:        style.Description,
					Style:              style.Style,
				}
				if err := matchParticipantPerkStyleEntity.Insert(db); err != nil {
					log.Error(err)
					return err
				}

				// insert new match participant perk style selections
				for _, selection := range style.Selections {
					matchParticipantPerkStyleSelectionEntity := models.MatchParticipantPerkStyleSelectionDAO{
						StyleId: styleId,
						Perk:    selection.Perk,
						Var1:    selection.Var1,
						Var2:    selection.Var2,
						Var3:    selection.Var3,
					}
					if err := matchParticipantPerkStyleSelectionEntity.Insert(db); err != nil {
						log.Error(err)
						return err
					}
				}
			}
		}

		// insert new match team
		for _, t := range match.Info.Teams {
			matchTeamEntity := models.MatchTeamDAO{
				MatchId:         matchId,
				TeamId:          t.TeamId,
				Win:             t.Win,
				BaronFirst:      t.Objectives.Baron.First,
				BaronKills:      t.Objectives.Baron.Kills,
				ChampionFirst:   t.Objectives.Champion.First,
				ChampionKills:   t.Objectives.Champion.Kills,
				DragonFirst:     t.Objectives.Dragon.First,
				DragonKills:     t.Objectives.Dragon.Kills,
				InhibitorFirst:  t.Objectives.Inhibitor.First,
				InhibitorKills:  t.Objectives.Inhibitor.Kills,
				RiftHeraldFirst: t.Objectives.RiftHerald.First,
				RiftHeraldKills: t.Objectives.RiftHerald.Kills,
				TowerFirst:      t.Objectives.Tower.First,
				TowerKills:      t.Objectives.Tower.Kills,
			}
			if err := matchTeamEntity.Insert(db); err != nil {
				log.Error(err)
				return err
			}

			// insert new match team bans
			for _, ban := range t.Bans {
				matchTeamBanEntity := models.MatchTeamBanDAO{
					MatchId:    matchId,
					TeamId:     t.TeamId,
					ChampionId: ban.ChampionId,
					PickTurn:   ban.PickTurn,
				}
				if err := matchTeamBanEntity.Insert(db); err != nil {
					log.Error(err)
					return err
				}
			}
		}
	} else {
		// ok, match exists in local db
		// check if match is connected -> summoner
		_, exists, err = models.GetSummonerMatchDAO(db, puuid, matchId)
		if err != nil {
			log.Error(err)
			return err
		}
		if !exists {
			// insert new summoner match (upsert)
			summonerMatchEntity := models.SummonerMatchDAO{
				Puuid:   puuid,
				MatchId: matchId,
			}
			if err := summonerMatchEntity.Upsert(db); err != nil {
				log.Error(err)
				return err
			}
		}
	}

	return nil
}
