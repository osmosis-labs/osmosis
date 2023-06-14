package keeper_test

import (
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	appParams "github.com/osmosis-labs/osmosis/v16/app/params"
	"github.com/osmosis-labs/osmosis/v16/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v16/x/lockup/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
)

var _ = suite.TestingSuite(nil)

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

	coinsToMint := sdk.NewCoins(sdk.NewCoin(defaultRewardDenom, sdk.NewInt(10000000)), sdk.NewCoin(appParams.BaseCoinUnit, sdk.NewInt(10000000)))
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

				for _, gauge := range gauges {
					for _, coin := range gauge.Coins {
						// get poolId from GaugeId
						poolId, err := s.App.PoolIncentivesKeeper.GetPoolIdFromGaugeId(s.Ctx, gauge.GetId(), currentEpoch.Duration)
						s.Require().NoError(err)

						// check that incentive record wasn't created
						_, err = s.App.ConcentratedLiquidityKeeper.GetIncentiveRecord(s.Ctx, poolId, coin.Denom, currentEpoch.Duration, s.App.AccountKeeper.GetModuleAddress(types.ModuleName))
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
					s.Require().Equal(tc.expectedDistributions.AmountOf(coin.Denom).Add(sdk.ZeroInt()), actualbalanceAfterDistribution.Add(sdk.ZeroInt()))
				}

				for _, gauge := range gauges {
					for _, coin := range gauge.Coins {
						gaugeId := gauge.GetId()

						// get poolId from GaugeId
						poolId, err := s.App.PoolIncentivesKeeper.GetPoolIdFromGaugeId(s.Ctx, gaugeId, currentEpoch.Duration)
						s.Require().NoError(err)

						// GetIncentiveRecord to see if pools received incentives properly
						incentiveRecord, err := s.App.ConcentratedLiquidityKeeper.GetIncentiveRecord(s.Ctx, poolId, defaultRewardDenom, types.DefaultConcentratedUptime, s.App.AccountKeeper.GetModuleAddress(types.ModuleName))
						s.Require().NoError(err)

						expectedEmissionRate := sdk.NewDecFromInt(coin.Amount).QuoTruncate(sdk.NewDec(int64(currentEpoch.Duration.Seconds())))

						// Check that gauge distribution state is updated.
						s.ValidateDistributedGauge(gaugeId, 1, tc.gaugeCoins)

						// check every parameter in incentiveRecord so that it matches what we created
						s.Require().Equal(poolId, incentiveRecord.PoolId)
						s.Require().Equal(defaultRewardDenom, incentiveRecord.IncentiveDenom)
						s.Require().Equal(s.App.AccountKeeper.GetModuleAddress(types.ModuleName).String(), incentiveRecord.IncentiveCreatorAddr)
						s.Require().Equal(expectedEmissionRate, incentiveRecord.GetIncentiveRecordBody().EmissionRate)
						s.Require().Equal(s.Ctx.BlockTime().UTC().String(), incentiveRecord.GetIncentiveRecordBody().StartTime.UTC().String())
						s.Require().Equal(types.DefaultConcentratedUptime, incentiveRecord.MinUptime)
						s.Require().Equal(fiveKRewardCoins.Amount, incentiveRecord.GetIncentiveRecordBody().RemainingAmount.RoundInt())
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
		expectedRemainingAmountIncentiveRecord: []sdk.Dec{sdk.NewDec(defaultAmount)},
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
			tc.expectedRemainingAmountIncentiveRecord[i] = sdk.NewDec(gaugeCoins[i].Amount.Int64())
		}
		return tc
	}

	withNumEpochs := func(tc test, numEpochs uint64) test {
		tc.numEpochsPaidOver = numEpochs

		// Do deep copies
		tempDistributions := make(sdk.Coins, len(tc.expectedDistributions))
		copy(tempDistributions, tc.expectedDistributions)

		tempRemainingAmountIncentiveRecord := make([]sdk.Dec, len(tc.expectedRemainingAmountIncentiveRecord))
		copy(tempRemainingAmountIncentiveRecord, tc.expectedRemainingAmountIncentiveRecord)

		for i := range tc.expectedRemainingAmountIncentiveRecord {
			// update expected distributions
			tempDistributions[i].Amount = tc.expectedDistributions[i].Amount.Quo(sdk.NewInt(int64(numEpochs)))

			// update expected remaining amount in incentive record
			tempRemainingAmountIncentiveRecord[i] = tc.expectedRemainingAmountIncentiveRecord[i].QuoTruncate(sdk.NewDec(int64(numEpochs))).TruncateDec()
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
		"error: balancer pool id":                    withPoolId(defaultTest, defaultBalancerPool),
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

				// Get module account
				moduleAccount := s.App.AccountKeeper.GetModuleAccount(s.Ctx, types.ModuleName)
				incentiveModuleAddress := moduleAccount.GetAddress()

				incentivesEpochDuration := s.App.IncentivesKeeper.GetEpochInfo(s.Ctx).Duration
				incentivesEpochDurationSeconds := sdk.NewDec(incentivesEpochDuration.Milliseconds()).QuoInt(sdk.NewInt(1000))

				// Check that incentive records were created
				for i, coin := range tc.expectedDistributions {
					incentiveRecords, err := s.App.ConcentratedLiquidityKeeper.GetIncentiveRecord(s.Ctx, tc.poolId, coin.Denom, time.Nanosecond, incentiveModuleAddress)
					s.Require().NoError(err)

					expectedEmissionRatePerEpoch := coin.Amount.ToDec().QuoTruncate(incentivesEpochDurationSeconds)

					s.Require().Equal(incentiveModuleAddress.String(), incentiveRecords.IncentiveCreatorAddr)
					s.Require().Equal(tc.startTime.UTC(), incentiveRecords.IncentiveRecordBody.StartTime.UTC())
					s.Require().Equal(coin.Denom, incentiveRecords.IncentiveDenom)
					s.Require().Equal(tc.expectedRemainingAmountIncentiveRecord[i], incentiveRecords.IncentiveRecordBody.RemainingAmount)
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
