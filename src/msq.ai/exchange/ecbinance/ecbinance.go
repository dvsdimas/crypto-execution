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
const orderNotExistError = -2013

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

	trade := func(request *proto.ExecRequest, response *proto.ExecResponse) *proto.ExecResponse {

		client := binance.NewClient(request.RawCmd.ApiKey, request.RawCmd.SecretKey)

		orderService := client.NewCreateOrderService().Symbol(request.RawCmd.Instrument)
		orderService = orderService.NewClientOrderID(request.RawCmd.Id)

		if request.RawCmd.OrderType == constants.OrderTypeMarketName {
			orderService = orderService.Type(binance.OrderTypeMarket)
		} else if request.RawCmd.Direction == constants.OrderTypeLimitName {
			orderService = orderService.Type(binance.OrderTypeLimit)
			orderService = orderService.Price(request.RawCmd.LimitPrice)
		} else {
			ctxLog.Fatal("Protocol violation! ExecRequest wrong OrderType with empty cmd ! ", request)
			return nil
		}

		if request.RawCmd.Direction == constants.OrderDirectionBuyName {
			orderService = orderService.Side(binance.SideTypeBuy)
		} else if request.RawCmd.Direction == constants.OrderDirectionSellName {
			orderService = orderService.Side(binance.SideTypeSell)
		} else {
			ctxLog.Fatal("Protocol violation! ExecRequest wrong Direction with empty cmd ! ", request)
			return nil
		}

		if request.RawCmd.TimeInForce == constants.TimeInForceGtcName {
			// orderService = orderService.TimeInForce(binance.TimeInForceGTC)
		} else {
			msg := "Protocol violation! ExecRequest has wrong TimeInForce. Binance supported only GTC !"
			ctxLog.Error(msg, request)
			response.Description = msg
			response.Status = proto.StatusError
			return response
		}

		orderService = orderService.Quantity(request.RawCmd.Amount)

		start := time.Now()

		order, err := orderService.Do(context.Background())

		response.OutsideExecution = time.Now().Sub(start)

		if err != nil {
			ctxLog.Error("Trade error ", err)
			response.Description = err.Error()
			response.Status = proto.StatusError
			return response
		}

		response.Description = orderToString(order)

		ctxLog.Trace("Order from Binance ", response.Description)

		if request.RawCmd.OrderType == constants.OrderTypeMarketName {

			if order.Status != filledValue {

				response.Status = proto.StatusError
				return response
			}

		} else { // constants.OrderTypeLimitName
			// TODO
		}

		response.Order = &cmd.Order{}

		response.Order.ExternalOrderId = order.OrderID

		response.Order.ExecutionId, err = strconv.ParseInt(order.ClientOrderID, 10, 64)

		if err != nil {
			return errorResponse(response, err)
		}

		response.Order.Price, err = strconv.ParseFloat(order.Fills[0].Price, 64)

		if err != nil {
			return errorResponse(response, err)
		}

		response.Order.Commission, err = strconv.ParseFloat(order.Fills[0].Commission, 64)

		if err != nil {
			return errorResponse(response, err)
		}

		response.Order.CommissionAsset = order.Fills[0].CommissionAsset

		response.Status = proto.StatusOk

		return response
	}

	check := func(request *proto.ExecRequest, response *proto.ExecResponse) *proto.ExecResponse {

		client := binance.NewClient(request.RawCmd.ApiKey, request.RawCmd.SecretKey)

		order, err := client.NewGetOrderService().Symbol(request.RawCmd.Instrument).OrigClientOrderID(request.RawCmd.Id).Do(context.Background())

		if err != nil {

			if binance.IsAPIError(err) && err.(*binance.APIError).Code == orderNotExistError {

				if request.Cmd.ExecuteTillTime.After(time.Now()) {
					return trade(request, response)
				} else {
					ctxLog.Info("Check error, order not exist and will be marked timed_out ", err)
					response.Description = err.Error()
					response.Status = proto.StatusTimedOut
					return response
				}

			} else {
				ctxLog.Error("Check error ", err)
				response.Description = err.Error()
				response.Status = proto.StatusError
				return response
			}
		}

		ctxLog.Trace("Order from Binance ", order)

		response.Description = fmt.Sprintf("%+v", order)

		if request.RawCmd.OrderType == constants.OrderTypeMarketName {

			if order.Status != filledValue {

				response.Status = proto.StatusError
				return response
			}

		} else { // constants.OrderTypeLimitName
			// TODO
		}

		response.Order = &cmd.Order{}

		response.Order.ExternalOrderId = order.OrderID

		response.Order.ExecutionId, err = strconv.ParseInt(order.ClientOrderID, 10, 64)

		if err != nil {
			return errorResponse(response, err)
		}

		response.Order.Price, err = strconv.ParseFloat(order.Price, 64)

		if err != nil {
			return errorResponse(response, err)
		}

		response.Order.Commission = 0

		response.Order.CommissionAsset = "UNKNOWN"

		response.Status = proto.StatusOk

		return response
	}

	info := func(request *proto.ExecRequest, response *proto.ExecResponse) *proto.ExecResponse {

		//client := binance.NewClient(request.RawCmd.ApiKey, request.RawCmd.SecretKey)

		// TODO

		return response
	}

	connector.RunConnector(ctxLog, in, out, execPoolSize, trade, check, info)
}
