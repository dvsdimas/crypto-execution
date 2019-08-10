package ecib

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"msq.ai/connectors/connector"
	"msq.ai/connectors/proto"
)

type msg struct {
	Header string `json:"header"`
	Body   string `json:"body"`
}

func RunIbConnector(in <-chan *proto.ExecRequest, out chan<- *proto.ExecResponse, wsUrl string) {

	ctxLog := log.WithFields(log.Fields{"id": "BinanceConnector"})

	log.Trace("Connecting to ws with URL ", wsUrl)

	c, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)

	if err != nil {
		log.Fatal("dial:", err)
	}

	var m1, m2 = &msg{Header: "Hello", Body: "World"}, &msg{}

	err = c.WriteJSON(m1)

	if err != nil {
		ctxLog.Fatal("WriteJSON error", err)
	}

	err = c.ReadJSON(m2)

	if err != nil {
		ctxLog.Fatal("ReadJSON error", err)
	}

	ctxLog.Info("Sent: [", m1, "], Get: [", m2, "]")

	_ = c.Close()

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
