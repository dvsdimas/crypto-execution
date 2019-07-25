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
	What ExecType
	Cmd  *cmd.RawCommand
}

type ExecResponse struct {
	Status      Status
	Description string
	OriginCmd   *cmd.RawCommand
}
