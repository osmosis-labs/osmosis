package keeper_test

import (
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/osmosis-labs/osmosis/v15/x/incentives/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ = suite.TestingSuite(nil)

var (
	seventyTokens = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(70000000)))
	tenTokens     = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000)))
)

func (suite *KeeperTestSuite) TestCreateGauge_Fee() {
	tests := []struct {
		name                 string
		accountBalanceToFund sdk.Coins
		gaugeAddition        sdk.Coins
		expectedEndBalance   sdk.Coins
		isPerpetual          bool
		isModuleAccount      bool
		expectErr            bool
	}{
		{
			name:                 "user creates a non-perpetual gauge and fills gauge with all remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(60000000))),
			gaugeAddition:        tenTokens,
		},
		{
			name:                 "user creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(70000000))),
			gaugeAddition:        tenTokens,
		},
		{
			name:                 "user with multiple denoms creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(70000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
			gaugeAddition:        tenTokens,
		},
		{
			name:                 "module account creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(70000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
			gaugeAddition:        tenTokens,
			isPerpetual:          true,
			isModuleAccount:      true,
		},
		{
			name:                 "user with multiple denoms creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(70000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
			gaugeAddition:        tenTokens,
			isPerpetual:          true,
		},
		{
			name:                 "user tries to create a non-perpetual gauge but does not have enough funds to pay for the create gauge fee",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(40000000))),
			gaugeAddition:        tenTokens,
			expectErr:            true,
		},
		{
			name:                 "user tries to create a non-perpetual gauge but does not have the correct fee denom",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(60000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10000000))),
			expectErr:            true,
		},
		{
			name:                 "one user tries to create a gauge, has enough funds to pay for the create gauge fee but not enough to fill the gauge",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(60000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(30000000))),
			expectErr:            true,
		},
	}

	for _, tc := range tests {
		suite.SetupTest()

		testAccountPubkey := secp256k1.GenPrivKeyFromSecret([]byte("acc")).PubKey()
		testAccountAddress := sdk.AccAddress(testAccountPubkey.Address())

		ctx := suite.Ctx
		bankKeeper := suite.App.BankKeeper
		accountKeeper := suite.App.AccountKeeper
		msgServer := keeper.NewMsgServerImpl(suite.App.IncentivesKeeper)

		suite.FundAcc(testAccountAddress, tc.accountBalanceToFund)

		if tc.isModuleAccount {
			modAcc := authtypes.NewModuleAccount(authtypes.NewBaseAccount(testAccountAddress, testAccountPubkey, 1, 0),
				"module",
				"permission",
			)
			accountKeeper.SetModuleAccount(ctx, modAcc)
		}

		suite.SetupManyLocks(1, defaultLiquidTokens, defaultLPTokens, defaultLockDuration)
		distrTo := lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         defaultLPDenom,
			Duration:      defaultLockDuration,
		}

		msg := &types.MsgCreateGauge{
			IsPerpetual:       tc.isPerpetual,
			Owner:             testAccountAddress.String(),
			DistributeTo:      distrTo,
			Coins:             tc.gaugeAddition,
			StartTime:         time.Now(),
			NumEpochsPaidOver: 1,
		}
		// System under test.
		_, err := msgServer.CreateGauge(sdk.WrapSDKContext(ctx), msg)

		if tc.expectErr {
			suite.Require().Error(err)
		} else {
			suite.Require().NoError(err)
		}

		balanceAmount := bankKeeper.GetAllBalances(ctx, testAccountAddress)

		if tc.expectErr {
			suite.Require().Equal(tc.accountBalanceToFund.String(), balanceAmount.String(), "test: %v", tc.name)
		} else {
			fee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, types.CreateGaugeFee))
			accountBalance := tc.accountBalanceToFund.Sub(tc.gaugeAddition)
			finalAccountBalance := accountBalance.Sub(fee)
			suite.Require().Equal(finalAccountBalance.String(), balanceAmount.String(), "test: %v", tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestAddToGauge_Fee() {
	tests := []struct {
		name                 string
		accountBalanceToFund sdk.Coins
		gaugeAddition        sdk.Coins
		nonexistentGauge     bool
		isPerpetual          bool
		isModuleAccount      bool
		isGaugeComplete      bool
		expectErr            bool
	}{
		{
			name:                 "user creates a non-perpetual gauge and fills gauge with all remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(35000000))),
			gaugeAddition:        tenTokens,
		},
		{
			name:                 "user creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: seventyTokens,
			gaugeAddition:        tenTokens,
		},
		{
			name:                 "user with multiple denoms creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(70000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
			gaugeAddition:        tenTokens,
		},
		{
			name:                 "module account creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(70000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
			gaugeAddition:        tenTokens,
			isPerpetual:          true,
			isModuleAccount:      true,
		},
		{
			name:                 "user with multiple denoms creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(70000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
			gaugeAddition:        tenTokens,
			isPerpetual:          true,
		},
		{
			name:                 "user tries to create a non-perpetual gauge but does not have enough funds to pay for the create gauge fee",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20000000))),
			gaugeAddition:        tenTokens,
			expectErr:            true,
		},
		{
			name:                 "user tries to add to a non-perpetual gauge but does not have the correct fee denom",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(60000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10000000))),
			expectErr:            true,
		},
		{
			name:                 "user tries to add to a finished gauge",
			accountBalanceToFund: seventyTokens,
			gaugeAddition:        tenTokens,
			isGaugeComplete:      true,
			expectErr:            true,
		},
	}

	for _, tc := range tests {
		suite.SetupTest()

		testAccountPubkey := secp256k1.GenPrivKeyFromSecret([]byte("acc")).PubKey()
		testAccountAddress := sdk.AccAddress(testAccountPubkey.Address())

		bankKeeper := suite.App.BankKeeper
		incentivesKeeper := suite.App.IncentivesKeeper
		accountKeeper := suite.App.AccountKeeper
		msgServer := keeper.NewMsgServerImpl(incentivesKeeper)

		suite.FundAcc(testAccountAddress, tc.accountBalanceToFund)

		if tc.isModuleAccount {
			modAcc := authtypes.NewModuleAccount(authtypes.NewBaseAccount(testAccountAddress, testAccountPubkey, 1, 0),
				"module",
				"permission",
			)
			accountKeeper.SetModuleAccount(suite.Ctx, modAcc)
		}

		// System under test.
		coins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(500000000)))
		gaugeID, gauge, _, _ := suite.SetupNewGauge(tc.isPerpetual, coins)
		if tc.nonexistentGauge {
			gaugeID = incentivesKeeper.GetLastGaugeID(suite.Ctx) + 1
		}
		// simulate times to complete the gauge.
		if tc.isGaugeComplete {
			suite.completeGauge(gauge, sdk.AccAddress([]byte("a___________________")))
		}
		msg := &types.MsgAddToGauge{
			Owner:   testAccountAddress.String(),
			GaugeId: gaugeID,
			Rewards: tc.gaugeAddition,
		}

		_, err := msgServer.AddToGauge(sdk.WrapSDKContext(suite.Ctx), msg)

		if tc.expectErr {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)
		}

		bal := bankKeeper.GetAllBalances(suite.Ctx, testAccountAddress)

		if !tc.expectErr {
			fee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, types.AddToGaugeFee))
			accountBalance := tc.accountBalanceToFund.Sub(tc.gaugeAddition)
			finalAccountBalance := accountBalance.Sub(fee)
			suite.Require().Equal(finalAccountBalance.String(), bal.String(), "test: %v", tc.name)
		} else if tc.expectErr && !tc.isGaugeComplete {
			suite.Require().Equal(tc.accountBalanceToFund.String(), bal.String(), "test: %v", tc.name)
		}
	}
}

func (suite *KeeperTestSuite) completeGauge(gauge *types.Gauge, sendingAddress sdk.AccAddress) {
	lockCoins := sdk.NewCoin(gauge.DistributeTo.Denom, sdk.NewInt(1000))
	suite.FundAcc(sendingAddress, sdk.NewCoins(lockCoins))
	suite.LockTokens(sendingAddress, sdk.NewCoins(lockCoins), gauge.DistributeTo.Duration)
	epochId := suite.App.IncentivesKeeper.GetEpochInfo(suite.Ctx).Identifier
	if suite.Ctx.BlockTime().Before(gauge.StartTime) {
		suite.Ctx = suite.Ctx.WithBlockTime(gauge.StartTime.Add(time.Hour))
	}
	suite.BeginNewBlock(false)
	for i := 0; i < int(gauge.NumEpochsPaidOver); i++ {
		err := suite.App.IncentivesKeeper.BeforeEpochStart(suite.Ctx, epochId, int64(i))
		suite.Require().NoError(err)
		err = suite.App.IncentivesKeeper.AfterEpochEnd(suite.Ctx, epochId, int64(i))
		suite.Require().NoError(err)
	}
	suite.BeginNewBlock(false)
	gauge2, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gauge.Id)
	suite.Require().NoError(err)
	suite.Require().True(gauge2.IsFinishedGauge(suite.Ctx.BlockTime()))
}
