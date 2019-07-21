package proto

import "msq.ai/data/cmd"

const ExecRequestCheckConnection = 1
const ExecRequestTrade = 2

const ExecResponseStatusOk = 1
const ExecResponseStatusError = 2

// GET_STATUS, TRADE, CHECK_CONNECTION

type ExecRequest struct {
	Id   int64
	What int16
	Cmd  *cmd.RawCommand
}

type ExecResponse struct {
	Id          int64
	Status      int16
	Description string

	// STATUS
	// RESULT
	// ERROR

}
