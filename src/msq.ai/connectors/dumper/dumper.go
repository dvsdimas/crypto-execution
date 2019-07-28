package dumper

import (
	"fmt"
	"github.com/go-errors/errors"
	log "github.com/sirupsen/logrus"
	"msq.ai/connectors/proto"
	"msq.ai/constants"
	"msq.ai/db/postgres/dao"
	dic "msq.ai/db/postgres/dictionaries"
	pgh "msq.ai/db/postgres/helper"
	"sync"
	"sync/atomic"
	"time"
)

func RunDumper(dburl string, dictionaries *dic.Dictionaries, in <-chan *proto.ExecResponse, out chan<- *proto.ExecResponse, execPoolSize int) {

	ctxLog := log.WithFields(log.Fields{"id": "Dumper"})

	ctxLog.Info("Dumper is going to start")

	logErrWithST := func(msg string, err error) {
		ctxLog.WithField("stacktrace", fmt.Sprintf("%+v", err.(*errors.Error).ErrorStack())).Error(msg)
	}

	if len(dburl) < 1 {
		ctxLog.Fatal("dburl is empty !")
		return
	}

	if dictionaries == nil {
		ctxLog.Fatal("dictionaries is nil !")
		return
	}

	if in == nil {
		ctxLog.Fatal("ExecResponse channel is nil !")
		return
	}

	if out == nil {
		ctxLog.Fatal("ExecResponse channel is nil !")
		return
	}

	//------------------------------------------------------------------------------------------------------------------

	db, err := pgh.GetDbByUrl(dburl)

	if err != nil {
		ctxLog.Fatal("Cannot connect to DB with URL ["+dburl+"] ", err)
		return
	}

	db.SetMaxIdleConns(execPoolSize + 1)
	db.SetMaxOpenConns(execPoolSize + 1)
	db.SetConnMaxLifetime(time.Hour)

	//------------------------------------------------------------------------------------------------------------------

	var lockOut = &sync.Mutex{}

	statusExecutingId := dictionaries.ExecutionStatuses().GetIdByName(constants.ExecutionStatusExecutingName)
	statusErrorId := dictionaries.ExecutionStatuses().GetIdByName(constants.ExecutionStatusErrorName)
	statusCompletedId := dictionaries.ExecutionStatuses().GetIdByName(constants.ExecutionStatusCompletedName)

	dumpResponse := func(response *proto.ExecResponse) error {

		ctxLog.Trace("Dumping response", response)

		if response.Request.What == proto.ExecuteCmd {

			ctxLog.Trace("Dumping response", response)

			if response.Status == proto.StatusOk {

				ctxLog.Trace("Dumping order", response.Order)

				return dao.FinishExecution(db, response.Request.Cmd.Id, int16(response.Request.Cmd.ConnectorId),
					statusExecutingId, statusCompletedId, response.Description) // TODO add order

			} else if response.Status == proto.StatusError {

				return dao.FinishExecution(db, response.Request.Cmd.Id, int16(response.Request.Cmd.ConnectorId),
					statusExecutingId, statusErrorId, response.Description)

			} else {
				ctxLog.Fatal("Protocol violation! Illegal response status", response)
			}

		} else if response.Request.What == proto.CheckCmd {

			ctxLog.Fatal("Check CMD !!!!!!!!!!!", response) // TODO

		} else {
			ctxLog.Fatal("Protocol violation! Illegal request What !!!!", response, response.Request)
		}

		return nil
	}

	performFunction := func(in <-chan *proto.ExecResponse) {

		var err error

		for {
			response := <-in

			for {
				err = dumpResponse(response)

				if err == nil {
					break
				}

				logErrWithST("Cannot save response error", err)

				time.Sleep(5 * time.Second)
			}

			lockOut.Lock()
			out <- response
			lockOut.Unlock()
		}
	}

	inChannels := make([]chan *proto.ExecResponse, execPoolSize)

	for i := 0; i < execPoolSize; i++ {

		inChannels[i] = make(chan *proto.ExecResponse)

		go performFunction(inChannels[i])
	}

	//------------------------------------------------------------------------------------------------------------------

	var pointer int32 = 0

	getNextChannel := func() chan<- *proto.ExecResponse {

		p := int(atomic.LoadInt32(&pointer))

		c := inChannels[p]

		if p+1 == execPoolSize {
			atomic.StoreInt32(&pointer, 0)
		} else {
			atomic.AddInt32(&pointer, 1)
		}

		return c
	}

	go func() {

		for {

			response := <-in

			if response == nil {
				ctxLog.Fatal("Protocol violation! ExecRequest is nil")
			}

			getNextChannel() <- response
		}
	}()

}
