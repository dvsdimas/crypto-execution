package main

import (
	_ "github.com/lib/pq"
	prop "github.com/magiconair/properties"
	log "github.com/sirupsen/logrus"
	"msq.ai/constants"
	"msq.ai/db/postgres/dao"
	pgh "msq.ai/db/postgres/helper"
	"msq.ai/rest/gin"
	"os"
	"time"
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

	ctxLog := log.WithFields(log.Fields{"id": "Execution"})

	ctxLog.Info("Execution is going to start")

	pwd, err := os.Getwd()

	if err != nil {
		ctxLog.Fatal("Getwd error", err)
	}

	ctxLog.Trace("Current folder is [" + pwd + "]")

	//------------------------------------------------------------------------------------------------------------------

	properties := prop.MustLoadFile(propertiesFileName, prop.UTF8)

	for k, v := range properties.Map() {
		ctxLog.Debug("key[" + k + "] value[" + v + "]")
	}

	//------------------------------------------------------------------------------------------------------------------

	url := properties.MustGet(constants.PostgresUrlPropertyName)

	db, err := pgh.GetDbByUrl(url)

	if err != nil {
		ctxLog.Fatal("Cannot connect to DB with URL ["+url+"] ", err)
	}

	//-------------------------------------- load dictionaries from DB -------------------------------------------------

	dictionaries, err := dao.LoadDictionaries(db)

	if err != nil {
		ctxLog.Fatal("Cannot load Dictionaries from DB with URL ["+url+"] ", err)
	}

	ctxLog.Trace(dictionaries)

	pgh.CloseDb(db)

	//------------------------------------------------------------------------------------------------------------------

	// TODO start execution timeout validator

	//------------------------------------------------------------------------------------------------------------------

	// TODO start execution execution history notifier

	//----------------------------------------- start REST provider ----------------------------------------------------

	timeForExecution := properties.GetInt(constants.CommandTimeForExecutionSecondsPropertyName, 60)

	//[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
	//- using env:	export GIN_MODE=release
	//- using code:	gin.SetMode(gin.ReleaseMode)

	gin.RunGinRestService(url, dictionaries, timeForExecution)

	//------------------------------------------------------------------------------------------------------------------

	// TODO perform some statistic calculation and print, send , something, ..... XZ

	for {
		time.Sleep(1 * time.Second)
	}
}
