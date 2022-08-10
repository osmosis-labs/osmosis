package ibc_rate_limit

import (
	"encoding/json"
	"fmt"
	"strings"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	"github.com/osmosis-labs/osmosis/v10/x/ibc-rate-limit/types"
)

func CheckRateLimits(ctx sdk.Context, wasmKeeper *wasmkeeper.Keeper, msgType, contractParam, channelValue, sourceChannel string, sender sdk.AccAddress, amount string) error {
	contract := strings.Trim(contractParam, `"`) // ToDo: Why is this stored with ""
	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return err
	}

	sendPacketMsg := BuildWasmExecMsg(
		msgType,
		sourceChannel,
		channelValue,
		amount,
	)

	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(wasmKeeper)
	// ToDo: Why doesn't this return a response
	_, err = contractKeeper.Execute(ctx, contractAddr, sender, []byte(sendPacketMsg), sdk.Coins{})

	if err != nil {
		return sdkerrors.Wrap(types.ErrRateLimitExceeded, err.Error())
	}
	return nil
}

func BuildWasmExecMsg(msgType, sourceChannel, channelValue, amount string) string {
	// ToDo: Do this with a struct
	return fmt.Sprintf(
		`{"%s": {"channel_id": "%s", "channel_value": "%s", "funds": "%s"}}`,
		msgType,
		sourceChannel,
		channelValue,
		amount,
	)
}

func GetFundsFromPacket(packet exported.PacketI) (string, string, error) {
	var packetData map[string]interface{} // ToDo: Do this with a struct
	err := json.Unmarshal(packet.GetData(), &packetData)
	if err != nil {
		return "", "", err
	}
	denom, ok := packetData["denom"].(string)
	if !ok {
		return "", "", sdkerrors.Wrap(transfertypes.ErrInvalidAmount, "bad denom in packet")
	}
	amount, ok := packetData["amount"].(string)
	if !ok {
		return "", "", sdkerrors.Wrap(transfertypes.ErrInvalidAmount, "bad amount in packet")
	}
	return amount, denom, nil
}
