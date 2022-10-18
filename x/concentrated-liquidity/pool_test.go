package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type testParams struct {
	wethbalance  int
	usdcbalance  int
	currentTick  sdk.Int
	lowerTick    sdk.Int
	upperTick    sdk.Int
	liquidity    sdk.Int
	currentSqrtP sdk.Int
}
