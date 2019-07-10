package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	prop "github.com/magiconair/properties"
	log "github.com/sirupsen/logrus"
	"msq.ai/constants"
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

	checkDbConnection(properties)

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

func checkDbConnection(properties *prop.Properties) {

	log.Trace("Try to check DB connection ...")

	db, err := sql.Open(constants.DbName, properties.MustGet(constants.PostgresUrlPropertyName))

	if err != nil {
		log.Fatal("Cannot open DB connection !", err)
	}

	err = db.Ping()

	if err != nil {
		log.Fatal("Cannot ping DB !", err)
	}

	err = db.Close()

	if err != nil {
		log.Fatal("Cannot close DB connection!", err)
	}

	log.Trace("Successfully connected to DB")
}
