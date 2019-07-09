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

	log.Info("execution is going to start")

}
