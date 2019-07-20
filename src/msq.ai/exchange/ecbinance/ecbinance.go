package ecbinance

import (
	"context"
	"github.com/adshao/go-binance"
	log "github.com/sirupsen/logrus"
	"msq.ai/connectors/proto"
	dic "msq.ai/db/postgres/dictionaries"
)

func RunBinanceConnector(dictionaries *dic.Dictionaries, apiKey string, secretKey string, in <-chan *proto.ExecRequest, out chan<- *proto.ExecResponse) {

	ctxLog := log.WithFields(log.Fields{"id": "BinanceConnector"})

	ctxLog.Info("BinanceConnector is going to start")

	if dictionaries == nil {
		ctxLog.Fatal("dictionaries is nil !")
	}

	if in == nil {
		ctxLog.Fatal("ExecRequest channel is nil !")
	}

	if out == nil {
		ctxLog.Fatal("ExecResponse channel is nil !")
	}

	client := binance.NewClient(apiKey, secretKey)

	var request *proto.ExecRequest

	//------------------------------------------------------------------------------------------------------------------

	checkConnection := func(request *proto.ExecRequest) {

		var response = proto.ExecResponse{Id: request.Id}

		_, err := client.NewListPricesService().Do(context.Background())

		if err != nil {
			ctxLog.Error("checkConnection error ", err)
			response.Description = err.Error()
			response.Status = proto.ExecResponseStatusError
		} else {
			//for _, p := range prices {
			//	log.Trace(p)
			//}

			response.Status = proto.ExecResponseStatusOk
		}

		out <- &response
	}

	//------------------------------------------------------------------------------------------------------------------

	go func() {

		for {

			request = <-in

			if request == nil {
				ctxLog.Fatal("Protocol violation! ExecRequest is nil")
			}

			if request.What == proto.ExecRequestCheckConnection {
				checkConnection(request)
			} else {
				ctxLog.Fatal("Protocol violation! ExecRequest with wrong type ! ", request.What)
			}

			// TODO

		}
	}()

}
