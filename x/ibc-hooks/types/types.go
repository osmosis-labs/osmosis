package types

import channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"

type ContractAck struct {
	ContractResult []byte `json:"contract_result"`
	IbcAck         []byte `json:"ibc_ack"`
}

type RequestIBCAckI struct {
	PacketSequence uint64 `json:"packet_sequence"`
	Channel        string `json:"channel"`
}

type RequestAck struct {
	RequestIBCAckI `json:"request_ack"`
}

type IBCAsync struct {
	RequestAck `json:"ibc_async"`
}

type IBCAckResponse struct {
	Packet      channeltypes.Packet `json:"packet"`
	ContractAck ContractAck         `json:"contract_ack"`
}
