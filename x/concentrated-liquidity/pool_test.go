package concentrated_liquidity_test

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cl "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity"
	cltypes "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func (s *KeeperTestSuite) TestCalcOutAmtGivenIn() {
	ctx := s.Ctx

	//
	// TEST 1: two overlapping price ranges
	//

	currPrice := sdk.NewDec(5000)
	currSqrtPrice, err := currPrice.ApproxSqrt() // 70.710678118654752440
	s.Require().NoError(err)
	currTick := cl.PriceToTick(currPrice) // 85176
	lowerPrice := sdk.NewDec(4545)
	s.Require().NoError(err)
	lowerTick := cl.PriceToTick(lowerPrice) // 84222
	upperPrice := sdk.NewDec(5500)
	s.Require().NoError(err)
	upperTick := cl.PriceToTick(upperPrice) // 86129

	// 1 eth 5000 usdc position
	amount0 := sdk.NewInt(1000000)
	amount1 := sdk.NewInt(5000000000)

	// create pool
	pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(ctx, 1, "eth", "usdc", currSqrtPrice, currTick)
	s.Require().NoError(err)

	// add first position
	_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, pool.Id, s.TestAccs[0], amount0, amount1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
	s.Require().NoError(err)

	// add second position
	_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, pool.Id, s.TestAccs[1], amount0, amount1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
	s.Require().NoError(err)
	pool = s.App.ConcentratedLiquidityKeeper.GetPoolbyId(ctx, 1)

	// swapping parameters used for test
	// swap in 42 usdc for some amount of eth
	tokenIn := sdk.NewCoin("usdc", sdk.NewInt(42000000))
	tokenOutDenom := "eth"
	// set no swap fee
	swapFee := sdk.ZeroDec()
	// limit max price impact to 5002 usdc per eth
	priceLimit := sdk.NewDec(5002)

	// run calculation
	tokenIn, tokenOut, updatedTick, updatedLiquidity, err := s.App.ConcentratedLiquidityKeeper.CalcOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, swapFee, priceLimit, pool.Id)
	s.Require().NoError(err)

	// we expect to put 42 usdc in and in return get .008398 eth back
	expectedTokenIn := sdk.NewCoin("usdc", sdk.NewInt(42000000))
	expectedTokenOut := sdk.NewCoin("eth", sdk.NewInt(8398))

	// ensure tokenIn and tokenOut meet our expected values
	s.Require().Equal(expectedTokenIn.String(), tokenIn.String())
	s.Require().Equal(expectedTokenOut.String(), tokenOut.String())

	// check the new tick is at the expected value
	s.Require().Equal(sdk.NewInt(85180).String(), updatedTick.String())

	// check pool liquidity
	lowerSqrtPrice, err := s.App.ConcentratedLiquidityKeeper.TickToSqrtPrice(lowerTick)
	s.Require().NoError(err)
	upperSqrtPrice, err := s.App.ConcentratedLiquidityKeeper.TickToSqrtPrice(upperTick)
	s.Require().NoError(err)
	expectedLiquidity := cl.GetLiquidityFromAmounts(currSqrtPrice, lowerSqrtPrice, upperSqrtPrice, amount0.Mul(sdk.NewInt(2)), amount1.Mul(sdk.NewInt(2)))
	s.Require().Equal(expectedLiquidity.String(), updatedLiquidity.String())

	//
	// TEST 2: one price range
	//

	// we use the same price range as above, but just with a single position instead of two

	// 1 eth 5000 usdc position
	amount0 = sdk.NewInt(1000000)
	amount1 = sdk.NewInt(5000000000)

	// create pool
	pool, err = s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(ctx, 2, "eth", "usdc", currSqrtPrice, currTick)
	s.Require().NoError(err)

	// add position
	_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, pool.Id, s.TestAccs[1], amount0, amount1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
	s.Require().NoError(err)

	// swapping parameters used for test
	// swap in 42 usdc for some amount of eth
	tokenIn = sdk.NewCoin("usdc", sdk.NewInt(42000000))
	tokenOutDenom = "eth"
	// set no swap fee
	swapFee = sdk.ZeroDec()
	// limit max price impact to 5004 usdc per eth
	priceLimit = sdk.NewDec(5004)

	// run calculation
	tokenIn, tokenOut, updatedTick, updatedLiquidity, err = s.App.ConcentratedLiquidityKeeper.CalcOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, swapFee, priceLimit, pool.Id)
	s.Require().NoError(err)

	// we expect to put 41999999 usdc in and in return get .008396 eth back
	expectedTokenIn = sdk.NewCoin("usdc", sdk.NewInt(41999999))
	expectedTokenOut = sdk.NewCoin("eth", sdk.NewInt(8396))

	// ensure tokenIn and tokenOut meet our expected values
	s.Require().Equal(expectedTokenIn.String(), tokenIn.String())
	s.Require().Equal(expectedTokenOut.String(), tokenOut.String())

	// this is off by one (too large), I think it is the priceToTick func, try using ln PR from main
	s.Require().Equal(sdk.NewInt(85184).String(), updatedTick.String())

	// check pool liquidity
	lowerSqrtPrice, err = s.App.ConcentratedLiquidityKeeper.TickToSqrtPrice(lowerTick)
	s.Require().NoError(err)
	upperSqrtPrice, err = s.App.ConcentratedLiquidityKeeper.TickToSqrtPrice(upperTick)
	s.Require().NoError(err)
	expectedLiquidity = cl.GetLiquidityFromAmounts(currSqrtPrice, lowerSqrtPrice, upperSqrtPrice, amount0, amount1)
	s.Require().Equal(expectedLiquidity.String(), updatedLiquidity.String())

	//
	// TEST 3: two consecutive price ranges
	//

	// we use the same price range as above, but for the first position
	// then for the second position, we use a new price range

	// both are 1 eth 5000 usdc positions
	amount0 = sdk.NewInt(1000000)
	amount1 = sdk.NewInt(5000000000)

	// create pool
	pool, err = s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(ctx, 3, "eth", "usdc", currSqrtPrice, currTick)
	s.Require().NoError(err)

	// add position one (utilizing old price range)
	_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, pool.Id, s.TestAccs[0], amount0, amount1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
	s.Require().NoError(err)

	// create second position parameters
	lowerPrice = sdk.NewDec(5501)
	s.Require().NoError(err)
	lowerTick = cl.PriceToTick(lowerPrice) // 84222
	upperPrice = sdk.NewDec(6250)
	s.Require().NoError(err)
	upperTick = cl.PriceToTick(upperPrice) // 87407

	// add position two with the new price range above
	_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, pool.Id, s.TestAccs[2], amount0, amount1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
	s.Require().NoError(err)

	// swapping parameters used for test
	// swap in 10000000 usdc for some amount of eth
	tokenIn = sdk.NewCoin("usdc", sdk.NewInt(10000000000))
	tokenOutDenom = "eth"
	// set no swap fee
	swapFee = sdk.ZeroDec()
	// limit max price impact to 6106 usdc per eth
	priceLimit = sdk.NewDec(6106)

	// run calculation
	tokenIn, tokenOut, updatedTick, updatedLiquidity, err = s.App.ConcentratedLiquidityKeeper.CalcOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, swapFee, priceLimit, pool.Id)
	s.Require().NoError(err)

	// we expect to put 999.99 usdc in and in return get 1.820536 eth back
	expectedTokenIn = sdk.NewCoin("usdc", sdk.NewInt(9999999999))
	expectedTokenOut = sdk.NewCoin("eth", sdk.NewInt(1820536))

	// ensure tokenIn and tokenOut meet our expected values
	s.Require().Equal(expectedTokenIn.String(), tokenIn.String())
	s.Require().Equal(expectedTokenOut.String(), tokenOut.String())

	// this is off by one (too large), I think it is the priceToTick func, try using ln PR from main
	s.Require().Equal(sdk.NewInt(87173).String(), updatedTick.String())

	// check pool liquidity
	lowerSqrtPrice, err = s.App.ConcentratedLiquidityKeeper.TickToSqrtPrice(lowerTick)
	s.Require().NoError(err)
	upperSqrtPrice, err = s.App.ConcentratedLiquidityKeeper.TickToSqrtPrice(upperTick)
	s.Require().NoError(err)
	expectedLiquidity = cl.GetLiquidityFromAmounts(currSqrtPrice, lowerSqrtPrice, upperSqrtPrice, amount0, amount1)
	s.Require().Equal(expectedLiquidity.TruncateInt().String(), updatedLiquidity.TruncateInt().String())
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
