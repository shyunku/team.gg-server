package service

import (
	log "github.com/shyunku-libraries/go-logger"
	"os"
	"os/signal"
	"syscall"
)

func PrepareFinalize() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)

	go receiveSignal(sigChan)
}

func receiveSignal(sigChan chan os.Signal) {
	for {
		sig := <-sigChan
		switch sig {
		case syscall.SIGTERM:
			log.Info("SIGTERM signal received in goroutine, shutting down.")
			// finalize (resource release, etc.)
			return
		default:
			log.Warn("Received unexpected signal:", sig)
		}
	}
}
