package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

func (k Keeper) EndBlockLogic(ctx sdk.Context) error {
	// Run all end block logic we want here
	tokenToPool := make(map[string] []uint64)

	for i := 0; i < 100; i++ {
		pool, poolErr := k.gammKeeper.GetPool(ctx, uint64(i))
		if poolErr != nil {
			return poolErr
		}
		assetsInPool := k.gammKeeper.GetTotalLiquidity(ctx)

		curToken := assetsInPool.GetDenomByIndex(0)
		if tokenToPool[curToken] == nil {
			tokenToPool[curToken] = []uint64{uint64(i)}
		} else {
			curArray := tokenToPool[curToken]
			curArray = append(curArray, uint64(i))
			tokenToPool[curToken] = curArray
		}

		curToken = assetsInPool.GetDenomByIndex(1)
		if tokenToPool[curToken] == nil {
			tokenToPool[curToken] = []uint64{uint64(i)}
		} else {
			curArray := tokenToPool[curToken]
			curArray = append(curArray, uint64(i))
			tokenToPool[curToken] = curArray
		}
		_ = pool
	}

	// Lets say I want to:
	// * swap 1 osmo through pool 1, to atom
	startingAmount := sdk.NewInt(1_000_000)
	curToken := "uosmo"
	pools := [1]uint64{1}

	swapInput := sdk.NewCoin(curToken, startingAmount)
	tokenOutMinAmount := sdk.ZeroInt() // accept full slippage for example
	sendingAddress := sdk.AccAddress{}
	receivedAmount := startingAmount

	for i := 0; i < len(pools); i++ {
		pool, poolErr := k.gammKeeper.GetPool(ctx, pools[i])
		if poolErr != nil {
			return poolErr
		}
		assetsInPool := k.gammKeeper.GetTotalLiquidity(ctx)
		tokenOutDenom := ""
		if curToken == assetsInPool.GetDenomByIndex(0) {
			tokenOutDenom = assetsInPool.GetDenomByIndex(1)
		} else {
			tokenOutDenom = assetsInPool.GetDenomByIndex(0)
		}

		sentCoins, err := k.gammKeeper.SwapExactAmountIn(ctx, sendingAddress, pools[i], swapInput, tokenOutDenom, tokenOutMinAmount)
		if err != nil {
			return err
		}

		receivedAmount = sentCoins
		swapInput = sdk.NewCoin(tokenOutDenom, sentCoins)
		curToken = tokenOutDenom
		_ = pool
	}

	priceDifference := sdk.Int.Sub(startingAmount, receivedAmount)

	// Hack to get around the golang required variables
	_ = priceDifference
	_ = tokenToPool

	return nil
}
