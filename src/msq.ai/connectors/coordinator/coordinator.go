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

func RunCoordinator(dburl string, dictionaries *dic.Dictionaries, out chan<- *proto.ExecRequest, in <-chan *proto.ExecResponse,
	exchangeId int16, connectorId int16, connectorExecPoolSize uint32) {

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

	makeExecRequest := func(command *cmd.Command, dic *dic.Dictionaries, eType proto.ExecType) *proto.ExecRequest {

		raw := cmd.ToRaw(command, dictionaries)

		var et = eType

		if raw.OrderType == constants.OrderTypeInfoName {
			et = proto.InfoCmd
		}

		return &proto.ExecRequest{What: et, RawCmd: raw, Cmd: command}
	}

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

			// TODO send to notification module
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

		dbTryGetCommandForRecovery := func() *cmd.Command {

			statusExecutingId := dictionaries.ExecutionStatuses().GetIdByName(constants.ExecutionStatusExecutingName)

			result, err := dao.TryGetCommandForRecovery(db, exchangeId, connectorId, statusExecutingId)

			if err != nil {
				logErrWithST("TryGetCommandForRecovery error ! ", err)
				time.Sleep(5 * time.Second)
				return nil
			}

			return result
		}

		dbTryGetCommandForExecution := func() *cmd.Command {

			statusCreatedId := dictionaries.ExecutionStatuses().GetIdByName(constants.ExecutionStatusCreatedName)
			statusExecutingId := dictionaries.ExecutionStatuses().GetIdByName(constants.ExecutionStatusExecutingName)

			result, err := dao.TryGetCommandForExecution(db, exchangeId, connectorId, time.Now().Add(future), statusCreatedId, statusExecutingId, 1)

			if err != nil {
				logErrWithST("dbTryGetCommandForExecution error ! ", err)
				time.Sleep(5 * time.Second)
				return nil
			}

			return result
		}

		ctxLog.Info("Start recovery procedure")

		for {

			forRecovery := dbTryGetCommandForRecovery() // TODO use start time !!!!!

			if forRecovery == nil {
				break
			}

			ctxLog.Trace("Has command for recovery ", forRecovery)

			for {

				s := atomic.LoadUint32(&sending)

				if s == 0 { // TODO fix with start time !!!!!

					atomic.AddUint32(&sending, 1)

					out <- makeExecRequest(forRecovery, dictionaries, proto.CheckCmd)

					for atomic.LoadUint32(&sending) != 0 {
						time.Sleep(100 * time.Millisecond)
					}

					break
				}

				time.Sleep(100 * time.Millisecond)
			}
		}

		ctxLog.Info("Recovery procedure finished")

		var command *cmd.Command
		var raw *cmd.RawCommand

		for {

			s := atomic.LoadUint32(&sending)

			if s <= connectorExecPoolSize {

				command = dbTryGetCommandForExecution()

				if command != nil {

					raw = cmd.ToRaw(command, dictionaries)

					ctxLog.Trace("New command for execution", raw)

					atomic.AddUint32(&sending, 1)

					out <- makeExecRequest(command, dictionaries, proto.ExecuteCmd)

					continue
				}
			}

			time.Sleep(100 * time.Millisecond)
		}

	}()

}
