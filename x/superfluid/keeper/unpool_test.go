package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

var (
	defaultSwapFee    = sdk.MustNewDecFromStr("0.025")
	defaultExitFee    = sdk.MustNewDecFromStr("0.025")
	defaultPoolParams = balancer.PoolParams{
		SwapFee: defaultSwapFee,
		ExitFee: defaultExitFee,
	}
	defaultFutureGovernor = ""

	// pool assets
	defaultFooAsset gammtypes.PoolAsset = gammtypes.PoolAsset{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
	}
	defaultBondDenomAsset gammtypes.PoolAsset = gammtypes.PoolAsset{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000)),
	}
	defaultPoolAssets []gammtypes.PoolAsset = []gammtypes.PoolAsset{defaultFooAsset, defaultBondDenomAsset}
	defaultAcctFunds  sdk.Coins             = sdk.NewCoins(
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
func (suite *KeeperTestSuite) TestUnpool() {
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

			// generate one delegator Addr, one gamm pool
			delAddrs := CreateRandomAccounts(2)
			poolCreateAcc := delAddrs[0]
			poolJoinAcc := delAddrs[1]
			for _, acc := range delAddrs {
				err := simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, acc, defaultAcctFunds)
				suite.Require().NoError(err)
			}

			// set up validator
			valAddr := suite.SetupValidator(stakingtypes.BondStatus(stakingtypes.Bonded))

			// create pool of "stake" and "foo"
			poolId, err := suite.App.GAMMKeeper.CreateBalancerPool(suite.Ctx, poolCreateAcc, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDec(0),
			}, defaultPoolAssets, defaultFutureGovernor)
			suite.Require().NoError(err)

			// join pool
			balanceBeforeJoin := suite.App.BankKeeper.GetAllBalances(suite.Ctx, poolJoinAcc)
			err = suite.App.GAMMKeeper.JoinPool(suite.Ctx, poolJoinAcc, poolId, gammtypes.OneShare.MulRaw(50), sdk.Coins{
				sdk.NewCoin("foo", sdk.NewInt(5000)),
			})
			suite.Require().NoError(err)
			balanceAfterJoin := suite.App.BankKeeper.GetAllBalances(suite.Ctx, poolJoinAcc)

			joinPoolAmt, _ := balanceBeforeJoin.SafeSub(balanceAfterJoin)

			pool, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, poolId)
			suite.Require().NoError(err)

			poolDenom := pool.GetTotalShares().Denom
			poolShareOut := suite.App.BankKeeper.GetBalance(suite.Ctx, poolJoinAcc, poolDenom)

			// register a LP token as a superfluid asset
			suite.App.SuperfluidKeeper.AddNewSuperfluidAsset(suite.Ctx, types.SuperfluidAsset{
				Denom:     poolDenom,
				AssetType: types.SuperfluidAssetTypeLPShare,
			})

			// whitelist designated pools
			// this should be done via `RunForkLogic` at upgrade
			whitelistedPool := []uint64{poolId}
			suite.App.SuperfluidKeeper.SetUnpoolAllowedPools(suite.Ctx, whitelistedPool)

			coinsToLock := poolShareOut
			unbondingDuration := suite.App.StakingKeeper.GetParams(suite.Ctx).UnbondingTime

			// create lock
			lockID := suite.LockTokens(poolJoinAcc, sdk.NewCoins(coinsToLock), unbondingDuration)

			// settings prior to testing for superfluid delegated cases
			intermediaryAcc := types.SuperfluidIntermediaryAccount{}
			if tc.superfluidDelegated {
				err = suite.App.SuperfluidKeeper.SuperfluidDelegate(suite.Ctx, poolJoinAcc.String(), lockID, valAddr.String())
				suite.Require().NoError(err)
				intermediaryAccConnection := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, lockID)
				intermediaryAcc = suite.App.SuperfluidKeeper.GetIntermediaryAccount(suite.Ctx, intermediaryAccConnection)
			}

			// settings prior to testing for superfluid undelegating cases
			if tc.superfluidUndelegating {
				err = suite.App.SuperfluidKeeper.SuperfluidUndelegate(suite.Ctx, poolJoinAcc.String(), lockID)
				suite.Require().NoError(err)
			}

			// settings prior to testing for unlocking cases
			if tc.unlocking {
				// if lock was superfluid staked, we can't unlock via `BeginUnlock`,
				// need to unlock lock via `SuperfluidUnbondLock`
				if tc.superfluidUndelegating {
					err = suite.App.SuperfluidKeeper.SuperfluidUnbondLock(suite.Ctx, lockID, poolJoinAcc.String())
					suite.Require().NoError(err)
				} else {
					lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockID)
					suite.Require().NoError(err)
					err = suite.App.LockupKeeper.BeginUnlock(suite.Ctx, *lock, lock.Coins)
					suite.Require().NoError(err)

					// add time to current time to test lock end time
					suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(time.Hour * 24))
				}
			}

			lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockID)
			suite.Require().NoError(err)

			// run unpooling logic
			newLockIDs, err := suite.App.SuperfluidKeeper.UnpoolAllowedPools(suite.Ctx, poolJoinAcc, poolId, lockID)
			suite.Require().NoError(err)

			cumulativeNewLockCoins := sdk.NewCoins()

			for _, newLockId := range newLockIDs {
				newLock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, newLockId)
				suite.Require().NoError(err)

				// check lock end time has been preserved after unpooling
				// if lock wasn't unlocking, it should be initiated unlocking
				// if lock was unlocking, lock end time should be preserved
				if tc.unlocking {
					suite.Require().Equal(lock.EndTime, newLock.EndTime)
				} else {
					suite.Require().Equal(suite.Ctx.BlockTime().Add(unbondingDuration), newLock.EndTime)
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
			_, err = suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockID)
			suite.Require().Error(err)

			// check for locks that were superfluid staked.
			if tc.superfluidDelegated {
				// check if unpooling deleted intermediary account connection
				addr := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, lockID)
				suite.Require().Equal(addr.String(), "")

				// check bonding synthetic lockup deletion
				_, err = suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lockID, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr.String()))
				suite.Require().Error(err)

				// check unbonding synthetic lockup creation
				unbondingDuration := suite.App.StakingKeeper.GetParams(suite.Ctx).UnbondingTime
				synthLock, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lockID, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr.String()))
				suite.Require().NoError(err)
				suite.Require().Equal(synthLock.UnderlyingLockId, lockID)
				suite.Require().Equal(synthLock.SynthDenom, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr.String()))
				suite.Require().Equal(synthLock.EndTime, suite.Ctx.BlockTime().Add(unbondingDuration))

				// check if delegation has reduced from intermediary account
				delegation, found := suite.App.StakingKeeper.GetDelegation(suite.Ctx, intermediaryAcc.GetAccAddress(), valAddr)
				suite.Require().False(found, "expected no delegation, found delegation w/ %d shares", delegation.Shares)
			}
		})
	}
}
