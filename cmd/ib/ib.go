package main

import (
	_ "github.com/lib/pq"
	prop "github.com/magiconair/properties"
	log "github.com/sirupsen/logrus"
	cord "msq.ai/connectors/coordinator"
	"msq.ai/connectors/dumper"
	"msq.ai/connectors/proto"
	"msq.ai/constants"
	"msq.ai/db/postgres/dao"
	pgh "msq.ai/db/postgres/helper"
	"msq.ai/exchange/ecib"
	"os"
	"time"
)

const propertiesFileName = "ib.properties"
const propertiesWsUrlName = "ws.url"
const connectorsExecPoolSize = 200
const dumperExecPoolSize = 10

func init() {

	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
}

func main() {

	ctxLog := log.WithFields(log.Fields{"id": "IB"})

	ctxLog.Info("IB is going to start")

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

	pgh.CloseDb(db)

	//------------------------------------ start IB connector  ----------------------------------------------------

	wsUrl := properties.MustGet(propertiesWsUrlName)

	if len(wsUrl) < 1 {
		log.Fatal("[" + propertiesWsUrlName + "] is wrong ! = [" + wsUrl + "]")
	}

	requests := make(chan *proto.ExecRequest)
	responses := make(chan *proto.ExecResponse)
	dump := make(chan *proto.ExecResponse)

	ecib.RunIbConnector(requests, responses, wsUrl)

	//----------------------------------------- start dumper ------------------------------------------------------

	dumper.RunDumper(url, dictionaries, responses, dump, dumperExecPoolSize)

	//----------------------------------------- start coordinator ------------------------------------------------------

	exchangeName := properties.MustGet(constants.ExchangeNamePropertyName)

	exchangeId := dictionaries.Exchanges().GetIdByName(exchangeName)

	if exchangeId < 1 {
		ctxLog.Fatal("Illegal Exchange name ! ", exchangeName)
	}

	connectorId := int16(properties.MustGetInt(constants.ConnectorIdPropertyName))

	if exchangeId < 1 {
		ctxLog.Fatal("Illegal connectorId ! ", connectorId)
	}

	cord.RunCoordinator(url, dictionaries, requests, dump, exchangeId, connectorId, connectorsExecPoolSize)

	//------------------------------------------------------------------------------------------------------------------

	for {
		time.Sleep(60 * time.Second)
	}

}
