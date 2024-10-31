package keeper_test

import (
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ = suite.TestingSuite(nil)

var (
	seventyTokens = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(70000000)))
	tenTokens     = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(10000000)))
)

func (s *KeeperTestSuite) TestCreateGauge_Fee() {
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
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(60000000))),
			gaugeAddition:        tenTokens,
		},
		{
			name:                 "user creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(70000000))),
			gaugeAddition:        tenTokens,
		},
		{
			name:                 "user with multiple denoms creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(70000000)), sdk.NewCoin("foo", osmomath.NewInt(70000000))),
			gaugeAddition:        tenTokens,
		},
		{
			name:                 "module account creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(70000000)), sdk.NewCoin("foo", osmomath.NewInt(70000000))),
			gaugeAddition:        tenTokens,
			isPerpetual:          true,
			isModuleAccount:      true,
		},
		{
			name:                 "user with multiple denoms creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(70000000)), sdk.NewCoin("foo", osmomath.NewInt(70000000))),
			gaugeAddition:        tenTokens,
			isPerpetual:          true,
		},
		{
			name:                 "user tries to create a non-perpetual gauge but does not have enough funds to pay for the create gauge fee",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(40000000))),
			gaugeAddition:        tenTokens,
			expectErr:            true,
		},
		{
			name:                 "user tries to create a non-perpetual gauge but does not have the correct fee denom",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(60000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(10000000))),
			expectErr:            true,
		},
		{
			name:                 "one user tries to create a gauge, has enough funds to pay for the create gauge fee but not enough to fill the gauge",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(60000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(30000000))),
			expectErr:            true,
		},
	}

	for _, tc := range tests {
		s.SetupTest()

		// Since this test creates or adds to a gauge, we need to ensure a route exists in protorev hot routes.
		// The pool doesn't need to actually exist for this test, so we can just ensure the denom pair has some entry.
		s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, sdk.DefaultBondDenom, 9999)

		testAccountPubkey := secp256k1.GenPrivKeyFromSecret([]byte("acc")).PubKey()
		testAccountAddress := sdk.AccAddress(testAccountPubkey.Address())

		bankKeeper := s.App.BankKeeper
		accountKeeper := s.App.AccountKeeper
		msgServer := keeper.NewMsgServerImpl(s.App.IncentivesKeeper)

		s.FundAcc(testAccountAddress, tc.accountBalanceToFund)

		if tc.isModuleAccount {
			modAcc := authtypes.NewModuleAccount(authtypes.NewBaseAccount(testAccountAddress, testAccountPubkey, s.App.AccountKeeper.NextAccountNumber(s.Ctx), 0),
				"module",
				"permission",
			)
			accountKeeper.SetModuleAccount(s.Ctx, modAcc)
		}

		s.SetupManyLocks(1, defaultLiquidTokens, defaultLPTokens, defaultLockDuration)
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
		_, err := msgServer.CreateGauge(s.Ctx, msg)

		if tc.expectErr {
			s.Require().Error(err)
		} else {
			s.Require().NoError(err)
		}

		balanceAmount := bankKeeper.GetAllBalances(s.Ctx, testAccountAddress)

		if tc.expectErr {
			s.Require().Equal(tc.accountBalanceToFund.String(), balanceAmount.String(), "test: %v", tc.name)
		} else {
			fee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, types.CreateGaugeFee))
			accountBalance := tc.accountBalanceToFund.Sub(tc.gaugeAddition...)
			finalAccountBalance := accountBalance.Sub(fee...)
			s.Require().Equal(finalAccountBalance.String(), balanceAmount.String(), "test: %v", tc.name)
		}
	}
}

func (s *KeeperTestSuite) TestAddToGauge_Fee() {
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
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(35000000))),
			gaugeAddition:        tenTokens,
		},
		{
			name:                 "user creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: seventyTokens,
			gaugeAddition:        tenTokens,
		},
		{
			name:                 "user with multiple denoms creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(70000000)), sdk.NewCoin("foo", osmomath.NewInt(70000000))),
			gaugeAddition:        tenTokens,
		},
		{
			name:                 "module account creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(70000000)), sdk.NewCoin("foo", osmomath.NewInt(70000000))),
			gaugeAddition:        tenTokens,
			isPerpetual:          true,
			isModuleAccount:      true,
		},
		{
			name:                 "user with multiple denoms creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(70000000)), sdk.NewCoin("foo", osmomath.NewInt(70000000))),
			gaugeAddition:        tenTokens,
			isPerpetual:          true,
		},
		{
			name:                 "user tries to create a non-perpetual gauge but does not have enough funds to pay for the create gauge fee",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(20000000))),
			gaugeAddition:        tenTokens,
			expectErr:            true,
		},
		{
			name:                 "user tries to add to a non-perpetual gauge but does not have the correct fee denom",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(60000000))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(10000000))),
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
		s.SetupTest()

		testAccountPubkey := secp256k1.GenPrivKeyFromSecret([]byte("acc")).PubKey()
		testAccountAddress := sdk.AccAddress(testAccountPubkey.Address())

		bankKeeper := s.App.BankKeeper
		incentivesKeeper := s.App.IncentivesKeeper
		accountKeeper := s.App.AccountKeeper
		msgServer := keeper.NewMsgServerImpl(incentivesKeeper)

		s.FundAcc(testAccountAddress, tc.accountBalanceToFund)

		if tc.isModuleAccount {
			modAcc := authtypes.NewModuleAccount(authtypes.NewBaseAccount(testAccountAddress, testAccountPubkey, s.App.AccountKeeper.NextAccountNumber(s.Ctx), 0),
				"module",
				"permission",
			)
			accountKeeper.SetModuleAccount(s.Ctx, modAcc)
		}

		// System under test.
		coins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(500000000)))
		gaugeID, gauge, _, _ := s.SetupNewGauge(tc.isPerpetual, coins)
		if tc.nonexistentGauge {
			gaugeID = incentivesKeeper.GetLastGaugeID(s.Ctx) + 1
		}
		// simulate times to complete the gauge.
		if tc.isGaugeComplete {
			s.completeGauge(gauge, sdk.AccAddress([]byte("a___________________")))
		}
		msg := &types.MsgAddToGauge{
			Owner:   testAccountAddress.String(),
			GaugeId: gaugeID,
			Rewards: tc.gaugeAddition,
		}

		_, err := msgServer.AddToGauge(s.Ctx, msg)

		if tc.expectErr {
			s.Require().Error(err, tc.name)
		} else {
			s.Require().NoError(err, tc.name)
		}

		bal := bankKeeper.GetAllBalances(s.Ctx, testAccountAddress)

		if !tc.expectErr {
			fee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, types.AddToGaugeFee))
			accountBalance := tc.accountBalanceToFund.Sub(tc.gaugeAddition...)
			finalAccountBalance := accountBalance.Sub(fee...)
			s.Require().Equal(finalAccountBalance.String(), bal.String(), "test: %v", tc.name)
		} else if tc.expectErr && !tc.isGaugeComplete {
			s.Require().Equal(tc.accountBalanceToFund.String(), bal.String(), "test: %v", tc.name)
		}
	}
}

func (s *KeeperTestSuite) TestCreateGroup_Fee() {
	tests := []struct {
		name                 string
		accountBalanceToFund sdk.Coins
		expectedEndBalance   sdk.Coins
		groupFunds           sdk.Coins
		isModuleAccount      bool
		numEpochsPaidOver    uint64
		expectErr            bool
	}{
		{
			name:                 "user creates a non-perpetual group and fills group with all remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(100000000)), sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(10000000))),
			groupFunds:           tenTokens,
			numEpochsPaidOver:    3,
		},
		{
			name:                 "user creates a perpetual group and fills group with all remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(100000000)), sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(10000000))),
			groupFunds:           tenTokens,
			numEpochsPaidOver:    0,
		},
		{
			name:                 "user creates a non-perpetual group and fills group with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(100000000)), sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(15000000))),
			groupFunds:           tenTokens,
			numEpochsPaidOver:    3,
		},
		{
			name:                 "module account creates a perpetual group",
			accountBalanceToFund: sdk.Coins{},
			groupFunds:           sdk.Coins{},
			isModuleAccount:      true,
			numEpochsPaidOver:    0,
		},
		{
			name:                 "user tries to create a non-perpetual group but does not have enough funds to pay for the create group fee",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(90000000)), sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(10000000))),
			groupFunds:           tenTokens,
			expectErr:            true,
			numEpochsPaidOver:    3,
		},
		{
			name:                 "one user tries to create a group, has enough funds to pay for the create group fee but not enough to fill the group funds",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(100000000)), sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(9000000))),
			groupFunds:           tenTokens,
			expectErr:            true,
			numEpochsPaidOver:    3,
		},
	}

	for _, tc := range tests {
		s.SetupTest()

		// Since this test creates or adds to a gauge, we need to ensure a route exists in protorev hot routes.
		// The pool doesn't need to actually exist for this test, so we can just ensure the denom pair has some entry.
		s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, sdk.DefaultBondDenom, 9999)

		testAccountPubkey := secp256k1.GenPrivKeyFromSecret([]byte("acc")).PubKey()
		testAccountAddress := sdk.AccAddress(testAccountPubkey.Address())

		bankKeeper := s.App.BankKeeper
		accountKeeper := s.App.AccountKeeper
		msgServer := keeper.NewMsgServerImpl(s.App.IncentivesKeeper)
		groupCreationFee := s.App.IncentivesKeeper.GetParams(s.Ctx).GroupCreationFee

		s.FundAcc(testAccountAddress, tc.accountBalanceToFund)

		if tc.isModuleAccount {
			testAccountAddress = accountKeeper.GetModuleAddress(types.ModuleName)
		}

		poolInfo := s.PrepareAllSupportedPools()

		poolIDs := []uint64{poolInfo.BalancerPoolID, poolInfo.ConcentratedPoolID, poolInfo.StableSwapPoolID}

		msg := &types.MsgCreateGroup{
			Coins:             tc.groupFunds,
			NumEpochsPaidOver: tc.numEpochsPaidOver,
			Owner:             testAccountAddress.String(),
			PoolIds:           poolIDs,
		}

		// setup volume so that the pool can be created
		s.overwriteVolumes(poolIDs, []osmomath.Int{defaultVolumeAmount, defaultVolumeAmount, defaultVolumeAmount})

		// System under test.
		_, err := msgServer.CreateGroup(s.Ctx, msg)

		if tc.expectErr {
			s.Require().Error(err)
		} else {
			s.Require().NoError(err)
			balanceAmount := bankKeeper.GetAllBalances(s.Ctx, testAccountAddress)

			accountBalance := tc.accountBalanceToFund.Sub(tc.groupFunds...)
			finalAccountBalance := accountBalance
			if !tc.isModuleAccount {
				finalAccountBalance = accountBalance.Sub(groupCreationFee...)
			}
			s.Require().Equal(finalAccountBalance.String(), balanceAmount.String(), "test: %v", tc.name)
		}
	}
}

func (s *KeeperTestSuite) completeGauge(gauge *types.Gauge, sendingAddress sdk.AccAddress) {
	lockCoins := sdk.NewCoin(gauge.DistributeTo.Denom, osmomath.NewInt(1000))
	s.FundAcc(sendingAddress, sdk.NewCoins(lockCoins))
	s.LockTokens(sendingAddress, sdk.NewCoins(lockCoins), gauge.DistributeTo.Duration)
	epochId := s.App.IncentivesKeeper.GetEpochInfo(s.Ctx).Identifier
	if s.Ctx.BlockTime().Before(gauge.StartTime) {
		s.Ctx = s.Ctx.WithBlockTime(gauge.StartTime.Add(time.Hour))
	}
	s.BeginNewBlock(false)
	for i := 0; i < int(gauge.NumEpochsPaidOver); i++ {
		err := s.App.IncentivesKeeper.BeforeEpochStart(s.Ctx, epochId, int64(i))
		s.Require().NoError(err)
		err = s.App.IncentivesKeeper.AfterEpochEnd(s.Ctx, epochId, int64(i))
		s.Require().NoError(err)
	}
	s.BeginNewBlock(false)
	gauge2, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gauge.Id)
	s.Require().NoError(err)
	s.Require().True(gauge2.IsFinishedGauge(s.Ctx.BlockTime()))
}
