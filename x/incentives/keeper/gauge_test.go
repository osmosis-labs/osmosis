package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v10/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v10/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGaugeTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// TestInvalidDurationGaugeCreationValidation tests error handling for creating a gauge with an invalid duration.
func (suite *KeeperTestSuite) TestInvalidDurationGaugeCreationValidation() {
	suite.SetupTest()

	addrs := suite.SetupManyLocks(1, defaultLiquidTokens, defaultLPTokens, defaultLockDuration)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         defaultLPDenom,
		Duration:      defaultLockDuration / 2, // 0.5 second, invalid duration
	}
	_, err := suite.App.IncentivesKeeper.CreateGauge(suite.Ctx, false, addrs[0], defaultLiquidTokens, distrTo, time.Time{}, 1)
	suite.Require().Error(err)

	distrTo.Duration = defaultLockDuration
	_, err = suite.App.IncentivesKeeper.CreateGauge(suite.Ctx, false, addrs[0], defaultLiquidTokens, distrTo, time.Time{}, 1)
	suite.Require().NoError(err)
}

// TestNonExistentDenomGaugeCreation tests error handling for creating a gauge with an invalid denom.
func (suite *KeeperTestSuite) TestNonExistentDenomGaugeCreation() {
	suite.SetupTest()

	addrNoSupply := sdk.AccAddress([]byte("Gauge_Creation_Addr_"))
	addrs := suite.SetupManyLocks(1, defaultLiquidTokens, defaultLPTokens, defaultLockDuration)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         defaultLPDenom,
		Duration:      defaultLockDuration,
	}
	_, err := suite.App.IncentivesKeeper.CreateGauge(suite.Ctx, false, addrNoSupply, defaultLiquidTokens, distrTo, time.Time{}, 1)
	suite.Require().Error(err)

	_, err = suite.App.IncentivesKeeper.CreateGauge(suite.Ctx, false, addrs[0], defaultLiquidTokens, distrTo, time.Time{}, 1)
	suite.Require().NoError(err)
}

// TestGaugeOperations tests perpetual and non-perpetual gauge distribution logic using the gauges by denom keeper.
func (suite *KeeperTestSuite) TestGaugeOperations() {
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
		suite.SetupTest()

		// initial module gauges check
		gauges := suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
		suite.Require().Len(gauges, 0)
		gaugeIdsByDenom := suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
		suite.Require().Len(gaugeIdsByDenom, 0)

		// setup lock and gauge
		lockOwners := suite.SetupManyLocks(tc.numLocks, defaultLiquidTokens, defaultLPTokens, time.Second)
		gaugeID, _, coins, startTime := suite.SetupNewGauge(tc.isPerpetual, sdk.Coins{sdk.NewInt64Coin("stake", 12)})
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
		gauges = suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
		suite.Require().Len(gauges, 1)
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
		suite.Require().Equal(expectedGauge.String(), gauges[0].String())

		// check gauge ids by denom
		gaugeIdsByDenom = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
		suite.Require().Len(gaugeIdsByDenom, 1)
		suite.Require().Equal(gaugeID, gaugeIdsByDenom[0])

		// check rewards estimation
		rewardsEst := suite.App.IncentivesKeeper.GetRewardsEst(suite.Ctx, lockOwners[0], []lockuptypes.PeriodLock{}, 100)
		suite.Require().Equal(expectedCoinsPerLock.String(), rewardsEst.String())

		// check gauges
		gauges = suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
		suite.Require().Len(gauges, 1)
		suite.Require().Equal(expectedGauge.String(), gauges[0].String())

		// check upcoming gauges
		gauges = suite.App.IncentivesKeeper.GetUpcomingGauges(suite.Ctx)
		suite.Require().Len(gauges, 1)

		// start distribution
		suite.Ctx = suite.Ctx.WithBlockTime(startTime)
		gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
		suite.Require().NoError(err)
		err = suite.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
		suite.Require().NoError(err)

		// check active gauges
		gauges = suite.App.IncentivesKeeper.GetActiveGauges(suite.Ctx)
		suite.Require().Len(gauges, 1)

		// check upcoming gauges
		gauges = suite.App.IncentivesKeeper.GetUpcomingGauges(suite.Ctx)
		suite.Require().Len(gauges, 0)

		// check gauge ids by denom
		gaugeIdsByDenom = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
		suite.Require().Len(gaugeIdsByDenom, 1)

		// check gauge ids by other denom
		gaugeIdsByDenom = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lpt")
		suite.Require().Len(gaugeIdsByDenom, 0)

		// distribute coins to stakers
		distrCoins, err := suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
		suite.Require().NoError(err)
		// We hardcoded 12 "stake" tokens when initializing gauge
		suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", int64(12/expectedNumEpochsPaidOver))}, distrCoins)

		if tc.isPerpetual {
			// distributing twice without adding more for perpetual gauge
			gauge, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
			suite.Require().NoError(err)
			distrCoins, err = suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
			suite.Require().NoError(err)
			suite.Require().True(distrCoins.Empty())

			// add to gauge
			addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
			suite.AddToGauge(addCoins, gaugeID)

			// distributing twice with adding more for perpetual gauge
			gauge, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
			suite.Require().NoError(err)
			distrCoins, err = suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
			suite.Require().NoError(err)
			suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 200)}, distrCoins)
		} else {
			// add to gauge
			addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
			suite.AddToGauge(addCoins, gaugeID)
		}

		// check active gauges
		gauges = suite.App.IncentivesKeeper.GetActiveGauges(suite.Ctx)
		suite.Require().Len(gauges, 1)

		// check gauge ids by denom
		gaugeIdsByDenom = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
		suite.Require().Len(gaugeIdsByDenom, 1)

		// finish distribution for non perpetual gauge
		if !tc.isPerpetual {
			err = suite.App.IncentivesKeeper.MoveActiveGaugeToFinishedGauge(suite.Ctx, *gauge)
			suite.Require().NoError(err)
		}

		// check non-perpetual gauges (finished + rewards estimate empty)
		if !tc.isPerpetual {

			// check finished gauges
			gauges = suite.App.IncentivesKeeper.GetFinishedGauges(suite.Ctx)
			suite.Require().Len(gauges, 1)

			// check gauge by ID
			gauge, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
			suite.Require().NoError(err)
			suite.Require().NotNil(gauge)
			suite.Require().Equal(gauges[0], *gauge)

			// check invalid gauge ID
			_, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID+1000)
			suite.Require().Error(err)
			rewardsEst = suite.App.IncentivesKeeper.GetRewardsEst(suite.Ctx, lockOwners[0], []lockuptypes.PeriodLock{}, 100)
			suite.Require().Equal(sdk.Coins{}, rewardsEst)

			// check gauge ids by denom
			gaugeIdsByDenom = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
			suite.Require().Len(gaugeIdsByDenom, 0)
		} else { // check perpetual gauges (not finished + rewards estimate empty)

			// check finished gauges
			gauges = suite.App.IncentivesKeeper.GetFinishedGauges(suite.Ctx)
			suite.Require().Len(gauges, 0)

			// check rewards estimation
			rewardsEst = suite.App.IncentivesKeeper.GetRewardsEst(suite.Ctx, lockOwners[0], []lockuptypes.PeriodLock{}, 100)
			suite.Require().Equal(sdk.Coins(nil), rewardsEst)

			// check gauge ids by denom
			gaugeIdsByDenom = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
			suite.Require().Len(gaugeIdsByDenom, 1)
		}
	}
}

<<<<<<< HEAD
// func (suite *KeeperTestSuite) TestCreateGaugeFee() {

// 	tests := []struct {
// 		name                 string
// 		accountBalanceToFund sdk.Coins
// 		gaugeAddition        sdk.Coins
// 		expectedEndBalance   sdk.Coins
// 		isPerpetual          bool
// 		isModuleAccount      bool
// 		expectErr            bool
// 	}{
// 		{
// 			name:                 "user creates a non-perpetual gauge and fills gauge with all remaining tokens",
// 			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(60000000))),
// 			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000))),
// 			expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(0))),
// 		},
// 		{
// 			name:                 "user creates a non-perpetual gauge and fills gauge with some remaining tokens",
// 			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(70000000))),
// 			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000))),
// 			expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000))),
// 		},
// 		{
// 			name:                 "user with multiple denoms creates a non-perpetual gauge and fills gauge with some remaining tokens",
// 			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(70000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
// 			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000))),
// 			expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
// 		},
// 		{
// 			name:                 "module account creates a perpetual gauge and fills gauge with some remaining tokens",
// 			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(70000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
// 			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000))),
// 			expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
// 			isPerpetual:          true,
// 			isModuleAccount:      true,
// 		},
// 		{
// 			name:                 "user with multiple denoms creates a perpetual gauge and fills gauge with some remaining tokens",
// 			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(70000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
// 			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000))),
// 			expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
// 			isPerpetual:          true,
// 		},
// 		{
// 			name:                 "user tries to create a non-perpetual gauge but does not have enough funds to pay for the create gauge fee",
// 			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(40000000))),
// 			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000))),
// 			expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(40000000))),
// 			expectErr:            true,
// 		},
// 		{
// 			name:                 "user tries to create a non-perpetual gauge but does not have the correct fee denom",
// 			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(60000000))),
// 			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10000000))),
// 			expectedEndBalance:   sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(60000000))),
// 			expectErr:            true,
// 		},
// 		// TODO: This is unexpected behavior
// 		// We need validation to not charge fee if user doesn't have enough funds
// 		// {
// 		// 	name:                 "one user tries to create a gauge, has enough funds to pay for the create gauge fee but not enough to fill the gauge",
// 		// 	accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(60000000))),
// 		// 	gaugeAddition:        sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(30000000))),
// 		// 	expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(60000000))),
// 		// 	expectErr:            true,
// 		// },

// 	}

// 	for _, tc := range tests {
// 		suite.SetupTest()

// 		//testAccount := suite.TestAccs[0]
// 		testAccountPubkey := secp256k1.GenPrivKeyFromSecret([]byte("acc")).PubKey()
// 		testAccountAddress := sdk.AccAddress(testAccountPubkey.Address())

// 		ctx := suite.Ctx
// 		incentivesKeepers := suite.App.IncentivesKeeper

// 		suite.FundAcc(testAccountAddress, tc.accountBalanceToFund)

// 		if tc.isModuleAccount {
// 			modAcc := authtypes.NewModuleAccount(authtypes.NewBaseAccount(testAccountAddress, testAccountPubkey, 1, 0),
// 				"module",
// 				"permission",
// 			)
// 			suite.App.AccountKeeper.SetModuleAccount(ctx, modAcc)
// 		}

// 		suite.SetupManyLocks(1, defaultLiquidTokens, defaultLPTokens, defaultLockDuration)
// 		distrTo := lockuptypes.QueryCondition{
// 			LockQueryType: lockuptypes.ByDuration,
// 			Denom:         defaultLPDenom,
// 			Duration:      defaultLockDuration,
// 		}

// 		// System under test.

// 		_, err := incentivesKeepers.CreateGauge(ctx, tc.isPerpetual, testAccountAddress, tc.gaugeAddition, distrTo, time.Time{}, 1)

// 		if tc.expectErr {
// 			suite.Require().Error(err)
// 		} else {
// 			suite.Require().NoError(err)
// 		}

// 		bal := suite.App.BankKeeper.GetAllBalances(suite.Ctx, testAccountAddress)
// 		suite.Require().Equal(tc.expectedEndBalance.String(), bal.String(), "test: %v", tc.name)

// 	}
// }
=======
func (suite *KeeperTestSuite) TestChargeFee() {
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
		suite.Run(name, func() {
			suite.SetupTest()

			testAccount := suite.TestAccs[0]

			ctx := suite.Ctx
			incentivesKeepers := suite.App.IncentivesKeeper
			bankKeeper := suite.App.BankKeeper

			// Pre-fund account.
			suite.FundAcc(testAccount, sdk.NewCoins(tc.accountBalanceToFund))

			oldBalanceAmount := bankKeeper.GetBalance(ctx, testAccount, sdk.DefaultBondDenom).Amount

			// System under test.
			err := incentivesKeepers.ChargeFee(ctx, testAccount, tc.feeToCharge, tc.gaugeCoins)

			// Assertions.
			newBalanceAmount := bankKeeper.GetBalance(ctx, testAccount, sdk.DefaultBondDenom).Amount
			if tc.expectError {
				suite.Require().Error(err)

				// check account balance unchanged
				suite.Require().Equal(oldBalanceAmount, newBalanceAmount)
			} else {
				suite.Require().NoError(err)

				// check account balance changed.
				expectedNewBalanceAmount := oldBalanceAmount.Sub(sdk.NewInt(tc.feeToCharge))
				suite.Require().Equal(expectedNewBalanceAmount.String(), newBalanceAmount.String())
			}
		})
	}
}
>>>>>>> roman/txfees-simplified
