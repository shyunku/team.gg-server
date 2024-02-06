package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/schollz/progressbar/v3"
	log "github.com/shyunku-libraries/go-logger"
	"io"
	"os"
	"regexp"
	"team.gg-server/core"
	"team.gg-server/libs/http"
	"team.gg-server/util"
)

// https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/ko_kr/v1/

var (
	DataDragonVersion      = ""
	LocalDataDragonVersion = ""

	Champions      = make(map[string]ChampionDataVO)      // key: champion key
	SummonerSpells = make(map[string]SummonerSpellDataVO) // key: summoner spell key
	Perks          = make(map[int]PerkInfoVO)             // key: perk id
	PerkStyles     = make(map[int]PerkStyleInfoVO)        // key: perk style id
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
	championsInfo, err := GetChampionData()
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

func SanitizeAndLoadDataDragonFile() error {
	// check if latest data dragon is downloaded
	projectRoot := util.GetProjectRootDirectory()

	// create tmp download dir if not exists
	tmpDownloadDirPath := fmt.Sprintf("%s/datafiles/tmp", projectRoot)
	err := os.MkdirAll(tmpDownloadDirPath, os.ModePerm)
	if err != nil {
		log.Error(err)
		return err
	}

	// create data dragon download dir if not exists
	dataDragonDirPath := fmt.Sprintf("%s/datafiles/data_dragon", projectRoot)
	err = os.MkdirAll(dataDragonDirPath, os.ModePerm)
	if err != nil {
		log.Error(err)
		return err
	}

	destDataDragonPath := fmt.Sprintf("%s/%s", dataDragonDirPath, DataDragonVersion)

	targetDataDragonDirPath := fmt.Sprintf("%s/%s", dataDragonDirPath, DataDragonVersion)
	targetTarGzPath := fmt.Sprintf("%s/dragontail-%s.tgz", tmpDownloadDirPath, DataDragonVersion)
	if _, err := os.Stat(targetDataDragonDirPath); errors.Is(err, os.ErrNotExist) {
		log.Infof("latest data dragon not found (%s)", DataDragonVersion)

		// check if tar.gz file exists
		if _, err := os.Stat(targetTarGzPath); errors.Is(err, os.ErrNotExist) {
			log.Infof("latest data dragon tar.gz file not found, downloading...")

			// tar.gz file not exists, download latest data dragon
			url := fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/dragontail-%s.tgz", DataDragonVersion)
			resp, err := http.StreamGet(http.GetRequest{
				Url: url,
			})
			if err != nil {
				log.Error(err)
				return err
			}
			defer resp.Stream.Close()

			// save data dragon tar.gz file
			out, err := os.Create(targetTarGzPath)
			if err != nil {
				return err
			}
			defer out.Close()

			// progress bar
			progressBar := progressbar.DefaultBytes(
				resp.ContentLength,
				"Downloading data dragon",
			)

			// copy data dragon
			_, err = io.Copy(io.MultiWriter(out, progressBar), resp.Stream)
			if err != nil {
				return err
			}
		}

		// extract data dragon
		log.Infof("extracting data dragon tar.gz file (%s)...", targetTarGzPath)
		if err := util.UnTarGz(targetTarGzPath, destDataDragonPath); err != nil {
			return err
		}

		log.Infof("Data dragon extraction done! dest: %s", destDataDragonPath)
	} else {
		log.Infof("Data dragon is already latest: %s", DataDragonVersion)
	}

	// remove sub files in tmp download dir (except latest)
	files, err := os.ReadDir(dataDragonDirPath)
	if err != nil {
		return err
	}
	var ddragonEntries []os.DirEntry
	var latestDdragonEntry *os.DirEntry
	for _, file := range files {
		ddragonEntries = append(ddragonEntries, file)
		entryName := file.Name()
		updateLatest := func() {
			latestDdragonEntry = &file
		}

		if latestDdragonEntry == nil {
			updateLatest()
		} else {
			latestVersion, err := version.NewVersion((*latestDdragonEntry).Name())
			if err != nil {
				log.Warnf("failed to parse version (%s)", (*latestDdragonEntry).Name())
				continue
			}
			currentVersion, err := version.NewVersion(entryName)
			if err != nil {
				log.Warnf("failed to parse version (%s)", entryName)
				continue
			}

			if currentVersion.GreaterThan(latestVersion) {
				updateLatest()
			}
		}
	}
	if latestDdragonEntry != nil {
		log.Infof("latest data dragon version: %s, keep alive", (*latestDdragonEntry).Name())
		LocalDataDragonVersion = (*latestDdragonEntry).Name()
		for _, file := range ddragonEntries {
			if file == *latestDdragonEntry {
				continue
			}
			removingDirPath := fmt.Sprintf("%s/%s", dataDragonDirPath, file.Name())
			log.Debugf("removing old data dragon dir (%s)...", removingDirPath)
			if err := os.RemoveAll(removingDirPath); err != nil {
				log.Warnf("failed to remove old data dragon dir (%s)", file.Name())
			}
			log.Debugf("remove complete", file.Name())
		}
		if len(ddragonEntries) > 1 {
			log.Debugf("%d old data dragon files removed", len(ddragonEntries)-1)
		}
	} else {
		if core.UrgentMode {
			log.Warnf("latest data dragon not found")
		} else {
			return errors.New("latest data dragon not found")
		}
	}

	// remove old data dragon if exists (except latest)
	// get all data dragon files
	files, err = os.ReadDir(tmpDownloadDirPath)
	if err != nil {
		return err
	}

	// remove old data dragon tar.gz files
	var ddragonTarGzEntries []os.DirEntry
	var latestDdragonTarGzEntry *os.DirEntry
	var latestDdragonTarGzVersion string
	for _, file := range files {
		ddragonTarGzEntries = append(ddragonTarGzEntries, file)
		entryName := file.Name()

		versionRegex := regexp.MustCompile(`\d+\.\d+\.\d+`)
		entryVersion := versionRegex.FindString(entryName)
		if entryVersion == "" {
			log.Warnf("failed to parse version (%s)", entryName)
			continue
		}

		updateLatest := func() {
			latestDdragonTarGzEntry = &file
			latestDdragonTarGzVersion = entryVersion
		}

		if latestDdragonTarGzEntry == nil || latestDdragonTarGzVersion == "" {
			updateLatest()
		} else {
			latestVersion, err := version.NewVersion(latestDdragonTarGzVersion)
			if err != nil {
				log.Warnf("failed to parse version (%s)", latestDdragonTarGzVersion)
				continue
			}
			currentVersion, err := version.NewVersion(entryVersion)
			if err != nil {
				log.Warnf("failed to parse version (%s)", entryName)
				continue
			}

			if currentVersion.GreaterThan(latestVersion) {
				updateLatest()
			}
		}
	}
	if latestDdragonTarGzEntry != nil {
		log.Infof("latest data dragon tar.gz version: %s, keep alive", (*latestDdragonTarGzEntry).Name())
		for _, file := range ddragonTarGzEntries {
			if file == *latestDdragonTarGzEntry {
				continue
			}
			if err := os.RemoveAll(fmt.Sprintf("%s/%s", dataDragonDirPath, file.Name())); err != nil {
				log.Warnf("failed to remove old data dragon tar.gz (%s)", file.Name())
			}
		}
		if len(ddragonTarGzEntries) > 1 {
			log.Debugf("%d old data dragon tar.gz removed", len(ddragonTarGzEntries)-1)
		}
	}

	return nil
}

func LoadSummonerSpellsData() error {
	// load summoner spells
	var SummonerSpellsDto DDragonSummonerJsonDto
	if err := LoadDDragonKorFile(&SummonerSpellsDto, "/summoner.json"); err != nil {
		return err
	}

	// save to memory
	SummonerSpells = map[string]SummonerSpellDataVO{}
	for _, summonerSpell := range SummonerSpellsDto.Data {
		SummonerSpells[summonerSpell.Key] = summonerSpell
	}

	log.Debugf("%d summoner spells loaded", len(SummonerSpells))
	return nil
}

func GetChampionData() (*ChampionsInfoDto, error) {
	resp, err := http.Get(http.GetRequest{
		Url: fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/data/ko_KR/champion.json", DataDragonVersion),
	})
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, resp.Err
	}

	var champions ChampionsInfoDto
	if err := json.Unmarshal(resp.Body, &champions); err != nil {
		return nil, err
	}

	return &champions, nil
}

func GetLatestDataDragonVersion() (string, error) {
	resp, err := http.Get(http.GetRequest{
		Url: "https://ddragon.leagueoflegends.com/api/versions.json",
	})
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", resp.Err
	}

	var versions []string
	if err := json.Unmarshal(resp.Body, &versions); err != nil {
		return "", err
	}

	return versions[0], nil
}

func GetCDragonPerksData() (PerksInfoDto, error) {
	path := "https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/ko_kr/v1/perks.json"
	resp, err := http.Get(http.GetRequest{
		Url: path,
	})
	if err != nil {
		return nil, err
	}

	var perksData PerksInfoDto
	if err := json.Unmarshal(resp.Body, &perksData); err != nil {
		return nil, err
	}

	return perksData, nil
}

func GetCDragonPerkStylesData() (*PerkStylesInfoDto, error) {
	path := "https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/ko_kr/v1/perkstyles.json"
	resp, err := http.Get(http.GetRequest{
		Url: path,
	})
	if err != nil {
		return nil, err
	}

	var perkStylesData PerkStylesInfoDto
	if err := json.Unmarshal(resp.Body, &perkStylesData); err != nil {
		return nil, err
	}

	return &perkStylesData, nil
}
