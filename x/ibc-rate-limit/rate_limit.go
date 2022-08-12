package ibc_rate_limit

import (
	"encoding/json"
	"strings"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	"github.com/osmosis-labs/osmosis/v10/x/ibc-rate-limit/types"
)

func CheckRateLimits(ctx sdk.Context, wasmKeeper *wasmkeeper.Keeper,
	msgType, contractParam string,
	channelValue sdk.Int, sourceChannel string,
	sender sdk.AccAddress, amount string) error {
	contract := strings.Trim(contractParam, `"`) // ToDo: Why is this stored with ""
	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return err
	}

	sendPacketMsg, _ := BuildWasmExecMsg(
		msgType,
		sourceChannel,
		channelValue,
		amount,
	)

	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(wasmKeeper)
	_, err = contractKeeper.Execute(ctx, contractAddr, sender, []byte(sendPacketMsg), sdk.Coins{})

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
	ChannelValue sdk.Int `json:"channel_value"`
	Funds        string  `json:"funds"`
}

func BuildWasmExecMsg(msgType, sourceChannel string, channelValue sdk.Int, amount string) (string, error) {
	content := RateLimitExecMsg{
		ChannelId:    sourceChannel,
		ChannelValue: channelValue,
		Funds:        amount,
	}

	var (
		asJson []byte
		err    error
	)
	switch {
	case msgType == "send_packet":
		msg := SendPacketMsg{SendPacket: content}
		asJson, err = json.Marshal(msg)
	case msgType == "recv_packet":
		msg := RecvPacketMsg{RecvPacket: content}
		asJson, err = json.Marshal(msg)
	default:
		return "", types.BadMessage
	}

	if err != nil {
		return "", err
	}

	return string(asJson), nil
}

type PacketData struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

func GetFundsFromPacket(packet exported.PacketI) (string, string, error) {
	var packetData PacketData
	err := json.Unmarshal(packet.GetData(), &packetData)
	if err != nil {
		return "", "", err
	}
	return packetData.Amount, packetData.Denom, nil
}
