package ibc_rate_limit

import (
	"encoding/json"
	"strings"

	errorsmod "cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/osmosis-labs/osmosis/v31/x/ibc-rate-limit/types"
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
		return wrapContractError(err)
	}

	return nil
}

// contractRateLimitExceededMarker is the start of the message the rate limiter
// contract produces for its RateLimitExceded error variant (see
// contracts/rate-limiter/src/error.rs). wasmd hands contract errors back as
// strings, so this text is the only way to tell a genuine quota rejection from
// any other contract failure. Keep the two in sync.
const contractRateLimitExceededMarker = "IBC Rate Limit exceeded for"

// wrapContractError maps an error returned by the rate limiter contract onto
// the module's error types. Only a quota rejection becomes
// ErrRateLimitExceeded; every other failure (a query the contract could not
// run, a malformed packet, a bug) becomes ErrContractError so it is not
// reported to users as a rate limit. The contract's own message is kept in
// both cases.
func wrapContractError(err error) error {
	if strings.Contains(err.Error(), contractRateLimitExceededMarker) {
		return errorsmod.Wrap(types.ErrRateLimitExceeded, err.Error())
	}
	return errorsmod.Wrap(types.ErrContractError, err.Error())
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

type RecordSendMsg struct {
	RecordSend UndoPacketMsg `json:"record_send"`
}

// RecordSendRateLimit tells the contract the sequence of a packet whose send it
// has just authorised, so the contract can refund it later if it fails while
// the quota window it was counted in is still active.
func RecordSendRateLimit(ctx sdk.Context, contractKeeper *wasmkeeper.PermissionedKeeper,
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

	msg := RecordSendMsg{RecordSend: UndoPacketMsg{Packet: unwrapped}}
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

type ConfirmSendMsg struct {
	ConfirmSend UndoPacketMsg `json:"confirm_send"`
}

// ConfirmSendRateLimit tells the contract that a sent packet was acknowledged
// successfully so it can settle the record it keeps for a possible refund.
func ConfirmSendRateLimit(ctx sdk.Context, contractKeeper *wasmkeeper.PermissionedKeeper,
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

	msg := ConfirmSendMsg{ConfirmSend: UndoPacketMsg{Packet: unwrapped}}
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
