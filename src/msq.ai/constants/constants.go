package constants

import "time"

const PostgresUrlPropertyName string = "postgres.url"
const CommandTimeForExecutionSecondsPropertyName string = "command.time.for.execution.seconds"
const ExchangeNamePropertyName string = "exchange.name"
const ConnectorIdPropertyName string = "connector.id"

const DbName string = "postgres"

const OrderTypeLimitName string = "LIMIT"
const OrderTypeMarketName string = "MARKET"
const OrderTypeInfoName string = "INFO"

const OrderDirectionBuyName string = "BUY"
const OrderDirectionSellName string = "SELL"

const TimeInForceFokName string = "FOK"
const TimeInForceGtcName string = "GTC"

const ExecutionStatusCreatedName string = "CREATED"
const ExecutionStatusExecutingName string = "EXECUTING"
const ExecutionStatusErrorName string = "ERROR"
const ExecutionStatusCompletedName string = "COMPLETED"
const ExecutionStatusTimedOutName string = "TIMED_OUT"
const ExecutionStatusRejectedName string = "REJECTED"

const DbErrorSleepTime = 10 * time.Second
