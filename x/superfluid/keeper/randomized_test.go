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
	epoch := time.Second * 2

	suite.app.StakingKeeper.SetParams(suite.ctx, stakingtypes.Params{
		UnbondingTime: unbondingTime,
	})

	state := struct {
		validators []sdk.AccAddress
		delegators []sdk.AccAddress
	}{
		validators: []sdk.AccAddress{
			sdk.AccAddress("vali0---------------"),
			sdk.AccAddress("vali1---------------"),
			sdk.AccAddress("vali2---------------"),
			sdk.AccAddress("vali3---------------"),
			sdk.AccAddress("vali4---------------"),
			sdk.AccAddress("vali5---------------"),
			sdk.AccAddress("vali6---------------"),
			sdk.AccAddress("vali7---------------"),
		},
		delegators: []sdk.AccAddress{
			sdk.AccAddress("dele0---------------"),
			sdk.AccAddress("dele1---------------"),
			sdk.AccAddress("dele2---------------"),
			sdk.AccAddress("dele3---------------"),
			sdk.AccAddress("dele4---------------"),
			sdk.AccAddress("dele5---------------"),
			sdk.AccAddress("dele6---------------"),
			sdk.AccAddress("dele7---------------"),
		},
	}

	poolId := uint64(1)
	sfDenom := "gamm/pool/1"

	operations := []struct {
		name      string
		operation func(delegator sdk.AccAddress, lockID uint64, valAddr string)
	}{
		{
			"delegate",
			func(delegator sdk.AccAddress, lockID uint64, valAddr string) {

				beforesupply := suite.app.BankKeeper.GetSupply(suite.ctx, bondDenom)
				beforedelegation := suite.app.StakingKeeper.GetDelegation(suite.ctx, delegator, valAddr)
				validator := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr)

				// run SuperfluidDelegate
				err := suite.app.SuperfluidKeeper.SuperfluidDelegate(suite.ctx, delegator.String(), lockID, valAddr)

				// ASSERT: should be error in any of the following conditions...
				{

					// delegator already used the lock for superfluid delegation
					delegationExists := suite.app.LockupKeeper.HasAnySyntheticLockups(suite.ctx, lockID)

					// lock is under unlocking state
					lock, _ := suite.app.LockupKeeper.GetLockByID(suite.ctx, lockID)
					isUnlocking := lock.IsUnlocking()

					// lock duration is less than unbondingtime
					tooSmallLockDuration := lock.Duration < unbondingTime

					// ...then should be error
					if delegationExists || isUnlocking || tooSmallLockDuration {
						suite.Require().Error(err)
						return
					}
				}

				// ASSERT: should not be error if else
				suite.Require().NoError(err)

				// ASSERT: total supply of bonddenom should be increased for OSMOEquivalent amount
				// TODO: too dependent on internal logic, refactor me
				pool, _ := suite.app.GAMMKeeper.GetPool(suite.ctx, poolId)
				asset, _ := pool.GetPoolAsset(bonddenom)
				osmoEquivalent := asset.Amount.Quo(pool.GetTotalShares().Amount).Mul
				sfasset := suite.app.SuperfluidKeeper.GetSuperfluidAsset(suite.ctx, sfDenom)
				riskAdjusted := suite.app.SuperfluidKeeper.GetRiskAdjustedOsmoValue(suite.ctx, sfasset, osmoEquivalent(lock.Coins[0].Amount))

				supply := suite.app.BankKeeper.GetSupply(suite.ctx, bonddenom)
				suite.Require().Equal(beforesupply.Add(riskAdjusted), supply)

				// ASSERT: delegation from the delegator should be increased for OSMOEquivalent amount
				// XXX check if the args are correct
				delegation := suite.app.StakingKeeper.GetDelegation(ctx, delegator, valAddr)
				sharesAdded, _ := validator.SharesFromTokens(riskAdjusted)
				suite.Require().Equal(beforedelegation.Shares.Add(sharesAdded), delegation.Shares)

				// ASSERT: balance of the delegator should not be changed
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
			"lock",
			func(_ sdk.AccAddress, lockId uint64, _ string) {

			},
		},
		{
			"beginunlock",
			func(_ sdk.AccAddress, lockId uint64, _ string) {
				// call BeginUnlock

				// ASSERT: should be error on one of the followings:
				//       - the lockup is used for superfluid delegation
				//
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
