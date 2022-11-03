package ibc_rate_limit

import (
	"encoding/json"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"strings"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channelkeeper "github.com/cosmos/ibc-go/v3/modules/core/04-channel/keeper"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	"github.com/osmosis-labs/osmosis/v12/x/ibc-rate-limit/types"
)

var (
	msgSend = "send_packet"
	msgRecv = "recv_packet"
)

func CheckAndUpdateRateLimits(ctx sdk.Context, contractKeeper *wasmkeeper.PermissionedKeeper,
	msgType, contract string, packet channeltypes.Packet, channelValue sdk.Int, localDenom string,
) error {
	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return err
	}

	sendPacketMsg, err := BuildWasmExecMsg(
		msgType,
		packet,
		channelValue,
		localDenom,
	)
	if err != nil {
		return err
	}

	_, err = contractKeeper.Sudo(ctx, contractAddr, sendPacketMsg)

	if err != nil {
		return sdkerrors.Wrap(types.ErrRateLimitExceeded, err.Error())
	}

	return nil
}

type UndoSendMsg struct {
	UndoSend UndoPacketMsg `json:"undo_send"`
}

type UndoPacketMsg struct {
	Packet     UnwrappedPacket `json:"packet"`
	LocalDenom string          `json:"local_denom"`
}

func UndoSendRateLimit(ctx sdk.Context, contractKeeper *wasmkeeper.PermissionedKeeper,
	contract string,
	packet channeltypes.Packet, denom string,
) error {
	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return err
	}

	unwrapped, err := unwrapPacket(packet)
	if err != nil {
		return err
	}

	msg := UndoSendMsg{UndoSend: UndoPacketMsg{Packet: unwrapped, LocalDenom: denom}}
	asJson, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = contractKeeper.Sudo(ctx, contractAddr, asJson)
	if err != nil {
		return sdkerrors.Wrap(types.ErrContractError, err.Error())
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
	Packet           UnwrappedPacket `json:"packet"`
	LocalDenom       string          `json:"local_denom"`
	ChannelValueHint sdk.Int         `json:"channel_value_hint"`
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

func unwrapPacket(packet channeltypes.Packet) (UnwrappedPacket, error) {
	var packetData transfertypes.FungibleTokenPacketData
	err := json.Unmarshal(packet.GetData(), &packetData)
	if err != nil {
		return UnwrappedPacket{}, err
	}
	return UnwrappedPacket{
		Sequence:           packet.Sequence,
		SourcePort:         packet.SourcePort,
		SourceChannel:      packet.SourceChannel,
		DestinationPort:    packet.DestinationPort,
		DestinationChannel: packet.DestinationChannel,
		Data:               packetData,
		TimeoutHeight:      packet.TimeoutHeight,
		TimeoutTimestamp:   packet.TimeoutTimestamp,
	}, nil

}

func BuildWasmExecMsg(msgType string, packet channeltypes.Packet, channelValue sdk.Int, localDenom string) ([]byte, error) {
	unwrapped, err := unwrapPacket(packet)
	if err != nil {
		return []byte{}, err
	}

	var asJson []byte
	switch {
	case msgType == msgSend:
		msg := SendPacketMsg{SendPacket: PacketMsg{
			Packet:           unwrapped,
			LocalDenom:       localDenom,
			ChannelValueHint: channelValue,
		}}
		asJson, err = json.Marshal(msg)
	case msgType == msgRecv:
		msg := RecvPacketMsg{RecvPacket: PacketMsg{
			Packet:           unwrapped,
			LocalDenom:       localDenom,
			ChannelValueHint: channelValue,
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

func GetFundsFromPacket(packet exported.PacketI) (string, string, error) {
	var packetData transfertypes.FungibleTokenPacketData
	err := json.Unmarshal(packet.GetData(), &packetData)
	if err != nil {
		return "", "", err
	}
	return packetData.Amount, GetLocalDenom(packetData.Denom), nil
}

func GetLocalDenom(denom string) string {
	// Expected denoms in the following cases:
	//
	// send non-native: transfer/channel-0/denom -> ibc/xxx
	// send native: denom -> denom
	// recv (B)non-native: denom
	// recv (B)native: transfer/channel-0/denom
	//
	if strings.HasPrefix(denom, "transfer/") {
		denomTrace := transfertypes.ParseDenomTrace(denom)
		return denomTrace.IBCDenom()
	} else {
		return denom
	}
}

func CalculateChannelValue(ctx sdk.Context, denom string, bankKeeper bankkeeper.Keeper, channelKeeper channelkeeper.Keeper) sdk.Int {
	// For non-native (ibc) tokens, return the supply if the token in osmosis
	if strings.HasPrefix(denom, "ibc/") {
		return bankKeeper.GetSupplyWithOffset(ctx, denom).Amount
	}

	return bankKeeper.GetSupplyWithOffset(ctx, denom).Amount

	// ToDo: The commented-out code bellow is what we want to happen, but we're temporarily
	//  using the whole supply for efficiency until there's a solution for
	//  https://github.com/cosmos/ibc-go/issues/2664

	// For native tokens, obtain the balance held in escrow for all potential channels
	//channels := channelKeeper.GetAllChannels(ctx)
	//balance := sdk.NewInt(0)
	//for _, channel := range channels {
	//	escrowAddress := transfertypes.GetEscrowAddress("transfer", channel.ChannelId)
	//	balance = balance.Add(bankKeeper.GetBalance(ctx, escrowAddress, denom).Amount)
	//
	//}
	//return balance
}
