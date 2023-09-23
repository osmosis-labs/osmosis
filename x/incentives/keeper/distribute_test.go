package keeper_test

import (
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	osmoutils "github.com/osmosis-labs/osmosis/osmoutils/coins"
	appParams "github.com/osmosis-labs/osmosis/v19/app/params"
	"github.com/osmosis-labs/osmosis/v19/x/incentives/types"
	incentivetypes "github.com/osmosis-labs/osmosis/v19/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v19/x/lockup/types"
	poolincentivetypes "github.com/osmosis-labs/osmosis/v19/x/pool-incentives/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
)

var _ = suite.TestingSuite(nil)

const (
	defaultGroupGaugeId = uint64(5)
)

var (
	defaultGaugeRecordOneRecord = types.InternalGaugeRecord{
		GaugeId:          1,
		CurrentWeight:    osmomath.NewInt(100),
		CumulativeWeight: osmomath.NewInt(200),
	}
	defaultGaugeRecordTwoRecords = types.InternalGaugeRecord{
		// Note that this is 4 and not 2 because we assume the second pool is a balancer pool
		// that creates three gauges (one for each lockable duration), only the last of which
		// we use.
		GaugeId:          4,
		CurrentWeight:    osmomath.NewInt(100),
		CumulativeWeight: osmomath.NewInt(200),
	}
	defaultGroupGauge = types.GroupGauge{
		GroupGaugeId: defaultGroupGaugeId,
		InternalGaugeInfo: types.InternalGaugeInfo{
			TotalWeight:  defaultGaugeRecordOneRecord.CurrentWeight.Add(defaultGaugeRecordTwoRecords.CurrentWeight),
			GaugeRecords: []types.InternalGaugeRecord{defaultGaugeRecordOneRecord, defaultGaugeRecordTwoRecords},
		},
		SplittingPolicy: types.Volume,
	}
	singleRecordGroupGauge = types.GroupGauge{
		GroupGaugeId: defaultGroupGaugeId,
		InternalGaugeInfo: types.InternalGaugeInfo{
			TotalWeight:  defaultGaugeRecordOneRecord.CurrentWeight,
			GaugeRecords: []types.InternalGaugeRecord{defaultGaugeRecordOneRecord},
		},
		SplittingPolicy: types.Volume,
	}

	emptyCoins          = sdk.Coins{}
	defaultVolumeAmount = osmomath.NewInt(300)
)

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
	// We skip these test until group gauge initialization refactor is complete
	s.T().Skip()

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

			// check internalGauges matches what we expect
			// TODO: assert initialization logic correctness once it is implemented
			// Tracked in issue https://github.com/osmosis-labs/osmosis/issues/6404

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

// deepCopyGroupGauge creates a deep copy of the passed in group gauge.
func deepCopyGroupGauge(src types.GroupGauge) types.GroupGauge {
	gaugeRecords := make([]types.InternalGaugeRecord, len(src.InternalGaugeInfo.GaugeRecords))
	for i, record := range src.InternalGaugeInfo.GaugeRecords {
		gaugeRecords[i] = types.InternalGaugeRecord{
			GaugeId:          record.GaugeId,
			CurrentWeight:    record.CurrentWeight,
			CumulativeWeight: record.CumulativeWeight,
		}
	}

	return types.GroupGauge{
		GroupGaugeId: src.GroupGaugeId,
		InternalGaugeInfo: types.InternalGaugeInfo{
			TotalWeight:  src.InternalGaugeInfo.TotalWeight,
			GaugeRecords: gaugeRecords,
		},
		SplittingPolicy: src.SplittingPolicy,
	}
}

// deepCopyGauge creates a deep copy of the passed in gauge.
func deepCopyGauge(src types.Gauge) types.Gauge {
	gauge := src
	gauge.Coins = sdk.NewCoins(src.Coins...)
	gauge.DistributedCoins = sdk.NewCoins(src.DistributedCoins...)
	return gauge
}

// withUpdatedVolumes takes in a group gauge and a list of updated cumulative volumes (ordered) and updates the contents of the gauge to
// reflect these new volumes.
// It is only intended to be used to set expected values for test cases.
func (s *KeeperTestSuite) withUpdatedVolumes(groupGauge types.GroupGauge, updatedCumulativeVolumes []osmomath.Int) types.GroupGauge {
	// Ensure there aren't more volumes to update than records in group gauge
	s.Require().True(len(updatedCumulativeVolumes) <= len(groupGauge.InternalGaugeInfo.GaugeRecords))

	// We make a deep copy of the group gauge to ensure we don't modify the original input/defaults
	updatedGroupGauge := deepCopyGroupGauge(groupGauge)

	newTotalWeight := osmomath.ZeroInt()
	for i, updatedVolume := range updatedCumulativeVolumes {
		currentRecord := groupGauge.InternalGaugeInfo.GaugeRecords[i]
		updatedRecord := types.InternalGaugeRecord{
			GaugeId:          currentRecord.GaugeId,
			CurrentWeight:    updatedVolume.Sub(currentRecord.CumulativeWeight),
			CumulativeWeight: updatedVolume,
		}
		updatedGroupGauge.InternalGaugeInfo.GaugeRecords[i] = updatedRecord
		newTotalWeight = newTotalWeight.Add(updatedRecord.CurrentWeight)
	}
	updatedGroupGauge.InternalGaugeInfo.TotalWeight = newTotalWeight

	return updatedGroupGauge
}

// withSplittingPolicy returns a deep copy of the passed in group gauge with the splitting policy set to the passed in value.
func withSplittingPolicy(groupGauge types.GroupGauge, splittingPolicy types.SplittingPolicy) types.GroupGauge {
	// We make a deep copy of the group gauge to ensure we don't modify the original input/defaults
	updatedGroupGauge := deepCopyGroupGauge(groupGauge)
	updatedGroupGauge.SplittingPolicy = splittingPolicy

	return updatedGroupGauge
}

// withGroupGaugeId returns a deep copy of the passed in group gauge with the group gauge id to the passed in value.
func withGroupGaugeId(groupGauge types.GroupGauge, groupGaugeId uint64) types.GroupGauge {
	// We make a deep copy of the group gauge to ensure we don't modify the original input/defaults
	updatedGroupGauge := deepCopyGroupGauge(groupGauge)
	updatedGroupGauge.GroupGaugeId = groupGaugeId
	return updatedGroupGauge
}

// withIsPerpetual sets the isPerpetual flag on the passed in gauge and returns a copy of the gauge.
func withIsPerpetual(gauge types.Gauge, isPerpetual bool) types.Gauge {
	gauge.IsPerpetual = isPerpetual
	return gauge
}

// withCoinsToDistribute sets total and distributed coins on the gauge
func withCoinsToDistribute(gauge types.Gauge, coins sdk.Coins, distributed sdk.Coins) types.Gauge {
	gauge.DistributedCoins = distributed
	gauge.Coins = coins
	return gauge
}

// withNonPerpetualEpochs sets filled epochs and num epochs paid over on the gauge
func withNonPerpetualEpochs(gauge types.Gauge, filledEpochs uint64, numEpochsPaidOver uint64) types.Gauge {
	gauge.FilledEpochs = filledEpochs
	gauge.NumEpochsPaidOver = numEpochsPaidOver
	return gauge
}

// withGaugeId sets the id of the gauge to given and returns the gauge.
func withGaugeId(gauge types.Gauge, id uint64) types.Gauge {
	gauge.Id = id
	return gauge
}

// setPoolVolumes takes in an array of pool IDs and volumes and sets each pool's volume to the corresponding volume amount.
// If there are more pool IDs than volumes, the extra pool IDs are ignored. This is to more simply accommodate cases where only the first k
// of n pools are updated without needing to pad the volumes array during test setup.
func (s *KeeperTestSuite) setPoolVolumes(poolIds []uint64, volumes []osmomath.Int) {
	s.Require().True(len(poolIds) >= len(volumes))

	for i, curVolume := range volumes {
		s.App.PoolManagerKeeper.SetVolume(s.Ctx, poolIds[i], sdk.NewCoins(sdk.NewCoin(s.App.StakingKeeper.BondDenom(s.Ctx), curVolume)))
	}
}

// TODO: rename this to syncVolumeSplitGroup as part of https://github.com/osmosis-labs/osmosis/pull/6446
func (s *KeeperTestSuite) TestSyncVolumeSplitGauge() {
	tests := map[string]struct {
		groupGaugeToSync types.GroupGauge

		// Each element updates either a CL or a balancer pool volume.
		// These pools are created at the beginning of each test.
		updatedPoolVolumes []osmomath.Int

		expectedSyncedGauge types.GroupGauge
		expectedError       error
	}{
		"happy path: valid update on group gauge with even volume growth": {
			groupGaugeToSync: deepCopyGroupGauge(defaultGroupGauge),
			updatedPoolVolumes: []osmomath.Int{
				osmomath.NewInt(300),
				osmomath.NewInt(300),
			},

			expectedSyncedGauge: s.withUpdatedVolumes(defaultGroupGauge, []osmomath.Int{osmomath.NewInt(300), osmomath.NewInt(300)}),
			expectedError:       nil,
		},
		"valid update on group gauge with different volume growth": {
			groupGaugeToSync: deepCopyGroupGauge(defaultGroupGauge),
			updatedPoolVolumes: []osmomath.Int{
				osmomath.NewInt(253),
				osmomath.NewInt(659),
			},

			expectedSyncedGauge: s.withUpdatedVolumes(defaultGroupGauge, []osmomath.Int{osmomath.NewInt(253), osmomath.NewInt(659)}),
			expectedError:       nil,
		},
		"valid update on group gauge with only one record to sync": {
			groupGaugeToSync: deepCopyGroupGauge(singleRecordGroupGauge),

			updatedPoolVolumes: []osmomath.Int{
				osmomath.NewInt(933),
			},

			expectedSyncedGauge: s.withUpdatedVolumes(singleRecordGroupGauge, []osmomath.Int{osmomath.NewInt(933)}),
			expectedError:       nil,
		},

		// Error catching
		"tracked volume has dropped to zero for a pool (no pool volume or volume cannot be found)": {
			groupGaugeToSync: deepCopyGroupGauge(defaultGroupGauge),
			updatedPoolVolumes: []osmomath.Int{
				osmomath.NewInt(300),
				osmomath.NewInt(0),
			},

			expectedError: types.NoPoolVolumeError{PoolId: uint64(2)},
		},
		"cumulative volume has decreased for a pool (impossible/invalid state)": {
			groupGaugeToSync: deepCopyGroupGauge(defaultGroupGauge),
			updatedPoolVolumes: []osmomath.Int{
				osmomath.NewInt(300),
				osmomath.NewInt(100),
			},

			expectedError: types.CumulativeVolumeDecreasedError{PoolId: uint64(2), PreviousVolume: osmomath.NewInt(200), NewVolume: osmomath.NewInt(100)},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			ik := s.App.IncentivesKeeper

			// Prepare pools so gauges and pool ids are set in state
			clPool := s.PrepareConcentratedPool()
			balPoolId := s.PrepareBalancerPool()

			poolIds := []uint64{clPool.GetId(), balPoolId}

			// Update cumulative volumes for pools
			s.setPoolVolumes(poolIds, tc.updatedPoolVolumes)

			// Save original input to help with mutation-related assertions
			originalGroupGauge := deepCopyGroupGauge(tc.groupGaugeToSync)

			// Set group gauge in state to make stronger assertions later
			ik.SetGroupGauge(s.Ctx, tc.groupGaugeToSync)

			// --- System under test ---

			err := ik.SyncVolumeSplitGauge(s.Ctx, tc.groupGaugeToSync)

			// --- Assertions ---

			if tc.expectedError != nil {
				s.Require().ErrorContains(tc.expectedError, err.Error())

				// Ensure original group gauge is not mutated
				s.Require().Equal(originalGroupGauge, tc.groupGaugeToSync)

				// Ensure group gauge is unchanged in state
				gaugeInState, err := ik.GetGroupGaugeById(s.Ctx, tc.groupGaugeToSync.GroupGaugeId)
				s.Require().NoError(err)
				s.Require().Equal(tc.groupGaugeToSync, gaugeInState)

				return
			}

			s.Require().NoError(err)

			updatedGauge, err := ik.GetGroupGaugeById(s.Ctx, tc.groupGaugeToSync.GroupGaugeId)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedSyncedGauge, updatedGauge)
		})
	}
}

// TODO: rename this to TestSyncGroupWeights as part of https://github.com/osmosis-labs/osmosis/pull/6446
func (s *KeeperTestSuite) TestSyncGroupGaugeWeights() {
	tests := map[string]struct {
		groupGaugeToSync types.GroupGauge

		expectedSyncedGauge types.GroupGauge
		expectedError       error
	}{
		"happy path: valid volume splitting group": {
			groupGaugeToSync: withSplittingPolicy(defaultGroupGauge, types.Volume),

			// Note: setup logic runs default setup based on groupGaugeToSync's splitting policy.
			// More involved tests related to syncing logic for specific splitting policies are in their respective tests.
			expectedSyncedGauge: s.withUpdatedVolumes(defaultGroupGauge, []osmomath.Int{defaultVolumeAmount, defaultVolumeAmount}),
			expectedError:       nil,
		},

		// Error catching
		"unsupported splitting policy": {
			groupGaugeToSync: withSplittingPolicy(defaultGroupGauge, types.SplittingPolicy(100)),

			expectedError: types.UnsupportedSplittingPolicyError{GroupGaugeId: uint64(5), SplittingPolicy: types.SplittingPolicy(100)},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			ik := s.App.IncentivesKeeper

			// Prepare pools so gauges and pool ids are set in state
			clPool := s.PrepareConcentratedPool()
			balPoolId := s.PrepareBalancerPool()

			poolIds := []uint64{clPool.GetId(), balPoolId}

			// Currently the only supported splitting policy is volume splitting.
			// When more are added in the future, setup logic should route to the appropriate setup function here.
			switch tc.groupGaugeToSync.SplittingPolicy {
			case types.Volume:
				s.setPoolVolumes(poolIds, []osmomath.Int{defaultVolumeAmount, defaultVolumeAmount})
			}

			// Save original input to help with mutation-related assertions
			originalGroupGauge := deepCopyGroupGauge(tc.groupGaugeToSync)

			// Set group gauge in state to make stronger assertions later
			ik.SetGroupGauge(s.Ctx, tc.groupGaugeToSync)

			// --- System under test ---

			err := ik.SyncGroupGaugeWeights(s.Ctx, tc.groupGaugeToSync)

			// --- Assertions ---

			if tc.expectedError != nil {
				s.Require().ErrorContains(tc.expectedError, err.Error())

				// Ensure original group gauge is not mutated
				s.Require().Equal(originalGroupGauge, tc.groupGaugeToSync)

				// Ensure group gauge is unchanged in state
				gaugeInState, err := ik.GetGroupGaugeById(s.Ctx, tc.groupGaugeToSync.GroupGaugeId)
				s.Require().NoError(err)
				s.Require().Equal(tc.groupGaugeToSync, gaugeInState)

				return
			}

			s.Require().NoError(err)

			updatedGauge, err := ik.GetGroupGaugeById(s.Ctx, tc.groupGaugeToSync.GroupGaugeId)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedSyncedGauge, updatedGauge)
		})
	}
}

// Tests allocations across gauges from groups.
//
// The test is set up as follows:
// Defaults and test-specific set up and expected value computation helpers at the top.
//
// Test cases follow.
// Each test case consists of:
// - groupConfig: the configuration of the group to be passed into the AllocateAcrossGauges function. It has the group itself, the 1:1 associated group gauge
// as well as the fields determining the expected behavior of the test case.
// - volumeToSet the volume to set for the pool associated with the group gauge. There are only 2 pools created - CL and balancer. For some test cases, we want
// to misconfigure the volume to trigger certain behavior.
// - shouldSkipFunding - whether the test case should skip funding the module account with the coins needed for the test case to succeed. Used for edge case setup.
//
// The structure of the test is as follows:
// - Set up the environment from the test configuration
// - Compute the expected distribtuion values based on the test configuration to minimize the number of testcase parameters.
// - Run system under test
// - Perform validations
//
// Validations include:
// - checking that groups are skipped in acceptable cases without failing other group distributions
// - checking that the group gauge is correctly updated - filled epochs increased and distributed coins increased
// - checking that the internal gauges are correctly updated - coins increased
func (s *KeeperTestSuite) TestAllocateAcrossGauges() {
	// consists of 1:1 mapped group and group gauge.
	type groupConfig struct {
		group      types.GroupGauge
		groupGauge types.Gauge

		expectedSkipped bool

		// This is auto-computed and set in the test
		// set up depending on the test case configuration.
		expectedTotalDistribution sdk.Coins
	}

	const (
		invalidUnderlyingGaugeId = uint64(100)
	)

	var (
		// local copies for this test to isolate failures if unexpected
		// mutations do occur.
		defaultGroupGauge      = deepCopyGroupGauge(defaultGroupGauge)
		singleRecordGroupGauge = deepCopyGroupGauge(singleRecordGroupGauge)

		baseTime = time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

		defaultCoins = sdk.NewCoins(sdk.NewCoin("uosmo", osmomath.NewInt(100_000_000)))

		// Volume pre-set configurations.
		balancerOnlyVolumeConfig  = []osmomath.Int{singleRecordGroupGauge.InternalGaugeInfo.GaugeRecords[0].CumulativeWeight, osmomath.ZeroInt()}
		balancerAndCLVolumeConfig = []osmomath.Int{defaultGroupGauge.InternalGaugeInfo.GaugeRecords[0].CumulativeWeight, defaultGroupGauge.InternalGaugeInfo.GaugeRecords[1].CumulativeWeight}

		// Remaining coins: 1 x defaultCoins (2x Coins - 1x DistributedCoins)
		perpetualGauge = types.Gauge{
			Id:               defaultGroupGaugeId,
			IsPerpetual:      true,
			Coins:            defaultCoins.Add(defaultCoins...),
			StartTime:        baseTime,
			FilledEpochs:     1,
			DistributedCoins: defaultCoins,
		}
	)

	// 2 changes: flip the isPerpetual flag and set the number of epochs paid over.
	nonPerpetualGauge := withIsPerpetual(deepCopyGauge(perpetualGauge), false)
	nonPerpetualGauge.NumEpochsPaidOver = 2

	// Configure to distribute to the invalid underlying gauge that is non-perpetual and finished.
	groupToInvalidUnderlying := deepCopyGroupGauge(singleRecordGroupGauge)
	groupToInvalidUnderlying.InternalGaugeInfo.GaugeRecords[0].GaugeId = invalidUnderlyingGaugeId

	////////////////////////////// Test-specific helpers

	// Setup the environment from the test configuration
	// Returns:
	// - inputGroups: the groups to pass into the AllocateAcrossGauges function
	// - preFundCoinsNeededForSuccess: the coins needed to be pre-funded to the module account for the test to succeed
	configureGroups := func(groups []groupConfig) (inputGroups []types.GroupGauge, preFundCoinsNeededForSuccess sdk.Coins) {
		// Set group gauges
		inputGroups = make([]types.GroupGauge, 0, len(groups))
		for _, groupConfig := range groups {
			group := groupConfig.group
			groupGauge := groupConfig.groupGauge

			s.App.IncentivesKeeper.SetGroupGauge(s.Ctx, group)
			inputGroups = append(inputGroups, group)

			// Create associated gauge for the group
			s.App.IncentivesKeeper.SetGauge(s.Ctx, &groupGauge)

			preFundCoinsNeededForSuccess = preFundCoinsNeededForSuccess.Add(groupGauge.Coins.Sub(groupGauge.DistributedCoins)...)
		}

		return inputGroups, preFundCoinsNeededForSuccess
	}

	// returns the expected coins that group is expected to distributed based on the gauge that it
	// is associated with.
	// WARNING: only use on the test configuration gauges.
	estimateDistributedGroupCoins := func(groupGauge types.Gauge) (expecteDistributedCoins sdk.Coins) {
		expecteDistributedCoins = groupGauge.Coins.Sub(groupGauge.DistributedCoins)
		if !groupGauge.IsPerpetual {
			remainingEpochs := groupGauge.NumEpochsPaidOver - groupGauge.FilledEpochs

			// Divide all coins by remainingEpochs.
			osmoutils.QuoRawMut(expecteDistributedCoins, int64(remainingEpochs))
		}
		return expecteDistributedCoins
	}

	// mutates the maps from gauge ids to coins to reflect the expected distributions of a given group and the total distribution amount this group
	// is expected to allocate.
	updateExpectedGaugeDistributionsMap := func(expectedGaugeDistributionsMap map[uint64]sdk.Coins, group types.GroupGauge, expectedAmountDistributedForGroup sdk.Coins) {
		totalWeight := group.InternalGaugeInfo.TotalWeight
		if group.SplittingPolicy == types.Volume {
			for _, underlyingGauge := range group.InternalGaugeInfo.GaugeRecords {

				// calculate expected amount distributed to this gauge
				expectedDistributedPerGauge := osmoutils.MulDec(expectedAmountDistributedForGroup, underlyingGauge.CurrentWeight.ToLegacyDec().Quo(totalWeight.ToLegacyDec()))

				if oldValue, ok := expectedGaugeDistributionsMap[underlyingGauge.GaugeId]; ok {
					expectedGaugeDistributionsMap[underlyingGauge.GaugeId] = oldValue.Add(expectedDistributedPerGauge...)
				} else {
					expectedGaugeDistributionsMap[underlyingGauge.GaugeId] = expectedDistributedPerGauge
				}
			}
		}

		// TODO: add support for other splitting policies
	}

	// Does 2 things:
	// 1. mutate input groups to set the expected total distribution amount for each group.
	// 2. return the map from gauge id to total distributions from all groups for that gauge.
	//
	// Note: groups are allowed to distribute to overlapping gauges.
	// As a result, we iteraet over all group's configurations to compute
	// the expected total distribution. Then, a the end of the test,
	// we iterate over all gauges to check that the expected distribution
	// matches the actual distribution.
	// map from gauge id to coins
	computeAndSetExpectedDistributions := func(groupConfigs []groupConfig) map[uint64]sdk.Coins {
		expectedGaugeDistributions := make(map[uint64]sdk.Coins, 0)

		for i, groupConfig := range groupConfigs {
			if groupConfig.expectedSkipped {
				continue
			}

			expectedAmountDistributed := estimateDistributedGroupCoins(groupConfig.groupGauge)
			groupConfigs[i].expectedTotalDistribution = expectedAmountDistributed

			// updates how much each gauge is expected to receive from the current group. Since we allow groups to distribute to overlapping gauges,
			// we need to keep track of the total expected distribution for each gauge and validate it at the end.
			updateExpectedGaugeDistributionsMap(expectedGaugeDistributions, groupConfig.group, expectedAmountDistributed)
		}

		return expectedGaugeDistributions
	}

	tests := map[string]struct {
		groups []groupConfig
		// index 0 for clPool, index 1 for balancer pool
		volumeToSet []osmomath.Int

		shouldSkipFunding bool
	}{

		"no groups": {
			groups: []groupConfig{},
		},

		///////////////// one gauge

		//// success

		"1: one group with perpetual gauge, one underlying gauge, double the volume": {
			groups: []groupConfig{
				{
					group:      singleRecordGroupGauge,
					groupGauge: perpetualGauge,
				},
			},

			volumeToSet: []osmomath.Int{singleRecordGroupGauge.InternalGaugeInfo.GaugeRecords[0].CumulativeWeight.MulRaw(2), osmomath.ZeroInt()},
		},
		"2: volume (weight) does not change - still distributes": {
			groups: []groupConfig{
				{
					group:      singleRecordGroupGauge,
					groupGauge: perpetualGauge,
				},
			},

			volumeToSet: balancerOnlyVolumeConfig,
		},

		"3: one group with non-perpetual gauge, one underlying gauge": {
			groups: []groupConfig{
				{
					group:      singleRecordGroupGauge,
					groupGauge: nonPerpetualGauge,
				},
			},

			volumeToSet: balancerOnlyVolumeConfig,
		},

		"4: one group gauge, multiple underlying gauges": {
			groups: []groupConfig{
				{
					group:      defaultGroupGauge,
					groupGauge: nonPerpetualGauge,
				},
			},

			volumeToSet: balancerAndCLVolumeConfig,
		},

		"5: no coins to distribute (the FilledEpoch is still updated)": {
			groups: []groupConfig{
				{
					group:      defaultGroupGauge,
					groupGauge: withCoinsToDistribute(deepCopyGauge(nonPerpetualGauge), defaultCoins, defaultCoins),
				},
			},

			volumeToSet: balancerAndCLVolumeConfig,
		},

		//// skipping

		// skipping on sync failure
		"6: skipped: synching fails due to no volume set": {
			groups: []groupConfig{
				{
					group:      defaultGroupGauge,
					groupGauge: withCoinsToDistribute(deepCopyGauge(nonPerpetualGauge), defaultCoins, defaultCoins),

					expectedSkipped: true,
				},
			},

			volumeToSet: []osmomath.Int{},
		},

		// skipping on gauge being inactive
		"7: skipping on gauge being inactive": {
			groups: []groupConfig{
				{
					group: defaultGroupGauge,
					// This makes non-perpetual gauge inactive
					groupGauge: withNonPerpetualEpochs(deepCopyGauge(nonPerpetualGauge), 1, 1),

					expectedSkipped: true,
				},
			},

			volumeToSet: balancerAndCLVolumeConfig,
		},

		// skipping because this gauge has no pool associated with it.
		// we only distributed to internal gauges.
		"8: associated group gauge is non perpetual and finished": {
			groups: []groupConfig{
				{
					group:      groupToInvalidUnderlying,
					groupGauge: perpetualGauge,

					expectedSkipped: true,
				},
			},

			volumeToSet: balancerOnlyVolumeConfig,
		},

		///////////////// multi-gauges

		// Note that groups distribute to overlapping gauges.
		"9: multiple groups with varying number of underlying gauges": {
			groups: []groupConfig{
				{
					group:      singleRecordGroupGauge,
					groupGauge: perpetualGauge,
				},
				{
					group:      withGroupGaugeId(defaultGroupGauge, 6),
					groupGauge: withGaugeId(deepCopyGauge(nonPerpetualGauge), 6),
				},
			},

			volumeToSet: []osmomath.Int{defaultGroupGauge.InternalGaugeInfo.GaugeRecords[0].CumulativeWeight, defaultGroupGauge.InternalGaugeInfo.GaugeRecords[1].CumulativeWeight},
		},

		"10: skipping one does not fail the other": {
			groups: []groupConfig{
				{
					group: defaultGroupGauge,
					// This makes non-perpetual gauge inactive
					groupGauge: withNonPerpetualEpochs(deepCopyGauge(nonPerpetualGauge), 1, 1),

					// skipped because inactive
					expectedSkipped: true,
				},
				{
					group:      withGroupGaugeId(defaultGroupGauge, 6),
					groupGauge: withGaugeId(deepCopyGauge(nonPerpetualGauge), 6),
				},
			},

			volumeToSet: balancerAndCLVolumeConfig,
		},

		///////////////// error cases

		"11: not enough funds in the module account causing error in AddToGaugeRewards": {
			groups: []groupConfig{
				{
					group:      singleRecordGroupGauge,
					groupGauge: deepCopyGauge(perpetualGauge),
				},
			},

			shouldSkipFunding: true,

			volumeToSet: balancerOnlyVolumeConfig,
		},

		// TODO: even splitting policy test cases once supported.
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()

			s.Ctx = s.Ctx.WithBlockTime(baseTime)

			incentivesKeeper := s.App.IncentivesKeeper

			// Note that this setup makes it so that the first gauge ID is an internal incentive CL gauge.
			// The next 3 gauges are internal incentive GAMM gauges.
			clPool := s.PrepareConcentratedPool()
			balPoolId := s.PrepareBalancerPool()

			preFundCoinsNeededForSuccess := sdk.NewCoins()

			// Setup the environment from the test configuration
			// Returns:
			// - inputGroups: the groups to pass into the AllocateAcrossGauges function
			// - preFundCoinsNeededForSuccess: the coins needed to be pre-funded to the module account for the test to succeed
			inputGroups, preFundCoinsNeededForSuccess := configureGroups(tc.groups)

			// Setup a gauge for testing the "invalid underlying gauge" case.
			nonPerpetualGaugeCopy := deepCopyGauge(nonPerpetualGauge)
			nonPerpetualGaugeCopy.Id = invalidUnderlyingGaugeId
			err := incentivesKeeper.SetGauge(s.Ctx, &nonPerpetualGaugeCopy)
			s.Require().NoError(err)

			// Setup volumes
			s.setupVolumes([]uint64{clPool.GetId(), balPoolId}, tc.volumeToSet)

			// Fund the right amounts depending on the test configuration.
			if !tc.shouldSkipFunding {
				s.FundModuleAcc(types.ModuleName, preFundCoinsNeededForSuccess)
			}

			// Compute expected distributions based on test configuration
			// See function definition for details.
			expectedGaugeDistributions := computeAndSetExpectedDistributions(tc.groups)

			// --- System under test ---
			err = incentivesKeeper.AllocateAcrossGauges(s.Ctx, inputGroups)

			if tc.shouldSkipFunding {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			for _, groupConfig := range tc.groups {

				// Get group gauge id
				groupGauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, groupConfig.group.GroupGaugeId)
				s.Require().NoError(err)

				if groupConfig.expectedSkipped {
					// Check that the group gauge was not filled
					s.Require().Equal(groupConfig.groupGauge.FilledEpochs, groupGauge.FilledEpochs)

					// Check that distributed coins have not changed
					s.Require().Equal(groupConfig.groupGauge.DistributedCoins, groupGauge.DistributedCoins)
					continue
				}

				// check that the group gauge distributed epoch updated
				s.Require().Equal(groupConfig.groupGauge.FilledEpochs+1, groupGauge.FilledEpochs)

				// check that the amounts distributed have updated
				actualDistributed := groupGauge.DistributedCoins.Sub(groupConfig.groupGauge.DistributedCoins)
				s.Require().Equal(groupConfig.expectedTotalDistribution, actualDistributed)

				// TODO: check that group gauge was moved to finished if applicable
			}

			// Validate that gauges received the expected amounts from all groups.
			for gaugeId, expectedDistributed := range expectedGaugeDistributions {
				internalGauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeId)
				s.Require().NoError(err)
				// check that the amounts distributed have updated
				s.Require().Equal(expectedDistributed.String(), internalGauge.Coins.String(), "gauge id: %d", gaugeId)
			}
		})
	}
}

// setupVolumes sets the volume for each pool in the passed in list of pool ids to the corresponding value in the passed in list of volumes.
func (s *KeeperTestSuite) setupVolumes(poolIds []uint64, updatedPoolVolumes []osmomath.Int) {
	// Update cumulative volumes for pools
	for i, updatedVolume := range updatedPoolVolumes {
		// Note that even though we deal with volumes as ints, they are tracked as coins to allow for tracking of more denoms in the future.
		s.App.PoolManagerKeeper.SetVolume(s.Ctx, poolIds[i], sdk.NewCoins(sdk.NewCoin(s.App.StakingKeeper.BondDenom(s.Ctx), updatedVolume)))
	}
}
