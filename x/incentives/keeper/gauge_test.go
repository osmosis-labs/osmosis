package keeper_test

import (
	"fmt"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v19/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v19/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ = suite.TestingSuite(nil)

// TestInvalidDurationGaugeCreationValidation tests error handling for creating a gauge with an invalid duration.
func (s *KeeperTestSuite) TestInvalidDurationGaugeCreationValidation() {
	s.SetupTest()

	addrs := s.SetupManyLocks(1, defaultLiquidTokens, defaultLPTokens, defaultLockDuration)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         defaultLPDenom,
		Duration:      defaultLockDuration / 2, // 0.5 second, invalid duration
	}
	_, err := s.App.IncentivesKeeper.CreateGauge(s.Ctx, false, addrs[0], defaultLiquidTokens, distrTo, time.Time{}, 1, 0)
	s.Require().Error(err)

	distrTo.Duration = defaultLockDuration
	_, err = s.App.IncentivesKeeper.CreateGauge(s.Ctx, false, addrs[0], defaultLiquidTokens, distrTo, time.Time{}, 1, 0)
	s.Require().NoError(err)
}

// TestNonExistentDenomGaugeCreation tests error handling for creating a gauge with an invalid denom.
func (s *KeeperTestSuite) TestNonExistentDenomGaugeCreation() {
	s.SetupTest()

	addrNoSupply := sdk.AccAddress([]byte("Gauge_Creation_Addr_"))
	addrs := s.SetupManyLocks(1, defaultLiquidTokens, defaultLPTokens, defaultLockDuration)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         defaultLPDenom,
		Duration:      defaultLockDuration,
	}
	_, err := s.App.IncentivesKeeper.CreateGauge(s.Ctx, false, addrNoSupply, defaultLiquidTokens, distrTo, time.Time{}, 1, 0)
	s.Require().Error(err)

	_, err = s.App.IncentivesKeeper.CreateGauge(s.Ctx, false, addrs[0], defaultLiquidTokens, distrTo, time.Time{}, 1, 0)
	s.Require().NoError(err)
}

// TestGaugeOperations tests perpetual and non-perpetual gauge distribution logic using the gauges by denom keeper.
func (s *KeeperTestSuite) TestGaugeOperations() {
	testCases := []struct {
		isPerpetual bool
		numLocks    int
	}{
		{
			isPerpetual: true,
			numLocks:    1,
		},
		{
			isPerpetual: false,
			numLocks:    1,
		},
		{
			isPerpetual: true,
			numLocks:    2,
		},
		{
			isPerpetual: false,
			numLocks:    2,
		},
	}
	for _, tc := range testCases {
		// test for module get gauges
		s.SetupTest()

		// initial module gauges check
		gauges := s.App.IncentivesKeeper.GetNotFinishedGauges(s.Ctx)
		s.Require().Len(gauges, 0)
		gaugeIdsByDenom := s.App.IncentivesKeeper.GetAllGaugeIDsByDenom(s.Ctx, "lptoken")
		s.Require().Len(gaugeIdsByDenom, 0)

		// setup lock and gauge
		lockOwners := s.SetupManyLocks(tc.numLocks, defaultLiquidTokens, defaultLPTokens, time.Second)
		gaugeID, _, coins, startTime := s.SetupNewGauge(tc.isPerpetual, sdk.Coins{sdk.NewInt64Coin("stake", 12)})
		// evenly distributed per lock
		expectedCoinsPerLock := sdk.Coins{sdk.NewInt64Coin("stake", 12/int64(tc.numLocks))}
		// set expected epochs
		var expectedNumEpochsPaidOver int
		if tc.isPerpetual {
			expectedNumEpochsPaidOver = 1
		} else {
			expectedNumEpochsPaidOver = 2
		}

		// check gauges
		gauges = s.App.IncentivesKeeper.GetNotFinishedGauges(s.Ctx)
		s.Require().Len(gauges, 1)
		expectedGauge := types.Gauge{
			Id:          gaugeID,
			IsPerpetual: tc.isPerpetual,
			DistributeTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.ByDuration,
				Denom:         "lptoken",
				Duration:      time.Second,
			},
			Coins:             coins,
			NumEpochsPaidOver: uint64(expectedNumEpochsPaidOver),
			FilledEpochs:      0,
			DistributedCoins:  sdk.Coins{},
			StartTime:         startTime,
		}
		s.Require().Equal(expectedGauge.String(), gauges[0].String())

		// check gauge ids by denom
		gaugeIdsByDenom = s.App.IncentivesKeeper.GetAllGaugeIDsByDenom(s.Ctx, "lptoken")
		s.Require().Len(gaugeIdsByDenom, 1)
		s.Require().Equal(gaugeID, gaugeIdsByDenom[0])

		// check rewards estimation
		rewardsEst := s.App.IncentivesKeeper.GetRewardsEst(s.Ctx, lockOwners[0], []lockuptypes.PeriodLock{}, 100)
		s.Require().Equal(expectedCoinsPerLock.String(), rewardsEst.String())

		// check gauges
		gauges = s.App.IncentivesKeeper.GetNotFinishedGauges(s.Ctx)
		s.Require().Len(gauges, 1)
		s.Require().Equal(expectedGauge.String(), gauges[0].String())

		// check upcoming gauges
		gauges = s.App.IncentivesKeeper.GetUpcomingGauges(s.Ctx)
		s.Require().Len(gauges, 1)

		// start distribution
		s.Ctx = s.Ctx.WithBlockTime(startTime)
		gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeID)
		s.Require().NoError(err)
		err = s.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(s.Ctx, *gauge)
		s.Require().NoError(err)

		// check active gauges
		gauges = s.App.IncentivesKeeper.GetActiveGauges(s.Ctx)
		s.Require().Len(gauges, 1)

		// check upcoming gauges
		gauges = s.App.IncentivesKeeper.GetUpcomingGauges(s.Ctx)
		s.Require().Len(gauges, 0)

		// check gauge ids by denom
		gaugeIdsByDenom = s.App.IncentivesKeeper.GetAllGaugeIDsByDenom(s.Ctx, "lptoken")
		s.Require().Len(gaugeIdsByDenom, 1)

		// check gauge ids by other denom
		gaugeIdsByDenom = s.App.IncentivesKeeper.GetAllGaugeIDsByDenom(s.Ctx, "lpt")
		s.Require().Len(gaugeIdsByDenom, 0)

		// distribute coins to stakers
		distrCoins, err := s.App.IncentivesKeeper.Distribute(s.Ctx, []types.Gauge{*gauge})
		s.Require().NoError(err)
		// We hardcoded 12 "stake" tokens when initializing gauge
		s.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", int64(12/expectedNumEpochsPaidOver))}, distrCoins)

		if tc.isPerpetual {
			// distributing twice without adding more for perpetual gauge
			gauge, err = s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeID)
			s.Require().NoError(err)
			distrCoins, err = s.App.IncentivesKeeper.Distribute(s.Ctx, []types.Gauge{*gauge})
			s.Require().NoError(err)
			s.Require().True(distrCoins.Empty())

			// add to gauge
			addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
			s.AddToGauge(addCoins, gaugeID)

			// distributing twice with adding more for perpetual gauge
			gauge, err = s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeID)
			s.Require().NoError(err)
			distrCoins, err = s.App.IncentivesKeeper.Distribute(s.Ctx, []types.Gauge{*gauge})
			s.Require().NoError(err)
			s.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 200)}, distrCoins)
		} else {
			// add to gauge
			addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
			s.AddToGauge(addCoins, gaugeID)
		}

		// check active gauges
		gauges = s.App.IncentivesKeeper.GetActiveGauges(s.Ctx)
		s.Require().Len(gauges, 1)

		// check gauge ids by denom
		gaugeIdsByDenom = s.App.IncentivesKeeper.GetAllGaugeIDsByDenom(s.Ctx, "lptoken")
		s.Require().Len(gaugeIdsByDenom, 1)

		// finish distribution for non perpetual gauge
		if !tc.isPerpetual {
			err = s.App.IncentivesKeeper.MoveActiveGaugeToFinishedGauge(s.Ctx, *gauge)
			s.Require().NoError(err)
		}

		// check non-perpetual gauges (finished + rewards estimate empty)
		if !tc.isPerpetual {
			// check finished gauges
			gauges = s.App.IncentivesKeeper.GetFinishedGauges(s.Ctx)
			s.Require().Len(gauges, 1)

			// check gauge by ID
			gauge, err = s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeID)
			s.Require().NoError(err)
			s.Require().NotNil(gauge)
			s.Require().Equal(gauges[0], *gauge)

			// check invalid gauge ID
			_, err = s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeID+1000)
			s.Require().Error(err)
			rewardsEst = s.App.IncentivesKeeper.GetRewardsEst(s.Ctx, lockOwners[0], []lockuptypes.PeriodLock{}, 100)
			s.Require().Equal(sdk.Coins{}, rewardsEst)

			// check gauge ids by denom
			gaugeIdsByDenom = s.App.IncentivesKeeper.GetAllGaugeIDsByDenom(s.Ctx, "lptoken")
			s.Require().Len(gaugeIdsByDenom, 0)
		} else { // check perpetual gauges (not finished + rewards estimate empty)
			// check finished gauges
			gauges = s.App.IncentivesKeeper.GetFinishedGauges(s.Ctx)
			s.Require().Len(gauges, 0)

			// check rewards estimation
			rewardsEst = s.App.IncentivesKeeper.GetRewardsEst(s.Ctx, lockOwners[0], []lockuptypes.PeriodLock{}, 100)
			s.Require().Equal(sdk.Coins(nil), rewardsEst)

			// check gauge ids by denom
			gaugeIdsByDenom = s.App.IncentivesKeeper.GetAllGaugeIDsByDenom(s.Ctx, "lptoken")
			s.Require().Len(gaugeIdsByDenom, 1)
		}
	}
}

func (s *KeeperTestSuite) TestChargeFeeIfSufficientFeeDenomBalance() {
	const baseFee = int64(100)

	testcases := map[string]struct {
		accountBalanceToFund sdk.Coin
		feeToCharge          int64
		gaugeCoins           sdk.Coins

		expectError bool
	}{
		"fee + base denom gauge coin == acount balance, success": {
			accountBalanceToFund: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseFee)),
			feeToCharge:          baseFee / 2,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseFee/2))),
		},
		"fee + base denom gauge coin < acount balance, success": {
			accountBalanceToFund: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseFee)),
			feeToCharge:          baseFee/2 - 1,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseFee/2))),
		},
		"fee + base denom gauge coin > acount balance, error": {
			accountBalanceToFund: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseFee)),
			feeToCharge:          baseFee/2 + 1,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseFee/2))),

			expectError: true,
		},
		"fee + base denom gauge coin < acount balance, custom values, success": {
			accountBalanceToFund: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(11793193112)),
			feeToCharge:          55,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(328812))),
		},
		"account funded with coins other than base denom, error": {
			accountBalanceToFund: sdk.NewCoin("usdc", sdk.NewInt(baseFee)),
			feeToCharge:          baseFee,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseFee/2))),

			expectError: true,
		},
		"fee == account balance, no gauge coins, success": {
			accountBalanceToFund: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseFee)),
			feeToCharge:          baseFee,
		},
		"gauge coins == account balance, no fee, success": {
			accountBalanceToFund: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseFee)),
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseFee))),
		},
		"fee == account balance, gauge coins in denom other than base, success": {
			accountBalanceToFund: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseFee)),
			feeToCharge:          baseFee,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(baseFee*2))),
		},
		"fee + gauge coins == account balance, multiple gauge coins, one in denom other than base, success": {
			accountBalanceToFund: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseFee)),
			feeToCharge:          baseFee / 2,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(baseFee*2)), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(baseFee/2))),
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.SetupTest()

			testAccount := s.TestAccs[0]

			ctx := s.Ctx
			incentivesKeepers := s.App.IncentivesKeeper
			bankKeeper := s.App.BankKeeper

			// Pre-fund account.
			s.FundAcc(testAccount, sdk.NewCoins(tc.accountBalanceToFund))

			oldBalanceAmount := bankKeeper.GetBalance(ctx, testAccount, sdk.DefaultBondDenom).Amount

			// System under test.
			err := incentivesKeepers.ChargeFeeIfSufficientFeeDenomBalance(ctx, testAccount, sdk.NewInt(tc.feeToCharge), tc.gaugeCoins)

			// Assertions.
			newBalanceAmount := bankKeeper.GetBalance(ctx, testAccount, sdk.DefaultBondDenom).Amount
			if tc.expectError {
				s.Require().Error(err)

				// check account balance unchanged
				s.Require().Equal(oldBalanceAmount, newBalanceAmount)
			} else {
				s.Require().NoError(err)

				// check account balance changed.
				expectedNewBalanceAmount := oldBalanceAmount.Sub(sdk.NewInt(tc.feeToCharge))
				s.Require().Equal(expectedNewBalanceAmount.String(), newBalanceAmount.String())
			}
		})
	}
}

func (s *KeeperTestSuite) TestAddToGaugeRewards() {
	testCases := []struct {
		name               string
		owner              sdk.AccAddress
		coinsToAdd         sdk.Coins
		gaugeId            uint64
		minimumGasConsumed uint64

		expectErr bool
	}{
		{
			name:  "valid case: valid gauge",
			owner: s.TestAccs[0],
			coinsToAdd: sdk.NewCoins(
				sdk.NewCoin("uosmo", sdk.NewInt(100000)),
				sdk.NewCoin("atom", sdk.NewInt(99999)),
			),
			gaugeId:            1,
			minimumGasConsumed: uint64(2 * types.BaseGasFeeForAddRewardToGauge),

			expectErr: false,
		},
		{
			name:  "valid case: valid gauge with >4 denoms",
			owner: s.TestAccs[0],
			coinsToAdd: sdk.NewCoins(
				sdk.NewCoin("uosmo", sdk.NewInt(100000)),
				sdk.NewCoin("atom", sdk.NewInt(99999)),
				sdk.NewCoin("mars", sdk.NewInt(88888)),
				sdk.NewCoin("akash", sdk.NewInt(77777)),
				sdk.NewCoin("eth", sdk.NewInt(6666)),
				sdk.NewCoin("usdc", sdk.NewInt(555)),
				sdk.NewCoin("dai", sdk.NewInt(4444)),
				sdk.NewCoin("ust", sdk.NewInt(3333)),
			),
			gaugeId:            1,
			minimumGasConsumed: uint64(8 * types.BaseGasFeeForAddRewardToGauge),

			expectErr: false,
		},
		{
			name:  "invalid case: gauge Id is not valid",
			owner: s.TestAccs[0],
			coinsToAdd: sdk.NewCoins(
				sdk.NewCoin("uosmo", sdk.NewInt(100000)),
				sdk.NewCoin("atom", sdk.NewInt(99999)),
			),
			gaugeId:            0,
			minimumGasConsumed: uint64(0),

			expectErr: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			_, _, existingGaugeCoins, _ := s.SetupNewGauge(true, sdk.Coins{sdk.NewInt64Coin("stake", 12)})

			s.FundAcc(tc.owner, tc.coinsToAdd)

			existingGasConsumed := s.Ctx.GasMeter().GasConsumed()

			err := s.App.IncentivesKeeper.AddToGaugeRewards(s.Ctx, tc.owner, tc.coinsToAdd, tc.gaugeId)
			if tc.expectErr {
				s.Require().Error(err)

				// balance shouldn't change in the module
				balance := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName))
				s.Require().Equal(existingGaugeCoins, balance)
			} else {
				s.Require().NoError(err)

				// Ensure that at least the minimum amount of gas was charged (based on number of additional gauge coins)
				gasConsumed := s.Ctx.GasMeter().GasConsumed() - existingGasConsumed
				fmt.Println(gasConsumed, tc.minimumGasConsumed)
				s.Require().True(gasConsumed >= tc.minimumGasConsumed)

				// existing coins gets added to the module when we create gauge and add to gauge
				expectedCoins := existingGaugeCoins.Add(tc.coinsToAdd...)

				// check module account balance, should go up
				balance := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName))
				s.Require().Equal(expectedCoins, balance)

				// check gauge coins should go up
				gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, tc.gaugeId)
				s.Require().NoError(err)

				s.Require().Equal(expectedCoins, gauge.Coins)
			}
		})
	}
}

// TestCreateGauge_NoLockGauges tests the CreateGauge function
// specifically focusing on the no lock gauge type and test cases around it.
// It tests the following:
// - For no lock gauges, a CL pool id must be given and then pool must exist
// - For no lock gauges, the denom must be set either to NoLockInternalGaugeDenom(<pool id>)
// or be unset. If set to anything other than the internal prefix, fails with error.
// Assumes that the gauge was created externally (via MsgCreateGauge) if the denom is unset and overwrites it
// with NoLockExternalGaugeDenom(<pool id>)
// - Otherwise, the given pool id must be zero. Errors if not.
func (s *KeeperTestSuite) TestCreateGauge_NoLockGauges() {
	const (
		zeroPoolId         = uint64(0)
		balancerPoolId     = uint64(1)
		concentratedPoolId = uint64(2)
		invalidPool        = uint64(3)
		// 3 are created for balancer pool and 1 for CL.
		// As a result, the next gauge id should be 5.
		defaultExpectedGaugeId = uint64(5)

		defaultIsPerpetualParam = false

		defaultNumEpochPaidOver = uint64(10)
	)

	var (
		defaultCoins = sdk.NewCoins(
			sdk.NewCoin("uosmo", sdk.NewInt(100000)),
			sdk.NewCoin("atom", sdk.NewInt(99999)),
		)

		defaultTime = time.Unix(0, 0)
	)
	testCases := []struct {
		name    string
		distrTo lockuptypes.QueryCondition
		poolId  uint64

		expectedGaugeId  uint64
		expectedDenomSet string
		expectErr        bool
	}{
		{
			name: "create valid no lock gauge with CL pool (no denom set)",
			distrTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.NoLock,
				// Note: this assumes the gauge is external
				Denom: "",
			},
			poolId: concentratedPoolId,

			expectedGaugeId:  defaultExpectedGaugeId,
			expectedDenomSet: types.NoLockExternalGaugeDenom(concentratedPoolId),
			expectErr:        false,
		},
		{
			name: "create valid no lock gauge with CL pool (denom set to no lock internal prefix)",
			distrTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.NoLock,
				// Note: this assumes the gauge is internal
				Denom: types.NoLockInternalGaugeDenom(concentratedPoolId),
			},
			poolId: concentratedPoolId,

			expectedGaugeId:  defaultExpectedGaugeId,
			expectedDenomSet: types.NoLockInternalGaugeDenom(concentratedPoolId),
			expectErr:        false,
		},
		{
			name: "fail to create gauge because invalid denom is set",
			distrTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.NoLock,
				// Note: this is invalid for NoLock gauges
				Denom: "uosmo",
			},
			poolId: concentratedPoolId,

			expectErr: true,
		},
		{
			name: "fail to create no lock gauge with balancer pool",
			distrTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.NoLock,
			},
			poolId: balancerPoolId,

			expectErr: true,
		},
		{
			name: "fail to create no lock gauge with non-existent pool",
			distrTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.NoLock,
			},
			poolId: invalidPool,

			expectErr: true,
		},
		{
			name: "fail to create no lock gauge with zero pool id",
			distrTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.NoLock,
			},
			poolId: zeroPoolId,

			expectErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			s.PrepareBalancerPool()
			s.PrepareConcentratedPool()

			s.FundAcc(s.TestAccs[0], defaultCoins)

			// System under test
			// Note that the default params are used for some inputs since they are not relevant to the test case.
			gaugeId, err := s.App.IncentivesKeeper.CreateGauge(s.Ctx, defaultIsPerpetualParam, s.TestAccs[0], defaultCoins, tc.distrTo, defaultTime, defaultNumEpochPaidOver, tc.poolId)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				s.Require().Equal(tc.expectedGaugeId, gaugeId)

				// Assert that pool id and gauge id link meant for internally incentivized gauges is unset.
				_, err := s.App.PoolIncentivesKeeper.GetPoolGaugeId(s.Ctx, tc.poolId, tc.distrTo.Duration)
				s.Require().Error(err)

				// Confirm that the general pool id to gauge id link is set.
				gaugeIds, err := s.App.PoolIncentivesKeeper.GetNoLockGaugeIdsFromPool(s.Ctx, tc.poolId)
				s.Require().NoError(err)
				// One must have been created at pool creation time for internal incentives.
				s.Require().Len(gaugeIds, 2)
				gaugeId := gaugeIds[1]

				s.Require().Equal(tc.expectedGaugeId, gaugeId)

				// Get gauge and check that the denom is set correctly
				gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, tc.expectedGaugeId)
				s.Require().NoError(err)

				s.Require().Equal(tc.expectedDenomSet, gauge.DistributeTo.Denom)
				s.Require().Equal(tc.distrTo.LockQueryType, gauge.DistributeTo.LockQueryType)
				s.Require().Equal(defaultIsPerpetualParam, gauge.IsPerpetual)
				s.Require().Equal(defaultCoins, gauge.Coins)
				s.Require().Equal(defaultTime.UTC(), gauge.StartTime.UTC())
				s.Require().Equal(defaultNumEpochPaidOver, gauge.NumEpochsPaidOver)
			}
		})
	}
}
