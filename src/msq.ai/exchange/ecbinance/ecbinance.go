package ecbinance

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance"
	log "github.com/sirupsen/logrus"
	"msq.ai/connectors/proto"
	"msq.ai/constants"
	"msq.ai/data/cmd"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const filledValue = "FILLED"

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

	orderToString := func(order *binance.CreateOrderResponse) string {

		if order == nil {
			return "nil"
		}

		var fill = ""

		if order.Fills != nil && order.Fills[0] != nil {
			fill = fmt.Sprintf("%+v", order.Fills[0])
		}

		return fmt.Sprintf("%+v %s", order, fill)
	}

	errorResponse := func(response *proto.ExecResponse, err error) *proto.ExecResponse {

		response.Status = proto.StatusError

		response.Description = response.Description + " Parse error [" + err.Error() + "]"

		return response
	}

	trade := func(request *proto.ExecRequest) *proto.ExecResponse {

		if request == nil {
			ctxLog.Fatal("Protocol violation! nil ExecRequest")
			return nil
		}

		c := request.RawCmd

		if c == nil {
			ctxLog.Fatal("Protocol violation! ExecRequest Trade with empty cmd ! ", request)
			return nil
		}

		var response = proto.ExecResponse{Request: request}

		client := binance.NewClient(c.ApiKey, c.SecretKey)

		orderService := client.NewCreateOrderService().Symbol(c.Instrument)
		orderService = orderService.NewClientOrderID(c.Id)

		if c.OrderType == constants.OrderTypeMarketName {
			orderService = orderService.Type(binance.OrderTypeMarket)
		} else if c.Direction == constants.OrderTypeLimitName {
			orderService = orderService.Type(binance.OrderTypeLimit)
			orderService = orderService.Price(c.LimitPrice)
		} else {
			ctxLog.Fatal("Protocol violation! ExecRequest wrong OrderType with empty cmd ! ", request)
			return nil
		}

		if c.Direction == constants.OrderDirectionBuyName {
			orderService = orderService.Side(binance.SideTypeBuy)
		} else if c.Direction == constants.OrderDirectionSellName {
			orderService = orderService.Side(binance.SideTypeSell)
		} else {
			ctxLog.Fatal("Protocol violation! ExecRequest wrong Direction with empty cmd ! ", request)
			return nil
		}

		if c.TimeInForce == constants.TimeInForceGtcName {
			// orderService = orderService.TimeInForce(binance.TimeInForceGTC)
		} else {
			msg := "Protocol violation! ExecRequest has wrong TimeInForce. Binance supported only GTC !"
			ctxLog.Error(msg, request)
			response.Description = msg
			response.Status = proto.StatusError
			return &response
		}

		orderService = orderService.Quantity(c.Amount)

		start := time.Now()

		order, err := orderService.Do(context.Background())

		response.OutsideExecution = time.Now().Sub(start)

		if err != nil {
			ctxLog.Error("Trade error ", err)
			response.Description = err.Error()
			response.Status = proto.StatusError
			return &response
		}

		response.Description = orderToString(order)

		ctxLog.Trace("Order from Binance ", response.Description)

		if c.OrderType == constants.OrderTypeMarketName {

			if order.Status != filledValue {

				response.Status = proto.StatusError
				return &response
			}

		} else { // constants.OrderTypeLimitName
			// TODO
		}

		response.Order = &cmd.Order{}

		response.Order.ExternalOrderId = order.OrderID

		response.Order.ExecutionId, err = strconv.ParseInt(order.ClientOrderID, 10, 64)

		if err != nil {
			return errorResponse(&response, err)
		}

		response.Order.Price, err = strconv.ParseFloat(order.Fills[0].Price, 64)

		if err != nil {
			return errorResponse(&response, err)
		}

		response.Order.Commission, err = strconv.ParseFloat(order.Fills[0].Commission, 64)

		if err != nil {
			return errorResponse(&response, err)
		}

		response.Order.CommissionAsset = order.Fills[0].CommissionAsset

		response.Status = proto.StatusOk

		return &response
	}

	//------------------------------------------------------------------------------------------------------------------

	var lockOut = &sync.Mutex{}

	performFunction := func(in <-chan *proto.ExecRequest) {

		for {
			request := <-in

			ctxLog.Trace("Start sending cmd to Binance .....", request)

			var response *proto.ExecResponse

			if request.What == proto.ExecuteCmd {
				response = trade(request)
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
