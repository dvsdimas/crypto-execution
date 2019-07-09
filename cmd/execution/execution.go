package main

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {

	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
}

func main() {

	log.Info("Execution is going to start")

	pwd, err := os.Getwd()

	if err != nil {
		log.Fatal("Getwd error", err)
	}

	log.Debug("Current folder is [" + pwd + "]")

}
