package constants

import "time"

const PostgresUrlPropertyName = "postgres.url"
const CommandTimeForExecutionSecondsPropertyName = "command.time.for.execution.seconds"
const ExchangeNamePropertyName = "exchange.name"
const ConnectorIdPropertyName = "connector.id"

const DbName = "postgres"

const OrderTypeLimitName = "LIMIT"
const OrderTypeMarketName = "MARKET"
const OrderTypeInfoName = "INFO"

const OrderDirectionBuyName = "BUY"
const OrderDirectionSellName = "SELL"

const TimeInForceFokName = "FOK"
const TimeInForceGtcName = "GTC"

const ExecutionStatusCreatedName = "CREATED"
const ExecutionStatusExecutingName = "EXECUTING"
const ExecutionStatusErrorName = "ERROR"
const ExecutionStatusCompletedName = "COMPLETED"
const ExecutionStatusTimedOutName = "TIMED_OUT"
const ExecutionStatusRejectedName = "REJECTED"

const DbErrorSleepTime = 10 * time.Second
