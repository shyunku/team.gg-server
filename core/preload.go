package core

import (
	log "github.com/shyunku-libraries/go-logger"
	"os"
	"team.gg-server/service"
	core "team.gg-server/util"
)

var (
	AppServerHost string
	AppServerPort = os.Getenv("APP_SERVER_PORT")
	DebugMode     = os.Getenv("DEBUG") == "true"

	DataDragonVersion = ""
)

func Preload() error {
	log.Debugf("Preload started...")

	// load public ip
	ipv4, err := core.GetPublicIp()
	if err != nil {
		return err
	}

	DataDragonVersion, err = service.GetLatestDataDragonVersion()
	if err != nil {
		return err
	}
	log.Debugf("DataDragon version: %s", DataDragonVersion)

	AppServerHost = ipv4
	AppServerPort = os.Getenv("APP_SERVER_PORT")
	log.Debugf("server is running on public ip: %s:%s", AppServerHost, AppServerPort)
	return nil
}
