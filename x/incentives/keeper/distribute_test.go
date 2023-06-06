package keeper_test

import (
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	appParams "github.com/osmosis-labs/osmosis/v16/app/params"
	cltypes "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v16/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v16/x/lockup/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
)

var _ = suite.TestingSuite(nil)

func (s *KeeperTestSuite) SetupCLPoolAndGauge(numPools uint64, currEpochDuration time.Duration, coins sdk.Coins, gaugeStartTime time.Time) []types.Gauge {
	var gauges []types.Gauge
	// create CL Pools
	for i := uint64(1); i <= numPools; i++ {
		poolId := s.PrepareConcentratedPool().GetId()

		// get the gaugeId corresponding to the CL pool
		gaugeId, err := s.App.PoolIncentivesKeeper.GetPoolGaugeId(s.Ctx, poolId, currEpochDuration)
		s.Require().NoError(err)

		// get the gauge from the gaudeId
		gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeId)
		s.Require().NoError(err)

		gauge.Coins = coins
		gauge.StartTime = gaugeStartTime

		gauges = append(gauges, *gauge)
	}

	return gauges
}

func (s *KeeperTestSuite) CheckIncentiveRecords(poolId uint64, denom string, moduleAddr string, emissionRate sdk.Dec, startTime time.Time, remainingAmt sdk.Int, incentiveRecord cltypes.IncentiveRecord) {
	s.Require().Equal(poolId, incentiveRecord.PoolId)
	s.Require().Equal(denom, incentiveRecord.IncentiveDenom)
	s.Require().Equal(moduleAddr, incentiveRecord.IncentiveCreatorAddr)
	s.Require().Equal(emissionRate, incentiveRecord.GetIncentiveRecordBody().EmissionRate)
	s.Require().Equal(startTime, incentiveRecord.GetIncentiveRecordBody().StartTime)
	s.Require().Equal(types.DefaultConcentratedUptime, incentiveRecord.MinUptime)
	s.Require().Equal(remainingAmt, incentiveRecord.GetIncentiveRecordBody().RemainingAmount.RoundInt())
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

func (s *KeeperTestSuite) TestDistributeToConcentratedLiquidityPools() {
	defaultGauge := perpGaugeDesc{
		lockDenom:    defaultLPDenom,
		lockDuration: defaultLockDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 3000)},
	}

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
		poolType           poolmanagertypes.PoolType
		lockExist          bool
		authorizedUptimes  []time.Duration

		// expected
		expectErr             bool
		expectedDistributions sdk.Coins
	}{
		"valid case: one poolId and gaugeId": {
			numPools:              1,
			gaugeStartTime:        defaultGaugeStartTime,
			poolType:              poolmanagertypes.Concentrated,
			gaugeCoins:            sdk.NewCoins(sdk.NewCoin(defaultRewardDenom, sdk.NewInt(5000))),
			authorizedUptimes:     cltypes.SupportedUptimes,
			expectedDistributions: sdk.NewCoins(fiveKRewardCoins),
			expectErr:             false,
		},
		"valid case: gauge with multiple coins": {
			numPools:              1,
			gaugeStartTime:        defaultGaugeStartTime,
			poolType:              poolmanagertypes.Concentrated,
			gaugeCoins:            sdk.NewCoins(sdk.NewCoin(defaultRewardDenom, sdk.NewInt(5000)), sdk.NewCoin(appParams.BaseCoinUnit, sdk.NewInt(5000))),
			authorizedUptimes:     cltypes.SupportedUptimes,
			expectedDistributions: sdk.NewCoins(fiveKRewardCoins, fiveKRewardCoinsUosmo),
			expectErr:             false,
		},
		"valid case: multiple gaugeId and poolId": {
			numPools:              3,
			gaugeStartTime:        defaultGaugeStartTime,
			poolType:              poolmanagertypes.Concentrated,
			gaugeCoins:            sdk.NewCoins(sdk.NewCoin(defaultRewardDenom, sdk.NewInt(5000))),
			authorizedUptimes:     cltypes.SupportedUptimes,
			expectedDistributions: sdk.NewCoins(fifteenKRewardCoins),
			expectErr:             false,
		},
		"valid case: attempt to create balancer pool": {
			numPools:              1,
			poolType:              poolmanagertypes.Balancer,
			gaugeCoins:            sdk.NewCoins(sdk.NewCoin(defaultRewardDenom, sdk.NewInt(5000))),
			gaugeStartTime:        defaultGaugeStartTime,
			authorizedUptimes:     cltypes.SupportedUptimes,
			expectedDistributions: sdk.NewCoins(),
			expectErr:             false, // still a valid case we just donot update the CL incentive parameters
		},
		"valid case: distributing to locks since no pool associated with gauge": {
			numPools:              0,
			poolType:              poolmanagertypes.Balancer,
			gaugeCoins:            sdk.NewCoins(),
			gaugeStartTime:        defaultGaugeStartTime,
			authorizedUptimes:     cltypes.SupportedUptimes,
			expectedDistributions: sdk.NewCoins(sdk.NewCoin(defaultRewardDenom, sdk.NewInt(3000))),
			lockExist:             true,
			expectErr:             false, // we do not expect error because we run the gauge distribution to lock logic
		},
		"valid case: one poolId and gaugeId, limited authorized uptimes": {
			numPools:              1,
			gaugeStartTime:        defaultGaugeStartTime,
			poolType:              poolmanagertypes.Concentrated,
			gaugeCoins:            sdk.NewCoins(sdk.NewCoin(defaultRewardDenom, sdk.NewInt(5000))),
			authorizedUptimes:     []time.Duration{time.Nanosecond, time.Hour * 24},
			expectedDistributions: sdk.NewCoins(fiveKRewardCoins),
			expectErr:             false,
		},
		"valid case: one poolId and gaugeId, default authorized uptimes (1ns)": {
			numPools:              1,
			gaugeStartTime:        defaultGaugeStartTime,
			poolType:              poolmanagertypes.Concentrated,
			gaugeCoins:            sdk.NewCoins(sdk.NewCoin(defaultRewardDenom, sdk.NewInt(5000))),
			expectedDistributions: sdk.NewCoins(fiveKRewardCoins),
			expectErr:             false,
		},
		"invalid case: attempt to createIncentiveRecord with starttime < currentBlockTime": {
			numPools:          1,
			poolType:          poolmanagertypes.Concentrated,
			gaugeCoins:        sdk.NewCoins(sdk.NewCoin(defaultRewardDenom, sdk.NewInt(5000))),
			gaugeStartTime:    defaultGaugeStartTime.Add(-5 * time.Hour),
			authorizedUptimes: cltypes.SupportedUptimes,
			expectErr:         true,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			// setup test
			s.SetupTest()
			// We fix blocktime to ensure tests are deterministic
			s.Ctx = s.Ctx.WithBlockTime(defaultGaugeStartTime)

			// Set up authorized CL uptimes to robustly test distribution
			if tc.authorizedUptimes != nil {
				clParams := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
				clParams.AuthorizedUptimes = tc.authorizedUptimes
				s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, clParams)
			}

			var gauges []types.Gauge
			// prepare the minting account
			addr := s.TestAccs[0]
			// mints coins so supply exists on chain
			s.FundAcc(addr, coinsToMint)

			// make sure the module has enough funds
			err := s.App.BankKeeper.SendCoinsFromAccountToModule(s.Ctx, addr, types.ModuleName, coinsToMint)
			s.Require().NoError(err)

			// prepare a CL Pool that creates gauge at the end of createPool
			if tc.poolType == poolmanagertypes.Concentrated {
				gauges = s.SetupCLPoolAndGauge(uint64(tc.numPools), currentEpoch.Duration, tc.gaugeCoins, tc.gaugeStartTime)
			}

			var addrs []sdk.AccAddress
			// this is the case where retrieving pool fails so we run the else logic where gauge is distributed via locks
			if tc.lockExist {
				gauges = s.SetupGauges([]perpGaugeDesc{defaultGauge}, defaultLPDenom)
				addrs = s.SetupUserLocks([]userLocks{oneLockupUser})
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

				// this check is specifically for CL pool gauges, because we donot create pools other than CL
				if tc.poolType == poolmanagertypes.Concentrated {
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
							// get poolId from GaugeId
							poolId, err := s.App.PoolIncentivesKeeper.GetPoolIdFromGaugeId(s.Ctx, gauge.GetId(), currentEpoch.Duration)
							s.Require().NoError(err)

							// GetIncentiveRecord to see if pools received incentives properly
							incentiveRecord, err := s.App.ConcentratedLiquidityKeeper.GetIncentiveRecord(s.Ctx, poolId, defaultRewardDenom, types.DefaultConcentratedUptime, s.App.AccountKeeper.GetModuleAddress(types.ModuleName))
							s.Require().NoError(err)

							expectedEmissionRate := sdk.NewDecFromInt(coin.Amount).QuoTruncate(sdk.NewDec(int64(currentEpoch.Duration.Seconds())))

							// check every parameter in incentiveRecord so that it matches what we created
							s.CheckIncentiveRecords(poolId, defaultRewardDenom, s.App.AccountKeeper.GetModuleAddress(types.ModuleName).String(), expectedEmissionRate, gauge.StartTime, fiveKRewardCoins.Amount, incentiveRecord)
						}
					}
				}

				// this check is specifically for gauge distribution via locks
				for i, addr := range addrs {
					bal := s.App.BankKeeper.GetAllBalances(s.Ctx, addr)
					s.Require().Equal(tc.expectedDistributions[i].String(), bal.String(), "test %v, person %d", name, i)
				}

				// check the totalAmount of tokens distributed, for both lock gauges and CL pool gauges
				s.Require().Equal(tc.expectedDistributions, totalDistributedCoins)
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

// TestNoLockPerpetualGaugeDistribution tests that the creation of a perp gauge that has no locks associated does not distribute any tokens.
func (s *KeeperTestSuite) TestNoLockPerpetualGaugeDistribution() {
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
}

// TestNoLockNonPerpetualGaugeDistribution tests that the creation of a non perp gauge that has no locks associated does not distribute any tokens.
func (s *KeeperTestSuite) TestNoLockNonPerpetualGaugeDistribution() {
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
				s.App.PoolIncentivesKeeper.SetPoolGaugeId(s.Ctx, validPoolId, duration, poolIdOne)
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

func (s *KeeperTestSuite) TestDistributeConcentratedLiquidity() {
	var (
		timeBeforeBlock   = time.Unix(0, 0)
		defaultBlockTime  = timeBeforeBlock.Add(10 * time.Second)
		defaultAmountCoin = sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 3000)}
		defaultGauge      = perpGaugeDesc{
			lockDenom:    defaultLPDenom,
			lockDuration: defaultLockDuration,
			rewardAmount: defaultAmountCoin,
		}
		withLength = func(gauge perpGaugeDesc, length time.Duration) perpGaugeDesc {
			gauge.lockDuration = length
			return gauge
		}
		withAmount = func(gauge perpGaugeDesc, amount sdk.Coins) perpGaugeDesc {
			gauge.rewardAmount = amount
			return gauge
		}
	)

	type distributeConcentratedLiquidityInternalTestCase struct {
		name              string
		poolId            uint64
		sender            sdk.AccAddress
		incentiveDenom    string
		incentiveAmount   sdk.Int
		emissionRate      sdk.Dec
		startTime         time.Time
		minUptime         time.Duration
		expectedCoins     sdk.Coins
		gauge             perpGaugeDesc
		authorizedUptimes []time.Duration
		expectedError     bool
	}

	testCases := []distributeConcentratedLiquidityInternalTestCase{
		{
			name:              "valid: valid incentive record with valid gauge",
			poolId:            1,
			sender:            s.TestAccs[0],
			incentiveDenom:    defaultRewardDenom,
			incentiveAmount:   sdk.NewInt(100),
			emissionRate:      sdk.NewDec(1),
			startTime:         defaultBlockTime,
			minUptime:         time.Hour * 24,
			gauge:             defaultGauge,
			authorizedUptimes: []time.Duration{time.Hour * 24},

			expectedCoins: sdk.NewCoins(sdk.NewCoin(defaultRewardDenom, sdk.NewInt(100))),
		},
		{
			name:            "valid: valid incentive record with valid gauge (default authorized uptimes)",
			poolId:          1,
			sender:          s.TestAccs[0],
			incentiveDenom:  defaultRewardDenom,
			incentiveAmount: sdk.NewInt(100),
			emissionRate:    sdk.NewDec(1),
			startTime:       defaultBlockTime,
			minUptime:       time.Nanosecond,
			gauge:           defaultGauge,

			expectedCoins: sdk.NewCoins(sdk.NewCoin(defaultRewardDenom, sdk.NewInt(100))),
		},
		{
			name:              "valid: valid incentive with double length record with valid gauge",
			poolId:            1,
			sender:            s.TestAccs[0],
			incentiveDenom:    defaultRewardDenom,
			incentiveAmount:   sdk.NewInt(100),
			emissionRate:      sdk.NewDec(1),
			startTime:         defaultBlockTime,
			minUptime:         time.Hour * 24,
			gauge:             withLength(defaultGauge, defaultGauge.lockDuration*2),
			authorizedUptimes: []time.Duration{time.Hour * 24},

			expectedCoins: sdk.NewCoins(sdk.NewCoin(defaultRewardDenom, sdk.NewInt(100))),
		},
		{
			name:              "valid: valid incentive with double amount record and valid gauge",
			poolId:            1,
			sender:            s.TestAccs[0],
			incentiveDenom:    defaultRewardDenom,
			incentiveAmount:   sdk.NewInt(100),
			emissionRate:      sdk.NewDec(1),
			startTime:         defaultBlockTime,
			minUptime:         time.Hour * 24,
			gauge:             withAmount(defaultGauge, defaultAmountCoin.Add(defaultAmountCoin...)),
			authorizedUptimes: []time.Duration{time.Hour * 24},

			expectedCoins: sdk.NewCoins(sdk.NewCoin(defaultRewardDenom, sdk.NewInt(100))),
		},
		{
			name:              "Invalid Case: invalid incentive Record with valid Gauge",
			poolId:            1,
			sender:            s.TestAccs[0],
			incentiveDenom:    defaultRewardDenom,
			incentiveAmount:   sdk.NewInt(200),
			emissionRate:      sdk.NewDec(2),
			startTime:         timeBeforeBlock,
			minUptime:         time.Hour * 2,
			gauge:             defaultGauge,
			authorizedUptimes: cltypes.SupportedUptimes,

			expectedError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultBlockTime)

			// Set up authorized CL uptimes to robustly test distribution
			if tc.authorizedUptimes != nil {
				clParams := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
				clParams.AuthorizedUptimes = tc.authorizedUptimes
				s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, clParams)
			}

			s.PrepareConcentratedPool()

			s.FundAcc(tc.sender, sdk.NewCoins(sdk.NewCoin(defaultRewardDenom, sdk.NewInt(10000))))
			gauges := s.SetupGauges([]perpGaugeDesc{tc.gauge}, defaultRewardDenom)

			err := s.App.IncentivesKeeper.DistributeConcentratedLiquidity(s.Ctx, tc.poolId, tc.sender, sdk.NewCoin(tc.incentiveDenom, tc.incentiveAmount), tc.emissionRate, tc.startTime, tc.minUptime, gauges[0])
			if tc.expectedError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gauges[0].Id)
				s.Require().NoError(err)

				s.Require().Equal(gauge.DistributedCoins, gauges[0].DistributedCoins.Add(tc.expectedCoins...))
			}
		})
	}
}

// TestFunctionalConcentratedLiquidityGaugeDistribute is a functional test that covers more complex scenarios relating to distributing incentives through gauges
// at the end of each epoch.
//
// we expect these events to occour at the end of each epoch;
// - mint and distribute coins according to the configuration.
// - distribution happens proportional to the gauge weight in the record.
// - moduleAccount holds the funds and handles distribution.
//
// Testing strategy:
// 1. Initialize variables.
// 2. Setup CL pool and gauge (gauge automatically gets created at the end of CL pool creation).
// 3. Update the gauge by adding rewards (should happen at end of each epoch).
// 4. get updated gauges and distribute the rewards that was added.
// 5. Check that incentive has been correctly created and gauge has been correctly updated.
func (s *KeeperTestSuite) TestFunctionalConcentratedLiquidityGaugeDistribute() {
	s.SetupTest()

	// get epoch information
	incentivesParams := s.App.IncentivesKeeper.GetParams(s.Ctx).DistrEpochIdentifier
	currentEpoch := s.App.EpochsKeeper.GetEpochInfo(s.Ctx, incentivesParams)

	// 1. Initialize variables
	poolIds := []uint64{1, 2, 3}
	startTime := s.Ctx.BlockTime()
	// 2. Setup CL pool and gauge (gauge automatically gets created at the end of CL pool creation).
	gauges := s.SetupCLPoolAndGauge(uint64(len(poolIds)), currentEpoch.Duration, sdk.NewCoins(), startTime)

	//Test1: distribute 2 coins using CL gauge
	gaugeRewards := sdk.NewCoins(sdk.NewCoin("uusdc", sdk.NewInt(10_000)), sdk.NewCoin("uosmo", sdk.NewInt(10_000)))
	expectedCoinsToDistribute := sdk.NewCoins(sdk.NewCoin("uusdc", sdk.NewInt(30_000)), sdk.NewCoin("uosmo", sdk.NewInt(30_000))) // 3 pool, 10k each token per pool = 30k each token
	s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10_000_000_000)),
		sdk.NewCoin("uusdc", sdk.NewInt(10_000_000_000)),
		sdk.NewCoin("uatom", sdk.NewInt(10_000_000_000)),
	))

	s.TestFunctionalConcentratedDistributeGaugeHelper(poolIds, gauges, startTime, gaugeRewards, expectedCoinsToDistribute)

	//Test2: distribute 3 coins using CL gauge
	poolIds1 := []uint64{4, 5, 6}
	gauges1 := s.SetupCLPoolAndGauge(uint64(len(poolIds1)), currentEpoch.Duration, sdk.NewCoins(), startTime)
	gaugeRewards1 := sdk.NewCoins(sdk.NewCoin("uusdc", sdk.NewInt(10_000)), sdk.NewCoin("uosmo", sdk.NewInt(10_000)), sdk.NewCoin("uatom", sdk.NewInt(10_000)))
	expectedCoinsToDistribute1 := sdk.NewCoins(sdk.NewCoin("uusdc", sdk.NewInt(30_000)), sdk.NewCoin("uosmo", sdk.NewInt(30_000)), sdk.NewCoin("uatom", sdk.NewInt(30_000))) // 3 pool, 10k each token per pool = 30k each token

	s.TestFunctionalConcentratedDistributeGaugeHelper(poolIds1, gauges1, startTime, gaugeRewards1, expectedCoinsToDistribute1)
}

func (s *KeeperTestSuite) TestFunctionalConcentratedDistributeGaugeHelper(poolIds []uint64, gauges_original []types.Gauge, startTime time.Time, gaugeRewards sdk.Coins, expectedCoinsToDistribute sdk.Coins) {
	// 3. update the gauge by adding rewards (should happen at end of each epoch)
	for _, gauge := range gauges_original {
		// adds token from a specific owner to the moduleAddress, which updated the gauge coins in the process
		err := s.App.IncentivesKeeper.AddToGaugeRewards(s.Ctx, s.TestAccs[0], gaugeRewards, gauge.Id)
		s.Require().NoError(err)
	}

	// make sure the module recieved the funds
	moduleAddress := s.App.AccountKeeper.GetModuleAddress(types.ModuleName)
	actualModuleCoins := s.App.BankKeeper.GetAllBalances(s.Ctx, moduleAddress)
	s.Require().Equal(expectedCoinsToDistribute, actualModuleCoins)

	// 4. get updated gauges and distribute the rewards that was added
	for _, gauge := range gauges_original {
		err := s.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(s.Ctx, gauge)
		s.Require().NoError(err)
	}

	gauges_afterEpoch := s.App.IncentivesKeeper.GetActiveGauges(s.Ctx)
	s.Require().Equal(len(gauges_original), len(gauges_afterEpoch))

	// we run distribute at the end of each epoch and expect all the coins in gauge to be distributed
	totalDistributedCoins, err := s.App.IncentivesKeeper.Distribute(s.Ctx, gauges_afterEpoch)
	s.Require().NoError(err)

	s.Require().Equal(expectedCoinsToDistribute, totalDistributedCoins)

	// 5. Check that incentive has been correctly created and gauge has been correctly updated
	for _, poolId := range poolIds {
		for _, gaugeReward := range gaugeRewards {
			incentivesParams := s.App.IncentivesKeeper.GetParams(s.Ctx).DistrEpochIdentifier
			currentEpoch := s.App.EpochsKeeper.GetEpochInfo(s.Ctx, incentivesParams)
			expectedEmissionRate := sdk.NewDecFromInt(gaugeReward.Amount).QuoTruncate(sdk.NewDec(int64(currentEpoch.Duration.Seconds())))

			incentiveRecord, err := s.App.ConcentratedLiquidityKeeper.GetIncentiveRecord(s.Ctx, poolId, gaugeReward.Denom, types.DefaultConcentratedUptime, moduleAddress)
			s.Require().NoError(err)

			s.CheckIncentiveRecords(poolId, gaugeReward.Denom, moduleAddress.String(), expectedEmissionRate, startTime, gaugeReward.Amount, incentiveRecord)
		}
	}

	// check gauge post distribution
	gauges_afterDistribute := s.App.IncentivesKeeper.GetActiveGauges(s.Ctx)
	s.Require().Equal(len(gauges_afterDistribute), len(gauges_afterEpoch))

	for idx, gauge := range gauges_afterDistribute {
		s.Require().Equal(gauges_afterEpoch[idx].FilledEpochs+1, gauge.FilledEpochs)
	}

	// check module account amount decreased
	moduleCoins_afterDistribute := s.App.BankKeeper.GetAllBalances(s.Ctx, moduleAddress)
	s.Require().Equal(sdk.Coins{}, moduleCoins_afterDistribute)

	for _, gauge := range gauges_afterDistribute {
		err := s.App.IncentivesKeeper.MoveActiveGaugeToFinishedGauge(s.Ctx, gauge)
		s.Require().NoError(err)
	}
}
