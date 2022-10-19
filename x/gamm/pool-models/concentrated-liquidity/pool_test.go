package concentrated_liquidity_test

import (
	fmt "fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	cl "github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/concentrated-liquidity"
)

func TestCalcOutAmtGivenIn(t *testing.T) {
	ctx := sdk.Context{}
	pool, _ := cl.NewConcentratedLiquidityPool(1)
	tokensIn := sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(13370)))
	tokenOutDenom := "usdc"
	swapFee := sdk.NewDec(0)

	amountOut, _ := pool.CalcOutAmtGivenIn(ctx, tokensIn, tokenOutDenom, swapFee)
	fmt.Printf("%v", amountOut)
	require.Equal(t, sdk.NewDec(66), amountOut.Amount.ToDec())

	tokensIn = sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(41999999)))
	tokenOutDenom = "eth"
	swapFee = sdk.NewDec(0)

	amountOut, _ = pool.CalcOutAmtGivenIn(ctx, tokensIn, tokenOutDenom, swapFee)
	fmt.Printf("%v", amountOut)
	require.Equal(t, sdk.NewDec(8396), amountOut.Amount.ToDec())
}

func TestCalcInAmtGivenOut(t *testing.T) {
	ctx := sdk.Context{}
	pool, _ := cl.NewConcentratedLiquidityPool(1)
	tokensOut := sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(41999999)))
	tokenInDenom := "eth"
	swapFee := sdk.NewDec(0)

	amountIn, _ := pool.CalcInAmtGivenOut(ctx, tokensOut, tokenInDenom, swapFee)
	fmt.Printf("%v", amountIn)
	require.Equal(t, sdk.NewDec(8396), amountIn.Amount.ToDec())

	tokensOut = sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(13370)))
	tokenInDenom = "usdc"
	swapFee = sdk.NewDec(0)

	amountIn, _ = pool.CalcOutAmtGivenIn(ctx, tokensOut, tokenInDenom, swapFee)
	fmt.Printf("%v", amountIn)
	require.Equal(t, sdk.NewDec(66), amountIn.Amount.ToDec())

	tokensOut = sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(13370)))
	tokenInDenom = "usdc"
	swapFee = sdk.NewDecWithPrec(2, 2)

	amountIn, _ = pool.CalcOutAmtGivenIn(ctx, tokensOut, tokenInDenom, swapFee)
	fmt.Printf("%v", amountIn)
	require.Equal(t, sdk.NewDec(65), amountIn.Amount.ToDec())
}
