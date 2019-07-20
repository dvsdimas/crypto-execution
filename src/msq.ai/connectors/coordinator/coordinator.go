package coordinator

import (
	"database/sql"
	log "github.com/sirupsen/logrus"
	"msq.ai/connectors/proto"
	"time"
)

func RunCoordinator(db *sql.DB, out chan<- *proto.ExecRequest, in <-chan *proto.ExecResponse, exchangeId int16) {

	ctxLog := log.WithFields(log.Fields{"id": "Coordinator"})

	ctxLog.Info("Coordinator is going to start")

	if db == nil {
		ctxLog.Fatal("db is nil !")
	}

	if in == nil {
		ctxLog.Fatal("ExecRequest channel is nil !")
	}

	if out == nil {
		ctxLog.Fatal("ExecResponse channel is nil !")
	}

	//------------------------------------------------------------------------------------------------------------------

	var pingTime = 30 * time.Second
	var prevOpTime time.Time
	var request *proto.ExecRequest
	var response *proto.ExecResponse
	var id int64 = 0

	// TODO configure DB pool

	incId := func() {
		id = id + 1
	}

	//------------------------------------------------------------------------------------------------------------------

	pingExchange := func() {

		for {

			incId()

			request = &proto.ExecRequest{Id: id, What: proto.ExecRequestCheckConnection}

			out <- request

			response = <-in

			if response == nil {
				ctxLog.Fatal("Protocol violation! ExecResponse is nil")
			}

			if response.Id != request.Id {
				ctxLog.Fatal("Protocol violation! response.Id doesn't equal request.Id")
			}

			if response.Status == proto.ExecResponseStatusOk {
				prevOpTime = time.Now()
				log.Info("Exchange successfully pinged")
				return
			} else if response.Status == proto.ExecResponseStatusError {
				log.Info("Exchange ping error ! ", response.Description)
			} else {
				ctxLog.Fatal("Protocol violation! Ping response has unknown status")
			}

			time.Sleep(5 * time.Second)
		}
	}

	//------------------------------------------------------------------------------------------------------------------

	go func() {

		// TODO restore state lost operations

		for {

			// TODO try get from DB

			// TODO execute

			// TODO save result to DB

			delta := time.Now().Sub(prevOpTime)

			if delta > pingTime {
				pingExchange()
			}

			time.Sleep(100 * time.Millisecond)
		}

	}()

}
