package ibc_rate_limit

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/osmosis-labs/osmosis/v27/x/ibc-rate-limit/types"
)

var (
	msgSend = "send_packet"
	msgRecv = "recv_packet"
)

func CheckAndUpdateRateLimits(ctx sdk.Context, contractKeeper *wasmkeeper.PermissionedKeeper,
	msgType, contract string, packet exported.PacketI,
) error {
	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return err
	}

	sendPacketMsg, err := BuildWasmExecMsg(
		msgType,
		packet,
	)
	if err != nil {
		return err
	}

	_, err = contractKeeper.Sudo(ctx, contractAddr, sendPacketMsg)

	if err != nil {
		return errorsmod.Wrap(types.ErrRateLimitExceeded, err.Error())
	}

	return nil
}

type UndoSendMsg struct {
	UndoSend UndoPacketMsg `json:"undo_send"`
}

type UndoPacketMsg struct {
	Packet UnwrappedPacket `json:"packet"`
}

func UndoSendRateLimit(ctx sdk.Context, contractKeeper *wasmkeeper.PermissionedKeeper,
	contract string,
	packet exported.PacketI,
) error {
	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return err
	}

	unwrapped, err := unwrapPacket(packet)
	if err != nil {
		return err
	}

	msg := UndoSendMsg{UndoSend: UndoPacketMsg{Packet: unwrapped}}
	asJson, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = contractKeeper.Sudo(ctx, contractAddr, asJson)
	if err != nil {
		return errorsmod.Wrap(types.ErrContractError, err.Error())
	}

	return nil
}

type SendPacketMsg struct {
	SendPacket PacketMsg `json:"send_packet"`
}

type RecvPacketMsg struct {
	RecvPacket PacketMsg `json:"recv_packet"`
}

type PacketMsg struct {
	Packet UnwrappedPacket `json:"packet"`
}

type UnwrappedPacket struct {
	Sequence           uint64                                `json:"sequence"`
	SourcePort         string                                `json:"source_port"`
	SourceChannel      string                                `json:"source_channel"`
	DestinationPort    string                                `json:"destination_port"`
	DestinationChannel string                                `json:"destination_channel"`
	Data               transfertypes.FungibleTokenPacketData `json:"data"`
	TimeoutHeight      clienttypes.Height                    `json:"timeout_height"`
	TimeoutTimestamp   uint64                                `json:"timeout_timestamp,omitempty"`
}

func unwrapPacket(packet exported.PacketI) (UnwrappedPacket, error) {
	var packetData transfertypes.FungibleTokenPacketData
	err := json.Unmarshal(packet.GetData(), &packetData)
	if err != nil {
		return UnwrappedPacket{}, err
	}
	height, ok := packet.GetTimeoutHeight().(clienttypes.Height)
	if !ok {
		return UnwrappedPacket{}, types.ErrBadMessage
	}
	return UnwrappedPacket{
		Sequence:           packet.GetSequence(),
		SourcePort:         packet.GetSourcePort(),
		SourceChannel:      packet.GetSourceChannel(),
		DestinationPort:    packet.GetDestPort(),
		DestinationChannel: packet.GetDestChannel(),
		Data:               packetData,
		TimeoutHeight:      height,
		TimeoutTimestamp:   packet.GetTimeoutTimestamp(),
	}, nil
}

func BuildWasmExecMsg(msgType string, packet exported.PacketI) ([]byte, error) {
	unwrapped, err := unwrapPacket(packet)
	if err != nil {
		return []byte{}, err
	}

	var asJson []byte
	switch {
	case msgType == msgSend:
		msg := SendPacketMsg{SendPacket: PacketMsg{
			Packet: unwrapped,
		}}
		asJson, err = json.Marshal(msg)
	case msgType == msgRecv:
		msg := RecvPacketMsg{RecvPacket: PacketMsg{
			Packet: unwrapped,
		}}
		asJson, err = json.Marshal(msg)
	default:
		return []byte{}, types.ErrBadMessage
	}

	if err != nil {
		return []byte{}, err
	}

	return asJson, nil
}
