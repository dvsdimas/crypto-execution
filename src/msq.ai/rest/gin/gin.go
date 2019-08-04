package gin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	log "github.com/sirupsen/logrus"
	con "msq.ai/constants"
	comd "msq.ai/data/cmd"
	"msq.ai/db/postgres/dao"
	dic "msq.ai/db/postgres/dictionaries"
	pgh "msq.ai/db/postgres/helper"
	"msq.ai/utils/math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func RunGinRestService(dburl string, dictionaries *dic.Dictionaries, timeForExecution int) {

	ctxLog := log.WithFields(log.Fields{"id": "GinRestService"})

	logErrWithST := func(msg string, err error) {
		ctxLog.WithField("stacktrace", fmt.Sprintf("%+v", err.(*errors.Error).ErrorStack())).Error(msg)
	}

	logErr := func(msg string) {
		ctxLog.Error(msg)
	}

	delta := time.Duration(timeForExecution) * time.Second

	ctxLog.Info("GinRestService is going to start")

	db, err := pgh.GetDbByUrl(dburl)

	if err != nil {
		ctxLog.Fatal("Cannot connect to DB with URL ["+dburl+"] ", err)
	}

	executionStatusCompletedId := dictionaries.ExecutionStatuses().GetIdByName(con.ExecutionStatusCompletedName)

	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(time.Hour)

	dbLoadCommandById := func(id int64) (*comd.Command, error) {
		return dao.LoadCommandById(db, id)
	}

	dbInsertCommand := func(exchangeId int16, instrumentVal string, directionId int16, orderTypeId int16, limitPrice float64,
		timeInForce int16, amount float64, executionTypeId int16, future time.Time, refPositionIdVal string, now time.Time,
		accountId int64, apiKey string, secretKey string, fingerPrint string) (int64, error) {

		statusCreatedId := dictionaries.ExecutionStatuses().GetIdByName(con.ExecutionStatusCreatedName)

		return dao.InsertCommand(db, exchangeId, instrumentVal, directionId, orderTypeId, limitPrice, timeInForce, amount,
			statusCreatedId, executionTypeId, future, refPositionIdVal, now, accountId, apiKey, secretKey, fingerPrint)
	}

	// curl -X GET localhost:8080/execution/v1/command/25

	var handlerGET = func(c *gin.Context) {

		idVal := c.Param("id")

		ctxLog.Trace("id [", idVal, "]")

		id, err := strconv.ParseInt(idVal, 10, 64)

		if err != nil {

			logErrWithST("Cannot parse id ["+idVal+"]", err)

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Wrong command 'id' [" + idVal + "]",
			})

			return
		}

		ctxLog.Trace("id [", id, "]")

		command, err := dbLoadCommandById(id)

		if err != nil {

			logErrWithST("Cannot LoadCommandById ["+idVal+"] ", err)

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Cannot LoadCommandById [" + idVal + "] ",
			})

			return
		}

		if command == nil {

			logErr("Not found Command with Id [" + idVal + "] ")

			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "Not found Command with Id [" + idVal + "] ",
			})

			return
		}

		if command.StatusId == executionStatusCompletedId {

			// TODO add balances to INFO command

			// TODO add order to BUY/SELL command
		}

		c.JSON(http.StatusOK, comd.ToRaw(command, dictionaries))
	}

	router := gin.Default()

	// BUY
	// curl -X PUT -d "cmd[exchange]=BINANCE&cmd[instrument]=BTTBTC&cmd[direction]=BUY&cmd[order_type]=MARKET&cmd[time_in_force]=GTC&cmd[amount]=10000&cmd[execution_type]=OPEN&cmd[account_id]=1&cmd[api_key]=JbOlqQXl&cmd[secret_key]=xfPip87D&cmd[finger_print]=asdas" localhost:8080/execution/v1/command/

	// SELL
	// curl -X PUT -d "cmd[exchange]=BINANCE&cmd[instrument]=BTTBTC&cmd[direction]=SELL&cmd[order_type]=MARKET&cmd[time_in_force]=GTC&cmd[amount]=10000&cmd[execution_type]=OPEN&cmd[account_id]=1&cmd[api_key]=B8U0U&cmd[secret_key]=0abqCy&cmd[finger_print]=qwe" localhost:8080/execution/v1/command/

	// INFO
	// curl -X PUT -d "cmd[exchange]=BINANCE&cmd[instrument]= &cmd[direction]=ACCOUNT&cmd[order_type]=INFO&cmd[time_in_force]=FOK&cmd[amount]=0&cmd[execution_type]=REQUEST&cmd[account_id]=1&cmd[api_key]=JbOlqQ&cmd[secret_key]=xfPip87&cmd[finger_print]=asdfda" localhost:8080/execution/v1/command/

	var handlerPUT = func(c *gin.Context) {

		cmd := c.PostFormMap("cmd")

		if cmd == nil || len(cmd) == 0 {

			logErr("Absent PostFormMap 'cmd' ")

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Absent PostFormMap 'cmd' ",
			})

			return
		}

		ctxLog.Trace(cmd)

		//--------------------------------------------------------------------------------------------------------------

		exchangeVal := strings.ToUpper(cmd["exchange"])

		ctxLog.Trace("exchange [", exchangeVal, "]")

		exchangeId := dictionaries.Exchanges().GetIdByName(exchangeVal)

		ctxLog.Trace("exchangeId [", exchangeId, "]")

		if exchangeId < 0 {

			logErr("Wrong 'exchange' parameter [" + exchangeVal + "]")

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Wrong 'exchange' parameter [" + exchangeVal + "]",
			})

			return
		}

		//--------------------------------------------------------------------------------------------------------------

		instrumentVal := strings.ToUpper(cmd["instrument"])

		ctxLog.Trace("instrument [", instrumentVal, "]")

		if len(instrumentVal) < 1 || len(instrumentVal) > 20 {

			logErr("Wrong 'instrument' parameter [" + instrumentVal + "]")

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Wrong 'instrument' parameter [" + instrumentVal + "]",
			})

			return
		}

		//--------------------------------------------------------------------------------------------------------------

		directionVal := strings.ToUpper(cmd["direction"])

		ctxLog.Trace("direction [", directionVal, "]")

		directionId := dictionaries.Directions().GetIdByName(directionVal)

		ctxLog.Trace("directionId [", directionId, "]")

		if directionId < 0 {

			logErr("Wrong 'direction' parameter [" + directionVal + "]")

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Wrong 'direction' parameter [" + directionVal + "]",
			})

			return
		}

		//--------------------------------------------------------------------------------------------------------------

		orderTypeVal := strings.ToUpper(cmd["order_type"])

		ctxLog.Trace("order_type [", orderTypeVal, "]")

		orderTypeId := dictionaries.OrderTypes().GetIdByName(orderTypeVal)

		ctxLog.Trace("orderTypeId [", orderTypeId, "]")

		if orderTypeId < 0 {

			logErr("Wrong 'order_type' parameter [" + orderTypeVal + "]")

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Wrong 'order_type' parameter [" + orderTypeVal + "]",
			})

			return
		}

		//--------------------------------------------------------------------------------------------------------------

		limitPriceVal := cmd["limit_price"]

		ctxLog.Trace("limit_price [", limitPriceVal, "]")

		var limitPrice float64 = -1

		if orderTypeId == dictionaries.OrderTypes().GetIdByName(con.OrderTypeLimitName) {

			limitPrice, err = strconv.ParseFloat(limitPriceVal, 64)

			if err != nil {

				logErrWithST("Cannot parse limit_price ["+limitPriceVal+"]", err)

				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": "Wrong 'limit_price' parameter [" + limitPriceVal + "]",
				})

				return
			}

			ctxLog.Trace("limit_price [", limitPrice, "]")

			if math.IsZero(limitPrice) || limitPrice < 0 {

				logErr("Wrong 'limit_price' parameter [" + limitPriceVal + "]")

				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": "Wrong 'limit_price' parameter [" + limitPriceVal + "]",
				})

				return
			}
		}

		//--------------------------------------------------------------------------------------------------------------

		timeInForceVal := strings.ToUpper(cmd["time_in_force"])

		ctxLog.Trace("time_in_force [", timeInForceVal, "]")

		timeInForceId := dictionaries.TimeInForces().GetIdByName(timeInForceVal)

		ctxLog.Trace("timeInForceId [", timeInForceId, "]")

		if timeInForceId < 0 {

			logErr("Wrong 'time_in_force' parameter [" + timeInForceVal + "]")

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Wrong 'time_in_force' parameter [" + timeInForceVal + "]",
			})

			return
		}

		//--------------------------------------------------------------------------------------------------------------

		amountVal := cmd["amount"]

		ctxLog.Trace("amount [", amountVal, "]")

		amount, err := strconv.ParseFloat(amountVal, 64)

		if err != nil {

			logErrWithST("Cannot parse amount ["+amountVal+"]", err)

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Wrong 'amount' parameter [" + amountVal + "]",
			})

			return
		}

		ctxLog.Trace("amount [", amount, "]")

		if amount < 0 {

			logErr("Wrong 'amount' parameter [" + amountVal + "]")

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Wrong 'amount' parameter [" + amountVal + "]",
			})

			return
		}

		//--------------------------------------------------------------------------------------------------------------

		executionTypeVal := strings.ToUpper(cmd["execution_type"])

		ctxLog.Trace("execution_type [", executionTypeVal, "]")

		executionTypeId := dictionaries.ExecutionTypes().GetIdByName(executionTypeVal)

		ctxLog.Trace("executionTypeId [", executionTypeId, "]")

		if executionTypeId < 0 {

			logErr("Wrong 'execution_type' parameter [" + executionTypeVal + "]")

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Wrong 'execution_type' parameter [" + executionTypeVal + "]",
			})

			return
		}

		//--------------------------------------------------------------------------------------------------------------

		refPositionIdVal := cmd["ref_position_id"]

		ctxLog.Trace("ref_position_id [", refPositionIdVal, "]")

		//--------------------------------------------------------------------------------------------------------------

		accountIdVal := cmd["account_id"]

		ctxLog.Trace("account_id [", accountIdVal, "]")

		accountId, err := strconv.ParseInt(accountIdVal, 10, 64)

		if err != nil {

			logErr("Wrong 'account_id' parameter [" + accountIdVal + "]")

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Wrong 'account_id' parameter [" + accountIdVal + "]",
			})

			return
		}

		ctxLog.Trace("accountId [", accountId, "]")

		//--------------------------------------------------------------------------------------------------------------

		apiKey := cmd["api_key"]

		ctxLog.Trace("api_key [", apiKey, "]")

		if len(apiKey) < 1 {

			logErr("Wrong 'api_key' parameter [" + apiKey + "]")

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Wrong 'api_key' parameter [" + apiKey + "]",
			})

			return
		}

		ctxLog.Trace("apiKey [", apiKey, "]")

		//--------------------------------------------------------------------------------------------------------------

		secretKey := cmd["secret_key"]

		ctxLog.Trace("secret_key [", secretKey, "]")

		if len(secretKey) < 1 {

			logErr("Wrong 'secret_key' parameter [" + secretKey + "]")

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Wrong 'secret_key' parameter [" + secretKey + "]",
			})

			return
		}

		ctxLog.Trace("secretKey [", secretKey, "]")

		//--------------------------------------------------------------------------------------------------------------

		fingerPrint := cmd["finger_print"]

		ctxLog.Trace("finger_print [", fingerPrint, "]")

		if len(fingerPrint) < 1 {

			logErr("Wrong 'finger_print' parameter [" + fingerPrint + "]")

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Wrong 'finger_print' parameter [" + fingerPrint + "]",
			})

			return
		}

		ctxLog.Trace("fingerPrint [", fingerPrint, "]")

		//--------------------------------------------------------------------------------------------------------------

		now := time.Now()

		future := now.Add(delta)

		id, err := dbInsertCommand(exchangeId, instrumentVal, directionId, orderTypeId, limitPrice, timeInForceId, amount,
			executionTypeId, future, refPositionIdVal, now, accountId, apiKey, secretKey, fingerPrint)

		if err != nil {

			logErrWithST("Cannot insert command into DB", err)

			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Cannot insert command into DB [" + err.Error() + "]",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": id})
	}

	v1 := router.Group("/execution/v1/command")
	{
		v1.PUT("/", handlerPUT)
		v1.GET("/:id", handlerGET)
	}

	go func() {

		err := router.Run()

		if err != nil {
			ctxLog.Fatal("GinRestService error", err)
		}
	}()

}
