package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v32/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v32/x/lockup/types"
)

// Direct callers, including superfluid, must skip idle perpetual gauges while
// non-perpetual gauges continue progressing toward their final epoch.
func (s *KeeperTestSuite) TestDistribute_SkipsIdlePerpetualGauges() {
	tests := []struct {
		name             string
		isPerpetual      bool
		amount           int64
		fullyDistributed bool
		expectedEpochs   uint64
		expectedPayout   int64
	}{
		{name: "empty perpetual", isPerpetual: true},
		{name: "fully distributed perpetual", isPerpetual: true, amount: 1000, fullyDistributed: true},
		{name: "dust perpetual", isPerpetual: true, amount: 100},
		{name: "funded perpetual", isPerpetual: true, amount: 1000, expectedEpochs: 1, expectedPayout: 1000},
		{name: "empty non-perpetual"},
		{name: "fully distributed non-perpetual", amount: 1000, fullyDistributed: true, expectedEpochs: 1},
		{name: "dust non-perpetual", amount: 100, expectedEpochs: 1},
		{name: "funded non-perpetual", amount: 1000, expectedEpochs: 1, expectedPayout: 100},
	}

	for _, synthetic := range []bool{false, true} {
		path := "regular"
		if synthetic {
			path = "synthetic"
		}
		s.Run(path, func() {
			for _, tc := range tests {
				s.Run(tc.name, func() {
					s.SetupTest()
					k := s.App.IncentivesKeeper
					k.SetParam(s.Ctx, types.KeyMinValueForDistr, sdk.NewInt64Coin(otherDenom, 1))

					denom := defaultLPDenom
					var owners []sdk.AccAddress
					if synthetic {
						denom = defaultLPSyntheticDenom
						owners = s.SetupUserSyntheticLocks([]userLocks{oneSyntheticLockupUser})
					} else {
						owners = s.SetupUserLocks([]userLocks{oneLockupUser})
					}

					coins := sdk.NewCoins()
					if tc.amount > 0 {
						coins = sdk.NewCoins(sdk.NewInt64Coin(otherDenom, tc.amount))
					}
					gaugeID, gauge := s.CreateGauge(tc.isPerpetual, s.TestAccs[0], coins, lockuptypes.QueryCondition{
						LockQueryType: lockuptypes.ByDuration,
						Denom:         denom,
						Duration:      defaultLockDuration,
					}, s.Ctx.BlockTime(), 10)
					if tc.fullyDistributed {
						gauge.DistributedCoins = gauge.Coins
						s.Require().NoError(k.SetGauge(s.Ctx, gauge))
					}
					s.Require().NoError(k.MoveUpcomingGaugeToActiveGauge(s.Ctx, *gauge))
					balanceBefore := s.App.BankKeeper.GetBalance(s.Ctx, owners[0], otherDenom)

					distributed, err := k.Distribute(s.Ctx, []types.Gauge{*gauge})
					s.Require().NoError(err)
					s.Require().Equal(osmomath.NewInt(tc.expectedPayout), distributed.AmountOf(otherDenom))
					updated, err := k.GetGaugeByID(s.Ctx, gaugeID)
					s.Require().NoError(err)
					s.Require().Equal(tc.expectedEpochs, updated.FilledEpochs)
					s.Require().Equal(gauge.DistributedCoins.AmountOf(otherDenom).AddRaw(tc.expectedPayout), updated.DistributedCoins.AmountOf(otherDenom))
					s.Require().Equal(balanceBefore.Amount.AddRaw(tc.expectedPayout), s.App.BankKeeper.GetBalance(s.Ctx, owners[0], otherDenom).Amount)
					s.Require().Contains(gaugeIDs(k.GetActiveGauges(s.Ctx)), gaugeID)
				})
			}
		})
	}
}
