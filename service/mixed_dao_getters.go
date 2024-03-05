package service

import (
	"database/sql"
	"errors"
	"team.gg-server/libs/db"
)

func GetSummonerSoloRankingMXDAO(puuid string) (*SummonerRankingMXDAO, error) {
	var rankingMXDAO SummonerRankingMXDAO
	if err := db.Root.Get(&rankingMXDAO, `
		WITH rank_data AS (
			SELECT
				s.puuid,
				s.game_name,
				s.tag_line,
				l.tier,
				l.league_rank,
				IF(ISNULL(l.league_rank), 0, l.league_points) as league_points,
				IF(ISNULL(l.league_rank), 0, (str.score + l.league_points)) as rating_points,
				ROW_NUMBER() OVER (ORDER BY IF(ISNULL(l.league_rank), 0, (str.score + l.league_points)) DESC) as ranking
			FROM
				summoners s
			LEFT JOIN
				leagues l ON s.puuid = l.puuid AND l.queue_type = ?
			LEFT JOIN
				static_tier_ranks str ON l.tier = str.tier_label AND l.league_rank = str.rank_label
		), total_rankers AS (
			SELECT COUNT(*) as total FROM rank_data
		)
		SELECT rank_data.*, total_rankers.total
		FROM rank_data, total_rankers
		WHERE rank_data.puuid = ?;
	`, RankTypeSolo, puuid); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &rankingMXDAO, nil
}

func GetChampionStatisticMXDAOs() ([]*ChampionStatisticMXDAO, error) {
	var statistics []*ChampionStatisticMXDAO
	if err := StatisticsDB.Select(&statistics, `
		WITH ChampionStats AS (
			SELECT
				mp.champion_id AS champion_id,
				SUM(t.win) AS win,
				COUNT(*) AS total,
				AVG(mp.total_minions_killed) as avg_minions_killed,
				AVG(mp.kills) as avg_kills,
				AVG(mp.deaths) as avg_deaths,
				AVG(mp.assists) as avg_assists,
				AVG(mp.gold_earned) as avg_gold_earned
			FROM match_participants mp
			LEFT JOIN matches m ON mp.match_id = m.match_id
			LEFT JOIN match_teams t ON mp.team_id = t.team_id AND m.match_id = t.match_id
			LEFT JOIN match_team_bans b ON b.match_id = m.match_id AND b.champion_id = mp.champion_id
			GROUP BY mp.champion_id
		), BanStats AS (
			SELECT
				champion_id,
				COUNT(*) as total_bans
			FROM match_team_bans
			GROUP BY champion_id
		), MatchCount AS (
			SELECT
				COUNT(*) as matches
			FROM matches
		)
		SELECT
			cs.*,
    		IF(ISNULL(bs.total_bans), 0, bs.total_bans / mc.matches) as ban_rate,
			cs.total / mc.matches as pick_rate
		FROM ChampionStats cs
		LEFT JOIN BanStats bs ON cs.champion_id = bs.champion_id
		CROSS JOIN MatchCount mc;
	`); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]*ChampionStatisticMXDAO, 0), nil
		}
		return nil, err
	}
	return statistics, nil
}

func GetTierStatisticsTierCountMXDAOs() ([]*TierStatisticsTierCountMXDAO, error) {
	var tierCounts []*TierStatisticsTierCountMXDAO
	if err := StatisticsDB.Select(&tierCounts, `
		SELECT l.queue_type, l.tier, l.league_rank, COUNT(*) AS count
		FROM leagues l
		LEFT JOIN summoners s ON l.puuid = s.puuid
		WHERE l.queue_type = 'RANKED_SOLO_5x5' OR l.queue_type = 'RANKED_FLEX_SR'
		GROUP BY l.queue_type, l.tier, l.league_rank;
	`); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]*TierStatisticsTierCountMXDAO, 0), nil
		}
		return nil, err
	}
	return tierCounts, nil
}

func GetTierStatisticsTopRankersMXDAOs(topRanks int) ([]*TierStatisticsTopRankersMXDAO, error) {
	var topRankers []*TierStatisticsTopRankersMXDAO
	if err := StatisticsDB.Select(&topRankers, `
		WITH RankedLeagues AS (
			SELECT
				l.queue_type,
				l.tier,
				l.league_rank,
				l.puuid,
				l.league_points,
				l.wins,
				l.losses,
				ROW_NUMBER() OVER (
					PARTITION BY l.queue_type, l.tier, l.league_rank
					ORDER BY l.league_points DESC, l.wins, l.losses
				) AS ranks
			FROM leagues l
			WHERE l.queue_type = 'RANKED_SOLO_5x5' OR l.queue_type = 'RANKED_FLEX_SR'
		)
		SELECT
			rl.queue_type,
			rl.tier,
			rl.league_rank,
			s.puuid,
			s.profile_icon_id,
			s.game_name,
			s.tag_line,
			rl.league_points,
			rl.wins,
			rl.losses,
			rl.ranks
		FROM RankedLeagues rl
		LEFT JOIN summoners s ON rl.puuid = s.puuid
		WHERE rl.ranks <= ?
		ORDER BY rl.queue_type, rl.tier, rl.league_rank, rl.ranks ASC;
	`, topRanks); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]*TierStatisticsTopRankersMXDAO, 0), nil
		}
		return nil, err
	}
	return topRankers, nil
}

func GetMasteryStatisticsMXDAOs() ([]*MasteryStatisticsMXDAO, error) {
	var statistics []*MasteryStatisticsMXDAO
	if err := StatisticsDB.Select(&statistics, `
		SELECT
			m.champion_id,
			MAX(m.champion_points) as max_mastery,
			AVG(m.champion_points) as avg_mastery,
			SUM(m.champion_points) as total_mastery,
			SUM(IF(m.champion_level >= 7, 1, 0)) as mastered_count,
			COUNT(*) as count
		FROM masteries m
		GROUP BY m.champion_id;
	`); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]*MasteryStatisticsMXDAO, 0), nil
		}
		return nil, err
	}
	return statistics, nil
}

func GetMasteryStatisticsTopRankersMXDAOs(topRanks int) ([]*MasteryStatisticsTopRankersMXDAO, error) {
	var topRankers []*MasteryStatisticsTopRankersMXDAO
	if err := StatisticsDB.Select(&topRankers, `
		WITH RankedMasteries AS (
			SELECT
			    puuid,
				champion_id,
				champion_points,
				ROW_NUMBER() OVER (PARTITION BY champion_id ORDER BY champion_points DESC) AS ranks
			FROM masteries
		)
		SELECT s.puuid, s.game_name, s.tag_line, s.profile_icon_id, rm.champion_id, rm.champion_points, rm.ranks
		FROM RankedMasteries rm
		LEFT JOIN summoners s ON rm.puuid = s.puuid
		WHERE ranks <= ?;
	`, topRanks); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]*MasteryStatisticsTopRankersMXDAO, 0), nil
		}
		return nil, err
	}
	return topRankers, nil
}
