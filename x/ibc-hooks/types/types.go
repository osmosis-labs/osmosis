package types

import (
	"encoding/json"

	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
)

// Async: The following types represent the response sent by a contract on OnRecvPacket when it wants the ack to be async

// OnRecvPacketAsyncAckResponse the response a contract sends to instruct the module to make the ack async
type OnRecvPacketAsyncAckResponse struct {
	IsAsyncAck bool `json:"is_async_ack"`
}

// Async The following types are used to ask a contract that has sent a packet to generate an ack for it

// RequestAckI internals of IBCAsync
type RequestAckI struct {
	PacketSequence uint64 `json:"packet_sequence"`
	SourceChannel  string `json:"source_channel"`
}

// RequestAck internals of IBCAsync
type RequestAck struct {
	RequestAckI `json:"request_ack"`
}

// IBCAsync is the sudo message to be sent to the contract for it to generate  an ack for a sent packet
type IBCAsync struct {
	RequestAck `json:"ibc_async"`
}

// General

// ContractAck is the response to be stored when a wasm hook is executed
type ContractAck struct {
	ContractResult []byte `json:"contract_result"`
	IbcAck         []byte `json:"ibc_ack"`
}

// IBCAckResponse is the response that a contract returns from the sudo() call on OnRecvPacket or RequestAck
type IBCAckResponse struct {
	Packet      channeltypes.Packet `json:"packet"`
	ContractAck ContractAck         `json:"contract_ack"`
}

// IBCAckError is the error that a contract returns from the sudo() call on RequestAck
type IBCAckError struct {
	Packet           channeltypes.Packet `json:"packet"`
	ErrorDescription string              `json:"error_description"`
	ErrorResponse    string              `json:"error_response"`
}

type IBCAck struct {
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
	// Note: These two fields have to be pointers so that they can be null
	// If they are not pointers, they will be empty structs when null,
	// which will cause issues with json.Unmarshal.
	AckResponse *IBCAckResponse `json:"response,omitempty"`
	AckError    *IBCAckError    `json:"error,omitempty"`
}

func UnmarshalIBCAck(bz []byte) (*IBCAck, error) {
	var ack IBCAck
	if err := json.Unmarshal(bz, &ack); err != nil {
		return nil, err
	}

	switch ack.Type {
	case "ack_response":
		ack.AckResponse = &IBCAckResponse{}
		if err := json.Unmarshal(ack.Content, ack.AckResponse); err != nil {
			return nil, err
		}
	case "ack_error":
		ack.AckError = &IBCAckError{}
		if err := json.Unmarshal(ack.Content, ack.AckError); err != nil {
			return nil, err
		}
	}

	return &ack, nil
}
