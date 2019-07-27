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
	ApiKey          string
	SecretKey       string
	FingerPrint     string
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
	ApiKey          string
	SecretKey       string
	FingerPrint     string
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
	raw.ApiKey = cmd.ApiKey
	raw.SecretKey = cmd.SecretKey
	raw.FingerPrint = cmd.FingerPrint

	return &raw
}

type RawOrder struct {
	ID               string // ""		 				TODO DB int64
	Symbol           string // "BTTBTC" 				TODO DB string
	OrderID          string // 7693572 					TODO DB int64
	ClientOrderID    string // "F35WUSPdFNQGSB9tGx1g8w" TODO DB int64
	TransactTime     string // 1564228612661 			TODO DB int64
	Price            string // "0.00000009"				TODO DB float64
	ExecutedQuantity string // "10000.00000000"			TODO DB float64
	Status           string // "FILLED"					TODO DB string
	TimeInForce      string // "GTC"					TODO DB int16
	Type             string // "MARKET"					TODO DB int16
	Side             string // "BUY"					TODO DB int16
	Commission       string // "10.00000000"			TODO DB	float64
	CommissionAsset  string // "BTT"					TODO DB string
}
