package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

func (k Keeper) EndBlockLogic(ctx sdk.Context) error {
	// Run all end block logic we want here

	// Lets say I want to:
	// * swap 1 osmo through pool 1, to atom
	poolID := uint64(1)
	swapInput := sdk.NewCoin("uosmo", sdk.NewInt(1_000_000))
	tokenOutDenom := "uatom"
	tokenOutMinAmount := sdk.ZeroInt() // accept full slippage for example
	sendingAddress := sdk.AccAddress{}

	sentCoins, err := k.gammKeeper.SwapExactAmountIn(ctx, sendingAddress, poolID, swapInput, tokenOutDenom, tokenOutMinAmount)
	if err != nil {
		return err
	}

	// Hack to get around the golang required variables
	_ = sentCoins

	return nil
}
