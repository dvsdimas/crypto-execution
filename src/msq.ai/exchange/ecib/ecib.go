package ecib

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"msq.ai/connectors/connector"
	"msq.ai/connectors/proto"
	"msq.ai/constants"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const sleepTime = time.Second * 5
const pingTime = 5
const timeOutTime = 15
const K1 = 1024
const cidName = "cid"

type ibMarketOrder struct {
	Code      string `json:"code"`
	Account   string `json:"account"`
	Op        string `json:"op"`
	Symbol    string `json:"symbol"`
	Qty       string `json:"qty"`
	OrderType string `json:"order_type"`
	// TODO time_in_force
	Cid string `json:"cid"`
}

type ibLimitOrder struct {
	Code      string `json:"code"`
	Account   string `json:"account"`
	Op        string `json:"op"`
	Symbol    string `json:"symbol"`
	Qty       string `json:"qty"`
	OrderType string `json:"order_type"`
	Price     string `json:"price"`
	// TODO time_in_force
	Cid string `json:"cid"`
}

type rsp struct {
	RawMap *map[string]interface{}
}

func RunIbConnector(in <-chan *proto.ExecRequest, out chan<- *proto.ExecResponse, wsUrl string, execPoolSize int) {

	ctxLog := log.WithFields(log.Fields{"id": "IbConnector"})

	var lastTime int64 = 0
	var connection *websocket.Conn = nil
	var lock sync.Mutex

	var bytesChanelLock sync.Mutex
	bytesChanel := make(chan *[]byte)

	sendBytes := func(bts *[]byte) {
		bytesChanelLock.Lock()
		bytesChanel <- bts
		bytesChanelLock.Unlock()
	}

	//------------------------------------------------------------------------------------------------------------------

	var inChannelsLock sync.Mutex

	inChannelsLock.Lock()

	channels := make(chan chan *rsp, execPoolSize)

	for i := 0; i < execPoolSize; i++ {
		channels <- make(chan *rsp)
	}

	inChannelsLock.Unlock()

	getChannel := func() chan *rsp {
		inChannelsLock.Lock()
		in := <-channels
		inChannelsLock.Unlock()
		return in
	}

	returnChannel := func(c chan *rsp) {
		inChannelsLock.Lock()
		channels <- c
		inChannelsLock.Unlock()
	}

	//------------------------------------------------------------------------------------------------------------------
	var dicLock sync.Mutex
	dic := make(map[int64]chan *rsp)

	addDic := func(id int64, c chan *rsp) {
		dicLock.Lock()
		dic[id] = c
		dicLock.Unlock()
	}

	rmDic := func(id int64) {
		dicLock.Lock()
		delete(dic, id)
		dicLock.Unlock()
	}

	getDic := func(id int64) chan *rsp {
		dicLock.Lock()
		c := dic[id]
		dicLock.Unlock()
		return c
	}

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

			c.SetReadLimit(K1)
			c.SetReadLimit(K1)

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

				var rawMap map[string]interface{}

				if err := json.Unmarshal(bytes, &rawMap); err != nil {
					ctxLog.Error("rawMap Unmarshal error [" + string(bytes) + "]")
					continue
				}

				cidStr := rawMap[cidName] // TODO check !!!

				if cidStr == nil {
					ctxLog.Error("cidStr is nil [" + string(bytes) + "]")
					continue
				}

				cid, err := strconv.ParseInt(cidStr.(string), 10, 64)

				if err != nil {
					ctxLog.Error("Cannot convert cidStr to int64 [" + string(bytes) + "]")
					continue
				}

				c := getDic(cid)

				if c == nil {
					ctxLog.Error("Cannot find out channel fir cid [" + string(bytes) + "]")
					continue
				}

				c <- &rsp{RawMap: &rawMap}

			} else {
				ctxLog.Error("Got BinaryMessage from WS !!!")
				ctxLog.Printf("BinaryMessage: [%s]", bytes)
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

	requestToBytes := func(request *proto.ExecRequest) (*[]byte, error) {

		if request.RawCmd.OrderType == constants.OrderTypeMarketName {

			var market = ibMarketOrder{
				Cid:       request.RawCmd.Id,
				Code:      "PLACE-ORDER",
				Account:   request.RawCmd.ApiKey,
				Op:        request.RawCmd.Direction,
				Symbol:    request.RawCmd.Instrument,
				Qty:       request.RawCmd.Amount,
				OrderType: "MKT",
			}

			bytes, err := json.Marshal(market)

			return &bytes, err

		} else if request.RawCmd.OrderType == constants.OrderTypeLimitName {

			var limit = ibLimitOrder{
				Cid:       request.RawCmd.Id,
				Code:      "PLACE-ORDER",
				Account:   request.RawCmd.ApiKey,
				Op:        request.RawCmd.Direction,
				Symbol:    request.RawCmd.Instrument,
				Qty:       request.RawCmd.Amount,
				OrderType: "LMT",
				Price:     request.RawCmd.LimitPrice,
			}

			bytes, err := json.Marshal(limit)

			return &bytes, err

		} else {
			ctxLog.Fatal("Protocol violation! ExecRequest wrong OrderType ! ", request)
			return nil, nil
		}
	}

	//------------------------------------------------------------------------------------------------------------------

	check := func(request *proto.ExecRequest, response *proto.ExecResponse) *proto.ExecResponse {

		// TODO

		return nil
	}

	//------------------------------------------------------------------------------------------------------------------

	trade := func(request *proto.ExecRequest, response *proto.ExecResponse) *proto.ExecResponse {

		bts, err := requestToBytes(request)

		if err != nil {
			log.Error("Marshal error", err)
			response.Description = "Marshal error [" + err.Error() + "]"
			return response
		}

		in := getChannel()

		addDic(request.Cmd.Id, in)

		sendBytes(bts)

		for {
			ticker := time.NewTicker(time.Second * 60)

			var result *rsp

			select {

			case <-ticker.C:
				{
					ctxLog.Error("Didn't get response from WS during 60 sec!", request)
				}

			case result = <-in:
				{
					ctxLog.Trace(result)
				}
			}

			if result == nil {
				return check(request, response)
			}

			// TODO check is it final order status, if not

			break // TODO
		}

		rmDic(request.Cmd.Id)

		returnChannel(in)

		// TODO prepare response

		response.Status = proto.StatusOk
		response.Order = nil // TODO

		return response
	}

	//------------------------------------------------------------------------------------------------------------------

	info := func(request *proto.ExecRequest, response *proto.ExecResponse) *proto.ExecResponse {

		// TODO

		return nil

	}

	//------------------------------------------------------------------------------------------------------------------

	connector.RunConnector(ctxLog, in, out, execPoolSize, trade, check, info)
}
