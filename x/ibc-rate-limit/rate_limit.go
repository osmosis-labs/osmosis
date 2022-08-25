package ibc_rate_limit

import (
	"encoding/json"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	"github.com/osmosis-labs/osmosis/v11/x/ibc-rate-limit/types"
)

var (
	msgSend = "send_packet"
	msgRecv = "recv_packet"
)

type PacketData struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

func CheckRateLimits(ctx sdk.Context, contractKeeper *wasmkeeper.PermissionedKeeper,
	msgType, contract string,
	channelValue sdk.Int, sourceChannel, denom string,
	amount string,
) error {
	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return err
	}

	sendPacketMsg, err := BuildWasmExecMsg(
		msgType,
		sourceChannel,
		denom,
		channelValue,
		amount,
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

type SendPacketMsg struct {
	SendPacket RateLimitExecMsg `json:"send_packet"`
}

type RecvPacketMsg struct {
	RecvPacket RateLimitExecMsg `json:"recv_packet"`
}

type RateLimitExecMsg struct {
	ChannelId    string  `json:"channel_id"`
	Denom        string  `json:"denom"`
	ChannelValue sdk.Int `json:"channel_value"`
	Funds        string  `json:"funds"`
}

func BuildWasmExecMsg(msgType, sourceChannel, denom string, channelValue sdk.Int, amount string) ([]byte, error) {
	content := RateLimitExecMsg{
		ChannelId:    sourceChannel,
		Denom:        denom,
		ChannelValue: channelValue,
		Funds:        amount,
	}

	var (
		asJson []byte
		err    error
	)
	switch {
	case msgType == msgSend:
		msg := SendPacketMsg{SendPacket: content}
		asJson, err = json.Marshal(msg)
	case msgType == msgRecv:
		msg := RecvPacketMsg{RecvPacket: content}
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
	var packetData PacketData
	err := json.Unmarshal(packet.GetData(), &packetData)
	if err != nil {
		return "", "", err
	}
	return packetData.Amount, packetData.Denom, nil
}
