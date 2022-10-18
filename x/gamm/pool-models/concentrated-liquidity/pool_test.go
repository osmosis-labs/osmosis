package concentrated_liquidity_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	cl "github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/concentrated-liquidity"
)

func TestCalcOutAmtGivenIn(t *testing.T) {
	ctx := sdk.Context{}
	pool, _ := cl.NewConcentratedLiquidityPool(1)
	tokensIn := sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(41999999)))
	tokenOutDenom := "testing"
	swapFee := sdk.NewDec(0)

	amountOut, _ := pool.CalcOutAmtGivenIn(ctx, tokensIn, tokenOutDenom, swapFee)
	require.Equal(t, sdk.NewDec(8396), amountOut.Amount.ToDec())
}

func TestCalcInAmtGivenOut(t *testing.T) {
	ctx := sdk.Context{}
	pool, _ := cl.NewConcentratedLiquidityPool(1)
	tokensOut := sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(41999999)))
	tokenInDenom := "testing"
	swapFee := sdk.NewDec(0)

	amountIn, _ := pool.CalcInAmtGivenOut(ctx, tokensOut, tokenInDenom, swapFee)
	require.Equal(t, sdk.NewDec(8396), amountIn.Amount.ToDec())
}
