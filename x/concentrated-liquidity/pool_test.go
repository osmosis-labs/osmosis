package concentrated_liquidity_test

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestCalcOutAmtGivenIn() {
	ctx := s.Ctx
	pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, 1, "eth", "usdc", sdk.NewInt(0), sdk.NewInt(0))
	s.Require().NoError(err)

	// test asset a to b logic
	tokenIn := sdk.NewCoin("eth", sdk.NewInt(133700))
	tokenOutDenom := "usdc"
	swapFee := sdk.NewDec(0)

	amountOut, err := pool.CalcOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, swapFee)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewDec(663944647).String(), amountOut.Amount.ToDec().String())

	// test asset b to a logic
	tokenIn = sdk.NewCoin("usdc", sdk.NewInt(4199999999))
	tokenOutDenom = "eth"
	swapFee = sdk.NewDec(0)

	amountOut, err = pool.CalcOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, swapFee)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewDec(805287), amountOut.Amount.ToDec())

	// test with swap fee
	tokenIn = sdk.NewCoin("usdc", sdk.NewInt(4199999999))
	tokenOutDenom = "eth"
	swapFee = sdk.NewDecWithPrec(2, 2)

	amountOut, err = pool.CalcOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, swapFee)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewDec(789834), amountOut.Amount.ToDec())
}

func (s *KeeperTestSuite) TestCalcInAmtGivenOut() {
	ctx := s.Ctx
	fmt.Println("===0")
	pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, 1, "eth", "usdc", sdk.NewInt(0), sdk.NewInt(0))
	s.Require().NoError(err)
	// test asset a to b logic
	tokensOut := sdk.NewCoin("usdc", sdk.NewInt(4199999999))
	tokenInDenom := "eth"
	swapFee := sdk.NewDec(0)

	amountIn, err := pool.CalcInAmtGivenOut(ctx, tokensOut, tokenInDenom, swapFee)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewDec(805287), amountIn.Amount.ToDec())

	// test asset b to a logic
	tokensOut = sdk.NewCoin("eth", sdk.NewInt(133700))
	tokenInDenom = "usdc"
	swapFee = sdk.NewDec(0)

	amountIn, err = pool.CalcInAmtGivenOut(ctx, tokensOut, tokenInDenom, swapFee)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewDec(663944647), amountIn.Amount.ToDec())

	// test asset a to b logic
	tokensOut = sdk.NewCoin("usdc", sdk.NewInt(4199999999))
	tokenInDenom = "eth"
	swapFee = sdk.NewDecWithPrec(2, 2)

	amountIn, err = pool.CalcInAmtGivenOut(ctx, tokensOut, tokenInDenom, swapFee)
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
