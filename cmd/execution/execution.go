package main

import (
	_ "github.com/lib/pq"
	prop "github.com/magiconair/properties"
	log "github.com/sirupsen/logrus"
	"msq.ai/constants"
	"msq.ai/db/postgres/dao"
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

	db, err := pgh.GetDbByUrl(url)

	biMap, err := dao.LoadExchanges(db)

	if err != nil {
		log.Fatal("LEE !", err)
	}

	log.Info(biMap)

	biMap, err = dao.LoadDirections(db)

	if err != nil {
		log.Fatal("LEE !", err)
	}

	log.Info(biMap)

	biMap, err = dao.LoadOrderTypes(db)

	if err != nil {
		log.Fatal("LEE !", err)
	}

	log.Info(biMap)

	biMap, err = dao.LoadTimeInForce(db)

	if err != nil {
		log.Fatal("LEE !", err)
	}

	log.Info(biMap)

	biMap, err = dao.LoadExecutionTypes(db)

	if err != nil {
		log.Fatal("LEE !", err)
	}

	log.Info(biMap)

	biMap, err = dao.LoadExecutionStatuses(db)

	if err != nil {
		log.Fatal("LEE !", err)
	}

	log.Info(biMap)

	pgh.CloseDb(db)

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
