package concentrated_liquidity_test

import sdk "github.com/cosmos/cosmos-sdk/types"

func (s *KeeperTestSuite) TestGetPoolById() {
	s.SetupTest()

	pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, 1, "token0", "token1", sdk.NewDec(1), sdk.NewInt(1))
	s.Require().NoError(err)

	getPool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, pool.GetId())
	s.Require().NoError(err)

	// ensure that the pool is the same
	s.Require().Equal(pool.GetId(), getPool.GetId())
	s.Require().Equal(pool.GetAddress(), getPool.GetAddress())
	s.Require().Equal(pool.GetCurrentSqrtPrice(), getPool.GetCurrentSqrtPrice())
	s.Require().Equal(pool.GetCurrentTick(), getPool.GetCurrentTick())
	s.Require().Equal(pool.GetLiquidity(), getPool.GetLiquidity())

	// try getting invalid pool
	_, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, uint64(2))
	s.Require().Error(err)
}
