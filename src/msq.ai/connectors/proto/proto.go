package proto

const ExecRequestCheckConnection = 1

const ExecResponseStatusOk = 1
const ExecResponseStatusError = 2

type ExecRequest struct {
	Id int64
	// GET_STATUS, TRADE, CHECK_CONNECTION

	What int16
}

type ExecResponse struct {
	Id          int64
	Status      int16
	Description string

	// STATUS
	// RESULT
	// ERROR

}
