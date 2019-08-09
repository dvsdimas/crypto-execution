package ecib

import (
	log "github.com/sirupsen/logrus"
	"msq.ai/connectors/connector"
	"msq.ai/connectors/proto"
)

func RunIbConnector(in <-chan *proto.ExecRequest, out chan<- *proto.ExecResponse) {

	ctxLog := log.WithFields(log.Fields{"id": "BinanceConnector"})

	trade := func(request *proto.ExecRequest, response *proto.ExecResponse) *proto.ExecResponse {

		// TODO

		return nil
	}

	check := func(request *proto.ExecRequest, response *proto.ExecResponse) *proto.ExecResponse {

		// TODO

		return nil
	}

	info := func(request *proto.ExecRequest, response *proto.ExecResponse) *proto.ExecResponse {

		// TODO

		return nil

	}

	connector.RunConnector(ctxLog, in, out, 1, trade, check, info)
}
