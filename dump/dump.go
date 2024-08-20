package dump

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/schollz/progressbar/v3"
	log "github.com/shyunku-libraries/go-logger"
	"team.gg-server/libs/db"
	"team.gg-server/models"
	"team.gg-server/models/mixed"
	"team.gg-server/types"
)

type DumpArgs struct {
	tx       *sqlx.Tx
	matchDAO models.MatchDAO
}

func DumpMatchesToLegacy() error {
	// collect recent versions
	recentMatchGameVersions, _, err := mixed.GetRecentMatchGameVersions_byDescendingShortVersion_withCount(db.Root, types.RecentVersionCount)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debugf("recentMatchGameVersions: %v", recentMatchGameVersions)

	if len(recentMatchGameVersions) == 0 {
		return fmt.Errorf("no recent match game versions")
	}

	// get total count
	//var totalCount int
	var totalCountRaw []int
	query, args, err := sqlx.In(`
		SELECT COUNT(*) FROM matches
		WHERE game_version NOT IN (?);
	`, recentMatchGameVersions)
	if err != nil {
		log.Error(err)
		return err
	}
	query = db.Root.Rebind(query)
	if err := db.Root.Select(&totalCountRaw, query, args...); err != nil {
		return err
	}
	if len(totalCountRaw) == 0 {
		log.Info("No matches to dump as legacy")
		return nil
	}

	totalCount := totalCountRaw[0]
	if totalCount == 0 {
		log.Info("No matches to dump as legacy")
		return nil
	}

	log.Debugf("totalCount: %d", totalCount)

	// cap
	capping := 5000
	if capping != -1 && totalCount > capping {
		log.Debugf("Capping to %d", capping)
		totalCount = capping
	}

	prog := progressbar.Default(
		int64(totalCount),
		"Dumping matches to legacy...",
	)

	// start tx
	tx, err := db.Root.Beginx()
	if err != nil {
		log.Error(err)
		return err
	}

	batchSize := 1024
	offset := 0
	count := 0
	for {
		forceDown := false

		// get matches batch
		query, args, err := sqlx.In(`
			SELECT * FROM matches
			WHERE game_version NOT IN (?)
			LIMIT ? OFFSET ?;
		`, recentMatchGameVersions, batchSize, offset)
		if err != nil {
			log.Error(err)
			_ = tx.Rollback()
			return err
		}

		var matchDAOs []models.MatchDAO
		query = tx.Rebind(query)
		if err := tx.Select(&matchDAOs, query, args...); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				break
			}
			_ = tx.Rollback
			return err
		}
		if len(matchDAOs) == 0 {
			break
		}

		// move to legacy
		for _, match := range matchDAOs {
			if err := dumpMatch(tx, match); err != nil {
				log.Error(err)
				_ = tx.Rollback
				return err
			}

			count++
			_ = prog.Add(1)

			if capping != -1 && count >= capping {
				forceDown = true
				break
			}
		}

		if forceDown {
			break
		}

		offset += batchSize
	}

	if err := tx.Commit(); err != nil {
		log.Error(err)
		return err
	}

	log.Infof("Dumped %d matches to legacy", totalCount)
	return nil
}

func dumpMatch(tx *sqlx.Tx, match models.MatchDAO) error {
	teamDAOs, err := models.GetMatchTeamDAOs_byMatchId(tx, match.MatchId)
	if err != nil {
		log.Error(err)
		return err
	}

	teamBanDAOs, err := models.GetMatchTeamBanDAOs_byMatchId(tx, match.MatchId)
	if err != nil {
		log.Error(err)
		return err
	}

	summonerMatchDAOs, err := models.GetSummonerMatchDAOs_byMatchId(tx, match.MatchId)
	if err != nil {
		log.Error(err)
		return err
	}

	participantsDAOs, err := models.GetMatchParticipantDAOs_byMatchId(tx, match.MatchId)
	if err != nil {
		log.Error(err)
		return err
	}

	participantPerkDAOs := make([]models.MatchParticipantPerkDAO, 0)
	participantPerkStyleDAOs := make([]models.MatchParticipantPerkStyleDAO, 0)
	participantPerkStyleSelectionDAOs := make([]models.MatchParticipantPerkStyleSelectionDAO, 0)
	participantDetailDAOs := make([]models.MatchParticipantDetailDAO, 0)
	for _, participantDAO := range participantsDAOs {
		perks, err := models.GetMatchParticipantPerkDAOs_byMatchParticipantId(tx, participantDAO.MatchParticipantId)
		if err != nil {
			log.Error(err)
			return err
		}
		participantPerkDAOs = append(participantPerkDAOs, perks...)

		perkStyles, err := models.GetMatchParticipantPerkStyleDAOs(tx, participantDAO.MatchParticipantId)
		if err != nil {
			log.Error(err)
			return err
		}
		participantPerkStyleDAOs = append(participantPerkStyleDAOs, perkStyles...)

		for _, perkStyle := range perkStyles {
			selections, err := models.GetMatchParticipantPerkStyleSelectionDAOs(tx, perkStyle.StyleId)
			if err != nil {
				log.Error(err)
				return err
			}
			participantPerkStyleSelectionDAOs = append(participantPerkStyleSelectionDAOs, selections...)
		}

		participantDetailDAO, err := models.GetMatchParticipantDetailDAOs_byMatchParticipantId(tx, participantDAO.MatchParticipantId)
		if err != nil {
			log.Error(err)
			return err
		}
		participantDetailDAOs = append(participantDetailDAOs, *participantDetailDAO)
	}

	/* ------------------ Dump to legacy ------------------ */
	// match
	legacy := match.ToLegacy()
	if err := legacy.Insert(tx); err != nil {
		log.Error(err)
		return err
	}

	// match teams
	for _, team := range teamDAOs {
		legacy := team.ToLegacy()
		if err := legacy.Insert(tx); err != nil {
			log.Error(err)
			return err
		}
	}

	// match team bans
	for _, ban := range teamBanDAOs {
		legacy := ban.ToLegacy()
		log.Debugf("ban: %v, %v", legacy, match.QueueId)
		if err := legacy.Insert(tx); err != nil {
			log.Error(err)
			return err
		}
	}

	// match participant
	for _, participant := range participantsDAOs {
		legacy := participant.ToLegacy()
		if err := legacy.Insert(tx); err != nil {
			log.Error(err)
			return err
		}
	}

	// match participant details
	for _, detail := range participantDetailDAOs {
		legacy := detail.ToLegacy()
		if err := legacy.Insert(tx); err != nil {
			log.Error(err)
			return err
		}
	}

	// match participant perks
	for _, perk := range participantPerkDAOs {
		legacy := perk.ToLegacy()
		if err := legacy.Insert(tx); err != nil {
			log.Error(err)
			return err
		}
	}

	// match participant perk styles
	for _, perkStyle := range participantPerkStyleDAOs {
		legacy := perkStyle.ToLegacy()
		if err := legacy.Insert(tx); err != nil {
			log.Error(err)
			return err
		}
	}

	// match participant perk style selections
	for _, perkStyleSelection := range participantPerkStyleSelectionDAOs {
		legacy := perkStyleSelection.ToLegacy()
		if err := legacy.Insert(tx); err != nil {
			log.Error(err)
			return err
		}
	}

	// summoner matches
	for _, summonerMatchDAO := range summonerMatchDAOs {
		legacy := summonerMatchDAO.ToLegacy()
		if err := legacy.Upsert(tx); err != nil {
			log.Error(err)
			return err
		}
	}

	/* ------------------ Delete original (reverse) ------------------ */
	// summoner matches
	for _, summonerMatchDAO := range summonerMatchDAOs {
		if err := summonerMatchDAO.Delete(tx); err != nil {
			log.Error(err)
			return err
		}
	}

	// match participant perk style selections
	for _, perkStyleSelection := range participantPerkStyleSelectionDAOs {
		if err := perkStyleSelection.Delete(tx); err != nil {
			log.Error(err)
			return err
		}
	}

	// match participant perk styles
	for _, perkStyle := range participantPerkStyleDAOs {
		if err := perkStyle.Delete(tx); err != nil {
			log.Error(err)
			return err
		}
	}

	// match participant perks
	for _, perk := range participantPerkDAOs {
		if err := perk.Delete(tx); err != nil {
			log.Error(err)
			return err
		}
	}

	// match participant details
	for _, detail := range participantDetailDAOs {
		if err := detail.Delete(tx); err != nil {
			log.Error(err)
			return err
		}
	}

	// match participant
	for _, participant := range participantsDAOs {
		if err := participant.Delete(tx); err != nil {
			log.Error(err)
			return err
		}
	}

	// match team bans
	for _, ban := range teamBanDAOs {
		if err := ban.Delete(tx); err != nil {
			log.Error(err)
			return err
		}
	}

	// match teams
	for _, team := range teamDAOs {
		if err := team.Delete(tx); err != nil {
			log.Error(err)
			return err
		}
	}

	// match
	if err := match.Delete(tx); err != nil {
		log.Error(err)
		return err
	}

	return nil
}
