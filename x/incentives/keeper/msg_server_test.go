package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	appparams "github.com/osmosis-labs/osmosis/v10/app/params"
	"github.com/osmosis-labs/osmosis/v10/x/incentives/keeper"
	incentiveskeeper "github.com/osmosis-labs/osmosis/v10/x/incentives/keeper"
	"github.com/osmosis-labs/osmosis/v10/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v10/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestCreateGaugeFee() {
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
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(60000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000000))),
			expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(0))),
		},
		{
			name:                 "user creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(70000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000000))),
			expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000000))),
		},
		{
			name:                 "user with multiple denoms creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(70000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000000))),
			expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
		},
		{
			name:                 "module account creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(70000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000000))),
			expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
			isPerpetual:          true,
			isModuleAccount:      true,
		},
		{
			name:                 "user with multiple denoms creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(70000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000000))),
			expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
			isPerpetual:          true,
		},
		{
			name:                 "user tries to create a non-perpetual gauge but does not have enough funds to pay for the create gauge fee",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(40000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000000))),
			expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(40000000))),
			expectErr:            true,
		},
		{
			name:                 "user tries to create a non-perpetual gauge but does not have the correct fee denom",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(60000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10000000))),
			expectedEndBalance:   sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(60000000))),
			expectErr:            true,
		},
		{
			name:                 "one user tries to create a gauge, has enough funds to pay for the create gauge fee but not enough to fill the gauge",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(60000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(30000000))),
			expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(60000000))),
			expectErr:            true,
		},
	}

	for _, tc := range tests {
		suite.SetupTest()

		testAccountPubkey := secp256k1.GenPrivKeyFromSecret([]byte("acc")).PubKey()
		testAccountAddress := sdk.AccAddress(testAccountPubkey.Address())

		ctx := suite.Ctx
		bankKeeper := suite.App.BankKeeper
		msgServer := keeper.NewMsgServerImpl(suite.App.IncentivesKeeper)

		suite.FundAcc(testAccountAddress, tc.accountBalanceToFund)

		if tc.isModuleAccount {
			modAcc := authtypes.NewModuleAccount(authtypes.NewBaseAccount(testAccountAddress, testAccountPubkey, 1, 0),
				"module",
				"permission",
			)
			suite.App.AccountKeeper.SetModuleAccount(ctx, modAcc)
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
		suite.Require().Equal(tc.expectedEndBalance.String(), balanceAmount.String(), "test: %v", tc.name)

		if tc.expectErr {
			suite.Require().Equal(tc.accountBalanceToFund.String(), balanceAmount.String(), "test: %v", tc.name)
		} else {
			suite.Require().Equal(tc.expectedEndBalance.String(), balanceAmount.String(), "test: %v", tc.name)
		}

	}
}

func (suite *KeeperTestSuite) TestAddToGaugeFee() {

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
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(60000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000000))),
			//expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(0))),
		},
		{
			name:                 "user creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(70000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000000))),
			//expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000000))),
		},
		{
			name:                 "user with multiple denoms creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(70000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000000))),
			//expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
		},
		{
			name:                 "module account creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(70000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000000))),
			//expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
			isPerpetual:     true,
			isModuleAccount: true,
		},
		{
			name:                 "user with multiple denoms creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(70000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000000))),
			//expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
			isPerpetual: true,
		},
		{
			name:                 "user tries to create a non-perpetual gauge but does not have enough funds to pay for the create gauge fee",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(40000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000000))),
			//expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(40000000))),
			expectErr: true,
		},
		{
			name:                 "user tries to create a non-perpetual gauge but does not have the correct fee denom",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(60000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10000000))),
			//expectedEndBalance:   sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(60000000))),
			expectErr: true,
		},
		// TODO: This is unexpected behavior
		// We need validation to not charge fee if user doesn't have enough funds
		// {
		// 	name:                 "one user tries to create a gauge, has enough funds to pay for the create gauge fee but not enough to fill the gauge",
		// 	accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(60000000))),
		// 	gaugeAddition:        sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(30000000))),
		// 	expectedEndBalance:   sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(60000000))),
		// 	expectErr:            true,
		// },

	}

	for _, tc := range tests {
		suite.SetupTest()

		testAccountPubkey := secp256k1.GenPrivKeyFromSecret([]byte("acc")).PubKey()
		testAccountAddress := sdk.AccAddress(testAccountPubkey.Address())

		ctx := suite.Ctx
		incentivesKeepers := suite.App.IncentivesKeeper

		suite.FundAcc(testAccountAddress, tc.accountBalanceToFund)

		if tc.isModuleAccount {
			modAcc := authtypes.NewModuleAccount(authtypes.NewBaseAccount(testAccountAddress, testAccountPubkey, 1, 0),
				"module",
				"permission",
			)
			suite.App.AccountKeeper.SetModuleAccount(ctx, modAcc)
		}

		suite.SetupManyLocks(1, defaultLiquidTokens, defaultLPTokens, defaultLockDuration)
		distrTo := lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         defaultLPDenom,
			Duration:      defaultLockDuration,
		}

		// System under test.

		_, err := incentivesKeepers.CreateGauge(ctx, tc.isPerpetual, testAccountAddress, tc.gaugeAddition, distrTo, time.Time{}, 1)
		//incentivesKeepers.AddToGaugeRewards(ctx, testAccountAddress, tc.gaugeAddition)

		if tc.expectErr {
			suite.Require().Error(err)
		} else {
			suite.Require().NoError(err)
		}

		bal := suite.App.BankKeeper.GetAllBalances(suite.Ctx, testAccountAddress)
		suite.Require().Equal(tc.expectedEndBalance.String(), bal.String(), "test: %v", tc.name)

		if tc.expectErr {
			suite.Require().Equal(tc.accountBalanceToFund.String(), bal.String(), "test: %v", tc.name)
		} else {
			finalAccountBalalance := tc.accountBalanceToFund.Sub(tc.gaugeAddition.Add(sdk.NewCoin(appparams.BaseCoinUnit, incentiveskeeper.CreateGaugeFee)))
			suite.Require().Equal(finalAccountBalalance.String(), bal.String(), "test: %v", tc.name)
		}

	}
}
