package concentrated_liquidity_test

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cl "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity"
	cltypes "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func (s *KeeperTestSuite) TestCalcOutAmtGivenIn() {
	ctx := s.Ctx

	currPrice := sdk.NewDec(5000)
	currSqrtPrice, err := currPrice.ApproxSqrt()
	s.Require().NoError(err)
	currTick := cl.PriceToTick(currPrice)
	lowerPrice := sdk.NewDec(4545)
	s.Require().NoError(err)
	lowerTick := cl.PriceToTick(lowerPrice)
	upperPrice := sdk.NewDec(5500)
	s.Require().NoError(err)
	upperTick := cl.PriceToTick(upperPrice)

	amount0 := sdk.NewInt(1)
	amount1 := sdk.NewInt(5000)

	// create pool
	pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(ctx, 1, "eth", "usdc", currSqrtPrice, currTick)
	s.Require().NoError(err)

	// add position
	_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, pool.Id, s.TestAccs[0], amount0, amount1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
	s.Require().NoError(err)

	// swapping parameters used for test
	tokenIn := sdk.NewCoin("usdc", sdk.NewInt(42))
	tokenOutDenom := "eth"
	swapFee := sdk.ZeroDec()
	priceLimit := sdk.NewDec(5004)
	tokenIn, tokenOut, updatedTick, updatedLiquidity, err := s.App.ConcentratedLiquidityKeeper.CalcOutAmtGivenIn(s.Ctx, tokenIn, tokenOutDenom, swapFee, priceLimit, pool.Id)
	expectedTokenIn := sdk.NewCoin("usdc", sdk.NewInt(0))
	expectedTokenOut := sdk.NewCoin("eth", sdk.NewInt(42))

	s.Require().NoError(err)
	s.Require().True(expectedTokenIn.Equal(tokenIn))
	s.Require().True(expectedTokenOut.Equal(tokenOut))
	// TODO: come back to tick calc: shouldn't this be 85183?
	s.Require().Equal(sdk.NewInt(85184), updatedTick)

	// check pool liquidity
	lowerSqrtPrice, err := lowerPrice.ApproxSqrt()
	s.Require().NoError(err)
	upperSqrtPrice, err := upperPrice.ApproxSqrt()
	s.Require().NoError(err)
	expectedLiquidity := cl.GetLiquidityFromAmounts(currSqrtPrice, lowerSqrtPrice, upperSqrtPrice, amount0, amount1)
	s.Require().True(updatedLiquidity.TruncateDec().Equal(expectedLiquidity.TruncateDec()))
}

// func (s *KeeperTestSuite) TestCalcInAmtGivenOut() {
// 	ctx := s.Ctx
// 	pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, 1, "eth", "usdc", sdk.MustNewDecFromStr("70.710678"), sdk.NewInt(85176))
// 	s.Require().NoError(err)
// 	s.SetupPosition(pool.Id)

// 	// test asset a to b logic
// 	tokenOut := sdk.NewCoin("usdc", sdk.NewInt(4199999999))
// 	tokenInDenom := "eth"
// 	swapFee := sdk.NewDec(0)
// 	minPrice := sdk.NewDec(4500)
// 	maxPrice := sdk.NewDec(5500)

// 	amountIn, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// 	s.Require().NoError(err)
// 	s.Require().Equal(sdk.NewDec(805287), amountIn.Amount.ToDec())

// 	// test asset b to a logic
// 	tokenOut = sdk.NewCoin("eth", sdk.NewInt(133700))
// 	tokenInDenom = "usdc"
// 	swapFee = sdk.NewDec(0)

// 	amountIn, _, _, _, err = s.App.ConcentratedLiquidityKeeper.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// 	s.Require().NoError(err)
// 	s.Require().Equal(sdk.NewDec(666975610), amountIn.Amount.ToDec())

// 	// test asset a to b logic
// 	tokenOut = sdk.NewCoin("usdc", sdk.NewInt(4199999999))
// 	tokenInDenom = "eth"
// 	swapFee = sdk.NewDecWithPrec(2, 2)

// 	amountIn, _, _, _, err = s.App.ConcentratedLiquidityKeeper.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// 	s.Require().NoError(err)
// 	s.Require().Equal(sdk.NewDec(821722), amountIn.Amount.ToDec())
// }

// func (s *KeeperTestSuite) TestSwapInAmtGivenOut() {
// 	ctx := s.Ctx
// 	pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(ctx, 1, "eth", "usdc", sdk.MustNewDecFromStr("70.710678"), sdk.NewInt(85176))
// 	s.Require().NoError(err)
// 	fmt.Printf("%v pool liq pre \n", pool.Liquidity)
// 	lowerTick := int64(84222)
// 	upperTick := int64(86129)
// 	amount0Desired := sdk.NewInt(1)
// 	amount1Desired := sdk.NewInt(5000)

// 	s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, pool.Id, s.TestAccs[0], amount0Desired, amount1Desired, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick, upperTick)

// 	// test asset a to b logic
// 	tokenOut := sdk.NewCoin("usdc", sdk.NewInt(4199999999))
// 	tokenInDenom := "eth"
// 	swapFee := sdk.NewDec(0)
// 	minPrice := sdk.NewDec(4500)
// 	maxPrice := sdk.NewDec(5500)

// 	amountIn, err := s.App.ConcentratedLiquidityKeeper.SwapInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// 	s.Require().NoError(err)
// 	fmt.Printf("%v amountIn \n", amountIn)
// 	pool = s.App.ConcentratedLiquidityKeeper.GetPoolbyId(ctx, pool.Id)

// // test asset a to b logic
// tokenOut := sdk.NewCoin("usdc", sdk.NewInt(4199999999))
// tokenInDenom := "eth"
// swapFee := sdk.NewDec(0)
// minPrice := sdk.NewDec(4500)
// maxPrice := sdk.NewDec(5500)

// amountIn, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// s.Require().NoError(err)
// s.Require().Equal(sdk.NewDec(805287), amountIn.Amount.ToDec())

// // test asset b to a logic
// tokenOut = sdk.NewCoin("eth", sdk.NewInt(133700))
// tokenInDenom = "usdc"
// swapFee = sdk.NewDec(0)

// amountIn, _, _, _, err = s.App.ConcentratedLiquidityKeeper.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// s.Require().NoError(err)
// s.Require().Equal(sdk.NewDec(666975610), amountIn.Amount.ToDec())

// // test asset a to b logic
// tokenOut = sdk.NewCoin("usdc", sdk.NewInt(4199999999))
// tokenInDenom = "eth"
// swapFee = sdk.NewDecWithPrec(2, 2)

// amountIn, _, _, _, err = s.App.ConcentratedLiquidityKeeper.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// s.Require().NoError(err)
// s.Require().Equal(sdk.NewDec(821722), amountIn.Amount.ToDec())
// }

func (s *KeeperTestSuite) TestSwapOutAmtGivenIn() {
	ctx := s.Ctx
	pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(ctx, 1, "eth", "usdc", sdk.MustNewDecFromStr("70.710678"), sdk.NewInt(85176))
	s.Require().NoError(err)
	fmt.Printf("%v pool liq pre \n", pool.Liquidity)
	lowerTick := int64(84222)
	upperTick := int64(86129)
	amount0Desired := sdk.NewInt(1)
	amount1Desired := sdk.NewInt(5000)

	s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, pool.Id, s.TestAccs[0], amount0Desired, amount1Desired, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick, upperTick)
	pool = s.App.ConcentratedLiquidityKeeper.GetPoolbyId(ctx, pool.Id)
	fmt.Printf("%v pool liq post 1 \n", pool.Liquidity)

	// tokenIn := sdk.NewCoin("eth", sdk.NewInt(133700))
	// tokenOutDenom := "usdc"
	// swapFee := sdk.NewDec(0)
	// minPrice := sdk.NewDec(4500)
	// maxPrice := sdk.NewDec(5500)

	// this is a test case for swapping within the tick
	// amountIn, err := s.App.ConcentratedLiquidityKeeper.SwapOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, swapFee, minPrice, maxPrice, pool.Id)
	// s.Require().NoError(err)
	// fmt.Printf("%v amountIn \n", amountIn)
	// pool = s.App.ConcentratedLiquidityKeeper.GetPoolbyId(ctx, pool.Id)

	// // calculation for this is tested in TestCalcOutAmtGivenInt
	// s.Require().Equal(sdk.NewInt(666975610), amountIn.Amount)

	// s.Require().Equal(sdk.MustNewDecFromStr("1517.818895638265328110"), pool.Liquidity)
	// // curr sqrt price and tick remains the same
	// s.Require().Equal(sdk.MustNewDecFromStr("70.710678000000000000"), pool.CurrentSqrtPrice)
	// s.Require().Equal(sdk.NewInt(85176), pool.CurrentTick)
}

func (s *KeeperTestSuite) TestOrderInitialPoolDenoms() {
	denom0, denom1, err := cltypes.OrderInitialPoolDenoms("axel", "osmo")
	s.Require().NoError(err)
	s.Require().Equal(denom0, "axel")
	s.Require().Equal(denom1, "osmo")

	denom0, denom1, err = cltypes.OrderInitialPoolDenoms("usdc", "eth")
	s.Require().NoError(err)
	s.Require().Equal(denom0, "eth")
	s.Require().Equal(denom1, "usdc")

	denom0, denom1, err = cltypes.OrderInitialPoolDenoms("usdc", "usdc")
	s.Require().Error(err)

}

func (suite *KeeperTestSuite) TestPriceToTick() {
	testCases := []struct {
		name         string
		price        sdk.Dec
		tickExpected string
	}{
		{
			"happy path",
			sdk.NewDec(5000),
			"85176",
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			tick := cl.PriceToTick(tc.price)
			suite.Require().Equal(tc.tickExpected, tick.String())
		})
	}
}
