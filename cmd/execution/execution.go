package main

import (
	_ "github.com/lib/pq"
	prop "github.com/magiconair/properties"
	log "github.com/sirupsen/logrus"
	"msq.ai/constants"
	pgh "msq.ai/db/postgres/helper"
	"os"
)

const propertiesFileName string = "execution.properties"

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

	log.Trace("Current folder is [" + pwd + "]")

	//------------------------------------------------------------------------------------------------------------------

	properties := prop.MustLoadFile(propertiesFileName, prop.UTF8)

	for k, v := range properties.Map() {
		log.Debug("key[" + k + "] value[" + v + "]")
	}

	//------------------------------------------------------------------------------------------------------------------

	url := properties.MustGet(constants.PostgresUrlPropertyName)

	err = pgh.CheckDbUrl(url)

	if err != nil {
		log.Fatal("Cannot connect to DB with URL ["+url+"] ", err)
	}

	//------------------------------------------------------------------------------------------------------------------

	// TODO load dictionaries from DB

	//------------------------------------------------------------------------------------------------------------------

	// TODO start execution timeout validator

	//------------------------------------------------------------------------------------------------------------------

	// TODO start execution execution history notifier

	//------------------------------------------------------------------------------------------------------------------

	// TODO start REST provider

	//------------------------------------------------------------------------------------------------------------------

	// TODO perform some statistic calculation and print, send , something, ..... XZ
}
