package dumper

import (
	log "github.com/sirupsen/logrus"
	"msq.ai/connectors/proto"
	dic "msq.ai/db/postgres/dictionaries"
	pgh "msq.ai/db/postgres/helper"
	"time"
)

func RunDumper(dburl string, dictionaries *dic.Dictionaries, in <-chan *proto.ExecResponse, out chan<- *proto.ExecResponse) {

	ctxLog := log.WithFields(log.Fields{"id": "Dumper"})

	ctxLog.Info("Dumper is going to start")

	if len(dburl) < 1 {
		ctxLog.Fatal("dburl is empty !")
	}

	if dictionaries == nil {
		ctxLog.Fatal("dictionaries is nil !")
	}

	if in == nil {
		ctxLog.Fatal("ExecResponse channel is nil !")
	}

	if out == nil {
		ctxLog.Fatal("ExecResponse channel is nil !")
	}

	//------------------------------------------------------------------------------------------------------------------

	db, err := pgh.GetDbByUrl(dburl)

	if err != nil {
		ctxLog.Fatal("Cannot connect to DB with URL ["+dburl+"] ", err)
	}

	db.SetMaxIdleConns(1) // TODO
	db.SetMaxOpenConns(3)
	db.SetConnMaxLifetime(time.Hour)

	//------------------------------------------------------------------------------------------------------------------

	go func() {

		for {
			response := <-in

			ctxLog.Trace("Dumped execution result to DB", response)

			// TODO dump

			out <- response
		}
	}()

}
