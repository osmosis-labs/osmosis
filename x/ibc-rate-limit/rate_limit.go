package ibc_rate_limit

import (
	"encoding/json"
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	"github.com/osmosis-labs/osmosis/v12/x/ibc-rate-limit/types"
)

var (
	msgSend = "send_packet"
	msgRecv = "recv_packet"
)

func CheckAndUpdateRateLimits(ctx sdk.Context, contractKeeper *wasmkeeper.PermissionedKeeper,
	msgType, contract string, packet channeltypes.Packet,
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

	fmt.Println(string(sendPacketMsg))

	r, err := contractKeeper.Sudo(ctx, contractAddr, sendPacketMsg)
	fmt.Println(r)
	if err != nil {
		return sdkerrors.Wrap(types.ErrRateLimitExceeded, err.Error())
	}
	return nil
}

//func CheckAndUpdateRateLimits(ctx sdk.Context, contractKeeper *wasmkeeper.PermissionedKeeper,
//	msgType, contract string,
//	channelValue sdk.Int, sourceChannel, denom string,
//	amount string,
//) error {
//	contractAddr, err := sdk.AccAddressFromBech32(contract)
//	if err != nil {
//		return err
//	}
//
//	sendPacketMsg, err := BuildWasmExecMsg(
//		msgType,
//		sourceChannel,
//		denom,
//		channelValue,
//		amount,
//	)
//	if err != nil {
//		return err
//	}
//
//	_, err = contractKeeper.Sudo(ctx, contractAddr, sendPacketMsg)
//	if err != nil {
//		return sdkerrors.Wrap(types.ErrRateLimitExceeded, err.Error())
//	}
//
//	return nil
//}

type UndoSendMsg struct {
	UndoSend UndoSendMsgContent `json:"undo_send"`
}

type UndoSendMsgContent struct {
	ChannelId string `json:"channel_id"`
	Denom     string `json:"denom"`
	Funds     string `json:"funds"`
}

func UndoSendRateLimit(ctx sdk.Context, contractKeeper *wasmkeeper.PermissionedKeeper,
	contract string,
	sourceChannel, denom string,
	amount string,
) error {
	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return err
	}
	msg := UndoSendMsg{UndoSend: UndoSendMsgContent{ChannelId: sourceChannel, Denom: denom, Funds: amount}}
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

func BuildWasmExecMsg(msgType string, packet channeltypes.Packet) ([]byte, error) {

	var packetData transfertypes.FungibleTokenPacketData
	err := json.Unmarshal(packet.GetData(), &packetData)
	if err != nil {
		return nil, err
	}

	unwrapped := UnwrappedPacket{
		Sequence:           packet.Sequence,
		SourcePort:         packet.SourcePort,
		SourceChannel:      packet.SourceChannel,
		DestinationPort:    packet.DestinationPort,
		DestinationChannel: packet.DestinationChannel,
		Data:               packetData,
		TimeoutHeight:      packet.TimeoutHeight,
		TimeoutTimestamp:   packet.TimeoutTimestamp,
	}

	var asJson []byte
	switch {
	case msgType == msgSend:
		msg := SendPacketMsg{SendPacket: PacketMsg{unwrapped}}
		asJson, err = json.Marshal(msg)
	case msgType == msgRecv:
		msg := RecvPacketMsg{RecvPacket: PacketMsg{unwrapped}}
		asJson, err = json.Marshal(msg)
	default:
		return []byte{}, types.ErrBadMessage
	}

	if err != nil {
		return []byte{}, err
	}

	return asJson, nil
}

// GetIBCDenom This is extracted from ibc/transfer and mostly unmodified
func GetIBCDenom(sourceChannel, destChannel, denom string) string {
	var denomTrace transfertypes.DenomTrace
	if transfertypes.ReceiverChainIsSource("transfer", sourceChannel, denom) {
		voucherPrefix := transfertypes.GetDenomPrefix("transfer", sourceChannel)
		unprefixedDenom := denom[len(voucherPrefix):]
		// The denomination used to send the coins is either the native denom or the hash of the path
		// if the denomination is not native.
		denomTrace = transfertypes.ParseDenomTrace(unprefixedDenom)
	} else {
		// since SendPacket did not prefix the denomination, we must prefix denomination here
		sourcePrefix := transfertypes.GetDenomPrefix("transfer", destChannel)
		// NOTE: sourcePrefix contains the trailing "/"
		prefixedDenom := sourcePrefix + denom

		// construct the denomination trace from the full raw denomination
		denomTrace = transfertypes.ParseDenomTrace(prefixedDenom)
	}

	return denomTrace.IBCDenom()
}

func GetFundsFromPacket(packet exported.PacketI) (amount, packetDenom, localDenom, ibcDenom string, error error) {
	var packetData transfertypes.FungibleTokenPacketData
	err := json.Unmarshal(packet.GetData(), &packetData)
	if err != nil {
		return "", "", "", "", err
	}
	ibcDenom = GetIBCDenom(packet.GetSourceChannel(), packet.GetDestChannel(), packetData.Denom)
	return packetData.Amount, packetData.Denom, "", ibcDenom, nil
}
