package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestCalcOutAmtGivenIn() {
	ctx := s.Ctx
	pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(ctx, 1, "eth", "usdc", sdk.NewInt(70710678), sdk.NewInt(85176))
	s.Require().NoError(err)

	// test asset a to b logic
	tokenIn := sdk.NewCoin("eth", sdk.NewInt(133700))
	tokenOutDenom := "usdc"
	swapFee := sdk.NewDec(0)
	minPrice := sdk.NewDec(4500)
	maxPrice := sdk.NewDec(5500)

	amountOut, err := pool.CalcOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, swapFee, minPrice, maxPrice)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewDec(663944645).String(), amountOut.Amount.ToDec().String())

	// test asset b to a logic
	tokenIn = sdk.NewCoin("usdc", sdk.NewInt(4199999999))
	tokenOutDenom = "eth"
	swapFee = sdk.NewDec(0)

	amountOut, err = pool.CalcOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, swapFee, minPrice, maxPrice)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewDec(805287), amountOut.Amount.ToDec())

	// test asset b to a logic
	tokenIn = sdk.NewCoin("usdc", sdk.NewInt(42000000))
	tokenOutDenom = "eth"
	swapFee = sdk.NewDec(0)

	amountOut, err = pool.CalcOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, swapFee, minPrice, maxPrice)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewDec(8396), amountOut.Amount.ToDec())

	// test with swap fee
	tokenIn = sdk.NewCoin("usdc", sdk.NewInt(4199999999))
	tokenOutDenom = "eth"
	swapFee = sdk.NewDecWithPrec(2, 2)

	amountOut, err = pool.CalcOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, swapFee, minPrice, maxPrice)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewDec(789834), amountOut.Amount.ToDec())
}

func (s *KeeperTestSuite) TestCalcInAmtGivenOut() {
	ctx := s.Ctx
	pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, 1, "eth", "usdc", sdk.NewInt(70710678), sdk.NewInt(85176))
	s.Require().NoError(err)
	// test asset a to b logic
	tokenOut := sdk.NewCoin("usdc", sdk.NewInt(4199999999))
	tokenInDenom := "eth"
	swapFee := sdk.NewDec(0)
	minPrice := sdk.NewDec(4500)
	maxPrice := sdk.NewDec(5500)

	amountIn, err := pool.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewDec(805287), amountIn.Amount.ToDec())

	// test asset b to a logic
	tokenOut = sdk.NewCoin("eth", sdk.NewInt(133700))
	tokenInDenom = "usdc"
	swapFee = sdk.NewDec(0)

	amountIn, err = pool.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewDec(663944645), amountIn.Amount.ToDec())

	// test asset a to b logic
	tokenOut = sdk.NewCoin("usdc", sdk.NewInt(4199999999))
	tokenInDenom = "eth"
	swapFee = sdk.NewDecWithPrec(2, 2)

	amountIn, err = pool.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewDec(821722), amountIn.Amount.ToDec())
}

func (s *KeeperTestSuite) TestOrderInitialPoolDenoms() {
	pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, 1, "eth", "usdc", sdk.NewInt(0), sdk.NewInt(0))
	s.Require().NoError(err)
	s.Require().Equal(pool.Token0, "eth")
	s.Require().Equal(pool.Token1, "usdc")

	err = pool.OrderInitialPoolDenoms("axel", "osmo")
	s.Require().NoError(err)
	s.Require().Equal(pool.Token0, "axel")
	s.Require().Equal(pool.Token1, "osmo")

	err = pool.OrderInitialPoolDenoms("usdc", "eth")
	s.Require().NoError(err)
	s.Require().Equal(pool.Token0, "eth")
	s.Require().Equal(pool.Token1, "usdc")

	err = pool.OrderInitialPoolDenoms("usdc", "usdc")
	s.Require().Error(err)

}
