package gin

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	con "msq.ai/constants"
	"msq.ai/db/postgres/dao"
	dic "msq.ai/db/postgres/dictionaries"
	pgh "msq.ai/db/postgres/helper"
	"msq.ai/utils/math"
	"net/http"
	"strconv"
	"strings"
)

func RunGinRestService(dburl string, dictionaries *dic.Dictionaries) {

	ctxLog := log.WithFields(log.Fields{"id": "GinRestService"})

	ctxLog.Info("GinRestService is going to start")

	db, err := pgh.GetDbByUrl(dburl) // TODO configure DB connection pool !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

	if err != nil {
		ctxLog.Fatal("Cannot connect to DB with URL ["+dburl+"] ", err)
	}

	statusCreatedId := dictionaries.ExecutionStatuses().GetIdByName(con.ExecutionStatusCreatedName)

	// ?exchange=BINANCE&instrument=EUR/USD&direction=BUY&order_type=MARKET&limit_price=1.1&amount=1.0&execution_type=OPEN&ref_position_id=12345&account_id=43542352

	var handlerGET = func(c *gin.Context) {

		exchangeVal := strings.ToUpper(c.Query("exchange"))

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

		instrumentVal := strings.ToUpper(c.Query("instrument"))

		ctxLog.Trace("instrument [", instrumentVal, "]")

		if len(instrumentVal) <= 1 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Wrong 'instrument' parameter [" + instrumentVal + "]",
			})
			return
		}

		//--------------------------------------------------------------------------------------------------------------

		directionVal := strings.ToUpper(c.Query("direction"))

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

		orderTypeVal := strings.ToUpper(c.Query("order_type"))

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

		limitPriceVal := c.Query("limit_price")

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

		amountVal := c.Query("amount")

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

		executionTypeVal := strings.ToUpper(c.Query("execution_type"))

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

		refPositionIdVal := c.Query("ref_position_id")

		ctxLog.Trace("ref_position_id [", refPositionIdVal, "]")

		//--------------------------------------------------------------------------------------------------------------

		accountIdVal := c.Query("account_id")

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

		id, err := dao.InsertCommand(db, exchangeId, instrumentVal, directionId, orderTypeId, limitPrice, amount, statusCreatedId,
			executionTypeId, refPositionIdVal, accountId)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Cannot insert command into DB [" + err.Error() + "]",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": id})
	}

	router := gin.Default()

	var handlerPUT = func(c *gin.Context) {

		names := c.PostFormMap("names")

		ctxLog.Info(names)
	}

	v1 := router.Group("/execution/v1/command")
	{
		//v1.POST("/", createTodo)

		v1.PUT("/", handlerPUT)

		v1.GET("/", handlerGET) // TODO replace with POST
	}

	go func() {

		err := router.Run()

		if err != nil {
			ctxLog.Fatal("GinRestService error", err)
		}
	}()

}
