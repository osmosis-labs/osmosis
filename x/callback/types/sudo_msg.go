package types

import "encoding/json"

// SudoMsg callback message sent to a contract.
// This is encoded as JSON input to the contract when executing the callback
type SudoMsg struct {
	// Callback is the endpoint name at the contract which is called
	Callback *CallbackMsg `json:"callback,omitempty"`
}

// CallbackMsg is the callback message sent to a contract.
type CallbackMsg struct {
	// JobID is the user specified job id
	JobID uint64 `json:"job_id"`
}

// NewCallbackMsg creates a new Callback instance.
func NewCallbackMsg(jobID uint64) SudoMsg {
	return SudoMsg{
		Callback: &CallbackMsg{
			JobID: jobID,
		},
	}
}

// Bytes returns the callback message as JSON bytes
func (s SudoMsg) Bytes() []byte {
	msgBz, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return msgBz
}

// String returns the callback message as JSON string
func (s SudoMsg) String() string {
	return string(s.Bytes())
}
