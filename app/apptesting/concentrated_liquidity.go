package apptesting

import (
	clmodel "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

var (
	ETH                = "eth"
	USDC               = "usdc"
	DefaultTickSpacing = uint64(1)
)

// PrepareConcentratedPool sets up an eth usdc concentrated liquidity pool with pool ID 1, tick spacing of 1, and no liquidity
func (s *KeeperTestHelper) PrepareConcentratedPool() types.ConcentratedPoolExtension {
	// Mint some assets to the account.
	s.FundAcc(s.TestAccs[0], DefaultAcctFunds)

	// Create a concentrated pool via the swaprouter
	poolID, err := s.App.SwapRouterKeeper.CreatePool(s.Ctx, clmodel.NewMsgCreateConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing))
	s.Require().NoError(err)

	// Retrieve the poolInterface via the poolID
	poolI, err := s.App.ConcentratedLiquidityKeeper.GetPool(s.Ctx, poolID)
	s.Require().NoError(err)

	// Type cast the PoolInterface to a ConcentratedPoolExtension
	pool, ok := poolI.(types.ConcentratedPoolExtension)
	s.Require().True(ok)

	return pool
}
