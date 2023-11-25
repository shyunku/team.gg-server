package service

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"
	log "github.com/shyunku-libraries/go-logger"
	"team.gg-server/models"
	"team.gg-server/third_party/riot"
	"team.gg-server/util"
	"time"
)

var (
	DataDragonVersion      = ""
	LocalDataDragonVersion = ""

	Champions      = make(map[string]ChampionInfo)      // key: champion key
	SummonerSpells = make(map[string]SummonerSpellInfo) // key: summoner spell key
	Perks          = make(map[int]PerkInfo)             // key: perk id
	PerkStyles     = make(map[int]PerkStyleInfo)        // key: perk style id
)

func Preload() error {
	log.Debugf("Service preload started...")

	// load data dragon version
	var err error
	DataDragonVersion, err = GetLatestDataDragonVersion()
	if err != nil {
		return err
	}
	log.Debugf("DataDragon version: %s", DataDragonVersion)

	// manage & load data dragon files
	if err := SanitizeAndLoadDataDragonFile(); err != nil {
		return err
	}

	// load summoner spell data
	if err := LoadSummonerSpellsData(); err != nil {
		return err
	}

	// load champion data
	championsInfo, err := GetLatestChampionData()
	if err != nil {
		return err
	}
	for _, champion := range championsInfo.Data {
		Champions[champion.Key] = champion
	}
	log.Debugf("%d Champion data loaded", len(Champions))

	// load perks data
	perksData, err := GetCDragonPerksData()
	if err != nil {
		return err
	}
	for _, perk := range perksData {
		Perks[perk.Id] = perk
	}

	// load perk styles data
	perkStylesData, err := GetCDragonPerkStylesData()
	if err != nil {
		return err
	}
	for _, perkStyle := range (*perkStylesData).Styles {
		PerkStyles[perkStyle.Id] = perkStyle
	}

	return nil
}

func RefreshSummonerInfoByName(tx *sql.Tx, name string) error {
	// get summoner by name
	summoner, err := riot.GetSummonerByName(name)
	if err != nil {
		log.Warnf("failed to get summoner by name (%s)", name)
		return err
	}

	// TODO :: update summoner?

	return RefreshSummonerInfoByPuuid(tx, summoner.Puuid)
}

func RefreshSummonerInfoByPuuid(tx *sql.Tx, puuid string) error {
	if puuid == "" {
		return errors.New("puuid is required")
	}

	/* -------------------- refresh summoner -------------------- */
	summoner, err := riot.GetSummonerByPuuid(puuid)
	if err != nil {
		log.Warnf("failed to get summoner by puuid (%s)", puuid)
		return err
	}
	if err := refreshSummoner(tx, summoner); err != nil {
		return err
	}

	/* -------------------- refresh summoner rank -------------------- */
	leagues, err := riot.GetLeaguesBySummonerId(summoner.Id)
	if err != nil {
		log.Warnf("failed to get league by summoner id (%s)", summoner.Id)
		return err
	}
	if err := refreshRank(tx, summoner, leagues); err != nil {
		return err
	}

	/* -------------------- refresh mastery -------------------- */
	mastery, err := riot.GetMasteryBySummonerId(summoner.Id)
	if err != nil {
		log.Warnf("failed to get mastery by summoner id (%s)", summoner.Id)
		return err
	}
	if err := refreshMastery(tx, summoner, mastery); err != nil {
		return err
	}

	/* -------------------- refresh match -------------------- */
	matches, err := riot.GetMatchIdsInterval(puuid, nil, nil)
	if err != nil {
		log.Warnf("failed to get match ids by puuid (%s)", puuid)
		return err
	}
	if err := refreshMatch(tx, summoner.Puuid, *matches); err != nil {
		return err
	}

	return nil
}

func refreshSummoner(tx *sql.Tx, summoner *riot.SummonerDto) error {
	// check if summoner exists in db
	summonerEntity, exists, err := models.StrictGetSummonerByPuuid(summoner.Puuid)
	if err != nil {
		return err
	}

	if !exists {
		// create new summoner
		summonerEntity = &models.SummonerEntity{
			AccountId:     summoner.AccountId,
			ProfileIconId: summoner.ProfileIconId,
			RevisionDate:  summoner.RevisionDate,
			Name:          summoner.Name,
			Id:            summoner.Id,
			Puuid:         summoner.Puuid,
			SummonerLevel: summoner.SummonerLevel,
			ShortenName:   util.ShortenSummonerName(summoner.Name),
			LastUpdatedAt: time.Now(),
			Hits:          0,
		}
	} else {
		// update summoner
		summonerEntity.AccountId = summoner.AccountId
		summonerEntity.ProfileIconId = summoner.ProfileIconId
		summonerEntity.RevisionDate = summoner.RevisionDate
		summonerEntity.Name = summoner.Name
		summonerEntity.Id = summoner.Id
		summonerEntity.Puuid = summoner.Puuid
		summonerEntity.SummonerLevel = summoner.SummonerLevel
		summonerEntity.ShortenName = util.ShortenSummonerName(summoner.Name)
		summonerEntity.LastUpdatedAt = time.Now()
		summonerEntity.Hits += 1
	}
	if err := summonerEntity.UpsertTx(tx); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func refreshRank(tx *sql.Tx, summoner *riot.SummonerDto, leagues *riot.LeagueDto) error {
	for _, league := range *leagues {
		if league.SummonerId != summoner.Id {
			log.Errorf("league summoner id (%s) != summoner id (%s)", league.SummonerId, summoner.Id)
			return errors.New("league summoner id is not equal to summoner id")
		}

		// create new league
		leagueEntity := &models.LeagueEntity{
			Puuid:      summoner.Puuid,
			LeagueId:   league.LeagueId,
			QueueType:  league.QueueType,
			Tier:       league.Tier,
			Rank:       league.Rank,
			Wins:       league.Wins,
			Losses:     league.Losses,
			HotStreak:  league.HotStreak,
			Veteran:    league.Veteran,
			FreshBlood: league.FreshBlood,
			Inactive:   league.Inactive,
			MsTarget:   league.MiniSeries.Target,
			MsWins:     league.MiniSeries.Wins,
			MsLosses:   league.MiniSeries.Losses,
			MsProgress: league.MiniSeries.Progress,
		}

		if err := leagueEntity.UpsertTx(tx); err != nil {
			return err
		}
	}

	return nil
}

func refreshMastery(tx *sql.Tx, summoner *riot.SummonerDto, masteries *riot.MasteryDto) error {
	for _, mastery := range *masteries {
		if mastery.Puuid != summoner.Puuid {
			log.Errorf("mastery puuid (%s) != summoner puuid (%s)", mastery.Puuid, summoner.Puuid)
			return errors.New("mastery puuid is not equal to summoner puuid")
		}

		// upsert mastery
		masteryEntity := &models.MasteryEntity{
			Puuid:                        summoner.Puuid,
			ChampionId:                   mastery.ChampionId,
			ChampionLevel:                mastery.ChampionLevel,
			ChampionPoints:               mastery.ChampionPoints,
			ChampionPointsSinceLastLevel: mastery.ChampionPointsSinceLastLevel,
			ChampionPointsUntilNextLevel: mastery.ChampionPointsUntilNextLevel,
			ChestGranted:                 mastery.ChestGranted,
			LastPlayTime:                 time.UnixMilli(mastery.LastPlayTime),
			TokensEarned:                 mastery.TokensEarned,
		}

		if err := masteryEntity.Upsert(tx); err != nil {
			return err
		}
	}

	return nil
}

func refreshMatch(tx *sql.Tx, puuid string, matchIds []string) error {
	for _, matchId := range matchIds {
		// check if match exists in db
		_, exists, err := models.StrictGetMatchByMatchId(matchId)
		if err != nil {
			log.Error(err)
			return err
		}

		if !exists {
			// get match by match id
			match, err := riot.GetMatchByMatchId(matchId)
			if err != nil {
				log.Error(err)
				return err
			}

			// insert new match
			matchEntity := models.MatchEntity{
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
			if err := matchEntity.InsertTx(tx); err != nil {
				log.Error(err)
				return err
			}

			// insert new match participants
			for _, p := range match.Info.Participants {
				matchParticipantId := uuid.New().String()

				if p.Puuid == puuid {
					// insert new summoner match (upsert)
					summonerMatchEntity := models.SummonerMatchEntity{
						Puuid:   p.Puuid,
						MatchId: matchId,
					}
					if err := summonerMatchEntity.UpsertTx(tx); err != nil {
						log.Error(err)
						return err
					}
				}

				// insert new match participant
				matchParticipantEntity := models.MatchParticipantEntity{
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
				if err := matchParticipantEntity.InsertTx(tx); err != nil {
					log.Error(err)
					return err
				}

				// insert new match participant detail
				matchParticipantDetailEntity := models.MatchParticipantDetailEntity{
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
				if err := matchParticipantDetailEntity.InsertTx(tx); err != nil {
					log.Error(err)
					return err
				}

				// insert new match participant perk
				matchParticipantPerkEntity := models.MatchParticipantPerkEntity{
					MatchParticipantId: matchParticipantId,
					StatPerkDefense:    p.Perks.StatPerks.Defense,
					StatPerkFlex:       p.Perks.StatPerks.Flex,
					StatPerkOffense:    p.Perks.StatPerks.Offense,
				}
				if err := matchParticipantPerkEntity.InsertTx(tx); err != nil {
					log.Error(err)
					return err
				}

				// insert new match participant perk style
				for _, style := range p.Perks.Styles {
					styleId := uuid.New().String()
					matchParticipantPerkStyleEntity := models.MatchParticipantPerkStyleEntity{
						MatchParticipantId: matchParticipantId,
						StyleId:            styleId,
						Description:        style.Description,
						Style:              style.Style,
					}
					if err := matchParticipantPerkStyleEntity.InsertTx(tx); err != nil {
						log.Error(err)
						return err
					}

					// insert new match participant perk style selections
					for _, selection := range style.Selections {
						matchParticipantPerkStyleSelectionEntity := models.MatchParticipantPerkStyleSelectionEntity{
							StyleId: styleId,
							Perk:    selection.Perk,
							Var1:    selection.Var1,
							Var2:    selection.Var2,
							Var3:    selection.Var3,
						}
						if err := matchParticipantPerkStyleSelectionEntity.InsertTx(tx); err != nil {
							log.Error(err)
							return err
						}
					}
				}
			}

			// insert new match team
			for _, t := range match.Info.Teams {
				matchTeamEntity := models.MatchTeamEntity{
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
				if err := matchTeamEntity.InsertTx(tx); err != nil {
					log.Error(err)
					return err
				}

				// insert new match team bans
				for _, ban := range t.Bans {
					matchTeamBanEntity := models.MatchTeamBanEntity{
						MatchId:    matchId,
						TeamId:     t.TeamId,
						ChampionId: ban.ChampionId,
						PickTurn:   ban.PickTurn,
					}
					if err := matchTeamBanEntity.InsertTx(tx); err != nil {
						log.Error(err)
						return err
					}
				}
			}
		} else {
			// ok, match exists in db
			// check if match is connected -> summoner
			_, exists, err := models.StrictGetSummonerMatch(tx, puuid, matchId)
			if err != nil {
				log.Error(err)
				return err
			}

			if !exists {
				// insert new summoner match (upsert)
				summonerMatchEntity := models.SummonerMatchEntity{
					Puuid:   puuid,
					MatchId: matchId,
				}
				if err := summonerMatchEntity.UpsertTx(tx); err != nil {
					log.Error(err)
					return err
				}
			}
		}
	}

	return nil
}
