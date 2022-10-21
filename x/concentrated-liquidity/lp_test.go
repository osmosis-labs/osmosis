package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestMint() {
	// testing params
	// current tick: 85176
	// lower tick: 84222
	// upper tick: 86129
	// liquidity(token in):1517882343751509868544
	// current sqrt price: 5602277097478614198912276234240
	// denom0: uosmo
	// denom1: usdc
	poolId := uint64(1)
	currentTick := sdk.NewInt(85176)
	lowerTick := int64(84222)
	upperTick := int64(86129)
	liquidity, ok := sdk.NewIntFromString("1517882343751509868544")
	s.Require().True(ok)
	currentSqrtP, ok := sdk.NewIntFromString("5602277097478614198912276234240")
	s.Require().True(ok)
	denom0 := "uosmo"
	denom1 := "usdc"

	s.SetupTest()

	s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, poolId, denom0, denom1, currentSqrtP, currentTick)

	_, _, err := s.App.ConcentratedLiquidityKeeper.Mint(s.Ctx, poolId, s.TestAccs[0], liquidity, lowerTick, upperTick)
	s.Require().NoError(err)
}
