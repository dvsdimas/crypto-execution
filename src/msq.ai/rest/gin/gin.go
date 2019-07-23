package gin

import (
	"github.com/gin-gonic/gin"
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

	delta := time.Duration(timeForExecution) * time.Second

	ctxLog.Info("GinRestService is going to start")

	db, err := pgh.GetDbByUrl(dburl)

	if err != nil {
		ctxLog.Fatal("Cannot connect to DB with URL ["+dburl+"] ", err)
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(30)
	db.SetConnMaxLifetime(time.Hour)

	dbLoadCommandById := func(id int64) (*comd.Command, error) {
		return dao.LoadCommandById(db, id)
	}

	dbInsertCommand := func(exchangeId int16, instrumentVal string, directionId int16, orderTypeId int16, limitPrice float64,
		amount float64, executionTypeId int16, future time.Time, refPositionIdVal string, now time.Time, accountId int64,
		apiKey string, secretKey string, fingerPrint string) (int64, error) {

		statusCreatedId := dictionaries.ExecutionStatuses().GetIdByName(con.ExecutionStatusCreatedName)

		return dao.InsertCommand(db, exchangeId, instrumentVal, directionId, orderTypeId, limitPrice, amount, statusCreatedId,
			executionTypeId, future, refPositionIdVal, now, accountId, apiKey, secretKey, fingerPrint)
	}

	// curl -X GET localhost:8080/execution/v1/command/25

	var handlerGET = func(c *gin.Context) {

		idVal := c.Param("id")

		ctxLog.Trace("id [", idVal, "]")

		id, err := strconv.ParseInt(idVal, 10, 64)

		if err != nil {

			ctxLog.Error("Cannot parse id ["+idVal+"]", err)

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Wrong command 'id' [" + idVal + "]",
			})
			return
		}

		ctxLog.Trace("id [", id, "]")

		command, err := dbLoadCommandById(id)

		if err != nil {

			ctxLog.Error("Cannot LoadCommandById ["+idVal+"] ", err)

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Cannot LoadCommandById [" + idVal + "] ",
			})
			return
		}

		c.JSON(http.StatusOK, comd.ToRaw(command, dictionaries))
	}

	router := gin.Default()

	// curl -X PUT -d "cmd[exchange]=BINANCE&cmd[instrument]=BTT/BTC&cmd[direction]=BUY&cmd[order_type]=MARKET&cmd[amount]=1000&cmd[execution_type]=OPEN&cmd[account_id]=1&cmd[api_key]=JbOl&cmd[secret_key]=xfPip8&cmd[finger_print]=sad" localhost:8080/execution/v1/command/

	var handlerPUT = func(c *gin.Context) {

		cmd := c.PostFormMap("cmd")

		if cmd == nil || len(cmd) == 0 {
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
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Wrong 'exchange' parameter [" + exchangeVal + "]",
			})
			return
		}

		//--------------------------------------------------------------------------------------------------------------

		instrumentVal := strings.ToUpper(cmd["instrument"])

		ctxLog.Trace("instrument [", instrumentVal, "]")

		if len(instrumentVal) <= 1 {
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
				ctxLog.Error("Cannot parse limit_price ["+limitPriceVal+"]", err)

				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": "Wrong 'limit_price' parameter [" + limitPriceVal + "]",
				})
				return
			}

			ctxLog.Trace("limit_price [", limitPrice, "]")

			if math.IsZero(limitPrice) || limitPrice < 0 {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": "Wrong 'limit_price' parameter [" + limitPriceVal + "]",
				})
				return
			}
		}

		//--------------------------------------------------------------------------------------------------------------

		amountVal := cmd["amount"]

		ctxLog.Trace("amount [", amountVal, "]")

		amount, err := strconv.ParseFloat(amountVal, 64)

		if err != nil {
			ctxLog.Error("Cannot parse amount ["+amountVal+"]", err)

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Wrong 'amount' parameter [" + amountVal + "]",
			})
			return
		}

		ctxLog.Trace("amount [", amount, "]")

		if math.IsZero(amount) || amount < 0 {
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
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Wrong 'finger_print' parameter [" + fingerPrint + "]",
			})
			return
		}

		ctxLog.Trace("fingerPrint [", fingerPrint, "]")

		//--------------------------------------------------------------------------------------------------------------

		now := time.Now()

		future := now.Add(delta)

		id, err := dbInsertCommand(exchangeId, instrumentVal, directionId, orderTypeId, limitPrice, amount, executionTypeId,
			future, refPositionIdVal, now, accountId, apiKey, secretKey, fingerPrint)

		if err != nil {
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
