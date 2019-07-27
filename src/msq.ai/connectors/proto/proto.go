package proto

import "msq.ai/data/cmd"

type ExecType uint32

const (
	ExecuteCmd ExecType = iota
	CheckCmd
)

type Status uint32

const (
	StatusError Status = iota
	StatusOk
)

type ExecRequest struct {
	What   ExecType
	RawCmd *cmd.RawCommand
	Cmd    *cmd.Command
}

type ExecResponse struct {
	Status       Status
	Description  string
	OriginRawCmd *cmd.RawCommand
	OriginCmd    *cmd.Command
	Order        *cmd.RawOrder
}
