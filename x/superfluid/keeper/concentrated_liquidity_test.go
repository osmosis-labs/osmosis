package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/types"
)

func (suite *KeeperTestSuite) TestAddToConcentratedLiquiditySuperfluidPosition() {
	defaultJoinTime := suite.Ctx.BlockTime()
	owner := suite.TestAccs[0]
	nonOwner := suite.TestAccs[1]
	type sendTest struct {
		superfluidDelegated    bool
		superfluidUndelegating bool
		unlocking              bool
		overwritePositionId    bool
		amount0Added           sdk.Int
		amount1Added           sdk.Int
		doNotFundAcc           bool
		isLastPositionInPool   bool
		overwriteExecutionAcc  bool
		expectedError          error
	}
	testCases := map[string]sendTest{
		"add to position that is superfluid delegated, not unlocking": {
			superfluidDelegated: true,
			amount0Added:        sdk.NewInt(100000000),
			amount1Added:        sdk.NewInt(100000000),
		},
		"error: negative amount 0": {
			superfluidDelegated: true,
			doNotFundAcc:        true,
			amount0Added:        sdk.NewInt(-100000000),
			amount1Added:        sdk.NewInt(100000000),
			expectedError:       cltypes.NegativeAmountAddedError{PositionId: 1, Asset0Amount: sdk.NewInt(-100000000), Asset1Amount: sdk.NewInt(100000000)},
		},
		"error: negative amount 1": {
			superfluidDelegated: true,
			doNotFundAcc:        true,
			amount0Added:        sdk.NewInt(100000000),
			amount1Added:        sdk.NewInt(-100000000),
			expectedError:       cltypes.NegativeAmountAddedError{PositionId: 1, Asset0Amount: sdk.NewInt(100000000), Asset1Amount: sdk.NewInt(-100000000)},
		},
		"error: not underlying lock owner of the position": {
			superfluidDelegated:   true,
			overwriteExecutionAcc: true,
			amount0Added:          sdk.NewInt(100000000),
			amount1Added:          sdk.NewInt(100000000),
			expectedError:         types.LockOwnerMismatchError{LockId: 1, LockOwner: owner.String(), ProvidedOwner: nonOwner.String()},
		},
		"error: not enough funds to add": {
			doNotFundAcc:        true,
			superfluidDelegated: true,
			amount0Added:        sdk.NewInt(100000000),
			amount1Added:        sdk.NewInt(100000000),
			expectedError:       fmt.Errorf("insufficient funds"),
		},
		"error: last position in pool": {
			superfluidDelegated:  true,
			isLastPositionInPool: true,
			amount0Added:         sdk.NewInt(100000000),
			amount1Added:         sdk.NewInt(100000000),
			expectedError:        cltypes.AddToLastPositionInPoolError{PoolId: 1, PositionId: 1},
		},
		"error: lock that is not superfluid delegated, not unlocking": {
			amount0Added:  sdk.NewInt(100000000),
			amount1Added:  sdk.NewInt(100000000),
			expectedError: types.ErrNotSuperfluidUsedLockup,
		},
		"error: lock that is not superfluid delegated, unlocking": {
			unlocking:     true,
			amount0Added:  sdk.NewInt(100000000),
			amount1Added:  sdk.NewInt(100000000),
			expectedError: types.LockImproperStateError{LockId: 1, UnbondingDuration: suite.App.StakingKeeper.UnbondingTime(suite.Ctx).String()},
		},
		"error: lock that is superfluid undelegating, not unlocking": {
			superfluidDelegated:    true,
			superfluidUndelegating: true,
			amount0Added:           sdk.NewInt(100000000),
			amount1Added:           sdk.NewInt(100000000),
			expectedError:          types.ErrNotSuperfluidUsedLockup,
		},
		"error: lock that is superfluid undelegating, unlocking": {
			superfluidDelegated:    true,
			superfluidUndelegating: true,
			unlocking:              true,
			amount0Added:           sdk.NewInt(100000000),
			amount1Added:           sdk.NewInt(100000000),
			expectedError:          types.LockImproperStateError{LockId: 1, UnbondingDuration: suite.App.StakingKeeper.UnbondingTime(suite.Ctx).String()},
		},
		"error: non-existent position ID": {
			overwritePositionId: true,
			amount0Added:        sdk.NewInt(100000000),
			amount1Added:        sdk.NewInt(100000000),
			expectedError:       cltypes.PositionIdNotFoundError{PositionId: 5},
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			suite.SetupTest()
			suite.Ctx = suite.Ctx.WithBlockTime(defaultJoinTime)
			ctx := suite.Ctx
			superfluidKeeper := suite.App.SuperfluidKeeper
			lockupKeeper := suite.App.LockupKeeper
			stakingKeeper := suite.App.StakingKeeper
			concentratedLiquidityKeeper := suite.App.ConcentratedLiquidityKeeper
			bankKeeper := suite.App.BankKeeper
			bondDenom := stakingKeeper.BondDenom(ctx)

			// Run test setup logic.
			positionId, lockId, amount0, amount1, valAddr, poolJoinAcc := suite.SetupSuperfluidConcentratedPosition(ctx, tc.superfluidDelegated, tc.superfluidUndelegating, tc.unlocking, owner)
			clPool, err := concentratedLiquidityKeeper.GetPoolFromPoolIdAndConvertToConcentrated(ctx, 1)
			suite.Require().NoError(err)
			clPoolAddress := clPool.GetAddress()

			executionAcc := poolJoinAcc

			if tc.overwriteExecutionAcc {
				executionAcc = nonOwner
			}

			if !tc.doNotFundAcc {
				suite.FundAcc(executionAcc, sdk.NewCoins(sdk.NewCoin(clPool.GetToken0(), tc.amount0Added), sdk.NewCoin(clPool.GetToken1(), tc.amount1Added)))
			}

			if !tc.isLastPositionInPool {
				fundCoins := sdk.NewCoins(sdk.NewCoin(clPool.GetToken0(), sdk.NewInt(100000000)), sdk.NewCoin(clPool.GetToken1(), sdk.NewInt(100000000)))
				suite.FundAcc(nonOwner, fundCoins)
				_, _, _, _, _, err := concentratedLiquidityKeeper.CreateFullRangePosition(ctx, clPool.GetId(), nonOwner, fundCoins)
				suite.Require().NoError(err)
			}

			if tc.overwritePositionId {
				positionId = 5
			}

			preAddToPositionStakeSupply := bankKeeper.GetSupply(ctx, bondDenom)
			preAddToPositionPoolFunds := bankKeeper.GetAllBalances(ctx, clPoolAddress)

			// System under test.
			newPositionId, finalAmount0, finalAmount1, newLiquidity, newLockId, err := superfluidKeeper.AddToConcentratedLiquiditySuperfluidPosition(ctx, executionAcc, positionId, tc.amount0Added, tc.amount1Added)
			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}
			suite.Require().NoError(err)

			// Define error tolerance
			var errTolerance osmomath.ErrTolerance
			errTolerance.AdditiveTolerance = sdk.NewDec(1)
			errTolerance.RoundingDir = osmomath.RoundDown

			postAddToPositionStakeSupply := bankKeeper.GetSupply(ctx, bondDenom)
			postAddToPositionPoolFunds := bankKeeper.GetAllBalances(ctx, clPoolAddress)

			// Check that bond denom supply changed by the amount of bond denom added (taking into consideration risk adjusted osmo value and err tolerance)
			diffInBondDenomSupply := postAddToPositionStakeSupply.Amount.Sub(preAddToPositionStakeSupply.Amount)
			expectedBondDenomSupplyDiff := superfluidKeeper.GetRiskAdjustedOsmoValue(ctx, tc.amount0Added)
			suite.Require().Equal(0, errTolerance.Compare(expectedBondDenomSupplyDiff, diffInBondDenomSupply), fmt.Sprintf("expected (%s), actual (%s)", expectedBondDenomSupplyDiff, diffInBondDenomSupply))

			// Check that the pool funds changed by the amount of tokens added (taking into consideration err tolerance)
			diffInPoolFundsToken0 := postAddToPositionPoolFunds.AmountOf(clPool.GetToken0()).Sub(preAddToPositionPoolFunds.AmountOf(clPool.GetToken0()))
			suite.Require().Equal(0, errTolerance.Compare(tc.amount0Added, diffInPoolFundsToken0), fmt.Sprintf("expected (%s), actual (%s)", tc.amount0Added, diffInPoolFundsToken0))
			diffInPoolFundsToken1 := postAddToPositionPoolFunds.AmountOf(clPool.GetToken1()).Sub(preAddToPositionPoolFunds.AmountOf(clPool.GetToken1()))
			suite.Require().Equal(0, errTolerance.Compare(tc.amount1Added, diffInPoolFundsToken1), fmt.Sprintf("expected (%s), actual (%s)", tc.amount1Added, diffInPoolFundsToken1))

			expectedNewCoins := sdk.NewCoins(sdk.NewCoin(clPool.GetToken0(), amount0.Add(tc.amount0Added)), sdk.NewCoin(clPool.GetToken1(), amount1.Add(tc.amount1Added)))

			clPoolDenom := cltypes.GetConcentratedLockupDenomFromPoolId(clPool.GetId())
			expectedLockCoins := sdk.NewCoins(sdk.NewCoin(clPoolDenom, newLiquidity.TruncateInt()))

			// Resulting position should have the expected amount of coins within one unit (rounding down).
			suite.Require().Equal(0, errTolerance.Compare(expectedNewCoins[0].Amount, finalAmount0), fmt.Sprintf("expected (%s), actual (%s)", expectedNewCoins[0].Amount, finalAmount0))
			suite.Require().Equal(0, errTolerance.Compare(expectedNewCoins[1].Amount, finalAmount1), fmt.Sprintf("expected (%s), actual (%s)", expectedNewCoins[1].Amount, finalAmount1))

			// Check the new lock.
			newLock, err := suite.App.LockupKeeper.GetLockByID(ctx, newLockId)
			suite.Require().NoError(err)
			suite.Require().Equal(suite.App.StakingKeeper.UnbondingTime(ctx), newLock.Duration)
			suite.Require().True(newLock.EndTime.IsZero())
			suite.Require().Equal(poolJoinAcc.String(), newLock.Owner)
			suite.Require().Equal(expectedLockCoins.String(), newLock.Coins.String())

			// Check that a new position and lock ID were generated.
			suite.Require().NotEqual(positionId, newPositionId)
			suite.Require().NotEqual(lockId, newLockId)

			// Check if intermediary account connection for the old lock ID is deleted.
			oldIntermediaryAcc := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, lockId)
			suite.Require().Equal(oldIntermediaryAcc.String(), "")

			// Check if intermediary account connection for the new lock ID is created.
			expAcc := types.NewSuperfluidIntermediaryAccount(clPoolDenom, valAddr.String(), 0)
			newIntermediaryAcc := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, newLockId)
			suite.Require().Equal(expAcc.GetAccAddress().String(), newIntermediaryAcc.String())

			// Check if synthetic lockup for the old lock ID is deleted.
			_, err = lockupKeeper.GetSyntheticLockup(ctx, lockId, keeper.StakingSyntheticDenom(clPoolDenom, valAddr.String()))
			suite.Require().Error(err)

			// Check if synthetic lockup for the new lock ID is created.
			_, err = lockupKeeper.GetSyntheticLockup(ctx, newLockId, keeper.StakingSyntheticDenom(clPoolDenom, valAddr.String()))
			suite.Require().NoError(err)

			// Check if the old intermediary account has no delegation.
			_, found := stakingKeeper.GetDelegation(ctx, oldIntermediaryAcc, valAddr)
			suite.Require().False(found)

			// Check if the new intermediary account has expected delegation amount.
			expectedDelegationAmt := superfluidKeeper.GetRiskAdjustedOsmoValue(ctx, finalAmount0)
			delegationAmt, found := stakingKeeper.GetDelegation(ctx, newIntermediaryAcc, valAddr)
			suite.Require().True(found)
			suite.Require().Equal(expectedDelegationAmt, delegationAmt.Shares.TruncateInt())

		})
	}
}

func (suite *KeeperTestSuite) SetupSuperfluidConcentratedPosition(ctx sdk.Context, superfluidDelegated, superfluidUndelegating, unlocking bool, owner sdk.AccAddress) (positionId, lockId uint64, amount0, amount1 sdk.Int, valAddr sdk.ValAddress, poolJoinAcc sdk.AccAddress) {
	bankKeeper := suite.App.BankKeeper
	superfluidKeeper := suite.App.SuperfluidKeeper
	lockupKeeper := suite.App.LockupKeeper
	stakingKeeper := suite.App.StakingKeeper

	fullRangeCoins := sdk.NewCoins(defaultPoolAssets[0].Token, defaultPoolAssets[1].Token)

	// Generate and fund two accounts.
	// Account 1 will be the account that creates the pool.
	// Account 2 will be the account that joins the pool.
	delAddrs := CreateRandomAccounts(1)
	poolCreateAcc := delAddrs[0]
	delAddrs = append(delAddrs, owner)
	poolJoinAcc = delAddrs[1]
	for _, acc := range delAddrs {
		err := simapp.FundAccount(bankKeeper, ctx, acc, defaultAcctFunds)
		suite.Require().NoError(err)
	}

	// Set up a single validator.
	valAddr = suite.SetupValidator(stakingtypes.BondStatus(stakingtypes.Bonded))

	// Create a cl pool.
	clPool := suite.PrepareCustomConcentratedPool(poolCreateAcc, defaultPoolAssets[0].Token.Denom, defaultPoolAssets[1].Token.Denom, 1, sdk.ZeroDec())
	clPoolId := clPool.GetId()

	// The lock duration is the same as the staking module's unbonding duration.
	unbondingDuration := stakingKeeper.GetParams(ctx).UnbondingTime

	// Create a full range position in the concentrated liquidity pool.
	positionId, amount0, amount1, _, _, lockId, err := suite.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(suite.Ctx, clPoolId, poolJoinAcc, fullRangeCoins, unbondingDuration)
	suite.Require().NoError(err)

	// Register the CL full range LP tokens as a superfluid asset.
	clPoolDenom := cltypes.GetConcentratedLockupDenomFromPoolId(clPoolId)
	err = suite.App.SuperfluidKeeper.AddNewSuperfluidAsset(suite.Ctx, types.SuperfluidAsset{
		Denom:     clPoolDenom,
		AssetType: types.SuperfluidAssetTypeConcentratedShare,
	})
	suite.Require().NoError(err)

	// Superfluid delegate the cl lock if the test case requires it.
	// Note the intermediary account that was created.
	if superfluidDelegated {
		err = superfluidKeeper.SuperfluidDelegate(ctx, poolJoinAcc.String(), lockId, valAddr.String())
		suite.Require().NoError(err)
	}

	// Superfluid undelegate the lock if the test case requires it.
	if superfluidUndelegating {
		err = superfluidKeeper.SuperfluidUndelegate(ctx, poolJoinAcc.String(), lockId)
		suite.Require().NoError(err)
	}

	// Unlock the cl lock if the test case requires it.
	if unlocking {
		// If lock was superfluid staked, we can't unlock via `BeginUnlock`,
		// we need to unlock lock via `SuperfluidUnbondLock`
		if superfluidUndelegating {
			err = superfluidKeeper.SuperfluidUnbondLock(ctx, lockId, poolJoinAcc.String())
			suite.Require().NoError(err)
		} else {
			lock, err := lockupKeeper.GetLockByID(ctx, lockId)
			suite.Require().NoError(err)
			_, err = lockupKeeper.BeginUnlock(ctx, lockId, lock.Coins)
			suite.Require().NoError(err)
		}
	}

	return positionId, lockId, amount0, amount1, valAddr, poolJoinAcc
}
