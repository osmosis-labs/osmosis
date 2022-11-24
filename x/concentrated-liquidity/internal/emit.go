package events

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/gamm/types"
)

// TODO: change as needed for CL rather than gamm
func EmitSwapEvent(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coin, output sdk.Coin) {
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		newSwapEvent(sender, poolId, input, output),
	})
}

func newSwapEvent(sender sdk.AccAddress, poolId uint64, input sdk.Coin, output sdk.Coin) sdk.Event {
	return sdk.NewEvent(
		types.TypeEvtTokenSwapped,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
		sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
		sdk.NewAttribute(types.AttributeKeyTokensIn, input.String()),
		sdk.NewAttribute(types.AttributeKeyTokensOut, output.String()),
	)
}

func EmitAddLiquidityEvent(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, liquidity sdk.Coins) {
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		newAddLiquidityEvent(sender, poolId, liquidity),
	})
}

func newAddLiquidityEvent(sender sdk.AccAddress, poolId uint64, liquidity sdk.Coins) sdk.Event {
	return sdk.NewEvent(
		types.TypeEvtPoolJoined,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
		sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
		sdk.NewAttribute(types.AttributeKeyTokensIn, liquidity.String()),
	)
}

func EmitRemoveLiquidityEvent(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, liquidity sdk.Coins) {
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		newRemoveLiquidityEvent(sender, poolId, liquidity),
	})
}

func newRemoveLiquidityEvent(sender sdk.AccAddress, poolId uint64, liquidity sdk.Coins) sdk.Event {
	return sdk.NewEvent(
		types.TypeEvtPoolExited,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
		sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
		sdk.NewAttribute(types.AttributeKeyTokensOut, liquidity.String()),
	)
}
