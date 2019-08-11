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
const pingTime = 5
const timeOutTime = 15
const K10 = 1024 * 10

type rsp struct {
	Order string
	Info  string
}

func RunIbConnector(in <-chan *proto.ExecRequest, out chan<- *proto.ExecResponse, wsUrl string, execPoolSize int) {

	ctxLog := log.WithFields(log.Fields{"id": "IbConnector"})

	var lastTime int64 = 0
	var connection *websocket.Conn = nil
	var lock sync.Mutex

	var bytesChanelLock sync.Mutex
	bytesChanel := make(chan *[]byte)

	//------------------------------------------------------------------------------------------------------------------

	var inChannelsLock sync.Mutex

	inChannelsLock.Lock()

	channels := make(chan chan *rsp, execPoolSize)

	for i := 0; i < execPoolSize; i++ {
		channels <- make(chan *rsp)
	}

	inChannelsLock.Unlock()

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
				time.Sleep(time.Millisecond * 100)
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

			c.SetReadLimit(K10)
			c.SetReadLimit(K10)

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

			case m := <-bytesChanel:
				{
					if err := con.WriteMessage(websocket.TextMessage, *m); err != nil {
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

	trade := func(request *proto.ExecRequest, response *proto.ExecResponse) *proto.ExecResponse {

		bts, err := json.Marshal(request)

		if err != nil {
			log.Error("Marshal error", err)
			response.Description = "Marshal error [" + err.Error() + "]"
			return response
		}

		// TODO LOCK
		// TODO get free channel
		// TODO put MAP ID -> channel
		// TODO UNLOCK

		bytesChanelLock.Lock()
		bytesChanel <- &bts
		bytesChanelLock.Unlock()

		// TODO await result on free channel

		// TODO LOCK
		// TODO remove MAP ID -> channel
		// TODO return free channel
		// TODO UNLOCK

		// TODO if nil

		// TODO parse response

		response.Status = proto.StatusOk
		response.Order = nil // TODO

		return response
	}

	//------------------------------------------------------------------------------------------------------------------

	check := func(request *proto.ExecRequest, response *proto.ExecResponse) *proto.ExecResponse {

		// TODO

		return nil
	}

	//------------------------------------------------------------------------------------------------------------------

	info := func(request *proto.ExecRequest, response *proto.ExecResponse) *proto.ExecResponse {

		// TODO

		return nil

	}

	//------------------------------------------------------------------------------------------------------------------

	connector.RunConnector(ctxLog, in, out, execPoolSize, trade, check, info)
}
