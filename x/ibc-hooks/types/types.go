package types

import channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"

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
