package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	incentiveskeeper "github.com/osmosis-labs/osmosis/v27/x/incentives/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
)

var _ = suite.TestingSuite(nil)

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

	feeDenom = "groupfee"
)

var (
	defaultEmptyGaugeInfo = types.InternalGaugeInfo{
		TotalWeight:  osmomath.ZeroInt(),
		GaugeRecords: []types.InternalGaugeRecord{},
	}

	defaultTime = time.Unix(1, 0).UTC()

	defaultGaugeCreationCoins = sdk.NewCoins(
		sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(100000)),
		sdk.NewCoin("atom", osmomath.NewInt(99999)),
	)

	customGroupCreationFee = sdk.NewCoins(sdk.NewInt64Coin(feeDenom, 1000000))

	errorNoCustomFeeInBalance = fmt.Errorf("0%s is smaller than %s: insufficient funds", feeDenom, customGroupCreationFee)
)

// TestInvalidDurationGaugeCreationValidation tests error handling for creating a gauge with an invalid duration.
func (s *KeeperTestSuite) TestInvalidDurationGaugeCreationValidation() {
	s.SetupTest()

	// Since this test creates or adds to a gauge, we need to ensure a route exists in protorev hot routes.
	// The pool doesn't need to actually exist for this test, so we can just ensure the denom pair has some entry.
	for _, coin := range defaultLiquidTokens {
		s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, coin.Denom, 9999)
	}

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

	// Since this test creates or adds to a gauge, we need to ensure a route exists in protorev hot routes.
	// The pool doesn't need to actually exist for this test, so we can just ensure the denom pair has some entry.
	for _, coin := range defaultLiquidTokens {
		s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, coin.Denom, 9999)
	}

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
			s.Require().Equal(sdk.Coins{}, rewardsEst)

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
		"fee + base denom gauge coin == account balance, success": {
			accountBalanceToFund: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(baseFee)),
			feeToCharge:          baseFee / 2,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(baseFee/2))),
		},
		"fee + base denom gauge coin < account balance, success": {
			accountBalanceToFund: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(baseFee)),
			feeToCharge:          baseFee/2 - 1,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(baseFee/2))),
		},
		"fee + base denom gauge coin > account balance, error": {
			accountBalanceToFund: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(baseFee)),
			feeToCharge:          baseFee/2 + 1,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(baseFee/2))),

			expectError: true,
		},
		"fee + base denom gauge coin < account balance, custom values, success": {
			accountBalanceToFund: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(11793193112)),
			feeToCharge:          55,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(328812))),
		},
		"account funded with coins other than base denom, error": {
			accountBalanceToFund: sdk.NewCoin("usdc", osmomath.NewInt(baseFee)),
			feeToCharge:          baseFee,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(baseFee/2))),

			expectError: true,
		},
		"fee == account balance, no gauge coins, success": {
			accountBalanceToFund: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(baseFee)),
			feeToCharge:          baseFee,
		},
		"gauge coins == account balance, no fee, success": {
			accountBalanceToFund: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(baseFee)),
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(baseFee))),
		},
		"fee == account balance, gauge coins in denom other than base, success": {
			accountBalanceToFund: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(baseFee)),
			feeToCharge:          baseFee,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin("usdc", osmomath.NewInt(baseFee*2))),
		},
		"fee + gauge coins == account balance, multiple gauge coins, one in denom other than base, success": {
			accountBalanceToFund: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(baseFee)),
			feeToCharge:          baseFee / 2,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin("usdc", osmomath.NewInt(baseFee*2)), sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(baseFee/2))),
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
			err := incentivesKeepers.ChargeFeeIfSufficientFeeDenomBalance(ctx, testAccount, osmomath.NewInt(tc.feeToCharge), tc.gaugeCoins)

			// Assertions.
			newBalanceAmount := bankKeeper.GetBalance(ctx, testAccount, sdk.DefaultBondDenom).Amount
			if tc.expectError {
				s.Require().Error(err)

				// check account balance unchanged
				s.Require().Equal(oldBalanceAmount, newBalanceAmount)
			} else {
				s.Require().NoError(err)

				// check account balance changed.
				expectedNewBalanceAmount := oldBalanceAmount.Sub(osmomath.NewInt(tc.feeToCharge))
				s.Require().Equal(expectedNewBalanceAmount.String(), newBalanceAmount.String())
			}
		})
	}
}

func (s *KeeperTestSuite) TestAddToGaugeRewards() {

	defaultCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 12))

	// since most of the same functionality and edge cases are tested by a higher level
	// AddToGaugeRewards down below, we only include a happy path test for the internal helper.
	s.Run("internal helper basic happy path test", func() {
		s.SetupTest()
		const defaultGaugeId = uint64(1)

		_, _, _, _ = s.SetupNewGauge(true, defaultCoins)

		err := s.App.IncentivesKeeper.AddToGaugeRewardsInternal(s.Ctx, defaultCoins, defaultGaugeId)
		s.Require().NoError(err)

		gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, defaultGaugeId)
		s.Require().NoError(err)

		// validate final coins were updated
		s.Require().Equal(defaultCoins.Add(defaultCoins...), gauge.Coins)
	})

	testCases := []struct {
		name               string
		owner              sdk.AccAddress
		coinsToAdd         sdk.Coins
		gaugeId            uint64
		minimumGasConsumed uint64

		skipSettingRoute bool
		expectErr        bool
	}{
		{
			name:  "valid case: valid gauge",
			owner: s.TestAccs[0],
			coinsToAdd: sdk.NewCoins(
				sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(100000)),
				sdk.NewCoin("atom", osmomath.NewInt(99999)),
			),
			gaugeId:            1,
			minimumGasConsumed: uint64(2 * types.BaseGasFeeForAddRewardToGauge),

			expectErr: false,
		},
		{
			name:  "valid case: valid gauge with >4 denoms",
			owner: s.TestAccs[0],
			coinsToAdd: sdk.NewCoins(
				sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(100000)),
				sdk.NewCoin("atom", osmomath.NewInt(99999)),
				sdk.NewCoin("mars", osmomath.NewInt(88888)),
				sdk.NewCoin("akash", osmomath.NewInt(77777)),
				sdk.NewCoin("eth", osmomath.NewInt(6666)),
				sdk.NewCoin("usdc", osmomath.NewInt(555)),
				sdk.NewCoin("dai", osmomath.NewInt(4444)),
				sdk.NewCoin("ust", osmomath.NewInt(3333)),
			),
			gaugeId:            1,
			minimumGasConsumed: uint64(8 * types.BaseGasFeeForAddRewardToGauge),

			expectErr: false,
		},
		{
			name:  "invalid case: gauge Id is not valid",
			owner: s.TestAccs[0],
			coinsToAdd: sdk.NewCoins(
				sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(100000)),
				sdk.NewCoin("atom", osmomath.NewInt(99999)),
			),
			gaugeId:            0,
			minimumGasConsumed: uint64(0),

			expectErr: true,
		},
		{
			name:  "invalid case: valid gauge, but errors due to no protorev route",
			owner: s.TestAccs[0],
			coinsToAdd: sdk.NewCoins(
				sdk.NewCoin("uosmo", osmomath.NewInt(100000)),
				sdk.NewCoin("atom", osmomath.NewInt(99999)),
			),
			gaugeId:            1,
			minimumGasConsumed: uint64(2 * types.BaseGasFeeForAddRewardToGauge),

			skipSettingRoute: true,
			expectErr:        true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Since this test creates or adds to a gauge, we need to ensure a route exists in protorev hot routes.
			// The pool doesn't need to actually exist for this test, so we can just ensure the denom pair has some entry.
			if !tc.skipSettingRoute {
				for _, coin := range tc.coinsToAdd {
					s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, coin.Denom, 9999)
				}
			}

			_, _, existingGaugeCoins, _ := s.SetupNewGauge(true, defaultCoins)

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
	testCases := []struct {
		name    string
		distrTo lockuptypes.QueryCondition
		poolId  uint64

		expectedGaugeId  uint64
		expectedDenomSet string
		skipSettingRoute bool
		expectErr        bool
	}{
		{
			name: "create valid no lock gauge with CL pool (no denom set)",
			distrTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.NoLock,
				// Note: this assumes the gauge is external
				Denom:    "",
				Duration: time.Nanosecond,
			},
			poolId: concentratedPoolId,

			expectedGaugeId:  defaultExpectedGaugeId,
			expectedDenomSet: types.NoLockExternalGaugeDenom(concentratedPoolId),
			expectErr:        false,
		},
		{
			name: "create valid no lock gauge with CL pool, but errors due to no protorev route",
			distrTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.NoLock,
				// Note: this assumes the gauge is external
				Denom:    "",
				Duration: time.Nanosecond,
			},
			poolId: concentratedPoolId,

			expectedGaugeId:  defaultExpectedGaugeId,
			expectedDenomSet: types.NoLockExternalGaugeDenom(concentratedPoolId),
			skipSettingRoute: true,
			expectErr:        true,
		},
		{
			name: "create valid no lock gauge with CL pool (denom set to no lock internal prefix)",
			distrTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.NoLock,
				// Note: this assumes the gauge is internal
				Denom:    types.NoLockInternalGaugeDenom(concentratedPoolId),
				Duration: s.App.IncentivesKeeper.GetEpochInfo(s.Ctx).Duration,
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
				Denom:    appparams.BaseCoinUnit,
				Duration: time.Nanosecond,
			},
			poolId: concentratedPoolId,

			expectErr: true,
		},
		{
			name: "fail to create no lock gauge with balancer pool",
			distrTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.NoLock,
				Duration:      defaultNoLockDuration,
			},
			poolId: balancerPoolId,

			expectErr: true,
		},
		{
			name: "fail to create no lock gauge with non-existent pool",
			distrTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.NoLock,
				Duration:      defaultNoLockDuration,
			},
			poolId: invalidPool,

			expectErr: true,
		},
		{
			name: "fail to create no lock gauge with zero pool id",
			distrTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.NoLock,
				Duration:      defaultNoLockDuration,
			},
			poolId: zeroPoolId,

			expectErr: true,
		},
		{
			name: "fail to create external no lock gauge due to unauthorized uptime",
			distrTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.NoLock,
				// Note: this assumes the gauge is external
				Denom: "",
				// 1h is a supported uptime that is not authorized
				Duration: time.Hour,
			},
			poolId:    concentratedPoolId,
			expectErr: true,
		},
		{
			name: "fail to create external no lock gauge due to entirely invalid uptime",
			distrTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.NoLock,
				// Note: this assumes the gauge is external
				Denom: "",
				// 2ns is an uptime that isn't supported at all (i.e. can't even be authorized)
				Duration: 2 * time.Nanosecond,
			},
			poolId:    concentratedPoolId,
			expectErr: true,
		},
		{
			name: "fail to create an internal gauge with an unexpected duration",
			distrTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.NoLock,
				// Note: this assumes the gauge is internal
				Denom:    types.NoLockInternalGaugeDenom(concentratedPoolId),
				Duration: time.Nanosecond,
			},
			poolId: concentratedPoolId,

			expectErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			// Since this test creates or adds to a gauge, we need to ensure a route exists in protorev hot routes.
			// The pool doesn't need to actually exist for this test, so we can just ensure the denom pair has some entry.
			if !tc.skipSettingRoute {
				for _, coin := range defaultGaugeCreationCoins {
					s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, coin.Denom, 9999)
				}
			}

			s.PrepareBalancerPool()
			s.PrepareConcentratedPool()

			s.FundAcc(s.TestAccs[0], defaultGaugeCreationCoins)

			// System under test
			// Note that the default params are used for some inputs since they are not relevant to the test case.
			gaugeId, err := s.App.IncentivesKeeper.CreateGauge(s.Ctx, defaultIsPerpetualParam, s.TestAccs[0], defaultGaugeCreationCoins, tc.distrTo, defaultTime, defaultNumEpochPaidOver, tc.poolId)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				s.Require().Equal(tc.expectedGaugeId, gaugeId)

				// Confirm that the general pool id to gauge id link is set.
				gaugeIds, err := s.App.PoolIncentivesKeeper.GetNoLockGaugeIdsFromPool(s.Ctx, tc.poolId)
				s.Require().NoError(err)
				// One must have been created at pool creation time for internal incentives.
				s.Require().Len(gaugeIds, 2)
				gaugeId := gaugeIds[1]

				s.Require().Equal(tc.expectedGaugeId, gaugeId)

				// Validate gauge
				tc.distrTo.Denom = tc.expectedDenomSet
				s.validateGauge(types.Gauge{
					Id:                tc.expectedGaugeId,
					DistributeTo:      tc.distrTo,
					IsPerpetual:       defaultIsPerpetualParam,
					StartTime:         defaultTime.UTC(),
					Coins:             defaultGaugeCreationCoins,
					NumEpochsPaidOver: defaultNumEpochPaidOver,
				})
			}
		})
	}
}

// Tests that CreateGauge can create ByGroup gauges correctly.
// Additionally, validates that no ref keys are created for the group gauge.
func (s *KeeperTestSuite) TestCreateGauge_Group() {
	testCases := []struct {
		name              string
		distrTo           lockuptypes.QueryCondition
		poolId            uint64
		isPerpetual       bool
		numEpochsPaidOver uint64

		expectedGaugeId  uint64
		expectedDenomSet string
		skipSettingRoute bool
		expectErr        error
	}{
		{
			name:              "create valid non-perpetual group gauge",
			distrTo:           incentiveskeeper.ByGroupQueryCondition,
			poolId:            zeroPoolId,
			isPerpetual:       false,
			numEpochsPaidOver: types.PerpetualNumEpochsPaidOver + 1,

			expectedGaugeId: defaultExpectedGaugeId,
		},
		{
			name:              "create valid perpetual group gauge",
			distrTo:           incentiveskeeper.ByGroupQueryCondition,
			poolId:            zeroPoolId,
			isPerpetual:       true,
			numEpochsPaidOver: types.PerpetualNumEpochsPaidOver,

			expectedGaugeId: defaultExpectedGaugeId,
		},

		// error cases

		{
			name:              "fail to create group gauge due to zero epochs paid over and non-perpetual",
			distrTo:           incentiveskeeper.ByGroupQueryCondition,
			poolId:            zeroPoolId,
			isPerpetual:       false,
			numEpochsPaidOver: 0,

			expectErr: types.ErrZeroNumEpochsPaidOver,
		},
		{
			name:              "create valid non-perpetual group gauge, but errors due to no protorev route",
			distrTo:           incentiveskeeper.ByGroupQueryCondition,
			poolId:            zeroPoolId,
			isPerpetual:       false,
			numEpochsPaidOver: types.PerpetualNumEpochsPaidOver + 1,

			skipSettingRoute: true,
			expectErr:        types.NoRouteForDenomError{Denom: "atom"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			// Since this test creates or adds to a gauge, we need to ensure a route exists in protorev hot routes.
			// The pool doesn't need to actually exist for this test, so we can just ensure the denom pair has some entry.
			if !tc.skipSettingRoute {
				for _, coin := range defaultGaugeCreationCoins {
					s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, coin.Denom, 9999)
				}
			}

			s.PrepareBalancerPool()
			s.PrepareConcentratedPool()

			s.FundAcc(s.TestAccs[0], defaultGaugeCreationCoins)

			// System under test
			// Note that the default params are used for some inputs since they are not relevant to the test case.
			gaugeId, err := s.App.IncentivesKeeper.CreateGauge(s.Ctx, tc.isPerpetual, s.TestAccs[0], defaultGaugeCreationCoins, tc.distrTo, defaultTime, tc.numEpochsPaidOver, tc.poolId)

			if tc.expectErr != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectErr)
			} else {
				s.Require().NoError(err)

				s.Require().Equal(tc.expectedGaugeId, gaugeId)

				// Assert that pool id and gauge id link meant for internally incentivized gauges is unset.
				_, err := s.App.PoolIncentivesKeeper.GetPoolGaugeId(s.Ctx, tc.poolId, tc.distrTo.Duration)
				s.Require().Error(err)

				// Confirm that for every lockable duration, there is not gauge ID and pool ID link.
				lockableDurations := s.App.PoolIncentivesKeeper.GetLockableDurations(s.Ctx)
				s.Require().NotEqual(0, len(lockableDurations))
				for _, duration := range lockableDurations {
					_, err = s.App.PoolIncentivesKeeper.GetPoolIdFromGaugeId(s.Ctx, gaugeId, duration)
					s.Require().Error(err)
				}

				// Confirm that for incentives epoch duration, there is no gauge ID and pool ID link.
				incentivesEpochDuration := s.App.IncentivesKeeper.GetEpochInfo(s.Ctx).Duration
				_, err = s.App.PoolIncentivesKeeper.GetPoolIdFromGaugeId(s.Ctx, gaugeId, incentivesEpochDuration)
				s.Require().Error(err)

				// Validate gauge
				s.validateGauge(types.Gauge{
					Id:                tc.expectedGaugeId,
					DistributeTo:      incentiveskeeper.ByGroupQueryCondition,
					IsPerpetual:       tc.isPerpetual,
					StartTime:         defaultTime.UTC(),
					Coins:             defaultGaugeCreationCoins,
					NumEpochsPaidOver: tc.numEpochsPaidOver,
				})

				// Validate that ref keys are not created for the group gauge

				// No upcoming ref keys
				upcomingGauges := s.App.IncentivesKeeper.GetUpcomingGauges(s.Ctx)
				s.validateNoGaugeIDInSlice(upcomingGauges, tc.expectedGaugeId)

				// No active ref keys
				activeGauges := s.App.IncentivesKeeper.GetActiveGauges(s.Ctx)
				s.validateNoGaugeIDInSlice(activeGauges, tc.expectedGaugeId)

				// No finished ref keys
				finishedGauges := s.App.IncentivesKeeper.GetActiveGauges(s.Ctx)
				s.validateNoGaugeIDInSlice(finishedGauges, tc.expectedGaugeId)
			}
		})
	}
}

// validates that the expected gauge info equals the actual gauge info
func (s *KeeperTestSuite) validateGaugeInfo(expected types.InternalGaugeInfo, actual types.InternalGaugeInfo) {
	s.Require().Equal(expected.TotalWeight.String(), actual.TotalWeight.String())
	s.Require().Equal(len(expected.GaugeRecords), len(actual.GaugeRecords))
	for i := range expected.GaugeRecords {
		s.Require().Equal(expected.GaugeRecords[i].GaugeId, actual.GaugeRecords[i].GaugeId)
		s.Require().Equal(expected.GaugeRecords[i].CurrentWeight.String(), actual.GaugeRecords[i].CurrentWeight.String())
		s.Require().Equal(expected.GaugeRecords[i].CumulativeWeight.String(), actual.GaugeRecords[i].CumulativeWeight.String())
	}
}

// retrieves the gauge of expectedGauge.ID from state and validates that it matches expectedGauge
func (s *KeeperTestSuite) validateGauge(expectedGauge types.Gauge) {
	// Get gauge and check that the denom is set correctly
	gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, expectedGauge.Id)
	s.Require().NoError(err)

	// No denom set for group gauges.
	s.Require().Equal(expectedGauge.DistributeTo.Denom, gauge.DistributeTo.Denom)
	// ByGroup type set
	s.Require().Equal(expectedGauge.DistributeTo.LockQueryType, gauge.DistributeTo.LockQueryType)
	s.Require().Equal(expectedGauge.IsPerpetual, gauge.IsPerpetual)
	s.Require().Equal(expectedGauge.Coins, gauge.Coins)
	s.Require().Equal(expectedGauge.StartTime, gauge.StartTime.UTC())
	s.Require().Equal(expectedGauge.NumEpochsPaidOver, gauge.NumEpochsPaidOver)
}

// test helper to create a gauge bypassing all checks and restrictions
// It is useful in edge case tests that rely on invalid gauges written to store (e.g. in Distribute())
func (s *KeeperTestSuite) createGaugeNoRestrictions(isPerpetual bool, coins sdk.Coins, distrTo lockuptypes.QueryCondition, startTime time.Time, numEpochsPaidOver uint64, poolID uint64) types.Gauge {
	// Fund incentives module account to simulate transfer from owner to module account
	s.FundModuleAcc(types.ModuleName, coins)
	lastGaugeID := s.App.IncentivesKeeper.GetLastGaugeID(s.Ctx)
	nextGaugeID := lastGaugeID + 1
	gauge := types.Gauge{
		Id:                nextGaugeID,
		IsPerpetual:       isPerpetual,
		Coins:             coins,
		DistributeTo:      distrTo,
		StartTime:         startTime,
		NumEpochsPaidOver: numEpochsPaidOver,
	}

	if poolID != 0 {
		s.App.PoolIncentivesKeeper.SetPoolGaugeIdNoLock(s.Ctx, poolID, nextGaugeID, distrTo.Duration)
	}

	err := s.App.IncentivesKeeper.SetGauge(s.Ctx, &gauge)
	s.Require().NoError(err)
	s.App.IncentivesKeeper.CreateGaugeRefKeys(s.Ctx, &gauge, incentiveskeeper.CombineKeys(types.KeyPrefixUpcomingGauges, incentiveskeeper.GetTimeKeys(startTime)))
	s.Require().NoError(err)

	// Retrieve from state and return
	gaugeFromState, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, nextGaugeID)
	s.Require().NoError(err)
	return *gaugeFromState
}

// validates that there is not gauge with the given ID in the slice
func (s *KeeperTestSuite) validateNoGaugeIDInSlice(slice []types.Gauge, gaugeID uint64) {
	gaugeMatch := osmoutils.Filter(func(gauge types.Gauge) bool {
		return gauge.Id == gaugeID
	}, slice)
	// No gauge matched ID.
	s.Require().Empty(gaugeMatch)
}

func (s *KeeperTestSuite) TestCheckIfDenomsAreDistributable() {
	s.SetupTest()

	coinWithRouteA := sdk.NewCoin("denom1", osmomath.NewInt(100))
	coinWithRouteB := sdk.NewCoin("denom2", osmomath.NewInt(100))
	coinWithoutRouteC := sdk.NewCoin("denom3", osmomath.NewInt(100))

	for _, coin := range []sdk.Coin{coinWithRouteA, coinWithRouteB} {
		s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, coin.Denom, 9999)
	}

	testCases := []struct {
		name        string
		coins       sdk.Coins
		expectedErr error
	}{
		{
			name:  "valid case: all denoms are distributable",
			coins: sdk.NewCoins(coinWithRouteA, coinWithRouteB),
		},
		{
			name:        "invalid case: one denom is not distributable",
			coins:       sdk.NewCoins(coinWithRouteA, coinWithoutRouteC),
			expectedErr: types.NoRouteForDenomError{Denom: coinWithoutRouteC.Denom},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			// System under test
			err := s.App.IncentivesKeeper.CheckIfDenomsAreDistributable(s.Ctx, tc.coins)

			if tc.expectedErr != nil {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
