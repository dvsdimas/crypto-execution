package coordinator

import (
	"fmt"
	"github.com/go-errors/errors"
	log "github.com/sirupsen/logrus"
	"msq.ai/connectors/proto"
	"msq.ai/constants"
	"msq.ai/data/cmd"
	"msq.ai/db/postgres/dao"
	dic "msq.ai/db/postgres/dictionaries"
	pgh "msq.ai/db/postgres/helper"
	"sync/atomic"
	"time"
)

const limit = 1

func RunCoordinator(dburl string, dictionaries *dic.Dictionaries, out chan<- *proto.ExecRequest, in <-chan *proto.ExecResponse, exchangeId int16, connectorId int16) {

	ctxLog := log.WithFields(log.Fields{"id": "Coordinator"})

	ctxLog.Info("Coordinator is going to start")

	if len(dburl) < 1 {
		ctxLog.Fatal("dburl is empty !")
	}

	if out == nil {
		ctxLog.Fatal("ExecResponse channel is nil !")
	}

	logErrWithST := func(msg string, err error) {
		ctxLog.WithField("stacktrace", fmt.Sprintf("%+v", err.(*errors.Error).ErrorStack())).Error(msg)
	}

	var sending uint32 = 0

	//------------------------------------------------------------------------------------------------------------------

	go func() {

		db, err := pgh.GetDbByUrl(dburl)

		if err != nil {
			ctxLog.Fatal("Cannot connect to DB with URL ["+dburl+"] ", err)
		}

		db.SetMaxIdleConns(1)
		db.SetMaxOpenConns(3)
		db.SetConnMaxLifetime(time.Hour)

		for {
			response := <-in

			atomic.AddUint32(&sending, ^uint32(0))

			ctxLog.Trace("Finished execution", response)
		}
	}()

	//------------------------------------------------------------------------------------------------------------------

	go func() {

		future := 50 * time.Millisecond

		db, err := pgh.GetDbByUrl(dburl)

		if err != nil {
			ctxLog.Fatal("Cannot connect to DB with URL ["+dburl+"] ", err)
		}

		db.SetMaxIdleConns(1)
		db.SetMaxOpenConns(3)
		db.SetConnMaxLifetime(time.Hour)

		dbTryGetCommandForExecution := func() *cmd.Command {

			statusCreatedId := dictionaries.ExecutionStatuses().GetIdByName(constants.ExecutionStatusCreatedName)
			statusExecutingId := dictionaries.ExecutionStatuses().GetIdByName(constants.ExecutionStatusExecutingName)

			result, err := dao.TryGetCommandForExecution(db, exchangeId, connectorId, time.Now().Add(future), statusCreatedId, statusExecutingId, limit)

			if err != nil {
				logErrWithST("dbTryGetCommandForExecution error ! ", err)
				time.Sleep(5 * time.Second)
				return nil
			}

			return result
		}

		var command *cmd.Command
		var raw *cmd.RawCommand

		// TODO restore state lost operations

		for {

			s := atomic.LoadUint32(&sending)

			if s == 0 {

				command = dbTryGetCommandForExecution()

				if command != nil {

					raw = cmd.ToRaw(command, dictionaries)

					ctxLog.Trace("New command for execution", raw)

					atomic.AddUint32(&sending, 1)

					out <- &proto.ExecRequest{What: proto.ExecuteCmd, RawCmd: raw, Cmd: command}

					continue
				}
			}

			time.Sleep(250 * time.Millisecond)
		}

	}()

}
