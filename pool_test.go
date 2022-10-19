package concentrated_liquidity_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	cl "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity"
)

func TestCalcOutAmtGivenIn(t *testing.T) {
	ctx := sdk.Context{}
	pool := cl.NewConcentratedLiquidityPool(1)

	// test asset a to b logic
	tokensIn := sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(133700)))
	tokenOutDenom := "usdc"
	swapFee := sdk.NewDec(0)

	amountOut, _ := pool.CalcOutAmtGivenIn(ctx, tokensIn, tokenOutDenom, swapFee)
	require.Equal(t, sdk.NewDec(664366278), amountOut.Amount.ToDec())

	// test asset b to a logic
	tokensIn = sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(4199999999)))
	tokenOutDenom = "eth"
	swapFee = sdk.NewDec(0)

	amountOut, _ = pool.CalcOutAmtGivenIn(ctx, tokensIn, tokenOutDenom, swapFee)
	require.Equal(t, sdk.NewDec(808398), amountOut.Amount.ToDec())

	// test with swap fee
	tokensIn = sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(4199999999)))
	tokenOutDenom = "eth"
	swapFee = sdk.NewDecWithPrec(2, 2)

	amountOut, _ = pool.CalcOutAmtGivenIn(ctx, tokensIn, tokenOutDenom, swapFee)
	require.Equal(t, sdk.NewDec(792827), amountOut.Amount.ToDec())
}

func TestCalcInAmtGivenOut(t *testing.T) {
	ctx := sdk.Context{}
	pool := cl.NewConcentratedLiquidityPool(1)

	// test asset a to b logic
	tokensOut := sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(4199999999)))
	tokenInDenom := "eth"
	swapFee := sdk.NewDec(0)

	amountIn, err := pool.CalcInAmtGivenOut(ctx, tokensOut, tokenInDenom, swapFee)
	require.NoError(t, err)
	require.Equal(t, sdk.NewDec(808398), amountIn.Amount.ToDec())

	// test asset b to a logic
	tokensOut = sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(133700)))
	tokenInDenom = "usdc"
	swapFee = sdk.NewDec(0)

	amountIn, err = pool.CalcInAmtGivenOut(ctx, tokensOut, tokenInDenom, swapFee)
	require.NoError(t, err)
	require.Equal(t, sdk.NewDec(664366278), amountIn.Amount.ToDec())

	// test asset a to b logic
	tokensOut = sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(4199999999)))
	tokenInDenom = "eth"
	swapFee = sdk.NewDecWithPrec(2, 2)

	amountIn, _ = pool.CalcInAmtGivenOut(ctx, tokensOut, tokenInDenom, swapFee)
	require.Equal(t, sdk.NewDec(824896), amountIn.Amount.ToDec())
}
