package keeper_test

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/types"
)

var (
	defaultFutureGovernor = ""

	// pool assets
	defaultFooAsset balancer.PoolAsset = balancer.PoolAsset{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
	}
	defaultBondDenomAsset balancer.PoolAsset = balancer.PoolAsset{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000)),
	}
	defaultPoolAssets []balancer.PoolAsset = []balancer.PoolAsset{defaultFooAsset, defaultBondDenomAsset}
	defaultAcctFunds  sdk.Coins            = sdk.NewCoins(
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000000)),
		sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
		sdk.NewCoin("foo", sdk.NewInt(10000000)),
		sdk.NewCoin("bar", sdk.NewInt(10000000)),
		sdk.NewCoin("baz", sdk.NewInt(10000000)),
	)
)

// we test unpooling in the following circumstances:
// 1. test unpooling lock that is not superfluid delegated, not unlocking
// 2. test unpooling lock that is not superfluid delegated, unlocking
// 3. test unpooling lock that is superfluid delegated, not unlocking
// 4. test unpooling lock that is superfluid undelegating, not unlocking
// 5. test unpooling lock that is superfluid undelegating, unlocking
func (s *KeeperTestSuite) TestUnpool() {
	testCases := []struct {
		name                   string
		superfluidDelegated    bool
		superfluidUndelegating bool
		unlocking              bool
	}{
		{
			"lock that is not superfluid delegated, not unlocking",
			false,
			false,
			false,
		},
		{
			"lock that is not superfluid delegated, unlocking",
			false,
			false,
			true,
		},
		{
			"lock that is superfluid delegated, not unlocking",
			true,
			false,
			false,
		},
		{
			"lock that is superfluid undelegating, not unlocking",
			true,
			true,
			false,
		},
		{
			"lock that is superfluid undelegating, unlocking",
			true,
			true,
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			ctx := suite.Ctx
			bankKeeper := suite.App.BankKeeper
			gammKeeper := suite.App.GAMMKeeper
			superfluidKeeper := suite.App.SuperfluidKeeper
			lockupKeeper := suite.App.LockupKeeper
			stakingKeeper := suite.App.StakingKeeper
			poolmanagerKeeper := suite.App.PoolManagerKeeper

			// generate one delegator Addr, one gamm pool
			delAddrs := CreateRandomAccounts(2)
			poolCreateAcc := delAddrs[0]
			poolJoinAcc := delAddrs[1]
			for _, acc := range delAddrs {
				err := simapp.FundAccount(bankKeeper, ctx, acc, defaultAcctFunds)
				suite.Require().NoError(err)
			}

			// set up validator
			valAddr := suite.SetupValidator(stakingtypes.Bonded)

			// create pool of "stake" and "foo"
			msg := balancer.NewMsgCreateBalancerPool(poolCreateAcc, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDec(0),
			}, defaultPoolAssets, defaultFutureGovernor)

			poolId, err := poolmanagerKeeper.CreatePool(ctx, msg)
			suite.Require().NoError(err)

			// join pool
			balanceBeforeJoin := bankKeeper.GetAllBalances(ctx, poolJoinAcc)
			_, _, err = gammKeeper.JoinPoolNoSwap(ctx, poolJoinAcc, poolId, gammtypes.OneShare.MulRaw(50), sdk.Coins{})
			suite.Require().NoError(err)
			balanceAfterJoin := bankKeeper.GetAllBalances(ctx, poolJoinAcc)

			joinPoolAmt, _ := balanceBeforeJoin.SafeSub(balanceAfterJoin)

			pool, err := gammKeeper.GetPoolAndPoke(ctx, poolId)
			suite.Require().NoError(err)

			poolDenom := gammtypes.GetPoolShareDenom(pool.GetId())
			poolShareOut := bankKeeper.GetBalance(ctx, poolJoinAcc, poolDenom)

			// register a LP token as a superfluid asset
			err = superfluidKeeper.AddNewSuperfluidAsset(ctx, types.SuperfluidAsset{
				Denom:     poolDenom,
				AssetType: types.SuperfluidAssetTypeLPShare,
			})
			suite.Require().NoError(err)

			// whitelist designated pools
			// this should be done via `RunForkLogic` at upgrade
			whitelistedPool := []uint64{poolId}
			superfluidKeeper.SetUnpoolAllowedPools(ctx, whitelistedPool)

			coinsToLock := poolShareOut
			unbondingDuration := stakingKeeper.GetParams(ctx).UnbondingTime

			// create lock
			lockID := suite.LockTokens(poolJoinAcc, sdk.NewCoins(coinsToLock), unbondingDuration)

			// settings prior to testing for superfluid delegated cases
			intermediaryAcc := types.SuperfluidIntermediaryAccount{}
			if tc.superfluidDelegated {
				err = superfluidKeeper.SuperfluidDelegate(ctx, poolJoinAcc.String(), lockID, valAddr.String())
				suite.Require().NoError(err)
				intermediaryAccConnection := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, lockID)
				intermediaryAcc = superfluidKeeper.GetIntermediaryAccount(ctx, intermediaryAccConnection)
			}

			// settings prior to testing for superfluid undelegating cases
			if tc.superfluidUndelegating {
				err = superfluidKeeper.SuperfluidUndelegate(ctx, poolJoinAcc.String(), lockID)
				suite.Require().NoError(err)
			}

			// settings prior to testing for unlocking cases
			if tc.unlocking {
				// if lock was superfluid staked, we can't unlock via `BeginUnlock`,
				// need to unlock lock via `SuperfluidUnbondLock`
				if tc.superfluidUndelegating {
					err = superfluidKeeper.SuperfluidUnbondLock(ctx, lockID, poolJoinAcc.String())
					suite.Require().NoError(err)
				} else {
					lock, err := lockupKeeper.GetLockByID(ctx, lockID)
					suite.Require().NoError(err)
					_, err = lockupKeeper.BeginUnlock(ctx, lockID, lock.Coins)
					suite.Require().NoError(err)

					// add time to current time to test lock end time
					ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Hour * 24))
				}
			}

			lock, err := lockupKeeper.GetLockByID(ctx, lockID)
			suite.Require().NoError(err)

			// run unpooling logic
			newLockIDs, err := superfluidKeeper.UnpoolAllowedPools(ctx, poolJoinAcc, poolId, lockID)
			suite.Require().NoError(err)

			suite.AssertEventEmitted(ctx, gammtypes.TypeEvtPoolExited, 1)

			cumulativeNewLockCoins := sdk.NewCoins()

			for _, newLockId := range newLockIDs {
				newLock, err := lockupKeeper.GetLockByID(ctx, newLockId)
				suite.Require().NoError(err)

				// check lock end time has been preserved after unpooling
				// if lock wasn't unlocking, it should be initiated unlocking
				// if lock was unlocking, lock end time should be preserved
				if tc.unlocking {
					suite.Require().Equal(lock.EndTime, newLock.EndTime)
				} else {
					suite.Require().Equal(ctx.BlockTime().Add(unbondingDuration), newLock.EndTime)
				}

				cumulativeNewLockCoins = cumulativeNewLockCoins.Add(newLock.Coins...)
			}

			// check if the new lock created has the same amount as pool exited

			// exitPool has rounding difference,
			// we test if correct amt has been exited and locked via comparing with rounding tolerance
			roundingToleranceCoins := sdk.NewCoins(sdk.NewCoin(defaultFooAsset.Token.Denom, sdk.NewInt(5)), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(5)))
			roundDownTolerance, _ := joinPoolAmt.SafeSub(roundingToleranceCoins)
			roundUpTolerance := joinPoolAmt.Add(roundingToleranceCoins...)
			suite.Require().True(cumulativeNewLockCoins.AmountOf("foo").GTE(roundDownTolerance.AmountOf("foo")))
			suite.Require().True(cumulativeNewLockCoins.AmountOf(sdk.DefaultBondDenom).GTE(roundDownTolerance.AmountOf(sdk.DefaultBondDenom)))
			suite.Require().True(cumulativeNewLockCoins.AmountOf("foo").LTE(roundUpTolerance.AmountOf("foo")))
			suite.Require().True(cumulativeNewLockCoins.AmountOf(sdk.DefaultBondDenom).LTE(roundUpTolerance.AmountOf(sdk.DefaultBondDenom)))

			// check if old lock is deleted
			_, err = lockupKeeper.GetLockByID(ctx, lockID)
			suite.Require().Error(err)

			// check for locks that were superfluid staked.
			if tc.superfluidDelegated {
				// check if unpooling deleted intermediary account connection
				addr := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, lockID)
				suite.Require().Equal(addr.String(), "")

				// check bonding synthetic lockup deletion
				_, err = lockupKeeper.GetSyntheticLockup(ctx, lockID, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr.String()))
				suite.Require().Error(err)

				// check unbonding synthetic lockup creation
				// unbondingDuration := stakingKeeper.GetParams(ctx).UnbondingTime
				// synthLock, err := lockupKeeper.GetSyntheticLockup(ctx, lockID, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr.String()))
				// suite.Require().NoError(err)
				// suite.Require().Equal(synthLock.UnderlyingLockId, lockID)
				// suite.Require().Equal(synthLock.SynthDenom, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr.String()))
				// suite.Require().Equal(synthLock.EndTime, ctx.BlockTime().Add(unbondingDuration))

				// check if delegation has reduced from intermediary account
				delegation, found := stakingKeeper.GetDelegation(ctx, intermediaryAcc.GetAccAddress(), valAddr)
				suite.Require().False(found, "expected no delegation, found delegation w/ %d shares", delegation.Shares)
			}
		})
	}
}

// TestUnpoolAllowedPools_WhiteList tests that unpooling does not work for pools not in a whitelist.
// Should fail immediately if pool is not in whitelist.
func (s *KeeperTestSuite) TestUnpoolAllowedPools_WhiteList() {
	// lock id does not matter in the context of this test.
	const (
		testLockId        = 1
		whiteListedPoolId = 1
	)
	suite.SetupTest()
	ctx := suite.Ctx
	superfluidKeeper := suite.App.SuperfluidKeeper

	// id 1
	suite.PrepareBalancerPool()
	// id 2
	suite.PrepareBalancerPool()

	// set id 1 to whitelist
	superfluidKeeper.SetUnpoolAllowedPools(ctx, []uint64{whiteListedPoolId})

	_, err := superfluidKeeper.UnpoolAllowedPools(ctx, suite.TestAccs[0], whiteListedPoolId, testLockId)

	// An error should still occur due to incorrect setup. However, it should be unrelated
	// to whitelist.
	suite.Error(err)
	suite.Require().NotErrorIs(err, types.ErrPoolNotWhitelisted)

	_, err = superfluidKeeper.UnpoolAllowedPools(ctx, suite.TestAccs[0], whiteListedPoolId+1, testLockId)

	// Here, we used a non-whitelisted pool id so it should fail with the whitelist error.
	suite.Error(err)
	suite.Require().ErrorIs(err, types.ErrPoolNotWhitelisted)
}

func (s *KeeperTestSuite) TestValidateGammLockForSuperfluid() {
	lockCreator := suite.TestAccs[0]
	nonLockCreator := suite.TestAccs[1]
	type sendTest struct {
		fundsToLock       sdk.Coins
		accountToValidate sdk.AccAddress
		poolIdToValidate  uint64
		lockIdToValidate  uint64
		expectedError     error
	}
	testCases := map[string]sendTest{
		"happy path": {
			fundsToLock:       sdk.NewCoins(sdk.NewCoin("gamm/pool/1", sdk.NewInt(100))),
			accountToValidate: lockCreator,
			poolIdToValidate:  1,
			lockIdToValidate:  1,
		},
		"error: non-existent lock ID": {
			fundsToLock:       sdk.NewCoins(sdk.NewCoin("gamm/pool/1", sdk.NewInt(100))),
			accountToValidate: lockCreator,
			poolIdToValidate:  1,
			lockIdToValidate:  2,
			expectedError:     errorsmod.Wrap(lockuptypes.ErrLockupNotFound, fmt.Sprintf("lock with ID %d does not exist", 2)),
		},
		"error: mismatched owner": {
			fundsToLock:       sdk.NewCoins(sdk.NewCoin("gamm/pool/1", sdk.NewInt(100))),
			accountToValidate: nonLockCreator,
			poolIdToValidate:  1,
			lockIdToValidate:  1,
			expectedError:     lockuptypes.ErrNotLockOwner,
		},
		"error: more than one coin in lock": {
			fundsToLock: sdk.NewCoins(sdk.NewCoin("gamm/pool/1", sdk.NewInt(100)),
				sdk.NewCoin("gamm/pool/2", sdk.NewInt(100))),
			accountToValidate: lockCreator,
			poolIdToValidate:  1,
			lockIdToValidate:  1,
			expectedError:     types.ErrMultipleCoinsLockupNotSupported,
		},
		"error: wrong pool ID provided when compared to lock denom": {
			fundsToLock:       sdk.NewCoins(sdk.NewCoin("gamm/pool/1", sdk.NewInt(100))),
			accountToValidate: lockCreator,
			poolIdToValidate:  2,
			lockIdToValidate:  1,
			expectedError:     types.UnexpectedDenomError{ExpectedDenom: gammtypes.GetPoolShareDenom(2), ProvidedDenom: "gamm/pool/1"},
		},
		"error: right pool ID provided but not gamm/pool/ prefix": {
			fundsToLock:       sdk.NewCoins(sdk.NewCoin("cl/pool/1", sdk.NewInt(100))),
			accountToValidate: lockCreator,
			poolIdToValidate:  1,
			lockIdToValidate:  1,
			expectedError:     types.UnexpectedDenomError{ExpectedDenom: gammtypes.GetPoolShareDenom(1), ProvidedDenom: "cl/pool/1"},
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			suite.SetupTest()

			ctx := suite.Ctx
			superfluidKeeper := suite.App.SuperfluidKeeper

			suite.FundAcc(lockCreator, tc.fundsToLock)
			_, err := suite.App.LockupKeeper.CreateLock(ctx, lockCreator, tc.fundsToLock, time.Hour)
			suite.Require().NoError(err)

			// System under test
			_, err = superfluidKeeper.ValidateGammLockForSuperfluidStaking(ctx, tc.accountToValidate, tc.poolIdToValidate, tc.lockIdToValidate)
			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}
			suite.Require().NoError(err)
		})
	}
}

func (s *KeeperTestSuite) TestGetExistingLockRemainingDuration() {
	defaultJoinTime := suite.Ctx.BlockTime()
	lockCreator := suite.TestAccs[0]
	type sendTest struct {
		isUnlocking               bool
		lockDuration              time.Duration
		timePassed                time.Duration
		expectedRemainingDuration time.Duration
		expectedError             error
	}
	testCases := map[string]sendTest{
		"lock that is not unlocking": {
			isUnlocking:               false,
			lockDuration:              time.Hour,
			timePassed:                time.Hour,
			expectedRemainingDuration: time.Hour,
		},
		"lock that is unlocking": {
			isUnlocking:               true,
			lockDuration:              time.Hour,
			timePassed:                time.Minute,
			expectedRemainingDuration: time.Hour - time.Minute,
		},
		"error: negative duration": {
			isUnlocking:   true,
			lockDuration:  time.Hour,
			timePassed:    time.Hour + time.Minute,
			expectedError: types.NegativeDurationError{Duration: -time.Minute},
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			suite.SetupTest()
			ctx := suite.Ctx.WithBlockTime(defaultJoinTime)

			superfluidKeeper := suite.App.SuperfluidKeeper

			suite.FundAcc(lockCreator, defaultAcctFunds)
			lock, err := suite.App.LockupKeeper.CreateLock(ctx, lockCreator, defaultAcctFunds, tc.lockDuration)
			suite.Require().NoError(err)

			if tc.isUnlocking {
				_, err = suite.App.LockupKeeper.BeginUnlock(ctx, lock.ID, defaultAcctFunds)
				suite.Require().NoError(err)
			}

			ctx = ctx.WithBlockTime(defaultJoinTime.Add(tc.timePassed))

			lockAfterTime, err := suite.App.LockupKeeper.GetLockByID(ctx, lock.ID)
			suite.Require().NoError(err)

			// System under test
			remainingDuration, err := superfluidKeeper.GetExistingLockRemainingDuration(ctx, lockAfterTime)
			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedRemainingDuration, remainingDuration)
		})
	}
}
