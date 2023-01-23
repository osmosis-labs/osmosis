package apptesting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	clmodel "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
)

var (
	ETH                       = "eth"
	USDC                      = "usdc"
	DefaultTickSpacing        = uint64(1)
	DefaultExponentAtPriceOne = sdk.NewInt(-4)
)

// PrepareConcentratedPool sets up an eth usdc concentrated liquidity pool with pool ID 1, tick spacing of 1,
// no liquidity and zero swap fee.
func (s *KeeperTestHelper) PrepareConcentratedPool() types.ConcentratedPoolExtension {
	return s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing, DefaultExponentAtPriceOne, sdk.ZeroDec())
}

// PrepareCustomConcentratedPool sets up a concentrated liquidity pool with the custom parameters.
func (s *KeeperTestHelper) PrepareCustomConcentratedPool(owner sdk.AccAddress, denom0, denom1 string, tickSpacing uint64, exponentAtPriceOne sdk.Int, swapFee sdk.Dec) types.ConcentratedPoolExtension {
	// Mint some assets to the account.
	s.FundAcc(s.TestAccs[0], DefaultAcctFunds)

	// Create a concentrated pool via the poolmanager
	poolID, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, clmodel.NewMsgCreateConcentratedPool(owner, denom0, denom1, tickSpacing, exponentAtPriceOne, swapFee))
	s.Require().NoError(err)

	// Retrieve the poolInterface via the poolID
	poolI, err := s.App.ConcentratedLiquidityKeeper.GetPool(s.Ctx, poolID)
	s.Require().NoError(err)

	// Type cast the PoolInterface to a ConcentratedPoolExtension
	pool, ok := poolI.(types.ConcentratedPoolExtension)
	s.Require().True(ok)

	return pool
}
