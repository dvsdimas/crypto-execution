package ecbinance

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance"
	log "github.com/sirupsen/logrus"
	"msq.ai/connectors/connector"
	"msq.ai/connectors/proto"
	"msq.ai/constants"
	"msq.ai/data/cmd"
	"strconv"
	"time"
)

const filledValue = "FILLED"

func RunBinanceConnector(in <-chan *proto.ExecRequest, out chan<- *proto.ExecResponse, execPoolSize int) {

	ctxLog := log.WithFields(log.Fields{"id": "BinanceConnector"})

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

	connector.RunConnector(ctxLog, in, out, execPoolSize, trade)
}
