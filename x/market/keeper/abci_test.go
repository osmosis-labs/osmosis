package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/market"
)

func (s *KeeperTestSuite) TestABCIReplenishPools() {
	symphonyDelta := sdk.NewDecWithPrec(17987573223725367, 3)
	s.App.MarketKeeper.SetOsmosisPoolDelta(s.Ctx, symphonyDelta)

	for i := 0; i < 100; i++ {
		symphonyDelta = s.App.MarketKeeper.GetOsmosisPoolDelta(s.Ctx)

		poolRecoveryPeriod := int64(s.App.MarketKeeper.PoolRecoveryPeriod(s.Ctx))
		symphonyRegressionAmt := symphonyDelta.QuoInt64(poolRecoveryPeriod)

		market.EndBlocker(s.Ctx, *s.App.MarketKeeper)

		sPoolDelta := s.App.MarketKeeper.GetOsmosisPoolDelta(s.Ctx)
		s.Require().Equal(symphonyDelta.Sub(symphonyRegressionAmt), sPoolDelta)
	}
}
