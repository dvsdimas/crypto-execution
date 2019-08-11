package ecib

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"msq.ai/connectors/connector"
	"msq.ai/connectors/proto"
	"sync"
	"sync/atomic"
	"time"
)

const sleepTime = time.Second * 5
const pingTime = 10
const timeOutTime = 30

type msg struct {
	Header string `json:"header"`
	Body   string `json:"body"`
}

func RunIbConnector(in <-chan *proto.ExecRequest, out chan<- *proto.ExecResponse, wsUrl string, execPoolSize int) {

	ctxLog := log.WithFields(log.Fields{"id": "IbConnector"})

	var lastTime int64 = 0
	var connection *websocket.Conn = nil
	var lock sync.Mutex

	//------------------------------------------------------------------------------------------------------------------

	updateLastReceiveTime := func() {
		atomic.StoreInt64(&lastTime, time.Now().Unix())
	}

	getLastReceiveTime := func() int64 {
		return atomic.LoadInt64(&lastTime)
	}

	//------------------------------------------------------------------------------------------------------------------

	getConnection := func() *websocket.Conn {

		var tmp *websocket.Conn = nil

		for {

			lock.Lock()
			tmp = connection
			lock.Unlock()

			if tmp == nil {
				time.Sleep(time.Second)
				continue
			}

			return tmp
		}
	}

	createConnection := func() *websocket.Conn {

		var tmp *websocket.Conn = nil

		lock.Lock()
		tmp = connection
		connection = nil
		lock.Unlock()

		if tmp != nil {
			err := tmp.Close()

			if err != nil {
				ctxLog.Error("websocket.Close error", err)
			}
		}

		for {

			log.Trace("Connecting to ws with URL ", wsUrl)

			c, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)

			if err != nil {
				ctxLog.Error("websocket.Dial error", err)
				time.Sleep(sleepTime)
				continue
			}

			updateLastReceiveTime()

			c.SetPongHandler(func(appData string) error {
				ctxLog.Trace("Pong msg")
				updateLastReceiveTime()
				return nil
			})

			lock.Lock()
			connection = c
			lock.Unlock()

			return c
		}
	}

	//------------------------------------------------------------------------------------------------------------------

	mc := make(chan *msg, 10)

	go func() {

		ticker := time.NewTicker(time.Second * 20)

		for {

			select {

			case <-ticker.C:
				{
					mc <- &msg{Header: "sadasd", Body: "12345"}
				}

			}

		}

	}()

	//------------------------------------------------------------------------------------------------------------------

	go func() {

		for {

			tp, bytes, err := getConnection().ReadMessage()

			if err != nil {
				ctxLog.Error("WS ReadMessage error", err)
				time.Sleep(time.Second)
				continue
			}

			ctxLog.Printf("recv: %s", bytes)

			updateLastReceiveTime()

			if tp == websocket.TextMessage {

				ctxLog.Trace("TextMessage")

			} else if tp == websocket.BinaryMessage {

				ctxLog.Trace("BinaryMessage")

			} else if tp == websocket.CloseMessage {

				ctxLog.Trace("CloseMessage")

			} else if tp == websocket.PingMessage {

				ctxLog.Trace("PingMessage")

			} else if tp == websocket.PongMessage {

				ctxLog.Trace("PongMessage")

			}
		}

	}()

	//------------------------------------------------------------------------------------------------------------------

	go func() {

		var lastSentTime int64 = 0

		con := createConnection()

		ticker := time.NewTicker(time.Second * 1)

		for {

			select {

			case <-ticker.C:
				{

					now := time.Now().Unix()

					if now-lastSentTime >= pingTime {

						if err := con.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
							ctxLog.Error("Ping error", err)
							con = createConnection()
							continue
						}

						lastSentTime = now
					}

					if now-getLastReceiveTime() > timeOutTime {
						con = createConnection()
						continue
					}

				}

			case m := <-mc:
				{

					bts, err := json.Marshal(m)

					if err != nil {
						log.Error("Marshal error", err)
						continue
					}

					if err := con.WriteMessage(websocket.TextMessage, bts); err != nil {
						ctxLog.Error("WriteMessage error", err)
						con = createConnection()
						continue
					}

					lastSentTime = time.Now().Unix()
				}

			}

		} // end for

	}()

	//------------------------------------------------------------------------------------------------------------------

	//var m1, m2 = &msg{Header: "Hello", Body: "World"}, &msg{}
	//
	//err = c.WriteJSON(m1)
	//err = c.SetReadDeadline()
	//
	//if err != nil {
	//	ctxLog.Fatal("WriteJSON error", err)
	//}
	//
	//err = c.ReadJSON(m2)
	//
	//if err != nil {
	//	ctxLog.Fatal("ReadJSON error", err)
	//}
	//
	//ctxLog.Info("Sent: [", m1, "], Get: [", m2, "]")
	//
	//_ = c.Close()

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

	connector.RunConnector(ctxLog, in, out, execPoolSize, trade, check, info)
}
