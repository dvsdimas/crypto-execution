package proto

import (
	"msq.ai/data/cmd"
	"time"
)

type ExecType uint32

const (
	ExecuteCmd ExecType = iota
	CheckCmd
	InfoCmd
)

type Status uint32

const (
	StatusError Status = iota
	StatusOk
	StatusTimedOut
	StatusRejected
)

type ExecRequest struct {
	What   ExecType
	RawCmd *cmd.RawCommand
	Cmd    *cmd.Command
}

type ExecResponse struct {
	Request          *ExecRequest
	Status           Status
	Description      string
	Order            *cmd.Order
	OutsideExecution time.Duration
	Balances         []cmd.Balance
}
