package keeper_test

import (
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	appParams "github.com/osmosis-labs/osmosis/v15/app/params"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

var _ = suite.TestingSuite(nil)

// TestDistribute tests that when the distribute command is executed on a provided gauge
// that the correct amount of rewards is sent to the correct lock owners.
func (suite *KeeperTestSuite) TestDistribute() {
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
	}
	for _, tc := range tests {
		suite.SetupTest()
		// setup gauges and the locks defined in the above tests, then distribute to them
		gauges := suite.SetupGauges(tc.gauges, defaultLPDenom)
		addrs := suite.SetupUserLocks(tc.users)

		_, err := suite.App.IncentivesKeeper.Distribute(suite.Ctx, gauges)
		suite.Require().NoError(err)
		// check expected rewards against actual rewards received
		for i, addr := range addrs {
			bal := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr)
			suite.Require().Equal(tc.expectedRewards[i].String(), bal.String(), "test %v, person %d", tc.name, i)
		}
	}
}

func (suite *KeeperTestSuite) TestDistributeToConcentratedLiquidityPools() {
	defaultGauge := perpGaugeDesc{
		lockDenom:    defaultLPDenom,
		lockDuration: defaultLockDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 3000)},
	}

	fiveKRewardCoins := sdk.NewInt64Coin(defaultRewardDenom, 5000)
	fiveKRewardCoinsUosmo := sdk.NewInt64Coin(appParams.BaseCoinUnit, 5000)
	fifteenKRewardCoins := sdk.NewInt64Coin(defaultRewardDenom, 15000)

	coinsToMint := sdk.NewCoins(sdk.NewCoin(defaultRewardDenom, sdk.NewInt(10000000)), sdk.NewCoin(appParams.BaseCoinUnit, sdk.NewInt(10000000)))
	defaultGaugeStartTime := suite.Ctx.BlockTime()

	incentivesParams := suite.App.IncentivesKeeper.GetParams(suite.Ctx).DistrEpochIdentifier
	currentEpoch := suite.App.EpochsKeeper.GetEpochInfo(suite.Ctx, incentivesParams)

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
		suite.Run(name, func() {
			// setup test
			suite.SetupTest()
			// We fix blocktime to ensure tests are deterministic
			suite.Ctx = suite.Ctx.WithBlockTime(defaultGaugeStartTime)

			// Set up authorized CL uptimes to robustly test distribution
			if tc.authorizedUptimes != nil {
				clParams := suite.App.ConcentratedLiquidityKeeper.GetParams(suite.Ctx)
				clParams.AuthorizedUptimes = tc.authorizedUptimes
				suite.App.ConcentratedLiquidityKeeper.SetParams(suite.Ctx, clParams)
			}

			var gauges []types.Gauge

			// prepare the minting account
			addr := suite.TestAccs[0]
			// mints coins so supply exists on chain
			suite.FundAcc(addr, coinsToMint)

			// make sure the module has enough funds
			err := suite.App.BankKeeper.SendCoinsFromAccountToModule(suite.Ctx, addr, types.ModuleName, coinsToMint)
			suite.Require().NoError(err)

			var poolId uint64
			// prepare a CL Pool that creates gauge at the end of createPool
			if tc.poolType == poolmanagertypes.Concentrated {
				for i := 0; i < tc.numPools; i++ {
					poolId = suite.PrepareConcentratedPool().GetId()

					// get the gaugeId corresponding to the CL pool
					gaugeId, err := suite.App.PoolIncentivesKeeper.GetPoolGaugeId(suite.Ctx, poolId, currentEpoch.Duration)
					suite.Require().NoError(err)

					// get the gauge from the gaudeId
					gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId)
					suite.Require().NoError(err)

					gauge.Coins = tc.gaugeCoins
					gauge.StartTime = tc.gaugeStartTime
					gauges = append(gauges, *gauge)
				}
			}

			var addrs []sdk.AccAddress
			// this is the case where retrieving pool fails so we run the else logic where gauge is distributed via locks
			if tc.lockExist {
				gauges = suite.SetupGauges([]perpGaugeDesc{defaultGauge}, defaultLPDenom)
				addrs = suite.SetupUserLocks([]userLocks{oneLockupUser})
			}

			// Distribute tokens from the gauge
			totalDistributedCoins, err := suite.App.IncentivesKeeper.Distribute(suite.Ctx, gauges)
			if tc.expectErr {
				suite.Require().Error(err)

				// module account amount must stay the same
				balance := suite.App.BankKeeper.GetAllBalances(suite.Ctx, suite.App.AccountKeeper.GetModuleAddress(types.ModuleName))
				suite.Require().Equal(coinsToMint, balance)

				for _, gauge := range gauges {
					for _, coin := range gauge.Coins {
						// get poolId from GaugeId
						poolId, err := suite.App.PoolIncentivesKeeper.GetPoolIdFromGaugeId(suite.Ctx, gauge.GetId(), currentEpoch.Duration)
						suite.Require().NoError(err)

						// check that incentive record wasn't created
						_, err = suite.App.ConcentratedLiquidityKeeper.GetIncentiveRecord(suite.Ctx, poolId, coin.Denom, currentEpoch.Duration, suite.App.AccountKeeper.GetModuleAddress(types.ModuleName))
						suite.Require().Error(err)
					}
				}
			} else {
				suite.Require().NoError(err)

				// this check is specifically for CL pool gauges, because we donot create pools other than CL
				if tc.poolType == poolmanagertypes.Concentrated {
					// check that gauge is not empty
					suite.Require().NotEqual(len(gauges), 0)

					// check if module amount got deducted correctly
					balance := suite.App.BankKeeper.GetAllBalances(suite.Ctx, suite.App.AccountKeeper.GetModuleAddress(types.ModuleName))
					for _, coin := range balance {
						actualbalanceAfterDistribution := coinsToMint.AmountOf(coin.Denom).Sub(coin.Amount)
						suite.Require().Equal(tc.expectedDistributions.AmountOf(coin.Denom).Add(sdk.ZeroInt()), actualbalanceAfterDistribution.Add(sdk.ZeroInt()))
					}

					for _, gauge := range gauges {
						for _, coin := range gauge.Coins {
							// get poolId from GaugeId
							poolId, err := suite.App.PoolIncentivesKeeper.GetPoolIdFromGaugeId(suite.Ctx, gauge.GetId(), currentEpoch.Duration)
							suite.Require().NoError(err)

							// GetIncentiveRecord to see if pools received incentives properly
							incentiveRecord, err := suite.App.ConcentratedLiquidityKeeper.GetIncentiveRecord(suite.Ctx, poolId, defaultRewardDenom, types.DefaultConcentratedUptime, suite.App.AccountKeeper.GetModuleAddress(types.ModuleName))
							suite.Require().NoError(err)

							expectedEmissionRate := sdk.NewDecFromInt(coin.Amount).QuoTruncate(sdk.NewDec(int64(currentEpoch.Duration.Seconds())))

							// check every parameter in incentiveRecord so that it matches what we created
							suite.Require().Equal(poolId, incentiveRecord.PoolId)
							suite.Require().Equal(defaultRewardDenom, incentiveRecord.IncentiveDenom)
							suite.Require().Equal(suite.App.AccountKeeper.GetModuleAddress(types.ModuleName).String(), incentiveRecord.IncentiveCreatorAddr)
							suite.Require().Equal(expectedEmissionRate, incentiveRecord.GetIncentiveRecordBody().EmissionRate)
							suite.Require().Equal(gauge.StartTime, incentiveRecord.GetIncentiveRecordBody().StartTime)
							suite.Require().Equal(types.DefaultConcentratedUptime, incentiveRecord.MinUptime)
							suite.Require().Equal(fiveKRewardCoins.Amount, incentiveRecord.GetIncentiveRecordBody().RemainingAmount.RoundInt())
						}
					}
				}

				// this check is specifically for gauge distribution via locks
				for i, addr := range addrs {
					bal := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr)
					suite.Require().Equal(tc.expectedDistributions[i].String(), bal.String(), "test %v, person %d", name, i)
				}

				// check the totalAmount of tokens distributed, for both lock gauges and CL pool gauges
				suite.Require().Equal(tc.expectedDistributions, totalDistributedCoins)
			}
		})
	}
}

// TestSyntheticDistribute tests that when the distribute command is executed on a provided gauge
// the correct amount of rewards is sent to the correct synthetic lock owners.
func (suite *KeeperTestSuite) TestSyntheticDistribute() {
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
		suite.SetupTest()
		// setup gauges and the synthetic locks defined in the above tests, then distribute to them
		gauges := suite.SetupGauges(tc.gauges, defaultLPSyntheticDenom)
		addrs := suite.SetupUserSyntheticLocks(tc.users)
		_, err := suite.App.IncentivesKeeper.Distribute(suite.Ctx, gauges)
		suite.Require().NoError(err)
		// check expected rewards against actual rewards received
		for i, addr := range addrs {
			var rewards string
			bal := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr)
			// extract the superbonding tokens from the rewards distribution
			// TODO: figure out a less hacky way of doing this
			if strings.Contains(bal.String(), "lptoken/superbonding,") {
				rewards = strings.Split(bal.String(), "lptoken/superbonding,")[1]
			}
			suite.Require().Equal(tc.expectedRewards[i].String(), rewards, "test %v, person %d", tc.name, i)
		}
	}
}

// TestGetModuleToDistributeCoins tests the sum of coins yet to be distributed for all of the module is correct.
func (suite *KeeperTestSuite) TestGetModuleToDistributeCoins() {
	suite.SetupTest()

	// check that the sum of coins yet to be distributed is nil
	coins := suite.App.IncentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	// setup a non perpetual lock and gauge
	_, gaugeID, gaugeCoins, startTime := suite.SetupLockAndGauge(false)

	// check that the sum of coins yet to be distributed is equal to the newly created gaugeCoins
	coins = suite.App.IncentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, gaugeCoins)

	// add coins to the previous gauge and check that the sum of coins yet to be distributed includes these new coins
	addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
	suite.AddToGauge(addCoins, gaugeID)
	coins = suite.App.IncentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, gaugeCoins.Add(addCoins...))

	// create a new gauge
	// check that the sum of coins yet to be distributed is equal to the gauge1 and gauge2 coins combined
	_, _, gaugeCoins2, _ := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 1000)})
	coins = suite.App.IncentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, gaugeCoins.Add(addCoins...).Add(gaugeCoins2...))

	// move all created gauges from upcoming to active
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	err = suite.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 105)})

	// check gauge changes after distribution
	coins = suite.App.IncentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, gaugeCoins.Add(addCoins...).Add(gaugeCoins2...).Sub(distrCoins))
}

// TestGetModuleDistributedCoins tests that the sum of coins that have been distributed so far for all of the module is correct.
func (suite *KeeperTestSuite) TestGetModuleDistributedCoins() {
	suite.SetupTest()

	// check that the sum of coins yet to be distributed is nil
	coins := suite.App.IncentivesKeeper.GetModuleDistributedCoins(suite.Ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	// setup a non perpetual lock and gauge
	_, gaugeID, _, startTime := suite.SetupLockAndGauge(false)

	// check that the sum of coins yet to be distributed is equal to the newly created gaugeCoins
	coins = suite.App.IncentivesKeeper.GetModuleDistributedCoins(suite.Ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	// move all created gauges from upcoming to active
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	err = suite.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 5)})

	// check gauge changes after distribution
	coins = suite.App.IncentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, distrCoins)
}

// TestNoLockPerpetualGaugeDistribution tests that the creation of a perp gauge that has no locks associated does not distribute any tokens.
func (suite *KeeperTestSuite) TestNoLockPerpetualGaugeDistribution() {
	suite.SetupTest()

	// setup a perpetual gauge with no associated locks
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	gaugeID, _, _, startTime := suite.SetupNewGauge(true, coins)

	// ensure the created gauge has not completed distribution
	gauges := suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
	suite.Require().Len(gauges, 1)

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
	suite.Require().Equal(gauges[0].String(), expectedGauge.String())

	// move the created gauge from upcoming to active
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	err = suite.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers, since it's perpetual distribute everything on single distribution
	distrCoins, err := suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins(nil))

	// check state is same after distribution
	gauges = suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
	suite.Require().Len(gauges, 1)
	suite.Require().Equal(gauges[0].String(), expectedGauge.String())
}

// TestNoLockNonPerpetualGaugeDistribution tests that the creation of a non perp gauge that has no locks associated does not distribute any tokens.
func (suite *KeeperTestSuite) TestNoLockNonPerpetualGaugeDistribution() {
	suite.SetupTest()

	// setup non-perpetual gauge with no associated locks
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	gaugeID, _, _, startTime := suite.SetupNewGauge(false, coins)

	// ensure the created gauge has not completed distribution
	gauges := suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
	suite.Require().Len(gauges, 1)

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
	suite.Require().Equal(gauges[0].String(), expectedGauge.String())

	// move the created gauge from upcoming to active
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	err = suite.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins(nil))

	// check state is same after distribution
	gauges = suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
	suite.Require().Len(gauges, 1)
	suite.Require().Equal(gauges[0].String(), expectedGauge.String())
}

func (suite *KeeperTestSuite) TestGetPoolFromGaugeId() {
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
		suite.Run(tc.name, func() {
			suite.SetupTest()

			incParams := suite.App.IncentivesKeeper.GetEpochInfo(suite.Ctx)
			duration := incParams.GetDuration()

			if !tc.shouldAvoidCreatingPool {
				if tc.expectedPoolType == poolmanagertypes.Concentrated {
					suite.PrepareConcentratedPool()
				} else {
					suite.PrepareBalancerPool()
				}
			}

			if tc.shouldSetPoolGaugeId {
				suite.App.PoolIncentivesKeeper.SetPoolGaugeId(suite.Ctx, validPoolId, duration, poolIdOne)
			}

			pool, err := suite.App.IncentivesKeeper.GetPoolFromGaugeId(suite.Ctx, tc.gaugeId, duration)
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Nil(pool)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(pool)
				suite.Require().Equal(pool.GetType(), tc.expectedPoolType)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestDistributeConcentratedLiquidity() {
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
			sender:            suite.TestAccs[0],
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
			sender:          suite.TestAccs[0],
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
			sender:            suite.TestAccs[0],
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
			sender:            suite.TestAccs[0],
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
			sender:            suite.TestAccs[0],
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
		suite.Run(tc.name, func() {
			suite.SetupTest()
			suite.Ctx = suite.Ctx.WithBlockTime(defaultBlockTime)

			// Set up authorized CL uptimes to robustly test distribution
			if tc.authorizedUptimes != nil {
				clParams := suite.App.ConcentratedLiquidityKeeper.GetParams(suite.Ctx)
				clParams.AuthorizedUptimes = tc.authorizedUptimes
				suite.App.ConcentratedLiquidityKeeper.SetParams(suite.Ctx, clParams)
			}

			suite.PrepareConcentratedPool()

			suite.FundAcc(tc.sender, sdk.NewCoins(sdk.NewCoin(defaultRewardDenom, sdk.NewInt(10000))))
			gauges := suite.SetupGauges([]perpGaugeDesc{tc.gauge}, defaultRewardDenom)

			err := suite.App.IncentivesKeeper.DistributeConcentratedLiquidity(suite.Ctx, tc.poolId, tc.sender, sdk.NewCoin(tc.incentiveDenom, tc.incentiveAmount), tc.emissionRate, tc.startTime, tc.minUptime, gauges[0])
			if tc.expectedError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)

				gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gauges[0].Id)
				suite.Require().NoError(err)

				suite.Require().Equal(gauge.DistributedCoins, gauges[0].DistributedCoins.Add(tc.expectedCoins...))
			}
		})
	}
}
