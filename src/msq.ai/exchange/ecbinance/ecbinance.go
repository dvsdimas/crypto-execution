package ecbinance

import (
	log "github.com/sirupsen/logrus"
	"msq.ai/connectors/proto"
	"sync"
	"sync/atomic"
)

func RunBinanceConnector(in <-chan *proto.ExecRequest, out chan<- *proto.ExecResponse, execPoolSize int) {

	ctxLog := log.WithFields(log.Fields{"id": "BinanceConnector"})

	ctxLog.Info("BinanceConnector is going to start")

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

	//------------------------------------------------------------------------------------------------------------------

	var lockOut = &sync.Mutex{}

	performFunction := func(in <-chan *proto.ExecRequest) {

		for {
			request := <-in

			ctxLog.Trace("Sending cmd to Binance .....", request)

			// TODO

			ctxLog.Trace("Sent cmd to Binance !!!!!!!!!!!")

			response := &proto.ExecResponse{OriginCmd: request.Cmd, Status: proto.StatusOk}

			lockOut.Lock()
			out <- response
			lockOut.Unlock()
		}
	}

	//------------------------------------------------------------------------------------------------------------------

	inChannels := make([]chan *proto.ExecRequest, execPoolSize)

	for i := 0; i < execPoolSize; i++ {

		inChannels[i] = make(chan *proto.ExecRequest)

		go performFunction(inChannels[i])
	}

	//client := binance.NewClient(apiKey, secretKey)

	//------------------------------------------------------------------------------------------------------------------

	//trade := func(request *proto.ExecRequest) {
	//
	//	var response = proto.ExecResponse{Id: request.Id}
	//
	//	c := request.Cmd
	//
	//	if c == nil {
	//		ctxLog.Fatal("Protocol violation! ExecRequest Trade with empty cmd ! ", request)
	//		return
	//	}
	//
	//	orderService := client.NewCreateOrderService().Symbol(request.Cmd.Instrument)
	//
	//	if c.OrderType == constants.OrderTypeMarketName {
	//		orderService = orderService.Type(binance.OrderTypeMarket)
	//	} else if c.Direction == constants.OrderTypeLimitName {
	//		orderService = orderService.Type(binance.OrderTypeLimit)
	//		orderService = orderService.Price(c.LimitPrice)
	//	} else {
	//		ctxLog.Fatal("Protocol violation! ExecRequest wrong OrderType with empty cmd ! ", request)
	//		return
	//	}
	//
	//	if c.Direction == constants.OrderDirectionBuyName {
	//		orderService = orderService.Side(binance.SideTypeBuy)
	//	} else if c.Direction == constants.OrderDirectionSellName {
	//		orderService = orderService.Side(binance.SideTypeSell)
	//	} else {
	//		ctxLog.Fatal("Protocol violation! ExecRequest wrong Direction with empty cmd ! ", request)
	//		return
	//	}
	//
	//	if c.TimeInForce == constants.TimeInForceFokName { // TODO add GTC
	//		//orderService = orderService.TimeInForce(binance.TimeInForceFOK)
	//	} else {
	//		ctxLog.Fatal("Protocol violation! ExecRequest wrong Direction with empty cmd ! ", request)
	//		return
	//	}
	//
	//	orderService = orderService.Quantity(c.Amount)
	//
	//	order, err := orderService.Do(context.Background())
	//
	//	ctxLog.Debug("Order from Binance ", order)
	//
	//	if err != nil {
	//		ctxLog.Error("Trade error ", err)
	//		response.Description = err.Error()
	//		response.Status = proto.ExecResponseStatusError
	//	} else {
	//		//for _, p := range prices {
	//		//	log.Trace(p)
	//		//}
	//
	//		response.Status = proto.ExecResponseStatusOk
	//	}
	//
	//	out <- &response
	//}

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

	go func() {

		var request *proto.ExecRequest

		for {

			request = <-in

			if request == nil {
				ctxLog.Fatal("Protocol violation! ExecRequest is nil")
			}

			ctxLog.Trace("Request for sending to Binance", request)

			getNextChannel() <- request
		}
	}()

}
