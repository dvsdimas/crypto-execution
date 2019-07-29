package connector

import (
	log "github.com/sirupsen/logrus"
	"msq.ai/connectors/proto"
	"sync"
	"sync/atomic"
)

func RunConnector(ctxLog *log.Entry, in <-chan *proto.ExecRequest, out chan<- *proto.ExecResponse, execPoolSize int,
	trade func(request *proto.ExecRequest) *proto.ExecResponse) {

	ctxLog.Info(" going to start")

	if in == nil {
		ctxLog.Fatal("ExecRequest channel is nil !")
	}

	if out == nil {
		ctxLog.Fatal("ExecResponse channel is nil !")
	}

	if execPoolSize < 1 {
		ctxLog.Fatal("execPoolSize less than 1")
	}

	if execPoolSize > 10000 {
		ctxLog.Fatal("execPoolSize more than 10000 !")
	}

	tradeInternal := func(request *proto.ExecRequest) *proto.ExecResponse {

		// TODO check time !!!!

		return trade(request)
	}

	//------------------------------------------------------------------------------------------------------------------

	var lockOut = &sync.Mutex{}

	performFunction := func(in <-chan *proto.ExecRequest) {

		for {
			request := <-in

			ctxLog.Trace("Start sending cmd to Binance .....", request)

			var response *proto.ExecResponse

			if request.What == proto.ExecuteCmd {
				response = tradeInternal(request)
			} else {
				ctxLog.Fatal("Unexpected ExecType", request)
			}

			ctxLog.Trace("Sent cmd to Binance !")

			lockOut.Lock()
			out <- response
			lockOut.Unlock()
		}
	}

	inChannels := make([]chan *proto.ExecRequest, execPoolSize)

	for i := 0; i < execPoolSize; i++ {

		inChannels[i] = make(chan *proto.ExecRequest)

		go performFunction(inChannels[i])
	}

	//------------------------------------------------------------------------------------------------------------------

	var pointer int32 = 0

	getNextChannel := func() chan<- *proto.ExecRequest {

		p := int(atomic.LoadInt32(&pointer))

		c := inChannels[p]

		if p+1 == execPoolSize {
			atomic.StoreInt32(&pointer, 0)
		} else {
			atomic.AddInt32(&pointer, 1)
		}

		return c
	}

	sendToExec := func(request *proto.ExecRequest) {
		for {
			sc := getNextChannel()

			select {
			case sc <- request:
				return
			default:
				continue
			}
		}
	}

	go func() {
		for {

			request := <-in

			if request == nil {
				ctxLog.Fatal("Protocol violation! ExecRequest is nil")
			}

			ctxLog.Trace("Request for sending ", request)

			sendToExec(request)
		}
	}()

}
