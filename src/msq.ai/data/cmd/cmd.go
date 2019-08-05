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

type RawCommandWithOrder struct {
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
	Order           RawOrder
}

type RawCommandWithBalances struct {
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
	Balances        []RawBalance
}

func ToRaw(cmd *Command, dictionaries *dic.Dictionaries) *RawCommand {

	var raw RawCommand

	raw.Id = strconv.FormatInt(cmd.Id, 10)
	raw.Exchange = dictionaries.Exchanges().GetNameById(cmd.ExchangeId)
	raw.Instrument = cmd.InstrumentName
	raw.Direction = dictionaries.Directions().GetNameById(cmd.DirectionId)
	raw.OrderType = dictionaries.OrderTypes().GetNameById(cmd.OrderTypeId)

	if cmd.LimitPrice < 0 {
		raw.LimitPrice = ""
	} else {
		raw.LimitPrice = fmt.Sprintf("%f", cmd.LimitPrice)
	}

	if cmd.Amount <= 0 {
		raw.Amount = ""
	} else {
		raw.Amount = fmt.Sprintf("%f", cmd.Amount)
	}

	raw.Status = dictionaries.ExecutionStatuses().GetNameById(cmd.StatusId)

	if cmd.ConnectorId < 0 {
		raw.ConnectorId = ""
	} else {
		raw.ConnectorId = strconv.FormatInt(cmd.ConnectorId, 10)
	}

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

func ToRawWithOrder(cmd *Command, dictionaries *dic.Dictionaries, order *Order) *RawCommandWithOrder {

	raw := ToRaw(cmd, dictionaries)

	var rawCommandWithOrder = RawCommandWithOrder{
		Id:              raw.Id,
		Exchange:        raw.Exchange,
		Instrument:      raw.Instrument,
		Direction:       raw.Direction,
		OrderType:       raw.OrderType,
		LimitPrice:      raw.LimitPrice,
		Amount:          raw.Amount,
		Status:          raw.Status,
		ConnectorId:     raw.ConnectorId,
		ExecutionType:   raw.ExecutionType,
		ExecuteTillTime: raw.ExecuteTillTime,
		RefPositionId:   raw.RefPositionId,
		TimeInForce:     raw.TimeInForce,
		UpdateTime:      raw.UpdateTime,
		AccountId:       raw.AccountId,
		Order:           *toRawOrder(order),
	}

	return &rawCommandWithOrder
}

func ToRawWithBalances(cmd *Command, dictionaries *dic.Dictionaries, balances *[]*Balance) *RawCommandWithBalances {

	raw := ToRaw(cmd, dictionaries)

	var rawCommandWithBalances = RawCommandWithBalances{
		Id:              raw.Id,
		Exchange:        raw.Exchange,
		Instrument:      raw.Instrument,
		Direction:       raw.Direction,
		OrderType:       raw.OrderType,
		LimitPrice:      raw.LimitPrice,
		Amount:          raw.Amount,
		Status:          raw.Status,
		ConnectorId:     raw.ConnectorId,
		ExecutionType:   raw.ExecutionType,
		ExecuteTillTime: raw.ExecuteTillTime,
		RefPositionId:   raw.RefPositionId,
		TimeInForce:     raw.TimeInForce,
		UpdateTime:      raw.UpdateTime,
		AccountId:       raw.AccountId,
		Balances:        *toRawBalances(balances),
	}

	return &rawCommandWithBalances
}

func toRawOrder(order *Order) *RawOrder {

	raw := RawOrder{
		Id:              strconv.FormatInt(order.Id, 10),
		ExternalOrderId: strconv.FormatInt(order.ExternalOrderId, 10),
		ExecutionId:     strconv.FormatInt(order.ExecutionId, 10),
		Price:           fmt.Sprintf("%f", order.Price),
		Commission:      fmt.Sprintf("%f", order.Commission), // TODO precision !!!!!!!!!!!!!!!!!!!
		CommissionAsset: order.CommissionAsset,
	}

	return &raw
}

func toRawBalances(balances *[]*Balance) *[]RawBalance {

	raw := make([]RawBalance, len(*balances))

	for i, val := range *balances {
		raw[i] = RawBalance{
			Asset:  val.Asset,
			Free:   fmt.Sprintf("%f", val.Free),
			Locked: fmt.Sprintf("%f", val.Locked),
		}
	}

	return &raw
}

type Order struct {
	Id              int64
	ExternalOrderId int64
	ExecutionId     int64
	Price           float64
	Commission      float64
	CommissionAsset string
}

type RawOrder struct {
	Id              string
	ExternalOrderId string
	ExecutionId     string
	Price           string
	Commission      string
	CommissionAsset string
}

type Balance struct {
	Asset  string
	Free   float64
	Locked float64
}

type RawBalance struct {
	Asset  string
	Free   string
	Locked string
}
