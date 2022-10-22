package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestMint() {
	// testing params
	// current tick: 85176
	// lower tick: 84222
	// upper tick: 86129
	// liquidity(token in):1517882323
	// current sqrt price: 70710678
	// denom0: eth
	// denom1: usdc
	poolId := uint64(1)
	currentTick := sdk.NewInt(85176)
	lowerTick := int64(84222)
	upperTick := int64(86129)
	liquidity, ok := sdk.NewIntFromString("1517882323")
	s.Require().True(ok)
	currentSqrtP, ok := sdk.NewIntFromString("70710678")
	s.Require().True(ok)
	denom0 := "eth"
	denom1 := "usdc"

	s.SetupTest()

	s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, poolId, denom0, denom1, currentSqrtP, currentTick)

	asset0, asset1, err := s.App.ConcentratedLiquidityKeeper.Mint(s.Ctx, poolId, s.TestAccs[0], liquidity, lowerTick, upperTick)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(998629), asset0)     // .998629 ETH
	s.Require().Equal(sdk.NewInt(5000208942), asset1) // 5000.20 USDC
}
