package core

import (
	log "github.com/shyunku-libraries/go-logger"
	"os"
	core "team.gg-server/util"
)

var (
	AppServerHost string
	AppServerPort = os.Getenv("APP_SERVER_PORT")
	DebugMode     = false
	DebugOnProd   = false
	UrgentMode    = false

	RsoClientId          = os.Getenv("RSO_CLIENT_ID")
	RsoClientSecret      = os.Getenv("RSO_CLIENT_SECRET")
	RsoClientCallbackUri = os.Getenv("RSO_CLIENT_CALLBACK_URI")
)

func Preload() error {
	log.Debugf("preload started...")

	// load public ip
	ipv4, err := core.GetPublicIp()
	if err != nil {
		return err
	}

	// load debug mode
	DebugMode = os.Getenv("DEBUG") == "true"

	// load debug on prod
	DebugOnProd = os.Getenv("DEBUG_ON_PRODUCTION") == "true"
	if DebugMode {
		DebugOnProd = true
	}

	// load urgent mode
	UrgentMode = os.Getenv("URGENT") == "true"

	AppServerHost = ipv4
	AppServerPort = os.Getenv("APP_SERVER_PORT")
	log.Debugf("server is active on public ip: %s:%s", AppServerHost, AppServerPort)
	return nil
}
