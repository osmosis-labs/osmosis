package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// randomly delegate and undelegate, simplified version of simulation
func (suite *KeeperTestSuite) TestRandomized() {
	suite.SetupTest()
	unbondingTime := time.Second * 10
	suite.app.StakingKeeper.SetParams(suite.ctx, stakingtypes.Params{
		UnbondingTime: unbondingTime,
	})

	state := struct {
		validators []sdk.AccAddress
	}{
		validators: []sdk.AccAddress{
			sdk.AccAddress("addr0---------------"),
			sdk.AccAddress("addr1---------------"),
			sdk.AccAddress("addr2---------------"),
			sdk.AccAddress("addr3---------------"),
			sdk.AccAddress("addr4---------------"),
			sdk.AccAddress("addr5---------------"),
			sdk.AccAddress("addr6---------------"),
			sdk.AccAddress("addr7---------------"),
		},
	}

	sfDenom := "gamm/pool/1"

	operations := []struct {
		name      string
		operation func(delegator sdk.AccAddress, lockID uint64, valAddr string)
	}{
		{
			"delegate",
			func(delegator sdk.AccAddress, lockID uint64, valAddr string) {

				// run SuperfluidDelegate
				err := suite.app.SuperfluidKeeper.SuperfluidDelegate(suite.ctx, delegator.String(), lockID, valAddr)

				// ASSERT: should be error in one of the following conditions:
				//       - delegator already used the lock for superfluid delegation
				//       - lock is under unlocking state
				//       - lock duration is less than unbondingtime

				// ASSERT: should not be error if else
				suite.Require().NoError(err)

				// ASSERT: only one synthlock should be created

				// ASSERT: intermediary account should be registered for the lock ID

				// ASSERT: total supply of bonddenom should be increased for OSMOEquivalent amount

				// ASSERT: delegation from the delegator should be increased for OSMOEquivalent amount
			},
		},
		{
			"undelegate",
			func(delegator sdk.AccAddress, _ uint64, validator string) {
				// call LockTokens(which internally calls SuperfluidDelegateMore)

				// ASSERT: synthlock should be replaced by an unlocking synthlock

				// ASSERT: intermediary account should be removed for the lock ID

				// ASSERT: total supply of bonddenom should be decreased for OSMOEquivalent amount

				// ASSERT: delegation from the delegator should be decreased for OSMOEquivalent amount
			},
		},
		{
			"delegatemore",
			func(delegator sdk.AccAddress, _ uint64, validator string) {
				// call LockTokens(which internally calls SuperfluidDelegateMore)

				// ASSERT: synthlock amount should be increased

				// ASSERT: total supply of bonddenom should be increased for OSMOEquivalent amount

				// ASSERT: delegation from the delegator should be increased for OSMOEquivalent amount
			},
		},
		{
			"slash",
			func(_ sdk.AccAddress, _ uint64, validator string) {
				// call SlashLockupsForValidatorSlash

				// decrease delegation amount manually

				// ASSERT: for all intermediary accounts for the slashed validator,
				//         for synthlocks that are staked to the intermediary account(regardless of unstaking),
				//         the underlying locks amount should be decreased

			},
		},
		{
			"beginunlock",
			func(delegator sdk.AccAddress, lockId uint64, _ string) {
				// XXX
			},
		},
		{
			"epochend",
			func(_ sdk.AccAddress, _ uint64, _ string) {
				// set new price for osmo

				// call AfterEpochEnd

				// ASSERT: delegation amount should be adjusted per new price

				// call incentives.AfterEpochEnd

				// ASSERT: all of the staking rewards are distributed to the delegators(perpetual gauge distributes all)
			},
		},
	}
}
