package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	appparams "github.com/osmosis-labs/osmosis/app/params"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

func (suite *KeeperTestSuite) TestSlashLockupsForSlashedOnDelegation() {
	type superfluidDelegation struct {
		valIndex int64
		lpDenom  string
	}
	testCases := []struct {
		name                  string
		validatorStats        []stakingtypes.BondStatus
		superDelegations      []superfluidDelegation
		slashedValIndexes     []int64
		expSlashedLockIndexes []int64
	}{
		{
			"happy path with single validator and delegator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]int64{0},
			[]int64{0},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			poolId := suite.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
			suite.Require().Equal(poolId, uint64(1))

			// setup validators
			valAddrs := []sdk.ValAddress{}
			for _, status := range tc.validatorStats {
				valAddr := suite.SetupValidator(status)
				valAddrs = append(valAddrs, valAddr)
			}

			intermediaryAccs := []types.SuperfluidIntermediaryAccount{}
			locks := []lockuptypes.PeriodLock{}
			slashFactor := sdk.NewDecWithPrec(5, 2)

			// setup superfluid delegations
			for _, del := range tc.superDelegations {
				valAddr := valAddrs[del.valIndex]
				lock := suite.SetupSuperfluidDelegate(valAddr, del.lpDenom)
				expAcc := types.SuperfluidIntermediaryAccount{
					Denom:   lock.Coins[0].Denom,
					ValAddr: valAddr.String(),
				}

				// check delegation from intermediary account to validator
				delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAddress(), valAddr)
				suite.Require().True(found)
				suite.Require().Equal(delegation.Shares, sdk.NewDec(19000000)) // 95% x 20 x 1000000

				// check delegated tokens
				validator, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr)
				suite.Require().True(found)
				delegatedTokens := validator.TokensFromShares(delegation.Shares).TruncateInt()
				suite.Require().Equal(delegatedTokens, sdk.NewInt(19000000))

				// save accounts and locks for future use
				intermediaryAccs = append(intermediaryAccs, expAcc)
				locks = append(locks, lock)
			}

			// slash validator
			for _, valIndex := range tc.slashedValIndexes {
				validator, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddrs[valIndex])
				suite.Require().True(found)
				suite.ctx = suite.ctx.WithBlockHeight(100)
				consAddr, err := validator.GetConsAddr()
				suite.Require().NoError(err)
				// slash by slash factor
				power := sdk.TokensToConsensusPower(validator.Tokens)
				suite.app.StakingKeeper.Slash(suite.ctx, consAddr, 80, power, slashFactor)
			}

			// refresh intermediary account delegations
			suite.NotPanics(func() {
				suite.app.SuperfluidKeeper.SlashLockupsForSlashedOnDelegation(suite.ctx)
			})

			// check lock changes after validator & lockups slashing
			for _, lockIndex := range tc.expSlashedLockIndexes {
				gotLock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, locks[lockIndex].ID)
				suite.Require().NoError(err)
				suite.Require().Equal(
					gotLock.Coins.AmountOf("gamm/pool/1").String(),
					sdk.NewDec(1000000).Mul(sdk.OneDec().Sub(slashFactor)).TruncateInt().String(),
				)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSlashLockupsForUnbondingDelegationSlash() {
	type superfluidDelegation struct {
		valIndex int64
		lpDenom  string
	}
	testCases := []struct {
		name                  string
		validatorStats        []stakingtypes.BondStatus
		superDelegations      []superfluidDelegation
		superUnbondingLockIds []uint64
		slashedValIndexes     []int64
		expSlashedLockIndexes []int64
	}{
		{
			"happy path with single validator and delegator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]uint64{1},
			[]int64{0},
			[]int64{0},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			poolId := suite.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
			suite.Require().Equal(poolId, uint64(1))

			// setup validators
			valAddrs := []sdk.ValAddress{}
			for _, status := range tc.validatorStats {
				valAddr := suite.SetupValidator(status)
				valAddrs = append(valAddrs, valAddr)
			}

			intermediaryAccs := []types.SuperfluidIntermediaryAccount{}
			locks := []lockuptypes.PeriodLock{}

			// setup superfluid delegations
			for _, del := range tc.superDelegations {
				valAddr := valAddrs[del.valIndex]
				lock := suite.SetupSuperfluidDelegate(valAddr, del.lpDenom)
				expAcc := types.SuperfluidIntermediaryAccount{
					Denom:   lock.Coins[0].Denom,
					ValAddr: valAddr.String(),
				}

				// check delegation from intermediary account to validator
				delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAddress(), valAddr)
				suite.Require().True(found)
				suite.Require().Equal(delegation.Shares, sdk.NewDec(19000000)) // 95% x 20 x 1000000

				// save accounts and locks for future use
				intermediaryAccs = append(intermediaryAccs, expAcc)
				locks = append(locks, lock)
			}

			for _, lockId := range tc.superUnbondingLockIds {
				// superfluid undelegate
				err := suite.app.SuperfluidKeeper.SuperfluidUndelegate(suite.ctx, lockId)
				suite.Require().NoError(err)
			}

			// slash unbonding lockups for all intermediary accounts
			for _, acc := range intermediaryAccs {
				suite.NotPanics(func() {
					suite.app.SuperfluidKeeper.SlashLockupsForUnbondingDelegationSlash(
						suite.ctx,
						acc.GetAddress().String(),
						acc.ValAddr,
						sdk.NewDecWithPrec(5, 2))
				})
			}

			// check check unbonding lockup changes
			for _, lockId := range tc.superUnbondingLockIds {
				gotLock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, lockId)
				suite.Require().NoError(err)
				suite.Require().Equal(gotLock.Coins.AmountOf("gamm/pool/1").String(), sdk.NewInt(950000).String())
			}
		})
	}
}
