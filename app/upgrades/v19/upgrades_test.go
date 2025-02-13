package v19_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/header"
	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v29/app/apptesting"
	gammtypes "github.com/osmosis-labs/osmosis/v29/x/gamm/types"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"

	superfluidtypes "github.com/osmosis-labs/osmosis/v29/x/superfluid/types"
)

const (
	v18UpgradeHeight = 5
	v19UpgradeHeight = 10
)

var (
	stakeAmt = osmomath.NewInt(100000000)
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
	preModule appmodule.HasPreBlocker
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
	s.preModule = upgrade.NewAppModule(s.App.UpgradeKeeper, addresscodec.NewBech32Codec("osmo"))
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	initialTokenBonded := sdk.DefaultPowerReduction
	s.SetupTest()

	// prepare superfluid delegation
	superfluidVal, lockDenom := s.setupSuperfluidDelegation()
	delegationBeforeV18Upgrade, _ := s.App.StakingKeeper.GetValidator(s.Ctx, superfluidVal)

	// run an epoch
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Hour * 24))
	s.Require().NotPanics(func() {
		_, err := s.App.BeginBlocker(s.Ctx)
		s.Require().NoError(err)
	})

	synthLockedPreV18, err := s.App.SuperfluidKeeper.GetTotalSyntheticAssetsLocked(s.Ctx, stakingSyntheticDenom(lockDenom, superfluidVal.String()))
	s.Require().NoError(err)

	// run v18 upgrade
	// by doing this, we should be having incorrect state of superfluid staking accumulator
	s.runv18Upgrade()
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Hour * 24 * 7))
	s.Require().NotPanics(func() {
		_, err := s.preModule.PreBlock(s.Ctx)
		s.Require().NoError(err)
		_, err = s.App.BeginBlocker(s.Ctx)
		s.Require().NoError(err)
	})

	// broken states (current status):
	// synth lock accumulator is set to 0
	totalSynthLocked, err := s.App.SuperfluidKeeper.GetTotalSyntheticAssetsLocked(s.Ctx, stakingSyntheticDenom(lockDenom, superfluidVal.String()))
	s.Require().NoError(err)
	s.Require().True(totalSynthLocked.Equal(osmomath.ZeroInt()))

	// superfluid delegated tokens have been undelegated from validator,
	// only have the initial bonded amount present
	delegationAfterV18Upgrade, _ := s.App.StakingKeeper.GetValidator(s.Ctx, superfluidVal)
	s.Require().True(delegationAfterV18Upgrade.Tokens.Equal(initialTokenBonded))

	// run v19 upgrade
	// this fix should set superfluid accumulators to the correct values
	s.runv19Upgrade()
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Hour * 24 * 7))
	s.Require().NotPanics(func() {
		_, err := s.preModule.PreBlock(s.Ctx)
		s.Require().NoError(err)
		_, err = s.App.BeginBlocker(s.Ctx)
		s.Require().NoError(err)
	})

	// synth lock accumulator should have been fixed after v19 upgrade,
	// and went back to normal state(pre-v18)
	synthLockAfterV19, err := s.App.SuperfluidKeeper.GetTotalSyntheticAssetsLocked(s.Ctx, stakingSyntheticDenom(lockDenom, superfluidVal.String()))
	s.Require().NoError(err)
	s.Require().True(synthLockAfterV19.Equal(synthLockedPreV18))

	// also check that we have the correct superfluid staked delegation back
	delegationAfterV19Upgrade, _ := s.App.StakingKeeper.GetValidator(s.Ctx, superfluidVal)
	s.Require().True(delegationBeforeV18Upgrade.Tokens.Equal(delegationAfterV19Upgrade.Tokens))
}

func (s *UpgradeTestSuite) setupSuperfluidDelegation() (val sdk.ValAddress, lockDenom string) {
	// set up validator that would be used for superfluid staking
	superfluidVal := s.SetupValidator(stakingtypes.Bonded)

	// create single pool with bond denom
	pools := s.SetupGammPoolsWithBondDenomMultiplier([]osmomath.Dec{osmomath.NewDec(20)})

	// we only created one pool, we will use this pool for all the continued tests
	pool := pools[0]

	// register a LP token as a superfluid asset
	denom := gammtypes.GetPoolShareDenom(pool.GetId())
	err := s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, superfluidtypes.SuperfluidAsset{
		Denom:     denom,
		AssetType: superfluidtypes.SuperfluidAssetTypeLPShare,
	})
	s.Require().NoError(err)

	stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
	unbondingDuration := stakingParams.UnbondingTime

	// set lockable duration so that we don't have errors upon creating gauge
	s.App.IncentivesKeeper.SetLockableDurations(s.Ctx, []time.Duration{
		time.Hour * 24 * 14,
		time.Hour,
		time.Hour * 3,
		time.Hour * 7,
		unbondingDuration,
	})

	// lock the tokens
	lockId := s.LockTokens(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin(denom, stakeAmt)), unbondingDuration)
	lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lockId)
	s.Require().NoError(err)

	// now we superfluid delegate
	err = s.App.SuperfluidKeeper.SuperfluidDelegate(s.Ctx, lock.Owner, lock.ID, superfluidVal.String())
	s.Require().NoError(err)

	return superfluidVal, denom
}

func (s *UpgradeTestSuite) runv18Upgrade() {
	s.Ctx = s.Ctx.WithBlockHeight(v18UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: "v18", Height: v18UpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, err = s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithHeaderInfo(header.Info{Height: v18UpgradeHeight, Time: s.Ctx.BlockTime().Add(time.Second)}).WithBlockHeight(v18UpgradeHeight)
}

func (s *UpgradeTestSuite) runv19Upgrade() {
	s.Ctx = s.Ctx.WithBlockHeight(v19UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: "v19", Height: v19UpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, err = s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithHeaderInfo(header.Info{Height: v19UpgradeHeight, Time: s.Ctx.BlockTime().Add(time.Second)}).WithBlockHeight(v19UpgradeHeight)
}

func stakingSyntheticDenom(denom, valAddr string) string {
	return fmt.Sprintf("%s/superbonding/%s", denom, valAddr)
}
