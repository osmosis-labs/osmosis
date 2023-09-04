package keeper_test

import (
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	appParams "github.com/osmosis-labs/osmosis/v19/app/params"
	"github.com/osmosis-labs/osmosis/v19/x/incentives/types"
	incentivetypes "github.com/osmosis-labs/osmosis/v19/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v19/x/lockup/types"
	poolincentivetypes "github.com/osmosis-labs/osmosis/v19/x/pool-incentives/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
)

var _ = suite.TestingSuite(nil)

type GroupGaugeCreationFields struct {
	coins            sdk.Coins
	numEpochPaidOver uint64
	owner            sdk.AccAddress
	internalGaugeIds []uint64
}

// TestDistribute tests that when the distribute command is executed on a provided gauge
// that the correct amount of rewards is sent to the correct lock owners.
func (s *KeeperTestSuite) TestDistribute() {
	defaultGauge := perpGaugeDesc{
		lockDenom:    defaultLPDenom,
		lockDuration: defaultLockDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 3000)},
	}
	doubleLengthGauge := perpGaugeDesc{
		lockDenom:    defaultLPDenom,
		lockDuration: 2 * defaultLockDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 3000)},
	}
	noRewardGauge := perpGaugeDesc{
		lockDenom:    defaultLPDenom,
		lockDuration: defaultLockDuration,
		rewardAmount: sdk.Coins{},
	}
	noRewardCoins := sdk.Coins{}
	oneKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 1000)}
	twoKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 2000)}
	threeKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 3000)}
	fiveKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 5000)}
	tests := []struct {
		name                 string
		users                []userLocks
		gauges               []perpGaugeDesc
		changeRewardReceiver []changeRewardReceiver
		expectedRewards      []sdk.Coins
	}{
		// gauge 1 gives 3k coins. three locks, all eligible. 1k coins per lock.
		// 1k should go to oneLockupUser and 2k to twoLockupUser.
		{
			name:            "One user with one lockup, another user with two lockups, single default gauge",
			users:           []userLocks{oneLockupUser, twoLockupUser},
			gauges:          []perpGaugeDesc{defaultGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, twoKRewardCoins},
		},
		// gauge 1 gives 3k coins. three locks, all eligible.
		// gauge 2 gives 3k coins. one lock, to twoLockupUser.
		// 1k should to oneLockupUser and 5k to twoLockupUser.
		{
			name:            "One user with one lockup (default gauge), another user with two lockups (double length gauge)",
			users:           []userLocks{oneLockupUser, twoLockupUser},
			gauges:          []perpGaugeDesc{defaultGauge, doubleLengthGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, fiveKRewardCoins},
		},
		// gauge 1 gives zero rewards.
		// both oneLockupUser and twoLockupUser should get no rewards.
		{
			name:            "One user with one lockup, another user with two lockups, both with no rewards gauge",
			users:           []userLocks{oneLockupUser, twoLockupUser},
			gauges:          []perpGaugeDesc{noRewardGauge},
			expectedRewards: []sdk.Coins{noRewardCoins, noRewardCoins},
		},
		// gauge 1 gives no rewards.
		// gauge 2 gives 3k coins. three locks, all eligible. 1k coins per lock.
		// 1k should to oneLockupUser and 2k to twoLockupUser.
		{
			name:            "One user with one lockup and another user with two lockups. No rewards and a default gauge",
			users:           []userLocks{oneLockupUser, twoLockupUser},
			gauges:          []perpGaugeDesc{noRewardGauge, defaultGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, twoKRewardCoins},
		},
		// gauge 1 gives 3k coins. three locks, all eligible. 1k coins per lock.
		// we change oneLockupUser lock's reward recepient to the twoLockupUser
		// none should go to oneLockupUser and 3k to twoLockupUser.
		{
			name:   "Change Reward Receiver: One user with one lockup, another user with two lockups, single default gauge",
			users:  []userLocks{oneLockupUser, twoLockupUser},
			gauges: []perpGaugeDesc{defaultGauge},
			changeRewardReceiver: []changeRewardReceiver{
				// change first lock's receiver address to the second account
				{
					lockId:              1,
					newReceiverAccIndex: 1,
				},
			},
			expectedRewards: []sdk.Coins{sdk.NewCoins(), threeKRewardCoins},
		},
		// gauge 1 gives 3k coins. three locks, all eligible. 1k coins per lock.
		// We change oneLockupUser's reward recepient to twoLockupUser, twoLockupUser's reward recepient to OneLockupUser.
		// Rewards should be reversed to the original test case, 2k should go to oneLockupUser and 1k to twoLockupUser.
		{
			name:   "Change Reward Receiver: One user with one lockup, another user with two lockups, single default gauge",
			users:  []userLocks{oneLockupUser, twoLockupUser},
			gauges: []perpGaugeDesc{defaultGauge},
			changeRewardReceiver: []changeRewardReceiver{
				// change first lock's receiver address to the second account
				{
					lockId:              1,
					newReceiverAccIndex: 1,
				},
				{
					lockId:              2,
					newReceiverAccIndex: 0,
				},
				{
					lockId:              3,
					newReceiverAccIndex: 0,
				},
			},
			expectedRewards: []sdk.Coins{twoKRewardCoins, oneKRewardCoins},
		},
		// gauge 1 gives 3k coins. three locks, all eligible.
		// gauge 2 gives 3k coins. one lock, to twoLockupUser.
		// Change all of oneLockupUser's reward recepient to twoLockupUser, vice versa.
		// Rewards should be reversed, 5k should to oneLockupUser and 1k to twoLockupUser.
		{
			name:   "Change Reward Receiver: One user with one lockup (default gauge), another user with two lockups (double length gauge)",
			users:  []userLocks{oneLockupUser, twoLockupUser},
			gauges: []perpGaugeDesc{defaultGauge, doubleLengthGauge},
			changeRewardReceiver: []changeRewardReceiver{
				{
					lockId:              1,
					newReceiverAccIndex: 1,
				},
				{
					lockId:              2,
					newReceiverAccIndex: 0,
				},
				{
					lockId:              3,
					newReceiverAccIndex: 0,
				},
			},
			expectedRewards: []sdk.Coins{fiveKRewardCoins, oneKRewardCoins},
		},
	}
	for _, tc := range tests {
		s.SetupTest()
		// setup gauges and the locks defined in the above tests, then distribute to them
		gauges := s.SetupGauges(tc.gauges, defaultLPDenom)
		addrs := s.SetupUserLocks(tc.users)

		// set up reward receiver if not nil
		if len(tc.changeRewardReceiver) != 0 {
			s.SetupChangeRewardReceiver(tc.changeRewardReceiver, addrs)
		}

		_, err := s.App.IncentivesKeeper.Distribute(s.Ctx, gauges)
		s.Require().NoError(err)
		// check expected rewards against actual rewards received
		for i, addr := range addrs {
			bal := s.App.BankKeeper.GetAllBalances(s.Ctx, addr)
			s.Require().Equal(tc.expectedRewards[i].String(), bal.String(), "test %v, person %d", tc.name, i)
		}
	}
}

func (s *KeeperTestSuite) TestDistribute_InternalIncentives_NoLock() {
	fiveKRewardCoins := sdk.NewInt64Coin(defaultRewardDenom, 5000)
	fiveKRewardCoinsUosmo := sdk.NewInt64Coin(appParams.BaseCoinUnit, 5000)
	fifteenKRewardCoins := sdk.NewInt64Coin(defaultRewardDenom, 15000)

	coinsToMint := sdk.NewCoins(sdk.NewCoin(defaultRewardDenom, osmomath.NewInt(10000000)), sdk.NewCoin(appParams.BaseCoinUnit, osmomath.NewInt(10000000)))
	defaultGaugeStartTime := s.Ctx.BlockTime()

	incentivesParams := s.App.IncentivesKeeper.GetParams(s.Ctx).DistrEpochIdentifier
	currentEpoch := s.App.EpochsKeeper.GetEpochInfo(s.Ctx, incentivesParams)

	tests := map[string]struct {
		// setup
		numPools           int
		tokensToAddToGauge sdk.Coins
		gaugeStartTime     time.Time
		gaugeCoins         sdk.Coins

		// expected
		expectErr             bool
		expectedDistributions sdk.Coins
	}{
		"valid case: one poolId and gaugeId": {
			numPools:              1,
			gaugeStartTime:        defaultGaugeStartTime,
			gaugeCoins:            sdk.NewCoins(fiveKRewardCoins),
			expectedDistributions: sdk.NewCoins(fiveKRewardCoins),
			expectErr:             false,
		},
		"valid case: gauge with multiple coins": {
			numPools:              1,
			gaugeStartTime:        defaultGaugeStartTime,
			gaugeCoins:            sdk.NewCoins(fiveKRewardCoins, fiveKRewardCoinsUosmo),
			expectedDistributions: sdk.NewCoins(fiveKRewardCoins, fiveKRewardCoinsUosmo),
			expectErr:             false,
		},
		"valid case: multiple gaugeId and poolId": {
			numPools:              3,
			gaugeStartTime:        defaultGaugeStartTime,
			gaugeCoins:            sdk.NewCoins(fiveKRewardCoins),
			expectedDistributions: sdk.NewCoins(fifteenKRewardCoins),
			expectErr:             false,
		},
		"valid case: one poolId and gaugeId, five 5000 coins": {
			numPools:              1,
			gaugeStartTime:        defaultGaugeStartTime,
			gaugeCoins:            sdk.NewCoins(fiveKRewardCoins),
			expectedDistributions: sdk.NewCoins(fiveKRewardCoins),
			expectErr:             false,
		},
		"valid case: attempt to createIncentiveRecord with start time < currentBlockTime - gets set to block time in incentive record": {
			numPools:              1,
			gaugeStartTime:        defaultGaugeStartTime.Add(-5 * time.Hour),
			gaugeCoins:            sdk.NewCoins(fiveKRewardCoins),
			expectedDistributions: sdk.NewCoins(fiveKRewardCoins),
			expectErr:             false,
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			// setup test
			s.SetupTest()
			// We fix blocktime to ensure tests are deterministic
			s.Ctx = s.Ctx.WithBlockTime(defaultGaugeStartTime)

			var gauges []types.Gauge

			// prepare the minting account
			addr := s.TestAccs[0]
			// mints coins so supply exists on chain
			s.FundAcc(addr, coinsToMint)

			// make sure the module has enough funds
			err := s.App.BankKeeper.SendCoinsFromAccountToModule(s.Ctx, addr, types.ModuleName, coinsToMint)
			s.Require().NoError(err)

			for i := 0; i < tc.numPools; i++ {
				var (
					poolId   uint64
					duration time.Duration
				)
				poolId = s.PrepareConcentratedPool().GetId()

				duration = currentEpoch.Duration

				// get the gaugeId corresponding to the CL pool
				gaugeId, err := s.App.PoolIncentivesKeeper.GetPoolGaugeId(s.Ctx, poolId, duration)
				s.Require().NoError(err)

				// get the gauge from the gaudeId
				gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeId)
				s.Require().NoError(err)

				gauge.Coins = tc.gaugeCoins
				gauge.StartTime = tc.gaugeStartTime
				gauges = append(gauges, *gauge)
			}

			// Distribute tokens from the gauge
			totalDistributedCoins, err := s.App.IncentivesKeeper.Distribute(s.Ctx, gauges)
			if tc.expectErr {
				s.Require().Error(err)

				// module account amount must stay the same
				balance := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName))
				s.Require().Equal(coinsToMint, balance)

				for i, gauge := range gauges {
					for j := range gauge.Coins {
						incentiveId := i*len(gauge.Coins) + j + 1

						// get poolId from GaugeId
						poolId, err := s.App.PoolIncentivesKeeper.GetPoolIdFromGaugeId(s.Ctx, gauge.GetId(), currentEpoch.Duration)
						s.Require().NoError(err)

						// check that incentive record wasn't created
						_, err = s.App.ConcentratedLiquidityKeeper.GetIncentiveRecord(s.Ctx, poolId, currentEpoch.Duration, uint64(incentiveId))
						s.Require().Error(err)
					}
				}
			} else {
				s.Require().NoError(err)

				// check that gauge is not empty
				s.Require().NotEqual(len(gauges), 0)

				// check if module amount got deducted correctly
				balance := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName))
				for _, coin := range balance {
					actualbalanceAfterDistribution := coinsToMint.AmountOf(coin.Denom).Sub(coin.Amount)
					s.Require().Equal(tc.expectedDistributions.AmountOf(coin.Denom).Add(osmomath.ZeroInt()), actualbalanceAfterDistribution.Add(osmomath.ZeroInt()))
				}

				for i, gauge := range gauges {
					for j, coin := range gauge.Coins {
						incentiveId := i*len(gauge.Coins) + j + 1

						gaugeId := gauge.GetId()

						// get poolId from GaugeId
						poolId, err := s.App.PoolIncentivesKeeper.GetPoolIdFromGaugeId(s.Ctx, gaugeId, currentEpoch.Duration)
						s.Require().NoError(err)

						// GetIncentiveRecord to see if pools received incentives properly
						incentiveRecord, err := s.App.ConcentratedLiquidityKeeper.GetIncentiveRecord(s.Ctx, poolId, types.DefaultConcentratedUptime, uint64(incentiveId))
						s.Require().NoError(err)

						expectedEmissionRate := osmomath.NewDecFromInt(coin.Amount).QuoTruncate(osmomath.NewDec(int64(currentEpoch.Duration.Seconds())))

						// Check that gauge distribution state is updated.
						s.ValidateDistributedGauge(gaugeId, 1, tc.gaugeCoins)

						// check every parameter in incentiveRecord so that it matches what we created
						incentiveRecordBody := incentiveRecord.GetIncentiveRecordBody()
						s.Require().Equal(poolId, incentiveRecord.PoolId)
						s.Require().Equal(expectedEmissionRate, incentiveRecordBody.EmissionRate)
						s.Require().Equal(s.Ctx.BlockTime().UTC().String(), incentiveRecordBody.StartTime.UTC().String())
						s.Require().Equal(types.DefaultConcentratedUptime, incentiveRecord.MinUptime)
						s.Require().Equal(coin.Amount, incentiveRecordBody.RemainingCoin.Amount.TruncateInt())
						s.Require().Equal(coin.Denom, incentiveRecordBody.RemainingCoin.Denom)
					}
				}
				// check the totalAmount of tokens distributed, for both lock gauges and CL pool gauges
				s.Require().Equal(tc.expectedDistributions, totalDistributedCoins)
			}
		})
	}
}

// TestDistribute_ExternalIncentives_NoLock tests the distribution of externally
// created NoLock gauges. It creates an external gauge with the correct configuration
// and uses it to attempt to distribute tokens to a concentrated liquidity pool.
// It attempts to distribute with all possible gauge configurations and with various tokens.
// However, it does not test distribution of NoLock gauges.
func (s *KeeperTestSuite) TestDistribute_ExternalIncentives_NoLock() {
	const (
		defaultCLPool       = uint64(1)
		defaultBalancerPool = uint64(2)

		defaultAmount = int64(5000)
	)

	fiveKRewardCoins := sdk.NewInt64Coin(defaultRewardDenom, defaultAmount)
	tenKOtherCoin := sdk.NewInt64Coin(otherDenom, defaultAmount+defaultAmount)

	defaultBothCoins := sdk.NewCoins(fiveKRewardCoins, tenKOtherCoin)

	defauBlockTime := time.Unix(123456789, 0)
	oneHourAfterDefault := defauBlockTime.Add(time.Hour)

	type test struct {
		// setup
		isPerpertual       bool
		tokensToAddToGauge sdk.Coins
		gaugeStartTime     time.Time
		gaugeCoins         sdk.Coins
		distrTo            lockuptypes.QueryCondition
		startTime          time.Time
		numEpochsPaidOver  uint64
		poolId             uint64

		// expected
		expectErr                              bool
		expectedDistributions                  sdk.Coins
		expectedRemainingAmountIncentiveRecord []sdk.Dec
	}

	defaultTest := test{
		isPerpertual:      false,
		gaugeStartTime:    defauBlockTime,
		gaugeCoins:        sdk.NewCoins(fiveKRewardCoins),
		distrTo:           lockuptypes.QueryCondition{LockQueryType: lockuptypes.NoLock},
		startTime:         oneHourAfterDefault,
		numEpochsPaidOver: 1,
		poolId:            defaultCLPool,
		expectErr:         false,

		expectedDistributions:                  sdk.NewCoins(fiveKRewardCoins),
		expectedRemainingAmountIncentiveRecord: []osmomath.Dec{osmomath.NewDec(defaultAmount)},
	}

	withIsPerpetual := func(tc test, isPerpetual bool) test {
		tc.isPerpertual = isPerpetual
		return tc
	}

	withGaugeCoins := func(tc test, gaugeCoins sdk.Coins) test {
		tc.gaugeCoins = gaugeCoins
		tc.expectedDistributions = gaugeCoins
		tc.expectedRemainingAmountIncentiveRecord = make([]sdk.Dec, len(gaugeCoins))
		for i := range tc.expectedRemainingAmountIncentiveRecord {
			tc.expectedRemainingAmountIncentiveRecord[i] = osmomath.NewDec(gaugeCoins[i].Amount.Int64())
		}
		return tc
	}

	withNumEpochs := func(tc test, numEpochs uint64) test {
		tc.numEpochsPaidOver = numEpochs
		if numEpochs == uint64(0) {
			return tc
		}

		// Do deep copies
		tempDistributions := make(sdk.Coins, len(tc.expectedDistributions))
		copy(tempDistributions, tc.expectedDistributions)

		tempRemainingAmountIncentiveRecord := make([]sdk.Dec, len(tc.expectedRemainingAmountIncentiveRecord))
		copy(tempRemainingAmountIncentiveRecord, tc.expectedRemainingAmountIncentiveRecord)

		for i := range tc.expectedRemainingAmountIncentiveRecord {
			// update expected distributions
			tempDistributions[i].Amount = tc.expectedDistributions[i].Amount.Quo(osmomath.NewInt(int64(numEpochs)))

			// update expected remaining amount in incentive record
			tempRemainingAmountIncentiveRecord[i] = tc.expectedRemainingAmountIncentiveRecord[i].QuoTruncate(osmomath.NewDec(int64(numEpochs))).TruncateDec()
		}

		tc.expectedDistributions = tempDistributions
		tc.expectedRemainingAmountIncentiveRecord = tempRemainingAmountIncentiveRecord
		return tc
	}

	withPoolId := func(tc test, poolId uint64) test {
		if poolId == defaultBalancerPool {
			// If we do not set it, SetPoolGaugeIdInternalIncentive(...) errors with
			// "zero duration is invalid"
			tc.distrTo.Duration = time.Hour
		}
		tc.poolId = poolId
		return tc
	}

	withError := func(tc test) test {
		tc.expectErr = true
		return tc
	}

	tests := map[string]test{
		"non-perpetual, 1 coin, paid over 1 epoch":   defaultTest,
		"perpetual, 1 coin, paid over 1 epoch":       withIsPerpetual(defaultTest, true),
		"non-perpetual, 2 coins, paid over 1 epoch":  withGaugeCoins(defaultTest, defaultBothCoins),
		"perpetual, 2 coins, paid over 1 epoch":      withIsPerpetual(withGaugeCoins(defaultTest, defaultBothCoins), true),
		"non-perpetual, 1 coin, paid over 2 epochs":  withNumEpochs(defaultTest, 2),
		"non-perpetual, 2 coins, paid over 3 epochs": withNumEpochs(withGaugeCoins(defaultTest, defaultBothCoins), 3),
		"error: balancer pool id":                    withError(withPoolId(defaultTest, defaultBalancerPool)),
		"error: inactive gauge":                      withError(withNumEpochs(defaultTest, 0)),
	}

	for name, tc := range tests {
		s.Run(name, func() {
			// setup test
			s.SetupTest()

			// We fix blocktime to ensure tests are deterministic
			s.Ctx = s.Ctx.WithBlockTime(defauBlockTime)

			// Create CL and Balancer pools
			s.PrepareConcentratedPool()
			s.PrepareBalancerPool()

			// Set block time one hour after block creation so that incentives logic
			// can function properly.
			s.Ctx = s.Ctx.WithBlockTime(oneHourAfterDefault)

			s.FundAcc(s.TestAccs[0], tc.gaugeCoins)

			// Create gauge and get it from state
			externalGaugeid, err := s.App.IncentivesKeeper.CreateGauge(s.Ctx, tc.isPerpertual, s.TestAccs[0], tc.gaugeCoins, tc.distrTo, tc.startTime, tc.numEpochsPaidOver, defaultCLPool)
			s.Require().NoError(err)
			externalGauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, externalGaugeid)
			s.Require().NoError(err)

			// Force gauge's pool id to balancer to trigger error
			if tc.poolId == defaultBalancerPool {
				err := s.App.PoolIncentivesKeeper.SetPoolGaugeIdInternalIncentive(s.Ctx, defaultBalancerPool, tc.distrTo.Duration, externalGaugeid)
				s.Require().NoError(err)
			}

			// Activate the gauge.
			err = s.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(s.Ctx, *externalGauge)
			s.Require().NoError(err)

			gauges := []types.Gauge{*externalGauge}

			// System under test.
			totalDistributedCoins, err := s.App.IncentivesKeeper.Distribute(s.Ctx, gauges)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				// check the totalAmount of tokens distributed, for both lock gauges and CL pool gauges
				s.Require().Equal(tc.expectedDistributions, totalDistributedCoins)

				incentivesEpochDuration := s.App.IncentivesKeeper.GetEpochInfo(s.Ctx).Duration
				incentivesEpochDurationSeconds := osmomath.NewDec(incentivesEpochDuration.Milliseconds()).QuoInt(osmomath.NewInt(1000))

				// Check that incentive records were created
				for i, coin := range tc.expectedDistributions {
					incentiveRecords, err := s.App.ConcentratedLiquidityKeeper.GetIncentiveRecord(s.Ctx, tc.poolId, time.Nanosecond, uint64(i+1))
					s.Require().NoError(err)

					expectedEmissionRatePerEpoch := coin.Amount.ToLegacyDec().QuoTruncate(incentivesEpochDurationSeconds)

					s.Require().Equal(tc.startTime.UTC(), incentiveRecords.IncentiveRecordBody.StartTime.UTC())
					s.Require().Equal(coin.Denom, incentiveRecords.IncentiveRecordBody.RemainingCoin.Denom)
					s.Require().Equal(tc.expectedRemainingAmountIncentiveRecord[i], incentiveRecords.IncentiveRecordBody.RemainingCoin.Amount)
					s.Require().Equal(expectedEmissionRatePerEpoch, incentiveRecords.IncentiveRecordBody.EmissionRate)
					s.Require().Equal(time.Nanosecond, incentiveRecords.MinUptime)
				}

				// Check that the gauge's distribution state was updated
				s.ValidateDistributedGauge(externalGaugeid, 1, tc.expectedDistributions)
			}
		})
	}
}

// TestSyntheticDistribute tests that when the distribute command is executed on a provided gauge
// the correct amount of rewards is sent to the correct synthetic lock owners.
func (s *KeeperTestSuite) TestSyntheticDistribute() {
	defaultGauge := perpGaugeDesc{
		lockDenom:    defaultLPSyntheticDenom,
		lockDuration: defaultLockDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 3000)},
	}
	doubleLengthGauge := perpGaugeDesc{
		lockDenom:    defaultLPSyntheticDenom,
		lockDuration: 2 * defaultLockDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 3000)},
	}
	noRewardGauge := perpGaugeDesc{
		lockDenom:    defaultLPSyntheticDenom,
		lockDuration: defaultLockDuration,
		rewardAmount: sdk.Coins{},
	}
	noRewardCoins := sdk.Coins{}
	oneKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 1000)}
	twoKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 2000)}
	fiveKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 5000)}
	tests := []struct {
		name            string
		users           []userLocks
		gauges          []perpGaugeDesc
		expectedRewards []sdk.Coins
	}{
		// gauge 1 gives 3k coins. three locks, all eligible. 1k coins per lock.
		// 1k should go to oneLockupUser and 2k to twoLockupUser.
		{
			name:            "One user with one synthetic lockup, another user with two synthetic lockups, both with default gauge",
			users:           []userLocks{oneSyntheticLockupUser, twoSyntheticLockupUser},
			gauges:          []perpGaugeDesc{defaultGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, twoKRewardCoins},
		},
		// gauge 1 gives 3k coins. three locks, all eligible.
		// gauge 2 gives 3k coins. one lock, to twoLockupUser.
		// 1k should to oneLockupUser and 5k to twoLockupUser.
		{
			name:            "One user with one synthetic lockup (default gauge), another user with two synthetic lockups (double length gauge)",
			users:           []userLocks{oneSyntheticLockupUser, twoSyntheticLockupUser},
			gauges:          []perpGaugeDesc{defaultGauge, doubleLengthGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, fiveKRewardCoins},
		},
		// gauge 1 gives zero rewards.
		// both oneLockupUser and twoLockupUser should get no rewards.
		{
			name:            "One user with one synthetic lockup, another user with two synthetic lockups, both with no rewards gauge",
			users:           []userLocks{oneSyntheticLockupUser, twoSyntheticLockupUser},
			gauges:          []perpGaugeDesc{noRewardGauge},
			expectedRewards: []sdk.Coins{noRewardCoins, noRewardCoins},
		},
		// gauge 1 gives no rewards.
		// gauge 2 gives 3k coins. three locks, all eligible. 1k coins per lock.
		// 1k should to oneLockupUser and 2k to twoLockupUser.
		{
			name:            "One user with one synthetic lockup (no rewards gauge), another user with two synthetic lockups (default gauge)",
			users:           []userLocks{oneSyntheticLockupUser, twoSyntheticLockupUser},
			gauges:          []perpGaugeDesc{noRewardGauge, defaultGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, twoKRewardCoins},
		},
	}
	for _, tc := range tests {
		s.SetupTest()
		// setup gauges and the synthetic locks defined in the above tests, then distribute to them
		gauges := s.SetupGauges(tc.gauges, defaultLPSyntheticDenom)
		addrs := s.SetupUserSyntheticLocks(tc.users)
		_, err := s.App.IncentivesKeeper.Distribute(s.Ctx, gauges)
		s.Require().NoError(err)
		// check expected rewards against actual rewards received
		for i, addr := range addrs {
			var rewards string
			bal := s.App.BankKeeper.GetAllBalances(s.Ctx, addr)
			// extract the superbonding tokens from the rewards distribution
			// TODO: figure out a less hacky way of doing this
			if strings.Contains(bal.String(), "lptoken/superbonding,") {
				rewards = strings.Split(bal.String(), "lptoken/superbonding,")[1]
			}
			s.Require().Equal(tc.expectedRewards[i].String(), rewards, "test %v, person %d", tc.name, i)
		}
	}
}

// TestGetModuleToDistributeCoins tests the sum of coins yet to be distributed for all of the module is correct.
func (s *KeeperTestSuite) TestGetModuleToDistributeCoins() {
	s.SetupTest()

	// check that the sum of coins yet to be distributed is nil
	coins := s.App.IncentivesKeeper.GetModuleToDistributeCoins(s.Ctx)
	s.Require().Equal(coins, sdk.Coins(nil))

	// setup a non perpetual lock and gauge
	_, gaugeID, gaugeCoins, startTime := s.SetupLockAndGauge(false)

	// check that the sum of coins yet to be distributed is equal to the newly created gaugeCoins
	coins = s.App.IncentivesKeeper.GetModuleToDistributeCoins(s.Ctx)
	s.Require().Equal(coins, gaugeCoins)

	// add coins to the previous gauge and check that the sum of coins yet to be distributed includes these new coins
	addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
	s.AddToGauge(addCoins, gaugeID)
	coins = s.App.IncentivesKeeper.GetModuleToDistributeCoins(s.Ctx)
	s.Require().Equal(coins, gaugeCoins.Add(addCoins...))

	// create a new gauge
	// check that the sum of coins yet to be distributed is equal to the gauge1 and gauge2 coins combined
	_, _, gaugeCoins2, _ := s.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 1000)})
	coins = s.App.IncentivesKeeper.GetModuleToDistributeCoins(s.Ctx)
	s.Require().Equal(coins, gaugeCoins.Add(addCoins...).Add(gaugeCoins2...))

	// move all created gauges from upcoming to active
	s.Ctx = s.Ctx.WithBlockTime(startTime)
	gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeID)
	s.Require().NoError(err)
	err = s.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(s.Ctx, *gauge)
	s.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := s.App.IncentivesKeeper.Distribute(s.Ctx, []types.Gauge{*gauge})
	s.Require().NoError(err)
	s.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 105)})

	// check gauge changes after distribution
	coins = s.App.IncentivesKeeper.GetModuleToDistributeCoins(s.Ctx)
	s.Require().Equal(coins, gaugeCoins.Add(addCoins...).Add(gaugeCoins2...).Sub(distrCoins))
}

// TestGetModuleDistributedCoins tests that the sum of coins that have been distributed so far for all of the module is correct.
func (s *KeeperTestSuite) TestGetModuleDistributedCoins() {
	s.SetupTest()

	// check that the sum of coins yet to be distributed is nil
	coins := s.App.IncentivesKeeper.GetModuleDistributedCoins(s.Ctx)
	s.Require().Equal(coins, sdk.Coins(nil))

	// setup a non perpetual lock and gauge
	_, gaugeID, _, startTime := s.SetupLockAndGauge(false)

	// check that the sum of coins yet to be distributed is equal to the newly created gaugeCoins
	coins = s.App.IncentivesKeeper.GetModuleDistributedCoins(s.Ctx)
	s.Require().Equal(coins, sdk.Coins(nil))

	// move all created gauges from upcoming to active
	s.Ctx = s.Ctx.WithBlockTime(startTime)
	gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeID)
	s.Require().NoError(err)
	err = s.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(s.Ctx, *gauge)
	s.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := s.App.IncentivesKeeper.Distribute(s.Ctx, []types.Gauge{*gauge})
	s.Require().NoError(err)
	s.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 5)})

	// check gauge changes after distribution
	coins = s.App.IncentivesKeeper.GetModuleToDistributeCoins(s.Ctx)
	s.Require().Equal(coins, distrCoins)
}

// TestByDurationPerpetualGaugeDistribution_NoLockNoOp tests that the creation of a perp gauge that has no locks associated does not distribute any tokens.
func (s *KeeperTestSuite) TestByDurationPerpetualGaugeDistribution_NoLockNoOp() {
	s.SetupTest()

	// setup a perpetual gauge with no associated locks
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	gaugeID, _, _, startTime := s.SetupNewGauge(true, coins)

	// ensure the created gauge has not completed distribution
	gauges := s.App.IncentivesKeeper.GetNotFinishedGauges(s.Ctx)
	s.Require().Len(gauges, 1)

	// ensure the not finished gauge matches the previously created gauge
	expectedGauge := types.Gauge{
		Id:          gaugeID,
		IsPerpetual: true,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             coins,
		NumEpochsPaidOver: 1,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	s.Require().Equal(gauges[0].String(), expectedGauge.String())

	// move the created gauge from upcoming to active
	s.Ctx = s.Ctx.WithBlockTime(startTime)
	gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeID)
	s.Require().NoError(err)
	err = s.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(s.Ctx, *gauge)
	s.Require().NoError(err)

	// distribute coins to stakers, since it's perpetual distribute everything on single distribution
	distrCoins, err := s.App.IncentivesKeeper.Distribute(s.Ctx, []types.Gauge{*gauge})
	s.Require().NoError(err)
	s.Require().Equal(distrCoins, sdk.Coins(nil))

	// check state is same after distribution
	gauges = s.App.IncentivesKeeper.GetNotFinishedGauges(s.Ctx)
	s.Require().Len(gauges, 1)
	s.Require().Equal(gauges[0].String(), expectedGauge.String())

	// Check that gauge distribution state is not updated.
	s.ValidateNotDistributedGauge(gaugeID)
}

// TestByDurationNonPerpetualGaugeDistribution_NoLockNoOp tests that the creation of a non perp gauge that has no locks associated does not distribute any tokens.
func (s *KeeperTestSuite) TestByDurationNonPerpetualGaugeDistribution_NoLockNoOp() {
	s.SetupTest()

	// setup non-perpetual gauge with no associated locks
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	gaugeID, _, _, startTime := s.SetupNewGauge(false, coins)

	// ensure the created gauge has not completed distribution
	gauges := s.App.IncentivesKeeper.GetNotFinishedGauges(s.Ctx)
	s.Require().Len(gauges, 1)

	// ensure the not finished gauge matches the previously created gauge
	expectedGauge := types.Gauge{
		Id:          gaugeID,
		IsPerpetual: false,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	s.Require().Equal(gauges[0].String(), expectedGauge.String())

	// move the created gauge from upcoming to active
	s.Ctx = s.Ctx.WithBlockTime(startTime)
	gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeID)
	s.Require().NoError(err)
	err = s.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(s.Ctx, *gauge)
	s.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := s.App.IncentivesKeeper.Distribute(s.Ctx, []types.Gauge{*gauge})
	s.Require().NoError(err)
	s.Require().Equal(distrCoins, sdk.Coins(nil))

	// check state is same after distribution
	gauges = s.App.IncentivesKeeper.GetNotFinishedGauges(s.Ctx)
	s.Require().Len(gauges, 1)
	s.Require().Equal(gauges[0].String(), expectedGauge.String())

	// Check that gauge distribution state is not updated.
	s.ValidateNotDistributedGauge(gaugeID)
}

func (s *KeeperTestSuite) TestGetPoolFromGaugeId() {
	const (
		poolIdOne   = uint64(1)
		validPoolId = poolIdOne
	)

	tests := []struct {
		name    string
		gaugeId uint64
		// For balancer pools, we do not create this link after pool
		// creation. As a result, for edge case testing we want
		// to manually set the link.
		shouldSetPoolGaugeId bool
		// this flag is necessary for edge case test where
		// there is a link between gauge and pool but the pool
		// does not exist. This case should not happen
		// in practice because the gauges must be created
		// after pool creation, via hook. However, we test
		// it for coverage.
		shouldAvoidCreatingPool bool
		expectedPoolType        poolmanagertypes.PoolType
		expectErr               bool
	}{
		{
			name:             "valid gaugeId and pool id link with concentrated pool",
			gaugeId:          poolIdOne,
			expectedPoolType: poolmanagertypes.Concentrated,
			expectErr:        false,
		},
		{
			name:                 "valid gaugeId and pool id link with balancer pool",
			gaugeId:              poolIdOne,
			expectedPoolType:     poolmanagertypes.Balancer,
			shouldSetPoolGaugeId: true,
			expectErr:            false,
		},
		{
			name:                 "invalid gaugeId and pool id link and balancer pool",
			gaugeId:              poolIdOne,
			expectedPoolType:     poolmanagertypes.Balancer,
			shouldSetPoolGaugeId: false,
			expectErr:            true,
		},
		{
			name:                    "valid gaugeId and pool id link with concentrated pool",
			gaugeId:                 poolIdOne,
			expectedPoolType:        poolmanagertypes.Concentrated,
			shouldAvoidCreatingPool: true,
			expectErr:               true,
		},
		{
			name:             "invalid gaugeId",
			gaugeId:          poolIdOne + 1,
			expectedPoolType: poolmanagertypes.Concentrated,
			expectErr:        true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()

			incParams := s.App.IncentivesKeeper.GetEpochInfo(s.Ctx)
			duration := incParams.GetDuration()

			if !tc.shouldAvoidCreatingPool {
				if tc.expectedPoolType == poolmanagertypes.Concentrated {
					s.PrepareConcentratedPool()
				} else {
					s.PrepareBalancerPool()
				}
			}

			if tc.shouldSetPoolGaugeId {
				err := s.App.PoolIncentivesKeeper.SetPoolGaugeIdInternalIncentive(s.Ctx, validPoolId, duration, poolIdOne)
				s.Require().NoError(err)
			}

			pool, err := s.App.IncentivesKeeper.GetPoolFromGaugeId(s.Ctx, tc.gaugeId, duration)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Nil(pool)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(pool)
				s.Require().Equal(pool.GetType(), tc.expectedPoolType)
			}
		})
	}
}

// TestFunctionalInternalExternalCLGauge is a functional test that covers more complex scenarios relating to distributing incentives through gauges
// at the end of each epoch.
//
// Testing strategy:
// 1. Initialize variables.
// 2. Setup CL pool and gauge (gauge automatically gets created at the end of CL pool creation).
// 3. Create external no-lock gauges for CL pools
// 4. Create Distribution records to incentivize internal CL no-lock gauges
// 5. let epoch 1 pass
//   - we only distribute external incentive in epoch 1.
//   - Check that incentive record has been correctly created and gauge has been correctly updated.
//   - all perpetual gauges must finish distributing records
//   - ClPool1 will recieve full 1Musdc, 1Meth in this epoch.
//   - ClPool2 will recieve 500kusdc, 500keth in this epoch.
//   - ClPool3 will recieve full 1Musdc, 1Meth in this epoch whereas
//
// 6. Remove distribution records for internal incentives using HandleReplacePoolIncentivesProposal
// 7. let epoch 2 pass
//   - We distribute internal incentive in epoch 2.
//   - check only external non-perpetual gauges with 2 epochs distributed
//   - check gauge has been correctly updated
//   - ClPool1 will already have 1Musdc, 1Meth (from epoch1) as external incentive. Will recieve 750Kstake as internal incentive.
//   - ClPool2 will already have 500kusdc, 500keth (from epoch1) as external incentive. Will recieve 500kusdc, 500keth (from epoch 2) as external incentive and 750Kstake as internal incentive.
//   - ClPool3 will already have 1M, 1M (from epoch1) as external incentive. This pool will not recieve any internal incentive.
//
// 8. let epoch 3 pass
//   - nothing distributes as non-perpetual gauges with 2 epochs have ended and perpetual gauges have not been reloaded
//   - nothing should change in terms of incentive records
func (s *KeeperTestSuite) TestFunctionalInternalExternalCLGauge() {
	// 1. Initialize variables
	s.SetupTest()
	const (
		defaultExternalGaugeValue int64 = 1_000_000
		defaultInternalGaugeValue int64 = 750_000
		numEpochsPaidOverGaugeTwo int64 = 2
	)
	var (
		epochInfo = s.App.IncentivesKeeper.GetEpochInfo(s.Ctx)

		requiredBalances         = sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(10_000_000)), sdk.NewCoin("usdc", osmomath.NewInt(10_000_000)))
		internalGaugeCoins       = sdk.NewCoins(sdk.NewCoin("stake", osmomath.NewInt(defaultInternalGaugeValue)))                                                                                                                    // distributed full sum at epoch
		externalGaugeCoins       = sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(defaultExternalGaugeValue)), sdk.NewCoin("usdc", osmomath.NewInt(defaultExternalGaugeValue)))                                                     // distributed full sum at epoch
		halfOfExternalGaugeCoins = sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(defaultExternalGaugeValue/numEpochsPaidOverGaugeTwo)), sdk.NewCoin("usdc", osmomath.NewInt(defaultExternalGaugeValue/numEpochsPaidOverGaugeTwo))) // distributed at each epoch for non-perp gauge with numEpoch = 2
	)

	s.FundAcc(s.TestAccs[1], requiredBalances)
	s.FundAcc(s.TestAccs[2], requiredBalances)
	s.FundModuleAcc(incentivetypes.ModuleName, requiredBalances)

	// 2. Setup CL pool and gauge (gauge automatically gets created at the end of CL pool creation).
	clPoolId1 := s.PrepareConcentratedPool() // creates internal no-lock gauge id = 1

	// check if the gauge is created
	clPoolInternalGaugeId1, err := s.App.PoolIncentivesKeeper.GetPoolGaugeId(s.Ctx, clPoolId1.GetId(), epochInfo.Duration)
	s.Require().NoError(err)

	clPoolId2 := s.PrepareConcentratedPool() // creates internal no-lock gauge id = 2

	// check if the gauge is created
	clPoolInternalGaugeId2, err := s.App.PoolIncentivesKeeper.GetPoolGaugeId(s.Ctx, clPoolId2.GetId(), epochInfo.Duration)
	s.Require().NoError(err)

	clPoolId3 := s.PrepareConcentratedPool() // creates internal no-lock gauge id = 3

	// check if the gauge is created
	clPoolInternalGaugeId3, err := s.App.PoolIncentivesKeeper.GetPoolGaugeId(s.Ctx, clPoolId3.GetId(), epochInfo.Duration)
	s.Require().NoError(err)

	// 3. Create external no-lock gauges for CL pools
	clPoolExternalGaugeIdPool1 := s.CreateNoLockExternalGauges(clPoolId1.GetId(), externalGaugeCoins, s.TestAccs[1], uint64(1))
	clPoolExternalGaugeIdPool2 := s.CreateNoLockExternalGauges(clPoolId2.GetId(), externalGaugeCoins, s.TestAccs[2], uint64(numEpochsPaidOverGaugeTwo))
	clPoolExternalGaugeIdPool3 := s.CreateNoLockExternalGauges(clPoolId3.GetId(), externalGaugeCoins, s.TestAccs[2], uint64(1))

	// 4. Create Distribution records to incentivize internal CL no-lock gauges
	// Note: We only internally incentivize ClPoolId1 and ClPoolId2
	s.IncentivizeInternalGauge([]uint64{clPoolId1.GetId(), clPoolId2.GetId()}, epochInfo.Duration, false)

	// 5. let epoch 1 pass
	// Note: we only distribute external incentive in epoch 1.
	// ******************** EPOCH 1 *********************
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(epochInfo.Duration))
	s.App.EpochsKeeper.AfterEpochEnd(s.Ctx, epochInfo.GetIdentifier(), 1)

	clPool1IncentiveRecordsAtEpoch1, err := s.App.ConcentratedLiquidityKeeper.GetAllIncentiveRecordsForPool(s.Ctx, clPoolId1.GetId())
	s.Require().NoError(err)

	clPool2IncentiveRecordsAtEpoch1, err := s.App.ConcentratedLiquidityKeeper.GetAllIncentiveRecordsForPool(s.Ctx, clPoolId2.GetId())
	s.Require().NoError(err)

	clPool3IncentiveRecordsAtEpoch1, err := s.App.ConcentratedLiquidityKeeper.GetAllIncentiveRecordsForPool(s.Ctx, clPoolId3.GetId())
	s.Require().NoError(err)

	// Validate Gauges
	// clPoolExternalGaugeIdPool1 expects full because the numEpochPaidOver is 1 for that gagueId
	// clPoolExternalGaugeIdPool2 expects half because the numEpochPaidOver is 2 for that gagueId
	s.ValidateDistributedGauge(clPoolExternalGaugeIdPool1, 1, externalGaugeCoins)
	s.ValidateDistributedGauge(clPoolExternalGaugeIdPool2, 1, halfOfExternalGaugeCoins)
	s.ValidateDistributedGauge(clPoolExternalGaugeIdPool3, 1, externalGaugeCoins)

	s.ValidateDistributedGauge(clPoolInternalGaugeId1, 1, sdk.Coins(nil))
	s.ValidateDistributedGauge(clPoolInternalGaugeId2, 1, sdk.Coins(nil))
	s.ValidateDistributedGauge(clPoolInternalGaugeId3, 1, sdk.Coins(nil))

	// check if incentives record got created.
	// Note: ClPool1 will recieve full 1Musdc, 1Meth in this epoch.
	s.Require().Equal(2, len(clPool1IncentiveRecordsAtEpoch1))
	s.Require().Equal(2, len(clPool2IncentiveRecordsAtEpoch1))
	s.Require().Equal(2, len(clPool3IncentiveRecordsAtEpoch1))

	s.ValidateIncentiveRecord(clPoolId1.GetId(), externalGaugeCoins[0], clPool1IncentiveRecordsAtEpoch1[0])
	s.ValidateIncentiveRecord(clPoolId1.GetId(), externalGaugeCoins[1], clPool1IncentiveRecordsAtEpoch1[1])

	// Note: ClPool2 will recieve 500kusdc, 500keth in this epoch.
	s.ValidateIncentiveRecord(clPoolId2.GetId(), halfOfExternalGaugeCoins[0], clPool2IncentiveRecordsAtEpoch1[0])
	s.ValidateIncentiveRecord(clPoolId2.GetId(), halfOfExternalGaugeCoins[1], clPool2IncentiveRecordsAtEpoch1[1])

	// Note: ClPool3 will recieve full 1Musdc, 1Meth in this epoch.
	// Note: emission rate is the same as CLPool1 because we are distributed same amount over 1 epoch.
	s.ValidateIncentiveRecord(clPoolId3.GetId(), externalGaugeCoins[0], clPool3IncentiveRecordsAtEpoch1[0])
	s.ValidateIncentiveRecord(clPoolId3.GetId(), externalGaugeCoins[1], clPool3IncentiveRecordsAtEpoch1[1])

	// 6. Remove distribution records for internal incentives using HandleReplacePoolIncentivesProposal
	s.IncentivizeInternalGauge([]uint64{clPoolId1.GetId(), clPoolId2.GetId()}, epochInfo.Duration, true)

	// 7. let epoch 2 pass
	// Note: we distribute internal incentive in epoch 2.
	// This is because at epoch 1 we first need to mint the tokens and distribute everything to the distr records. As a result,
	// internal gauges only get updated by epoch 2 and not one
	// ******************** EPOCH 2 *********************
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(epochInfo.Duration))
	s.App.EpochsKeeper.AfterEpochEnd(s.Ctx, epochInfo.GetIdentifier(), 2)

	clPool1IncentiveRecordsAtEpoch2, err := s.App.ConcentratedLiquidityKeeper.GetAllIncentiveRecordsForPool(s.Ctx, clPoolId1.GetId())
	s.Require().NoError(err)

	clPool2IncentiveRecordsAtEpoch2, err := s.App.ConcentratedLiquidityKeeper.GetAllIncentiveRecordsForPool(s.Ctx, clPoolId2.GetId())
	s.Require().NoError(err)

	clPool3IncentiveRecordsAtEpoch2, err := s.App.ConcentratedLiquidityKeeper.GetAllIncentiveRecordsForPool(s.Ctx, clPoolId3.GetId())
	s.Require().NoError(err)

	s.ValidateDistributedGauge(clPoolExternalGaugeIdPool1, 2, externalGaugeCoins)
	s.ValidateDistributedGauge(clPoolExternalGaugeIdPool2, 2, externalGaugeCoins)
	s.ValidateDistributedGauge(clPoolExternalGaugeIdPool3, 2, externalGaugeCoins)

	s.ValidateDistributedGauge(clPoolInternalGaugeId1, 2, internalGaugeCoins)
	s.ValidateDistributedGauge(clPoolInternalGaugeId2, 2, internalGaugeCoins)
	s.ValidateDistributedGauge(clPoolInternalGaugeId3, 2, sdk.Coins(nil))

	// check if incentives record got created.
	s.Require().Equal(3, len(clPool1IncentiveRecordsAtEpoch2))
	s.Require().Equal(5, len(clPool2IncentiveRecordsAtEpoch2))
	s.Require().Equal(2, len(clPool3IncentiveRecordsAtEpoch2))

	// Note: ClPool1 will recieve 1Musdc, 1Meth (from epoch1) as external incentive, 750Kstake as internal incentive.
	s.ValidateIncentiveRecord(clPoolId1.GetId(), externalGaugeCoins[0], clPool1IncentiveRecordsAtEpoch2[0])
	s.ValidateIncentiveRecord(clPoolId1.GetId(), externalGaugeCoins[1], clPool1IncentiveRecordsAtEpoch2[1])
	s.ValidateIncentiveRecord(clPoolId1.GetId(), internalGaugeCoins[0], clPool1IncentiveRecordsAtEpoch2[2])

	// Note: ClPool2 will recieve 500kusdc, 500keth (from epoch1) as external incentive, 500kusdc, 500keth (from epoch 2) as external incentive and 750Kstake as internal incentive.
	s.ValidateIncentiveRecord(clPoolId2.GetId(), halfOfExternalGaugeCoins[1], clPool2IncentiveRecordsAtEpoch2[0]) // new record
	s.ValidateIncentiveRecord(clPoolId2.GetId(), halfOfExternalGaugeCoins[0], clPool2IncentiveRecordsAtEpoch2[1]) // new record
	s.ValidateIncentiveRecord(clPoolId2.GetId(), halfOfExternalGaugeCoins[1], clPool2IncentiveRecordsAtEpoch2[2]) // new record
	s.ValidateIncentiveRecord(clPoolId2.GetId(), internalGaugeCoins[0], clPool2IncentiveRecordsAtEpoch2[3])       // old record
	s.ValidateIncentiveRecord(clPoolId2.GetId(), halfOfExternalGaugeCoins[0], clPool2IncentiveRecordsAtEpoch2[4]) // old record

	// all incentive for ClPoolId3 have already been distributed in epoch1. There is nothing left to distribute.
	s.Require().Equal(clPool3IncentiveRecordsAtEpoch1, clPool3IncentiveRecordsAtEpoch2)

	// 8. let epoch 3 pass
	// Note: All internal and external incentives have been distributed already.
	// Therefore we shouldn't distribue anything in this epoch.
	// ******************** EPOCH 3 *********************
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(epochInfo.Duration))
	s.App.EpochsKeeper.AfterEpochEnd(s.Ctx, epochInfo.GetIdentifier(), 3)

	clPool1IncentiveRecordsAtEpoch3, err := s.App.ConcentratedLiquidityKeeper.GetAllIncentiveRecordsForPool(s.Ctx, clPoolId1.GetId())
	s.Require().NoError(err)

	clPool2IncentiveRecordsAtEpoch3, err := s.App.ConcentratedLiquidityKeeper.GetAllIncentiveRecordsForPool(s.Ctx, clPoolId2.GetId())
	s.Require().NoError(err)

	s.ValidateDistributedGauge(clPoolExternalGaugeIdPool1, 3, externalGaugeCoins)
	s.ValidateDistributedGauge(clPoolExternalGaugeIdPool2, 2, externalGaugeCoins) // reason why this is 2 is because it is a non-perp gauge and it has finished distribution.
	s.ValidateDistributedGauge(clPoolInternalGaugeId1, 3, internalGaugeCoins)
	s.ValidateDistributedGauge(clPoolInternalGaugeId2, 3, internalGaugeCoins)

	// Since there is no incentive distributed in this epoch, the incentive Record for ClPool1 and ClPool2 after Epoch3
	// should be the same as the one from after Epoch2.
	s.Require().Equal(clPool1IncentiveRecordsAtEpoch2, clPool1IncentiveRecordsAtEpoch3)
	s.Require().Equal(clPool2IncentiveRecordsAtEpoch2, clPool2IncentiveRecordsAtEpoch3)

}

func (s *KeeperTestSuite) CreateNoLockExternalGauges(clPoolId uint64, externalGaugeCoins sdk.Coins, gaugeCreator sdk.AccAddress, numEpochsPaidOver uint64) uint64 {
	// Create 1 external no-lock gauge perpetual over 1 epochs MsgCreateGauge
	clPoolExternalGaugeId, err := s.App.IncentivesKeeper.CreateGauge(s.Ctx, numEpochsPaidOver == 1, gaugeCreator, externalGaugeCoins,
		lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.NoLock,
		},
		s.Ctx.BlockTime(),
		numEpochsPaidOver,
		clPoolId,
	)
	s.Require().NoError(err)

	return clPoolExternalGaugeId
}

func (s *KeeperTestSuite) IncentivizeInternalGauge(poolIds []uint64, epochDuration time.Duration, removeDistrRecord bool) {
	var weight osmomath.Int
	if !removeDistrRecord {
		weight = osmomath.NewInt(100)
	} else {
		weight = osmomath.ZeroInt()
	}

	var gaugeIds []uint64
	var poolIncentiveRecords []poolincentivetypes.DistrRecord
	for _, poolId := range poolIds {
		gaugeIdForPoolId, err := s.App.PoolIncentivesKeeper.GetPoolGaugeId(s.Ctx, poolId, epochDuration)
		s.Require().NoError(err)

		gaugeIds = append(gaugeIds, gaugeIdForPoolId)
		poolIncentiveRecords = append(poolIncentiveRecords, poolincentivetypes.DistrRecord{
			GaugeId: gaugeIdForPoolId,
			Weight:  weight,
		})
	}

	// incentivize both CL pools to recieve internal incentives
	err := s.App.PoolIncentivesKeeper.HandleReplacePoolIncentivesProposal(s.Ctx, &poolincentivetypes.ReplacePoolIncentivesProposal{
		Title:       "",
		Description: "",
		Records:     poolIncentiveRecords,
	},
	)
	s.Require().NoError(err)
}
func (s *KeeperTestSuite) TestAllocateAcrossGauges() {
	tests := []struct {
		name                               string
		GroupGauge                         types.GroupGauge
		expectedAllocationPerGroupGauge    sdk.Coins
		expectedAllocationPerInternalGauge sdk.Coins
		expectError                        bool
	}{
		{
			name: "Happy case: Valid perp Group Gauge",
			GroupGauge: types.GroupGauge{
				GroupGaugeId:    9,
				InternalIds:     []uint64{2, 3, 4},
				SplittingPolicy: types.Evenly,
			},
			expectedAllocationPerGroupGauge:    sdk.NewCoins(sdk.NewCoin("uosmo", osmomath.NewInt(100_000_000))),
			expectedAllocationPerInternalGauge: sdk.NewCoins(sdk.NewCoin("uosmo", osmomath.NewInt(33_333_333))),
			expectError:                        false,
		},
		{
			name: "Happy Case: Valid non-perp Group Gauge",
			GroupGauge: types.GroupGauge{
				GroupGaugeId:    10,
				InternalIds:     []uint64{5, 6, 7},
				SplittingPolicy: types.Evenly,
			},
			expectedAllocationPerGroupGauge:    sdk.NewCoins(sdk.NewCoin("uosmo", osmomath.NewInt(50_000_000))),
			expectedAllocationPerInternalGauge: sdk.NewCoins(sdk.NewCoin("uosmo", osmomath.NewInt(16_666_666))),
			expectError:                        false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			s.FundAcc(s.TestAccs[1], sdk.NewCoins(sdk.NewCoin("uosmo", osmomath.NewInt(200_000_000))))
			clPool := s.PrepareConcentratedPool()

			// create 3 internal Gauge
			internalGauges := s.setupNoLockInternalGauge(clPool.GetId(), uint64(6)) // gauge Id = 2,3,4,5,6,7

			// create non-perp internal Gauge
			s.CreateNoLockExternalGauges(clPool.GetId(), sdk.NewCoins(), s.TestAccs[1], uint64(2)) // gaugeid= 8

			// create perp group gauge
			_, err := s.App.IncentivesKeeper.CreateGroupGauge(s.Ctx, sdk.NewCoins(sdk.NewCoin("uosmo", osmomath.NewInt(100_000_000))), uint64(1), s.TestAccs[1], internalGauges[:3], lockuptypes.ByGroup, types.Evenly) // gauge id = 2,3,4
			s.Require().NoError(err)

			// create non-perp group gauge
			_, err = s.App.IncentivesKeeper.CreateGroupGauge(s.Ctx, sdk.NewCoins(sdk.NewCoin("uosmo", osmomath.NewInt(100_000_000))), uint64(2), s.TestAccs[1], internalGauges[len(internalGauges)-3:], lockuptypes.ByGroup, types.Evenly) // gauge id = 5,6,7
			s.Require().NoError(err)

			// Call Testing function
			err = s.App.IncentivesKeeper.AllocateAcrossGauges(s.Ctx)
			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				groupGaugePostAllocate, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, tc.GroupGauge.GroupGaugeId)
				s.Require().NoError(err)

				s.Require().Equal(groupGaugePostAllocate.DistributedCoins, tc.expectedAllocationPerGroupGauge)

				for _, gauge := range tc.GroupGauge.InternalIds {
					internalGauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gauge)
					s.Require().NoError(err)

					s.Require().Equal(internalGauge.Coins, tc.expectedAllocationPerInternalGauge)
				}
			}
		})
	}
}

func (s *KeeperTestSuite) WithBaseCaseDifferentCoins(baseCase GroupGaugeCreationFields, newCoins sdk.Coins) GroupGaugeCreationFields {
	baseCase.coins = newCoins
	return baseCase
}

func (s *KeeperTestSuite) WithBaseCaseDifferentEpochPaidOver(baseCase GroupGaugeCreationFields, numEpochPaidOver uint64) GroupGaugeCreationFields {
	baseCase.numEpochPaidOver = numEpochPaidOver
	return baseCase
}

func (s *KeeperTestSuite) WithBaseCaseDifferentInternalGauges(baseCase GroupGaugeCreationFields, internalGauges []uint64) GroupGaugeCreationFields {
	baseCase.internalGaugeIds = internalGauges
	return baseCase
}

func (s *KeeperTestSuite) TestCreateGroupGaugeAndDistribute() {
	hundredKUosmo := sdk.NewCoin("uosmo", osmomath.NewInt(100_000_000))
	hundredKUatom := sdk.NewCoin("uatom", osmomath.NewInt(100_000_000))
	fifetyKUosmo := sdk.NewCoin("uosmo", osmomath.NewInt(50_000_000))
	fifetyKUatom := sdk.NewCoin("uatom", osmomath.NewInt(50_000_000))
	twentyfiveKUosmo := sdk.NewCoin("uosmo", osmomath.NewInt(25_000_000))
	twentyfiveKUatom := sdk.NewCoin("uatom", osmomath.NewInt(25_000_000))

	baseCase := &GroupGaugeCreationFields{
		coins:            sdk.NewCoins(hundredKUosmo),
		numEpochPaidOver: 1,
		owner:            s.TestAccs[1],
		internalGaugeIds: []uint64{2, 3, 4, 5},
	}

	tests := []struct {
		name                                 string
		createGauge                          GroupGaugeCreationFields
		expectedCoinsPerInternalGauge        sdk.Coins
		expectedCoinsDistributedPerEpoch     sdk.Coins
		expectCreateGroupGaugeError          bool
		expectDistributeToInternalGaugeError bool
	}{
		{
			name:                             "Valid case: Valid perp-GroupGauge Creation and Distribution",
			createGauge:                      *baseCase,
			expectedCoinsPerInternalGauge:    sdk.NewCoins(twentyfiveKUosmo), // 100osmo / 4 = 25osmo
			expectedCoinsDistributedPerEpoch: sdk.NewCoins(hundredKUosmo),
		},
		{
			name:                             "Valid case: Valid perp-GroupGauge Creation with only CL internal gauges and Distribution",
			createGauge:                      s.WithBaseCaseDifferentInternalGauges(*baseCase, []uint64{2, 3, 4}),
			expectedCoinsPerInternalGauge:    sdk.NewCoins(sdk.NewCoin("uosmo", osmomath.NewInt(33_333_333))),
			expectedCoinsDistributedPerEpoch: sdk.NewCoins(hundredKUosmo),
		},
		{
			name:                             "Valid case: Valid perp-GroupGauge Creation with only GAMM internal gauge and Distribution",
			createGauge:                      s.WithBaseCaseDifferentInternalGauges(*baseCase, []uint64{5}),
			expectedCoinsPerInternalGauge:    sdk.NewCoins(hundredKUosmo),
			expectedCoinsDistributedPerEpoch: sdk.NewCoins(hundredKUosmo),
		},
		{
			name:                             "Valid case: Valid non-perpGroupGauge Creation with and Distribution",
			createGauge:                      s.WithBaseCaseDifferentEpochPaidOver(*baseCase, uint64(4)),
			expectedCoinsPerInternalGauge:    sdk.NewCoins(sdk.NewCoin("uosmo", osmomath.NewInt(6_250_000))),
			expectedCoinsDistributedPerEpoch: sdk.NewCoins(twentyfiveKUosmo),
		},
		{
			name:                             "Valid case: Valid perp-GroupGauge Creation with 2 coins and Distribution",
			createGauge:                      s.WithBaseCaseDifferentCoins(*baseCase, sdk.NewCoins(hundredKUosmo, hundredKUatom)),
			expectedCoinsPerInternalGauge:    sdk.NewCoins(twentyfiveKUosmo, twentyfiveKUatom),
			expectedCoinsDistributedPerEpoch: sdk.NewCoins(hundredKUosmo, hundredKUatom),
		},
		{
			name:                             "Valid case: Valid non-perp GroupGauge Creation with 2 coins and Distribution",
			createGauge:                      s.WithBaseCaseDifferentEpochPaidOver(s.WithBaseCaseDifferentCoins(*baseCase, sdk.NewCoins(hundredKUosmo, hundredKUatom)), uint64(2)),
			expectedCoinsPerInternalGauge:    sdk.NewCoins(sdk.NewCoin("uosmo", osmomath.NewInt(12_500_000)), sdk.NewCoin("uatom", osmomath.NewInt(12_500_000))),
			expectedCoinsDistributedPerEpoch: sdk.NewCoins(fifetyKUosmo, fifetyKUatom),
		},
		{
			name:                        "InValid case: Creating a GroupGauge with invalid internalIds",
			createGauge:                 s.WithBaseCaseDifferentInternalGauges(*baseCase, []uint64{100, 101}),
			expectCreateGroupGaugeError: true,
		},
		{
			name:                        "InValid case: Creating a GroupGauge with non-perpetual internalId",
			createGauge:                 s.WithBaseCaseDifferentInternalGauges(*baseCase, []uint64{2, 3, 4, 6}),
			expectCreateGroupGaugeError: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			s.FundAcc(s.TestAccs[1], sdk.NewCoins(hundredKUosmo, hundredKUatom)) // 100osmo, 100atom

			// Setup
			clPool := s.PrepareConcentratedPool()
			lockOwner := sdk.AccAddress([]byte("addr1---------------"))
			epochInfo := s.App.IncentivesKeeper.GetEpochInfo(s.Ctx)
			s.SetupGroupGauge(clPool.GetId(), lockOwner, uint64(3), uint64(1))

			//create 1 non-perp internal Gauge
			s.CreateNoLockExternalGauges(clPool.GetId(), sdk.NewCoins(), s.TestAccs[1], uint64(2)) // gauge id = 6

			groupGaugeId, err := s.App.IncentivesKeeper.CreateGroupGauge(s.Ctx, tc.createGauge.coins, tc.createGauge.numEpochPaidOver, tc.createGauge.owner, tc.createGauge.internalGaugeIds, lockuptypes.ByGroup, types.Evenly) // gauge id = 6
			if tc.expectCreateGroupGaugeError {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)

			groupGaugeObj, err := s.App.IncentivesKeeper.GetGroupGaugeById(s.Ctx, groupGaugeId)
			s.Require().NoError(err)

			// check internalGauges matches what we expect
			s.Require().Equal(groupGaugeObj.InternalIds, tc.createGauge.internalGaugeIds)

			for epoch := uint64(1); epoch <= tc.createGauge.numEpochPaidOver; epoch++ {
				// ******************** EPOCH PASSED ********************* //
				s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(epochInfo.Duration))
				s.App.EpochsKeeper.AfterEpochEnd(s.Ctx, epochInfo.GetIdentifier(), int64(epoch))

				// Validate GroupGauge
				groupGauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, groupGaugeId)
				s.Require().NoError(err)

				var expectedDistributedCoins []sdk.Coin
				for _, coin := range tc.expectedCoinsDistributedPerEpoch {
					expectedDistributedCoins = append(expectedDistributedCoins, sdk.NewCoin(coin.Denom, coin.Amount.Mul(osmomath.NewIntFromUint64(epoch))))
				}

				s.ValidateDistributedGauge(groupGauge.Id, epoch, expectedDistributedCoins)

				// Validate Internal Gauges
				internalGauges, err := s.App.IncentivesKeeper.GetGaugeFromIDs(s.Ctx, tc.createGauge.internalGaugeIds)
				s.Require().NoError(err)

				for _, internalGauge := range internalGauges {
					var expectedDistributedCoinsPerInternalGauge []sdk.Coin
					for _, coin := range tc.expectedCoinsPerInternalGauge {
						expectedDistributedCoinsPerInternalGauge = append(expectedDistributedCoinsPerInternalGauge, (sdk.NewCoin(coin.Denom, coin.Amount.Mul(osmomath.NewIntFromUint64(epoch)))))
					}
					s.ValidateDistributedGauge(internalGauge.Id, epoch, expectedDistributedCoinsPerInternalGauge)
				}

				// Validate CL Incentive distribution
				poolIncentives, err := s.App.ConcentratedLiquidityKeeper.GetAllIncentiveRecordsForPool(s.Ctx, clPool.GetId())
				s.Require().NoError(err)

				for i := 0; i < len(poolIncentives); i++ {
					idx := 0
					// the logic below is for indexing incentiveRecord, flips idx from 0,1,0,1 or 1,0,1,0 etc.
					if len(tc.expectedCoinsPerInternalGauge) > 1 {
						if epoch == 2 {
							idx = 1 - (i % 2)
						} else {
							idx = i % 2
						}
					}
					s.ValidateIncentiveRecord(clPool.GetId(), tc.expectedCoinsPerInternalGauge[idx], poolIncentives[i])
				}

				// Validate GAMM incentive distribution
				balances := s.App.BankKeeper.GetAllBalances(s.Ctx, lockOwner)
				if len(balances) != 0 {
					var coins sdk.Coins
					for _, bal := range tc.expectedCoinsPerInternalGauge {
						coin := sdk.NewCoin(bal.Denom, bal.Amount.Mul(osmomath.NewIntFromUint64(epoch)))
						coins = append(coins, coin)
					}

					s.Require().Equal(balances, coins)
				}
			}
		})
	}

}
