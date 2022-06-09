package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	TypeEvtPoolJoined   = "pool_joined"
	TypeEvtPoolExited   = "pool_exited"
	TypeEvtPoolCreated  = "pool_created"
	TypeEvtTokenSwapped = "token_swapped"
	TypeEvtSetSwapFee   = "set_swap_fee"
	TypeEvtSetExitFee   = "set_exit_fee"

	AttributeValueCategory = ModuleName
	AttributeKeyPoolId     = "pool_id"
	AttributeKeySwapFee    = "swap_fee"
	AttributeKeyTokensIn   = "tokens_in"
	AttributeKeyTokensOut  = "tokens_out"
)

func CreateSwapEvent(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) sdk.Event {
	return sdk.NewEvent(
		TypeEvtTokenSwapped,
		sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
		sdk.NewAttribute(AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
		sdk.NewAttribute(AttributeKeyTokensIn, input.String()),
		sdk.NewAttribute(AttributeKeyTokensOut, output.String()),
	)
}

func CreateAddLiquidityEvent(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, liquidity sdk.Coins) sdk.Event {
	return sdk.NewEvent(
		TypeEvtPoolJoined,
		sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
		sdk.NewAttribute(AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
		sdk.NewAttribute(AttributeKeyTokensIn, liquidity.String()),
	)
}

func CreateRemoveLiquidityEvent(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, liquidity sdk.Coins) sdk.Event {
	return sdk.NewEvent(
		TypeEvtPoolExited,
		sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
		sdk.NewAttribute(AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
		sdk.NewAttribute(AttributeKeyTokensOut, liquidity.String()),
	)
}
