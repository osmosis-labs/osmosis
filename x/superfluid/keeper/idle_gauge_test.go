package keeper_test

import (
	"time"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v32/app/params"
)

// A fully paid perpetual gauge stays active and resumes distribution when
// superfluid transfers new delegation rewards into it at the next epoch.
func (s *KeeperTestSuite) TestDistributeSuperfluidGauges_ResumesAfterRewards() {
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, STAKE, 9999)
	valAddrs := s.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded})
	denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]osmomath.Dec{osmomath.NewDec(20)})
	owners, intermediaryAccs, _ := s.setupSuperfluidDelegations(valAddrs, []superfluidDelegation{{0, 0, 0, 1000000}}, denoms)
	gaugeID := intermediaryAccs[0].GaugeId
	k := s.App.IncentivesKeeper
	sf := s.App.SuperfluidKeeper

	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Minute))
	s.Require().NoError(k.AfterEpochEnd(s.Ctx, k.GetEpochInfo(s.Ctx).Identifier, 1))

	// Fund and distribute once so the gauge has a nonempty reward history.
	s.AllocateRewardsToValidator(valAddrs[0], osmomath.NewInt(20000))
	sf.AfterEpochStartBeginBlock(s.Ctx)
	paid, err := k.GetGaugeByID(s.Ctx, gaugeID)
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), paid.FilledEpochs)
	s.Require().False(paid.Coins.Empty())
	s.Require().True(paid.Coins.Sub(paid.DistributedCoins...).Empty())

	// With no new rewards, the shared distribution path must leave it untouched.
	sf.DistributeSuperfluidGauges(s.Ctx)
	idle, err := k.GetGaugeByID(s.Ctx, gaugeID)
	s.Require().NoError(err)
	s.Require().Equal(paid, idle)

	// The normal epoch path adds delegation rewards before checking for idleness.
	bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)
	before := s.App.BankKeeper.GetBalance(s.Ctx, owners[0], bondDenom)
	s.AllocateRewardsToValidator(valAddrs[0], osmomath.NewInt(20000))
	sf.AfterEpochStartBeginBlock(s.Ctx)
	refunded, err := k.GetGaugeByID(s.Ctx, gaugeID)
	s.Require().NoError(err)
	payout := refunded.DistributedCoins.Sub(paid.DistributedCoins...)
	s.Require().True(payout.AmountOf(bondDenom).IsPositive())
	s.Require().Equal(paid.FilledEpochs+1, refunded.FilledEpochs)
	s.Require().True(refunded.Coins.Sub(refunded.DistributedCoins...).Empty())
	s.Require().Equal(before.Amount.Add(payout.AmountOf(bondDenom)), s.App.BankKeeper.GetBalance(s.Ctx, owners[0], bondDenom).Amount)
	active := false
	for _, gauge := range k.GetActiveGauges(s.Ctx) {
		if gauge.Id == gaugeID {
			active = true
			break
		}
	}
	s.Require().True(active, "the perpetual gauge must remain active for future rewards")
}
