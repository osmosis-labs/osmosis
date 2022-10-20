package concentrated_liquidity_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	cl "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity"
)

func TestCalcOutAmtGivenIn(t *testing.T) {
	ctx := sdk.Context{}
	poolDenoms := []string{"eth", "usdc"}
	pool, err := cl.NewConcentratedLiquidityPool(1, poolDenoms)
	require.NoError(t, err)

	// test asset a to b logic
	tokenIn := sdk.NewCoin("eth", sdk.NewInt(133700))
	tokenOutDenom := "usdc"
	swapFee := sdk.NewDec(0)

	amountOut, err := pool.CalcOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, swapFee)
	require.NoError(t, err)
	require.Equal(t, sdk.NewDec(663944647).String(), amountOut.Amount.ToDec().String())

	// test asset b to a logic
	tokenIn = sdk.NewCoin("usdc", sdk.NewInt(4199999999))
	tokenOutDenom = "eth"
	swapFee = sdk.NewDec(0)

	amountOut, err = pool.CalcOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, swapFee)
	require.NoError(t, err)
	require.Equal(t, sdk.NewDec(805287), amountOut.Amount.ToDec())

	// test with swap fee
	tokenIn = sdk.NewCoin("usdc", sdk.NewInt(4199999999))
	tokenOutDenom = "eth"
	swapFee = sdk.NewDecWithPrec(2, 2)

	amountOut, err = pool.CalcOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, swapFee)
	require.NoError(t, err)
	require.Equal(t, sdk.NewDec(789834), amountOut.Amount.ToDec())
}

func TestCalcInAmtGivenOut(t *testing.T) {
	ctx := sdk.Context{}
	poolDenoms := []string{"eth", "usdc"}
	pool, err := cl.NewConcentratedLiquidityPool(1, poolDenoms)
	require.NoError(t, err)

	// test asset a to b logic
	tokensOut := sdk.NewCoin("usdc", sdk.NewInt(4199999999))
	tokenInDenom := "eth"
	swapFee := sdk.NewDec(0)

	amountIn, err := pool.CalcInAmtGivenOut(ctx, tokensOut, tokenInDenom, swapFee)
	require.NoError(t, err)
	require.Equal(t, sdk.NewDec(805287), amountIn.Amount.ToDec())

	// test asset b to a logic
	tokensOut = sdk.NewCoin("eth", sdk.NewInt(133700))
	tokenInDenom = "usdc"
	swapFee = sdk.NewDec(0)

	amountIn, err = pool.CalcInAmtGivenOut(ctx, tokensOut, tokenInDenom, swapFee)
	require.NoError(t, err)
	require.Equal(t, sdk.NewDec(663944647), amountIn.Amount.ToDec())

	// test asset a to b logic
	tokensOut = sdk.NewCoin("usdc", sdk.NewInt(4199999999))
	tokenInDenom = "eth"
	swapFee = sdk.NewDecWithPrec(2, 2)

	amountIn, err = pool.CalcInAmtGivenOut(ctx, tokensOut, tokenInDenom, swapFee)
	require.NoError(t, err)
	require.Equal(t, sdk.NewDec(821721), amountIn.Amount.ToDec())
}
