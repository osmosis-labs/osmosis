package apptesting

import (
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

var (
	ETH                = "eth"
	USDC               = "usdc"
	DefaultTickSpacing = uint64(1)
)

// PrepareConcentratedPool sets up an eth usdc concentrated liquidity pool with pool ID 1, tick spacing of 1, and no liquidity
func (s *KeeperTestHelper) PrepareConcentratedPool() types.ConcentratedPoolExtension {
	pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, 1, ETH, USDC, DefaultTickSpacing)
	s.Require().NoError(err)
	return pool
}
