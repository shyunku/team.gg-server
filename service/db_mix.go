package service

import (
	"database/sql"
	"errors"
	"team.gg-server/libs/database"
)

type SummonerRecentMatchSummaryEntity struct {
	MatchId            string `db:"match_id" json:"matchId"`
	DataVersion        string `db:"data_version" json:"dataVersion"`
	GameCreation       int64  `db:"game_creation" json:"gameCreation"`
	GameDuration       int64  `db:"game_duration" json:"gameDuration"`
	GameEndTimestamp   int64  `db:"game_end_timestamp" json:"gameEndTimestamp"`
	GameId             int64  `db:"game_id" json:"gameId"`
	GameMode           string `db:"game_mode" json:"gameMode"`
	GameName           string `db:"game_name" json:"gameName"`
	GameStartTimestamp int64  `db:"game_start_timestamp" json:"gameStartTimestamp"`
	GameType           string `db:"game_type" json:"gameType"`
	GameVersion        string `db:"game_version" json:"gameVersion"`
	MapId              int    `db:"map_id" json:"mapId"`
	PlatformId         string `db:"platform_id" json:"platformId"`
	QueueId            int    `db:"queue_id" json:"queueId"`
	TournamentCode     string `db:"tournament_code" json:"tournamentCode"`

	None0                          string `db:"mp.match_id" json:"none0"`
	ParticipantId                  int    `db:"participant_id" json:"participantId"`
	MatchParticipantId             string `db:"match_participant_id" json:"matchParticipantId"`
	Puuid                          string `db:"puuid" json:"puuid"`
	Kills                          int    `db:"kills" json:"kills"`
	Deaths                         int    `db:"deaths" json:"deaths"`
	Assists                        int    `db:"assists" json:"assists"`
	ChampionId                     int    `db:"champion_id" json:"championId"`
	ChampionLevel                  int    `db:"champion_level" json:"championLevel"`
	ChampionName                   string `db:"champion_name" json:"championName"`
	ChampExperience                int    `db:"champ_experience" json:"champExperience"`
	SummonerLevel                  int    `db:"summoner_level" json:"summonerLevel"`
	SummonerName                   string `db:"summoner_name" json:"summonerName"`
	RiotIdName                     string `db:"riot_id_name" json:"riotIdName"`
	RiotIdTagLine                  string `db:"riot_id_tag_line" json:"riotIdTagLine"`
	ProfileIcon                    int    `db:"profile_icon" json:"profileIcon"`
	MagicDamageDealtToChampions    int    `db:"magic_damage_dealt_to_champions" json:"magicDamageDealtToChampions"`
	PhysicalDamageDealtToChampions int    `db:"physical_damage_dealt_to_champions" json:"physicalDamageDealtToChampions"`
	TrueDamageDealtToChampions     int    `db:"true_damage_dealt_to_champions" json:"trueDamageDealtToChampions"`
	TotalDamageDealtToChampions    int    `db:"total_damage_dealt_to_champions" json:"totalDamageDealtToChampions"`
	MagicDamageTaken               int    `db:"magic_damage_taken" json:"magicDamageTaken"`
	PhysicalDamageTaken            int    `db:"physical_damage_taken" json:"physicalDamageTaken"`
	TrueDamageTaken                int    `db:"true_damage_taken" json:"trueDamageTaken"`
	TotalDamageTaken               int    `db:"total_damage_taken" json:"totalDamageTaken"`
	TotalHeal                      int    `db:"total_heal" json:"totalHeal"`
	TotalHealsOnTeammates          int    `db:"total_heals_on_teammates" json:"totalHealsOnTeammates"`
	Item0                          int    `db:"item0" json:"item0"`
	Item1                          int    `db:"item1" json:"item1"`
	Item2                          int    `db:"item2" json:"item2"`
	Item3                          int    `db:"item3" json:"item3"`
	Item4                          int    `db:"item4" json:"item4"`
	Item5                          int    `db:"item5" json:"item5"`
	Item6                          int    `db:"item6" json:"item6"`
	Spell1Casts                    int    `db:"spell1_casts" json:"spell1Casts"`
	Spell2Casts                    int    `db:"spell2_casts" json:"spell2Casts"`
	Spell3Casts                    int    `db:"spell3_casts" json:"spell3Casts"`
	Spell4Casts                    int    `db:"spell4_casts" json:"spell4Casts"`
	Summoner1Casts                 int    `db:"summoner1_casts" json:"summoner1Casts"`
	Summoner1Id                    int    `db:"summoner1_id" json:"summoner1Id"`
	Summoner2Casts                 int    `db:"summoner2_casts" json:"summoner2Casts"`
	Summoner2Id                    int    `db:"summoner2_id" json:"summoner2Id"`
	FirstBloodAssist               bool   `db:"first_blood_assist" json:"firstBloodAssist"`
	FirstBloodKill                 bool   `db:"first_blood_kill" json:"firstBloodKill"`
	DoubleKills                    int    `db:"double_kills" json:"doubleKills"`
	TripleKills                    int    `db:"triple_kills" json:"tripleKills"`
	QuadraKills                    int    `db:"quadra_kills" json:"quadraKills"`
	PentaKills                     int    `db:"penta_kills" json:"pentaKills"`
	TotalMinionsKilled             int    `db:"total_minions_killed" json:"totalMinionsKilled"`
	TotalTimeCCDealt               int    `db:"total_time_cc_dealt" json:"totalTimeCCDealt"`
	NeutralMinionsKilled           int    `db:"neutral_minions_killed" json:"neutralMinionsKilled"`
	GoldSpent                      int    `db:"gold_spent" json:"goldSpent"`
	GoldEarned                     int    `db:"gold_earned" json:"goldEarned"`
	IndividualPosition             string `db:"individual_position" json:"individualPosition"`
	TeamPosition                   string `db:"team_position" json:"teamPosition"`
	Lane                           string `db:"lane" json:"lane"`
	Role                           string `db:"role" json:"role"`
	TeamId                         int    `db:"team_id" json:"teamId"`
	VisionScore                    int    `db:"vision_score" json:"visionScore"`
	Win                            bool   `db:"win" json:"win"`
	GameEndedInEarlySurrender      bool   `db:"game_ended_in_early_surrender" json:"gameEndedInEarlySurrender"`
	GameEndedInSurrender           bool   `db:"game_ended_in_surrender" json:"gameEndedInSurrender"`
	TeamEarlySurrendered           bool   `db:"team_early_surrendered" json:"teamEarlySurrendered"`
}

func GetSummonerRecentMatchSummaries(puuid string) ([]*SummonerRecentMatchSummaryEntity, error) {
	var summaries []*SummonerRecentMatchSummaryEntity
	if err := database.DB.Select(&summaries, `
		SELECT m.*, mp.*
		FROM summoner_matches sm
		LEFT JOIN matches m ON sm.match_id = m.match_id
		LEFT JOIN match_participants mp ON mp.match_id = m.match_id AND mp.puuid = sm.puuid
		WHERE sm.puuid = ?
		ORDER BY m.game_end_timestamp DESC
		LIMIT 20`, puuid); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return summaries, nil
}
