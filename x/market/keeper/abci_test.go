package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/market"
)

func (s *KeeperTestSuite) TestABCIReplenishPools() {
	osmosisDelta := sdk.NewDecWithPrec(17987573223725367, 3)
	s.App.MarketKeeper.SetOsmosisPoolDelta(s.Ctx, osmosisDelta)

	for i := 0; i < 100; i++ {
		osmosisDelta = s.App.MarketKeeper.GetOsmosisPoolDelta(s.Ctx)

		poolRecoveryPeriod := int64(s.App.MarketKeeper.PoolRecoveryPeriod(s.Ctx))
		osmosisRegressionAmt := osmosisDelta.QuoInt64(poolRecoveryPeriod)

		market.EndBlocker(s.Ctx, *s.App.MarketKeeper)

		osmosisPoolDelta := s.App.MarketKeeper.GetOsmosisPoolDelta(s.Ctx)
		s.Require().Equal(osmosisDelta.Sub(osmosisRegressionAmt), osmosisPoolDelta)
	}
}
