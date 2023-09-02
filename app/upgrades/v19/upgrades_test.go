package v19_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/app/apptesting"
	v19 "github.com/osmosis-labs/osmosis/v19/app/upgrades/v19"
	"github.com/osmosis-labs/osmosis/v19/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v19/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v19/x/gamm/types/migration"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v19/x/pool-incentives/types"

	"github.com/osmosis-labs/osmosis/v19/x/superfluid/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v19/x/superfluid/types"
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
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.Setup()
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	initialTokenBonded := osmomath.NewInt(100)
	s.Setup()

	var (
		superfluidVal              sdk.ValAddress
		lockDenom                  string
		delegationBeforeV18Upgrade stakingtypes.Validator
	)

	// Get max pool ID
	highestBalancerPoolId := uint64(0)
	highestCLPoolId := uint64(0)
	for _, record := range v19.Records {
		if record.BalancerPoolId > highestBalancerPoolId {
			highestBalancerPoolId = record.BalancerPoolId
		}
		if record.ClPoolId > highestCLPoolId {
			highestCLPoolId = record.ClPoolId
		}
	}

	// Create balancer pools
	for i := uint64(1); i <= highestBalancerPoolId; i++ {

		// 2 is a pool ID that does not exist in the linked balancer <> CL pool record.
		// The reason is that we must have a bond denom for superfluid delegation
		// but all balancer and CL pools to have the same ETH/USDC denoms.
		if i == 2 {
			// prepare superfluid delegation
			superfluidVal, lockDenom = s.setupSuperfluidDelegation()
			delegationBeforeV18Upgrade, _ = s.App.StakingKeeper.GetValidator(s.Ctx, superfluidVal)
		} else {
			s.PrepareCustomBalancerPool(
				[]balancer.PoolAsset{
					{
						Token:  sdk.NewCoin(apptesting.ETH, sdk.NewInt(100000000000)),
						Weight: sdk.NewInt(5),
					},
					{
						Token:  sdk.NewCoin(apptesting.USDC, sdk.NewInt(100000000000)),
						Weight: sdk.NewInt(5),
					},
				},
				balancer.PoolParams{
					SwapFee: sdk.ZeroDec(),
					ExitFee: sdk.ZeroDec(),
				},
			)
		}

	}

	// Create CL pools
	for i := uint64(highestBalancerPoolId + 1); i <= highestCLPoolId; i++ {
		s.PrepareConcentratedPool()
	}

	// Setup migration recods per mainnet state
	err := s.App.GAMMKeeper.ReplaceMigrationRecords(s.Ctx, []migration.BalancerToConcentratedPoolLink{
		{
			BalancerPoolId: 803,
			ClPoolId:       1136,
		},
	})
	s.Require().NoError(err)

	firstRecord := v19.Records[0]
	secondRecord := v19.Records[1]

	// Get CFMM gauges
	cfmmGauges, err := s.App.PoolIncentivesKeeper.GetGaugesForCFMMPool(s.Ctx, firstRecord.BalancerPoolId)
	s.Require().NoError(err)

	// Get longest gauge duration from CFMM pool.
	balancerLongestDurationGauge := cfmmGauges[0]
	for i := 1; i < len(cfmmGauges); i++ {
		if cfmmGauges[i].DistributeTo.Duration > balancerLongestDurationGauge.DistributeTo.Duration {
			balancerLongestDurationGauge = cfmmGauges[i]
		}
	}

	incentivesEpochDuration := s.App.IncentivesKeeper.GetEpochInfo(s.Ctx).Duration
	clGaugeIDSecondLink, err := s.App.PoolIncentivesKeeper.GetPoolGaugeId(s.Ctx, secondRecord.ClPoolId, incentivesEpochDuration)
	s.Require().NoError(err)

	// Setup distr records
	s.App.PoolIncentivesKeeper.SetDistrInfo(s.Ctx, poolincentivestypes.DistrInfo{
		TotalWeight: sdk.NewInt(1000),
		Records: []poolincentivestypes.DistrRecord{
			{
				GaugeId: balancerLongestDurationGauge.Id, // As a sanity check set distr record to balancer gauge id
				Weight:  sdk.NewInt(500),
			},
			{
				GaugeId: clGaugeIDSecondLink, // And to CL gauge ID, to make sure that there are no panics
				Weight:  sdk.NewInt(500),
			},
		},
	})

	// run an epoch
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Hour * 24))
	s.App.BeginBlocker(s.Ctx, abci.RequestBeginBlock{})

	synthLockedPreV18 := s.App.SuperfluidKeeper.GetTotalSyntheticAssetsLocked(s.Ctx, stakingSyntheticDenom(lockDenom, superfluidVal.String()))

	// run v18 upgrade
	// by doing this, we should be having incorrect state of superfluid staking accumulator
	s.runv18Upgrade()
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Hour * 24 * 7))
	s.App.BeginBlocker(s.Ctx, abci.RequestBeginBlock{})

	// broken states (current status):
	// synth lock accumulator is set to 0
	totalSynthLocked := s.App.SuperfluidKeeper.GetTotalSyntheticAssetsLocked(s.Ctx, stakingSyntheticDenom(lockDenom, superfluidVal.String()))
	s.Require().Equal(totalSynthLocked.String(), osmomath.ZeroInt().String())

	// superfluid delegated tokens have been undelegated from validator,
	// only have the initial bonded amount present
	delegationAfterV18Upgrade, _ := s.App.StakingKeeper.GetValidator(s.Ctx, superfluidVal)
	s.Require().True(delegationAfterV18Upgrade.Tokens.Equal(initialTokenBonded))

	// run v19 upgrade
	// this fix should set superfluid accumulators to the correct values
	s.runv19Upgrade()
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Hour * 24 * 7))
	s.App.BeginBlocker(s.Ctx, abci.RequestBeginBlock{})

	// synth lock accumulator should have been fixed after v19 upgrade,
	// and went back to normal state(pre-v18)
	synthLockAfterV19 := s.App.SuperfluidKeeper.GetTotalSyntheticAssetsLocked(s.Ctx, stakingSyntheticDenom(lockDenom, superfluidVal.String()))
	s.Require().True(synthLockAfterV19.Equal(synthLockedPreV18))

	// also check that we have the correct superfluid staked delegation back
	delegationAfterV19Upgrade, _ := s.App.StakingKeeper.GetValidator(s.Ctx, superfluidVal)
	s.Require().True(delegationBeforeV18Upgrade.Tokens.Equal(delegationAfterV19Upgrade.Tokens))

	// Validate that all migration records were set
	newMigrationRecords, err := s.App.GAMMKeeper.GetAllMigrationInfo(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(newMigrationRecords.BalancerToConcentratedPoolLinks, len(v19.Records))

	// Check that distr records point to CL gauges
	distrInfo := s.App.PoolIncentivesKeeper.GetDistrInfo(s.Ctx)
	s.Require().Len(distrInfo.Records, 2)

	clGaugeIDFirstLink, err := s.App.PoolIncentivesKeeper.GetPoolGaugeId(s.Ctx, firstRecord.ClPoolId, incentivesEpochDuration)
	s.Require().NoError(err)

	s.Require().Equal(clGaugeIDFirstLink, distrInfo.Records[0].GaugeId)
	s.Require().Equal(clGaugeIDSecondLink, distrInfo.Records[1].GaugeId)
}

func (s *UpgradeTestSuite) setupNormalDelegation() sdk.ValAddress {
	bondDenom := s.App.StakingKeeper.BondDenom(s.Ctx)
	vanillaVal := s.SetupValidator(stakingtypes.Bonded)
	val, found := s.App.StakingKeeper.GetValidator(s.Ctx, vanillaVal)
	s.Require().True(found)
	s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin(bondDenom, stakeAmt)))
	shares, err := s.App.StakingKeeper.Delegate(s.Ctx, s.TestAccs[0], stakeAmt, stakingtypes.Unbonded, val, true)
	s.Require().NoError(err)
	s.Require().True(shares.TruncateInt().Equal(stakeAmt))

	return vanillaVal
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
		AssetType: types.SuperfluidAssetTypeLPShare,
	})
	s.Require().NoError(err)

	unbondingDuration := s.App.StakingKeeper.GetParams(s.Ctx).UnbondingTime

	// set lockable duration so that we don't have erros upon creating gauge
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

func (suite *UpgradeTestSuite) runv18Upgrade() {
	suite.Ctx = suite.Ctx.WithBlockHeight(v18UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: "v18", Height: v18UpgradeHeight}
	err := suite.App.UpgradeKeeper.ScheduleUpgrade(suite.Ctx, plan)
	suite.Require().NoError(err)
	_, exists := suite.App.UpgradeKeeper.GetUpgradePlan(suite.Ctx)
	suite.Require().True(exists)

	suite.Ctx = suite.Ctx.WithBlockHeight(v18UpgradeHeight)
}

func (suite *UpgradeTestSuite) runv19Upgrade() {
	suite.Ctx = suite.Ctx.WithBlockHeight(v19UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: "v19", Height: v19UpgradeHeight}
	err := suite.App.UpgradeKeeper.ScheduleUpgrade(suite.Ctx, plan)
	suite.Require().NoError(err)
	_, exists := suite.App.UpgradeKeeper.GetUpgradePlan(suite.Ctx)
	suite.Require().True(exists)

	suite.Ctx = suite.Ctx.WithBlockHeight(v19UpgradeHeight)
}

func stakingSyntheticDenom(denom, valAddr string) string {
	return fmt.Sprintf("%s/superbonding/%s", denom, valAddr)
}
