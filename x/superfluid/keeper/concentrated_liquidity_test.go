package keeper_test

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
)

func (s *KeeperTestSuite) TestAddToConcentratedLiquiditySuperfluidPosition() {
	defaultJoinTime := s.Ctx.BlockTime()
	owner, nonOwner := apptesting.CreateRandomAccounts(1)[0], apptesting.CreateRandomAccounts(1)[0]
	unbondingTime, err := s.App.StakingKeeper.UnbondingTime(s.Ctx)
	s.Require().NoError(err)
	type sendTest struct {
		superfluidDelegated    bool
		superfluidUndelegating bool
		unlocking              bool
		overwritePositionId    bool
		amount0Added           osmomath.Int
		amount1Added           osmomath.Int
		doNotFundAcc           bool
		isLastPositionInPool   bool
		overwriteExecutionAcc  bool
		expectedError          error
	}
	testCases := map[string]sendTest{
		"add to position that is superfluid delegated, not unlocking": {
			superfluidDelegated: true,
			amount0Added:        osmomath.NewInt(100000000),
			amount1Added:        osmomath.NewInt(100000000),
		},
		"error: negative amount 0": {
			superfluidDelegated: true,
			doNotFundAcc:        true,
			amount0Added:        osmomath.NewInt(-100000000),
			amount1Added:        osmomath.NewInt(100000000),
			expectedError:       cltypes.NegativeAmountAddedError{PositionId: 1, Asset0Amount: osmomath.NewInt(-100000000), Asset1Amount: osmomath.NewInt(100000000)},
		},
		"error: negative amount 1": {
			superfluidDelegated: true,
			doNotFundAcc:        true,
			amount0Added:        osmomath.NewInt(100000000),
			amount1Added:        osmomath.NewInt(-100000000),
			expectedError:       cltypes.NegativeAmountAddedError{PositionId: 1, Asset0Amount: osmomath.NewInt(100000000), Asset1Amount: osmomath.NewInt(-100000000)},
		},
		"error: not underlying lock owner of the position": {
			superfluidDelegated:   true,
			overwriteExecutionAcc: true,
			amount0Added:          osmomath.NewInt(100000000),
			amount1Added:          osmomath.NewInt(100000000),
			expectedError:         types.LockOwnerMismatchError{LockId: 1, LockOwner: owner.String(), ProvidedOwner: nonOwner.String()},
		},
		"error: not enough funds to add": {
			doNotFundAcc:        true,
			superfluidDelegated: true,
			amount0Added:        osmomath.NewInt(100000000),
			amount1Added:        osmomath.NewInt(100000000),
			expectedError:       errors.New("insufficient funds"),
		},
		"error: last position in pool": {
			superfluidDelegated:  true,
			isLastPositionInPool: true,
			amount0Added:         osmomath.NewInt(100000000),
			amount1Added:         osmomath.NewInt(100000000),
			expectedError:        cltypes.AddToLastPositionInPoolError{PoolId: 1, PositionId: 1},
		},
		"error: lock that is not superfluid delegated, not unlocking": {
			amount0Added:  osmomath.NewInt(100000000),
			amount1Added:  osmomath.NewInt(100000000),
			expectedError: types.ErrNotSuperfluidUsedLockup,
		},
		"error: lock that is not superfluid delegated, unlocking": {
			unlocking:     true,
			amount0Added:  osmomath.NewInt(100000000),
			amount1Added:  osmomath.NewInt(100000000),
			expectedError: types.LockImproperStateError{LockId: 1, UnbondingDuration: unbondingTime.String()},
		},
		"error: lock that is superfluid undelegating, not unlocking": {
			superfluidDelegated:    true,
			superfluidUndelegating: true,
			amount0Added:           osmomath.NewInt(100000000),
			amount1Added:           osmomath.NewInt(100000000),
			expectedError:          types.ErrNotSuperfluidUsedLockup,
		},
		"error: lock that is superfluid undelegating, unlocking": {
			superfluidDelegated:    true,
			superfluidUndelegating: true,
			unlocking:              true,
			amount0Added:           osmomath.NewInt(100000000),
			amount1Added:           osmomath.NewInt(100000000),
			expectedError:          types.LockImproperStateError{LockId: 1, UnbondingDuration: unbondingTime.String()},
		},
		"error: non-existent position ID": {
			overwritePositionId: true,
			amount0Added:        osmomath.NewInt(100000000),
			amount1Added:        osmomath.NewInt(100000000),
			expectedError:       cltypes.PositionIdNotFoundError{PositionId: 5},
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)
			ctx := s.Ctx
			superfluidKeeper := s.App.SuperfluidKeeper
			lockupKeeper := s.App.LockupKeeper
			stakingKeeper := s.App.StakingKeeper
			concentratedLiquidityKeeper := s.App.ConcentratedLiquidityKeeper
			bankKeeper := s.App.BankKeeper
			bondDenom, err := stakingKeeper.BondDenom(ctx)
			s.Require().NoError(err)

			// Run test setup logic.
			positionId, lockId, amount0, amount1, valAddr, poolJoinAcc := s.SetupSuperfluidConcentratedPosition(ctx, tc.superfluidDelegated, tc.superfluidUndelegating, tc.unlocking, owner)
			clPool, err := concentratedLiquidityKeeper.GetConcentratedPoolById(ctx, 1)
			s.Require().NoError(err)
			clPoolAddress := clPool.GetAddress()

			executionAcc := poolJoinAcc

			if tc.overwriteExecutionAcc {
				executionAcc = nonOwner
			}

			if !tc.doNotFundAcc {
				s.FundAcc(executionAcc, sdk.NewCoins(sdk.NewCoin(clPool.GetToken0(), tc.amount0Added), sdk.NewCoin(clPool.GetToken1(), tc.amount1Added)))
			}

			if !tc.isLastPositionInPool {
				fundCoins := sdk.NewCoins(sdk.NewCoin(clPool.GetToken0(), osmomath.NewInt(100000000)), sdk.NewCoin(clPool.GetToken1(), osmomath.NewInt(100000000)))
				s.FundAcc(nonOwner, fundCoins)
				_, err := concentratedLiquidityKeeper.CreateFullRangePosition(ctx, clPool.GetId(), nonOwner, fundCoins)
				s.Require().NoError(err)
			}

			if tc.overwritePositionId {
				positionId = 5
			}

			preAddToPositionStakeSupply := bankKeeper.GetSupply(ctx, bondDenom)
			preAddToPositionPoolFunds := bankKeeper.GetAllBalances(ctx, clPoolAddress)

			// System under test.
			positionData, newLockId, err := superfluidKeeper.AddToConcentratedLiquiditySuperfluidPosition(ctx, executionAcc, positionId, tc.amount0Added, tc.amount1Added)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}
			s.Require().NoError(err)

			// We allow for an downward additive tolerance of 101 to accommodate our single happy path case while efficiently checking exact balance diffs.
			//
			// Using our full range asset amount equations, we get the following:
			//
			// expectedAsset0 = floor((liquidityDelta * (maxSqrtPrice - curSqrtPrice)) / (maxSqrtPrice * curSqrtPrice)) = 99999998.000000000000000000
			// expectedAsset1 = floor(liquidityDelta * (curSqrtPrice - minSqrtPrice)) =  99999899.000000000000000000
			//
			// Note that the expected difference valid additive difference of 101 on asset 1.
			var errTolerance osmomath.ErrTolerance
			errTolerance.AdditiveTolerance = osmomath.NewDec(101)
			errTolerance.RoundingDir = osmomath.RoundDown

			postAddToPositionStakeSupply := bankKeeper.GetSupply(ctx, bondDenom)
			postAddToPositionPoolFunds := bankKeeper.GetAllBalances(ctx, clPoolAddress)

			// Check that bond denom supply changed by the amount of bond denom added (taking into consideration risk adjusted osmo value and err tolerance)
			diffInBondDenomSupply := postAddToPositionStakeSupply.Amount.Sub(preAddToPositionStakeSupply.Amount)
			expectedBondDenomSupplyDiff := superfluidKeeper.GetRiskAdjustedOsmoValue(ctx, tc.amount0Added)
			osmoassert.Equal(s.T(), errTolerance, expectedBondDenomSupplyDiff, diffInBondDenomSupply)
			// Check that the pool funds changed by the amount of tokens added (taking into consideration err tolerance)
			diffInPoolFundsToken0 := postAddToPositionPoolFunds.AmountOf(clPool.GetToken0()).Sub(preAddToPositionPoolFunds.AmountOf(clPool.GetToken0()))
			osmoassert.Equal(s.T(), errTolerance, tc.amount0Added, diffInPoolFundsToken0)
			diffInPoolFundsToken1 := postAddToPositionPoolFunds.AmountOf(clPool.GetToken1()).Sub(preAddToPositionPoolFunds.AmountOf(clPool.GetToken1()))
			osmoassert.Equal(s.T(), errTolerance, tc.amount1Added, diffInPoolFundsToken1)

			expectedNewCoins := sdk.NewCoins(sdk.NewCoin(clPool.GetToken0(), amount0.Add(tc.amount0Added)), sdk.NewCoin(clPool.GetToken1(), amount1.Add(tc.amount1Added)))

			clPoolDenom := cltypes.GetConcentratedLockupDenomFromPoolId(clPool.GetId())
			expectedLockCoins := sdk.NewCoins(sdk.NewCoin(clPoolDenom, positionData.Liquidity.TruncateInt()))

			// Resulting position should have the expected amount of coins within one unit (rounding down).
			osmoassert.Equal(s.T(), errTolerance, expectedNewCoins[0].Amount, positionData.Amount0)
			osmoassert.Equal(s.T(), errTolerance, expectedNewCoins[1].Amount, positionData.Amount1)

			// Check the new lock.
			unbondingTime, err = s.App.StakingKeeper.UnbondingTime(s.Ctx)
			s.Require().NoError(err)
			newLock, err := s.App.LockupKeeper.GetLockByID(ctx, newLockId)
			s.Require().NoError(err)
			s.Require().Equal(unbondingTime, newLock.Duration)
			s.Require().True(newLock.EndTime.IsZero())
			s.Require().Equal(poolJoinAcc.String(), newLock.Owner)
			s.Require().Equal(expectedLockCoins.String(), newLock.Coins.String())

			// Check that a new position and lock ID were generated.
			s.Require().NotEqual(positionId, positionData.ID)
			s.Require().NotEqual(lockId, newLockId)

			// Check if intermediary account connection for the old lock ID is deleted.
			oldIntermediaryAcc := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, lockId)
			s.Require().Equal(oldIntermediaryAcc.String(), "")

			// Check if intermediary account connection for the new lock ID is created.
			expAcc := types.NewSuperfluidIntermediaryAccount(clPoolDenom, valAddr.String(), 0)
			newIntermediaryAcc := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, newLockId)
			s.Require().Equal(expAcc.GetAccAddress().String(), newIntermediaryAcc.String())

			// Check if synthetic lockup for the old lock ID is deleted.
			_, err = lockupKeeper.GetSyntheticLockup(ctx, lockId, keeper.StakingSyntheticDenom(clPoolDenom, valAddr.String()))
			s.Require().Error(err)

			// Check if synthetic lockup for the new lock ID is created.
			_, err = lockupKeeper.GetSyntheticLockup(ctx, newLockId, keeper.StakingSyntheticDenom(clPoolDenom, valAddr.String()))
			s.Require().NoError(err)

			// Check if the old intermediary account has no delegation.
			_, err = stakingKeeper.GetDelegation(ctx, oldIntermediaryAcc, valAddr)
			s.Require().Error(err)

			// Check if the new intermediary account has expected delegation amount.
			expectedDelegationAmt := superfluidKeeper.GetRiskAdjustedOsmoValue(ctx, positionData.Amount0)
			delegationAmt, err := stakingKeeper.GetDelegation(ctx, newIntermediaryAcc, valAddr)
			s.Require().NoError(err)
			s.Require().Equal(expectedDelegationAmt, delegationAmt.Shares.TruncateInt())
		})
	}
}

func (s *KeeperTestSuite) SetupSuperfluidConcentratedPosition(ctx sdk.Context, superfluidDelegated, superfluidUndelegating, unlocking bool, owner sdk.AccAddress) (positionId, lockId uint64, amount0, amount1 osmomath.Int, valAddr sdk.ValAddress, poolJoinAcc sdk.AccAddress) { //nolint:revive // TODO: refactor this function
	bankKeeper := s.App.BankKeeper
	superfluidKeeper := s.App.SuperfluidKeeper
	lockupKeeper := s.App.LockupKeeper
	stakingKeeper := s.App.StakingKeeper

	fullRangeCoins := sdk.NewCoins(defaultPoolAssets[0].Token, defaultPoolAssets[1].Token)

	// Generate and fund two accounts.
	// Account 1 will be the account that creates the pool.
	// Account 2 will be the account that joins the pool.
	delAddrs := CreateRandomAccounts(1)
	poolCreateAcc := delAddrs[0]
	delAddrs = append(delAddrs, owner)
	poolJoinAcc = delAddrs[1]
	for _, acc := range delAddrs {
		err := testutil.FundAccount(ctx, bankKeeper, acc, defaultAcctFunds)
		s.Require().NoError(err)
	}

	// Set up a single validator.
	valAddr = s.SetupValidator(stakingtypes.Bonded)

	// Create a cl pool.
	clPool := s.PrepareCustomConcentratedPool(poolCreateAcc, defaultPoolAssets[0].Token.Denom, defaultPoolAssets[1].Token.Denom, 1, osmomath.ZeroDec())
	clPoolId := clPool.GetId()

	// The lock duration is the same as the staking module's unbonding duration.
	stakingParams, err := stakingKeeper.GetParams(ctx)
	s.Require().NoError(err)
	unbondingDuration := stakingParams.UnbondingTime

	// Create a full range position in the concentrated liquidity pool.
	positionData, lockId, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(s.Ctx, clPoolId, poolJoinAcc, fullRangeCoins, unbondingDuration)
	s.Require().NoError(err)
	positionId = positionData.ID
	amount0 = positionData.Amount0
	amount1 = positionData.Amount1

	// Register the CL full range LP tokens as a superfluid asset.
	clPoolDenom := cltypes.GetConcentratedLockupDenomFromPoolId(clPoolId)
	err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{
		Denom:     clPoolDenom,
		AssetType: types.SuperfluidAssetTypeConcentratedShare,
	})
	s.Require().NoError(err)

	// Superfluid delegate the cl lock if the test case requires it.
	// Note the intermediary account that was created.
	if superfluidDelegated {
		err = superfluidKeeper.SuperfluidDelegate(ctx, poolJoinAcc.String(), lockId, valAddr.String())
		s.Require().NoError(err)
	}

	// Superfluid undelegate the lock if the test case requires it.
	if superfluidUndelegating {
		err = superfluidKeeper.SuperfluidUndelegate(ctx, poolJoinAcc.String(), lockId)
		s.Require().NoError(err)
	}

	// Unlock the cl lock if the test case requires it.
	if unlocking {
		// If lock was superfluid staked, we can't unlock via `BeginUnlock`,
		// we need to unlock lock via `SuperfluidUnbondLock`
		if superfluidUndelegating {
			err = superfluidKeeper.SuperfluidUnbondLock(ctx, lockId, poolJoinAcc.String())
			s.Require().NoError(err)
		} else {
			lock, err := lockupKeeper.GetLockByID(ctx, lockId)
			s.Require().NoError(err)
			_, err = lockupKeeper.BeginUnlock(ctx, lockId, lock.Coins)
			s.Require().NoError(err)
		}
	}

	return positionId, lockId, amount0, amount1, valAddr, poolJoinAcc
}
