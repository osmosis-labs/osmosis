package keeper_test

import (
	"testing"

	"github.com/osmosis-labs/osmosis/v23/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v23/app/params"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	// Set the bond denom to be uosmo to make volume tracking tests more readable.
	skParams := s.App.StakingKeeper.GetParams(s.Ctx)
	skParams.BondDenom = "uosmo"
	s.App.StakingKeeper.SetParams(s.Ctx, skParams)
	s.App.TxFeesKeeper.SetBaseDenom(s.Ctx, "uosmo")
	marketParams := s.App.MarketKeeper.GetParams(s.Ctx)
	s.App.MarketKeeper.SetParams(s.Ctx, marketParams)
}

func (s *KeeperTestSuite) TestOsmosisPoolDeltaUpdate() {
	terraPoolDelta := s.App.MarketKeeper.GetOsmosisPoolDelta(s.Ctx)
	s.Require().Equal(sdk.ZeroDec(), terraPoolDelta)

	diff := sdk.NewDec(10)
	s.App.MarketKeeper.SetOsmosisPoolDelta(s.Ctx, diff)

	terraPoolDelta = s.App.MarketKeeper.GetOsmosisPoolDelta(s.Ctx)
	s.Require().Equal(diff, terraPoolDelta)
}

// TestReplenishPools tests that
// each pools move towards base pool
func (s *KeeperTestSuite) TestReplenishPools() {
	s.App.OracleKeeper.SetLunaExchangeRate(s.Ctx, appparams.MicroSDRDenom, sdk.OneDec())

	basePool := s.App.MarketKeeper.BasePool(s.Ctx)
	terraPoolDelta := s.App.MarketKeeper.GetOsmosisPoolDelta(s.Ctx)
	s.Require().True(terraPoolDelta.IsZero())

	// Positive delta
	diff := basePool.QuoInt64((int64)(appparams.BlocksPerDay))
	s.App.MarketKeeper.SetOsmosisPoolDelta(s.Ctx, diff)

	s.App.MarketKeeper.ReplenishPools(s.Ctx)

	terraPoolDelta = s.App.MarketKeeper.GetOsmosisPoolDelta(s.Ctx)
	replenishAmt := diff.QuoInt64((int64)(s.App.MarketKeeper.PoolRecoveryPeriod(s.Ctx)))
	expectedDelta := diff.Sub(replenishAmt)
	s.Require().Equal(expectedDelta, terraPoolDelta)

	// Negative delta
	diff = diff.Neg()
	s.App.MarketKeeper.SetOsmosisPoolDelta(s.Ctx, diff)

	s.App.MarketKeeper.ReplenishPools(s.Ctx)

	osmosisPoolDelta := s.App.MarketKeeper.GetOsmosisPoolDelta(s.Ctx)
	replenishAmt = diff.QuoInt64((int64)(s.App.MarketKeeper.PoolRecoveryPeriod(s.Ctx)))
	expectedDelta = diff.Sub(replenishAmt)
	s.Require().Equal(expectedDelta, osmosisPoolDelta)
}
