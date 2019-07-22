package coordinator

import (
	log "github.com/sirupsen/logrus"
	"msq.ai/connectors/proto"
	dic "msq.ai/db/postgres/dictionaries"
	pgh "msq.ai/db/postgres/helper"
	"time"
)

func RunCoordinator(dburl string, dictionaries *dic.Dictionaries, out chan<- *proto.ExecRequest, in <-chan *proto.ExecResponse, exchangeId int16, connectorId int16) {

	ctxLog := log.WithFields(log.Fields{"id": "Coordinator"})

	ctxLog.Info("Coordinator is going to start")

	if len(dburl) < 1 {
		ctxLog.Fatal("dburl is empty !")
	}

	if in == nil {
		ctxLog.Fatal("ExecRequest channel is nil !")
	}

	if out == nil {
		ctxLog.Fatal("ExecResponse channel is nil !")
	}

	//------------------------------------------------------------------------------------------------------------------

	//var pingTime = 30 * time.Second
	//var prevOpTime time.Time
	//var id int64 = 0
	dump := make(chan *proto.ExecResponse, 10)

	//incId := func() {
	//	id = id + 1
	//}

	//------------------------------------------------------------------------------------------------------------------

	//pingExchange := func() {
	//
	//	for {
	//
	//		incId()
	//
	//		request := &proto.ExecRequest{Id: id, What: proto.ExecRequestCheckConnection}
	//
	//		out <- request
	//
	//		response := <-in
	//
	//		if response == nil {
	//			ctxLog.Fatal("Protocol violation! ExecResponse is nil")
	//		}
	//
	//		if response.Id != request.Id {
	//			ctxLog.Fatal("Protocol violation! response.Id doesn't equal request.Id")
	//		}
	//
	//		if response.Status == proto.ExecResponseStatusOk {
	//			prevOpTime = time.Now()
	//			log.Info("Exchange successfully pinged")
	//			return
	//		} else if response.Status == proto.ExecResponseStatusError {
	//			log.Info("Exchange ping error ! ", response.Description)
	//		} else {
	//			ctxLog.Fatal("Protocol violation! Ping response has unknown status")
	//		}
	//
	//		time.Sleep(5 * time.Second)
	//	}
	//}

	//------------------------------------------------------------------------------------------------------------------

	//tradeExchange := func(raw *cmd.RawCommand) {
	//
	//	incId()
	//
	//	request := &proto.ExecRequest{Id: id, What: proto.ExecRequestTrade, Cmd: raw}
	//
	//	out <- request
	//
	//	response := <-in
	//
	//	if response == nil {
	//		ctxLog.Fatal("Protocol violation! ExecResponse is nil")
	//		return
	//	}
	//
	//	if response.Id != request.Id {
	//		ctxLog.Fatal("Protocol violation! response.Id doesn't equal request.Id")
	//	}
	//
	//	if response.Status == proto.ExecResponseStatusOk {
	//		prevOpTime = time.Now()
	//		log.Trace("Trade operation successfully finished")
	//	} else if response.Status == proto.ExecResponseStatusError {
	//		log.Info("Exchange Trade error ! ", response.Description)
	//	} else {
	//		ctxLog.Fatal("Protocol violation! Trade response has unknown status")
	//	}
	//
	//	dump <- response
	//}

	go func() {

		db, err := pgh.GetDbByUrl(dburl)

		if err != nil {
			ctxLog.Fatal("Cannot connect to DB with URL ["+dburl+"] ", err)
		}

		db.SetMaxIdleConns(1)
		db.SetMaxOpenConns(3)
		db.SetConnMaxLifetime(time.Hour)

		for {

			response := <-dump

			ctxLog.Trace("Dump execution result to DB ", response)

			// TODO !!!!!!!!!!!!!!!!!!
		}

	}()

	//------------------------------------------------------------------------------------------------------------------

	go func() {

		//future := 100 * time.Millisecond

		db, err := pgh.GetDbByUrl(dburl)

		if err != nil {
			ctxLog.Fatal("Cannot connect to DB with URL ["+dburl+"] ", err)
		}

		db.SetMaxIdleConns(1)
		db.SetMaxOpenConns(3)
		db.SetConnMaxLifetime(time.Hour)

		//dbTryGetCommandForExecution := func() *cmd.Command {
		//
		//	statusCreatedId := dictionaries.ExecutionStatuses().GetIdByName(constants.ExecutionStatusCreatedName)
		//	statusExecutingId := dictionaries.ExecutionStatuses().GetIdByName(constants.ExecutionStatusExecutingName)
		//
		//	result, err := dao.TryGetCommandForExecution(db, exchangeId, connectorId, time.Now().Add(future), statusCreatedId, statusExecutingId)
		//
		//	if err != nil {
		//		ctxLog.Error("dbTryGetCommandForExecution error ! ", err)
		//		time.Sleep(5 * time.Second)
		//		return nil
		//	}
		//
		//	return result
		//}

		//var command *cmd.Command
		//var raw *cmd.RawCommand

		// TODO restore state lost operations

		for {

			time.Sleep(250 * time.Millisecond)

			//command = dbTryGetCommandForExecution()
			//
			//if command != nil {
			//
			//	raw = cmd.ToRaw(command, dictionaries)
			//
			//	ctxLog.Info("Just got new command for execution", raw)
			//
			//	//tradeExchange(raw)
			//
			//} else {
			//
			//	delta := time.Now().Sub(prevOpTime)
			//
			//	if delta > pingTime {
			//		//pingExchange()
			//	}
			//
			//	time.Sleep(250 * time.Millisecond)
			//}
		}

	}()

}
