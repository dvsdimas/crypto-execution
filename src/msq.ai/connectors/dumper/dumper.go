package dumper

import (
	log "github.com/sirupsen/logrus"
	"msq.ai/connectors/proto"
	dic "msq.ai/db/postgres/dictionaries"
	pgh "msq.ai/db/postgres/helper"
	"sync"
	"sync/atomic"
	"time"
)

func RunDumper(dburl string, dictionaries *dic.Dictionaries, in <-chan *proto.ExecResponse, out chan<- *proto.ExecResponse, execPoolSize int) {

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

	db.SetMaxIdleConns(execPoolSize + 1)
	db.SetMaxOpenConns(execPoolSize + 1)
	db.SetConnMaxLifetime(time.Hour)

	//------------------------------------------------------------------------------------------------------------------

	var lockOut = &sync.Mutex{}

	performFunction := func(in <-chan *proto.ExecResponse) {

		for {
			response := <-in

			ctxLog.Trace("Dumping response", response)

			if response.Status == proto.StatusOk {
				ctxLog.Trace("Dumping order", response.Order)
			}

			// TODO

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
