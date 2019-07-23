package cmd

import (
	"fmt"
	dic "msq.ai/db/postgres/dictionaries"
	"strconv"
	"time"
)

type Command struct {
	Id              int64
	ExchangeId      int16
	InstrumentName  string
	DirectionId     int16
	OrderTypeId     int16
	LimitPrice      float64
	Amount          float64
	StatusId        int16
	ConnectorId     int64
	ExecutionTypeId int16
	ExecuteTillTime time.Time
	RefPositionId   string
	TimeInForceId   int16
	UpdateTimestamp time.Time
	AccountId       int64
	Description     string
	ApiKey          string
	SecretKey       string
	ResultOrderId   string
}

type RawCommand struct {
	Id              string
	Exchange        string
	Instrument      string
	Direction       string
	OrderType       string
	LimitPrice      string
	Amount          string
	Status          string
	ConnectorId     string
	ExecutionType   string
	ExecuteTillTime string
	RefPositionId   string
	TimeInForce     string
	UpdateTime      string
	AccountId       string
	Description     string
	ApiKey          string
	SecretKey       string
	ResultOrderId   string
}

func ToRaw(cmd *Command, dictionaries *dic.Dictionaries) *RawCommand {

	var raw RawCommand

	raw.Id = strconv.FormatInt(cmd.Id, 10)
	raw.Exchange = dictionaries.Exchanges().GetNameById(cmd.ExchangeId)
	raw.Instrument = cmd.InstrumentName
	raw.Direction = dictionaries.Directions().GetNameById(cmd.DirectionId)
	raw.OrderType = dictionaries.OrderTypes().GetNameById(cmd.OrderTypeId)
	raw.LimitPrice = fmt.Sprintf("%f", cmd.LimitPrice)
	raw.Amount = fmt.Sprintf("%f", cmd.Amount)
	raw.Status = dictionaries.ExecutionStatuses().GetNameById(cmd.StatusId)
	raw.ConnectorId = strconv.FormatInt(cmd.ConnectorId, 10)
	raw.ExecutionType = dictionaries.ExecutionTypes().GetNameById(cmd.ExecutionTypeId)
	raw.ExecuteTillTime = cmd.ExecuteTillTime.Format(time.RFC3339)
	raw.RefPositionId = cmd.RefPositionId
	raw.TimeInForce = dictionaries.TimeInForces().GetNameById(cmd.TimeInForceId)
	raw.UpdateTime = cmd.UpdateTimestamp.Format(time.RFC3339)
	raw.AccountId = strconv.FormatInt(cmd.AccountId, 10)
	raw.Description = cmd.Description
	raw.ApiKey = cmd.ApiKey
	raw.SecretKey = cmd.SecretKey
	raw.ResultOrderId = cmd.ResultOrderId

	return &raw
}
