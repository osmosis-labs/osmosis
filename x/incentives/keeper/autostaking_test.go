package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
)

func (suite *KeeperTestSuite) TestAutostakingManagement() {
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	addr3 := sdk.AccAddress([]byte("addr3---------------"))

	valAddr1 := sdk.ValAddress(addr1)
	valAddr2 := sdk.ValAddress(addr2)

	suite.SetupTest()
	err := suite.app.IncentivesKeeper.SetAutostaking(suite.ctx, &types.AutoStaking{
		Address:              addr1.String(),
		AutostakingValidator: valAddr1.String(),
		// AutostakingRate:      sdk.NewDecWithPrec(5, 1),
	})
	suite.Require().NoError(err)
	err = suite.app.IncentivesKeeper.SetAutostaking(suite.ctx, &types.AutoStaking{
		Address:              addr2.String(),
		AutostakingValidator: valAddr2.String(),
		// AutostakingRate:      sdk.NewDecWithPrec(5, 1),
	})
	suite.Require().NoError(err)

	autostaking1 := suite.app.IncentivesKeeper.GetAutostakingByAddress(suite.ctx, addr1.String())
	suite.Require().NotNil(autostaking1)
	suite.Require().Equal(*autostaking1, types.AutoStaking{
		Address:              addr1.String(),
		AutostakingValidator: valAddr1.String(),
		// AutostakingRate:      sdk.NewDecWithPrec(5, 1),
	})

	autostaking2 := suite.app.IncentivesKeeper.GetAutostakingByAddress(suite.ctx, addr2.String())
	suite.Require().NotNil(autostaking2)
	suite.Require().Equal(*autostaking2, types.AutoStaking{
		Address:              addr2.String(),
		AutostakingValidator: valAddr2.String(),
		// AutostakingRate:      sdk.NewDecWithPrec(5, 1),
	})

	autostaking3 := suite.app.IncentivesKeeper.GetAutostakingByAddress(suite.ctx, addr3.String())
	suite.Require().Nil(autostaking3)

	err = suite.app.IncentivesKeeper.SetAutostaking(suite.ctx, &types.AutoStaking{
		Address:              addr1.String(),
		AutostakingValidator: valAddr2.String(),
		// AutostakingRate:      sdk.NewDecWithPrec(1, 1),
	})
	suite.Require().NoError(err)

	autostaking1 = suite.app.IncentivesKeeper.GetAutostakingByAddress(suite.ctx, addr1.String())
	suite.Require().NotNil(autostaking1)
	suite.Require().Equal(*autostaking1, types.AutoStaking{
		Address:              addr1.String(),
		AutostakingValidator: valAddr2.String(),
		// AutostakingRate:      sdk.NewDecWithPrec(1, 1),
	})

	autostakings := suite.app.IncentivesKeeper.AllAutoStakings(suite.ctx)
	suite.Require().Len(autostakings, 2)

	autostakingIters := []types.AutoStaking{}
	suite.app.IncentivesKeeper.IterateAutoStaking(suite.ctx, func(index int64, autostaking types.AutoStaking) (stop bool) {
		autostakingIters = append(autostakingIters, autostaking)
		return false
	})
	suite.Require().Len(autostakingIters, 2)
}

func (suite *KeeperTestSuite) TestAutostakeRewards() {
	PKS := simapp.CreateTestPubKeys(5)
	valConsPk1 := PKS[0]

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	addr3 := sdk.AccAddress([]byte("addr3---------------"))

	valAddr1 := sdk.ValAddress(addr1)
	valAddr2 := sdk.ValAddress(addr2)

	suite.SetupTest()
	bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)

	acc1 := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc1)
	err := suite.app.BankKeeper.SetBalance(suite.ctx, addr1, sdk.NewInt64Coin(bondDenom, 100))
	suite.Require().NoError(err)

	// create validator with 50% commission
	tstaking := teststaking.NewHelper(suite.T(), suite.ctx, suite.app.StakingKeeper)
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	tstaking.CreateValidator(sdk.ValAddress(addr1), valConsPk1, sdk.NewInt(100), true)

	err = suite.app.IncentivesKeeper.SetAutostaking(suite.ctx, &types.AutoStaking{
		Address:              addr1.String(),
		AutostakingValidator: valAddr1.String(),
		// AutostakingRate:      sdk.NewDecWithPrec(5, 1),
	})
	suite.Require().NoError(err)
	err = suite.app.IncentivesKeeper.SetAutostaking(suite.ctx, &types.AutoStaking{
		Address:              addr2.String(),
		AutostakingValidator: valAddr2.String(),
		// AutostakingRate:      sdk.NewDecWithPrec(5, 1),
	})
	suite.Require().NoError(err)

	coins1 := sdk.Coins{
		sdk.NewInt64Coin("alt", 10000),
		sdk.NewInt64Coin(bondDenom, 10000),
	}
	coins2 := sdk.Coins{
		sdk.NewInt64Coin(bondDenom, 10000),
	}
	coins3 := sdk.Coins{
		sdk.NewInt64Coin("alt", 10000),
	}
	coins4 := sdk.Coins{}

	osmo5k := sdk.NewInt64Coin(bondDenom, 5000)
	osmoZero := sdk.NewInt64Coin(bondDenom, 0)

	testCases := []struct {
		addr          sdk.AccAddress
		coins         sdk.Coins
		expDelegation sdk.Coin
		expLock       sdk.Coin
	}{
		{addr1, coins1, osmo5k, osmoZero},
		{addr1, coins2, osmo5k, osmoZero},
		{addr1, coins3, osmoZero, osmoZero},
		{addr1, coins4, osmoZero, osmoZero},
		{addr2, coins1, osmoZero, osmo5k},
		{addr2, coins2, osmoZero, osmo5k},
		{addr2, coins3, osmoZero, osmoZero},
		{addr2, coins4, osmoZero, osmoZero},
		{addr3, coins1, osmoZero, osmo5k},
		{addr3, coins2, osmoZero, osmo5k},
		{addr3, coins3, osmoZero, osmoZero},
		{addr3, coins4, osmoZero, osmoZero},
	}

	for _, tc := range testCases {
		initLocks := suite.app.LockupKeeper.GetAccountPeriodLocks(suite.ctx, tc.addr)
		initDelegation := sdk.NewInt(0)
		delegations := suite.app.StakingKeeper.GetDelegatorDelegations(suite.ctx, tc.addr, 100)
		for _, del := range delegations {
			valAddr, err := sdk.ValAddressFromBech32(del.ValidatorAddress)
			suite.Require().NoError(err)
			val, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr)
			suite.Require().True(found)
			delAmt := val.Tokens.ToDec().Mul(del.Shares).Quo(val.DelegatorShares)
			initDelegation = initDelegation.Add(delAmt.RoundInt())
		}

		err = suite.app.BankKeeper.SetBalances(suite.ctx, tc.addr, tc.coins)
		suite.Require().NoError(err)
		err = suite.app.IncentivesKeeper.AutostakeRewards(suite.ctx, tc.addr, tc.coins)
		suite.Require().NoError(err)

		finalLocks := suite.app.LockupKeeper.GetAccountPeriodLocks(suite.ctx, tc.addr)
		finalDelegation := sdk.NewInt(0)
		delegations = suite.app.StakingKeeper.GetDelegatorDelegations(suite.ctx, tc.addr, 100)
		for _, del := range delegations {
			valAddr, err := sdk.ValAddressFromBech32(del.ValidatorAddress)
			suite.Require().NoError(err)
			val, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr)
			suite.Require().True(found)
			delAmt := val.Tokens.ToDec().Mul(del.Shares).Quo(val.DelegatorShares)
			finalDelegation = finalDelegation.Add(delAmt.RoundInt())
		}

		if tc.expDelegation.Amount.IsPositive() {
			suite.Require().Equal(finalDelegation, initDelegation.Add(tc.expDelegation.Amount))
		}
		if tc.expLock.Amount.IsPositive() {
			suite.Require().Equal(len(finalLocks), len(initLocks)+1)
			suite.Require().Equal(finalLocks[len(finalLocks)-1].Coins, sdk.Coins{tc.expLock})
			suite.Require().Equal(finalLocks[len(finalLocks)-1].Duration, time.Hour*24*7*2)
		}
	}
}
