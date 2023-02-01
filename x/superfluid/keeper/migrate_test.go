package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	cl "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v14/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v14/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v14/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v14/x/superfluid/types"
)

// We test migrating in the following circumstances:
// 1. Migrating lock that is not superfluid delegated, not unlocking.
// 2. Migrating lock that is not superfluid delegated, unlocking.
// 3. Migrating lock that is superfluid delegated, not unlocking.
// 4. Migrating lock that is superfluid undelegating, not unlocking.
// 5. Migrating lock that is superfluid undelegating, unlocking.
func (suite *KeeperTestSuite) TestUnlockAndMigrate() {
	testCases := []struct {
		name                     string
		superfluidDelegated      bool
		superfluidUndelegating   bool
		unlocking                bool
		percentOfSharesToMigrate sdk.Dec
	}{
		{
			"lock that is not superfluid delegated, not unlocking",
			false,
			false,
			false,
			sdk.MustNewDecFromStr("0.9"),
		},
		{
			"lock that is not superfluid delegated, unlocking",
			false,
			false,
			true,
			sdk.MustNewDecFromStr("0.6"),
		},
		{
			"lock that is superfluid delegated, not unlocking",
			true,
			false,
			false,
			sdk.MustNewDecFromStr("1"),
		},
		{
			"lock that is superfluid undelegating, not unlocking",
			true,
			true,
			false,
			sdk.MustNewDecFromStr("0.5"),
		},
		{
			"lock that is superfluid undelegating, unlocking",
			true,
			true,
			true,
			sdk.MustNewDecFromStr("0.3"),
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

			// Generate and fund two accounts.
			// Account 1 will be the account that creates the pool.
			// Account 2 will be the account that joins the pool.
			delAddrs := CreateRandomAccounts(2)
			poolCreateAcc := delAddrs[0]
			poolJoinAcc := delAddrs[1]
			for _, acc := range delAddrs {
				err := simapp.FundAccount(bankKeeper, ctx, acc, defaultAcctFunds)
				suite.Require().NoError(err)
			}

			// Set up a single validator.
			valAddr := suite.SetupValidator(stakingtypes.BondStatus(stakingtypes.Bonded))

			// Create a balancer pool of "stake" and "foo".
			msg := balancer.NewMsgCreateBalancerPool(poolCreateAcc, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDec(0),
			}, defaultPoolAssets, defaultFutureGovernor)
			balancerPooId, err := poolmanagerKeeper.CreatePool(ctx, msg)
			suite.Require().NoError(err)

			// Join the balancer pool.
			// Note the account balance before and after joining the pool.
			balanceBeforeJoin := bankKeeper.GetAllBalances(ctx, poolJoinAcc)
			_, _, err = gammKeeper.JoinPoolNoSwap(ctx, poolJoinAcc, balancerPooId, gammtypes.OneShare.MulRaw(50), sdk.Coins{})
			suite.Require().NoError(err)
			balanceAfterJoin := bankKeeper.GetAllBalances(ctx, poolJoinAcc)

			// The balancer join pool amount is the difference between the account balance before and after joining the pool.
			joinPoolAmt, _ := balanceBeforeJoin.SafeSub(balanceAfterJoin)

			// Determine the pool's LP token denomination.
			balancerPool, err := gammKeeper.GetPoolAndPoke(ctx, balancerPooId)
			suite.Require().NoError(err)
			poolDenom := gammtypes.GetPoolShareDenom(balancerPool.GetId())

			// Register the LP token as a superfluid asset
			err = superfluidKeeper.AddNewSuperfluidAsset(ctx, types.SuperfluidAsset{
				Denom:     poolDenom,
				AssetType: types.SuperfluidAssetTypeLPShare,
			})
			suite.Require().NoError(err)

			// Note how much of the LP token the account that joined the pool has.
			poolShareOut := bankKeeper.GetBalance(ctx, poolJoinAcc, poolDenom)

			// Create a cl pool with the same underlying assets as the balancer pool.
			clPool := suite.PrepareCustomConcentratedPool(poolCreateAcc, defaultPoolAssets[0].Token.Denom, defaultPoolAssets[1].Token.Denom, 1, sdk.NewInt(-6), sdk.ZeroDec())

			// Add a sanctioned link between the balancer and concentrated liquidity pool.
			migrationRecord := gammtypes.MigrationRecords{BalancerToConcentratedPoolLinks: []gammtypes.BalancerToConcentratedPoolLink{
				{BalancerPoolId: balancerPool.GetId(), ClPoolId: clPool.GetId()},
			}}
			gammKeeper.SetMigrationInfo(ctx, migrationRecord)

			// The amount of coins we lock is the amount of LP tokens the account that joined the pool has
			// multiplied by the percent of shares to migrate.
			poolShareOut.Amount = poolShareOut.Amount.ToDec().Mul(tc.percentOfSharesToMigrate).RoundInt()
			coinsToLock := poolShareOut

			// The unbonding duration is the same as the staking module's unbonding duration.
			unbondingDuration := stakingKeeper.GetParams(ctx).UnbondingTime

			// Lock the LP tokens for the duration of the unbonding period.
			lockID := suite.LockTokens(poolJoinAcc, sdk.NewCoins(coinsToLock), unbondingDuration)

			// Superfluid delegate the lock if the test case requires it.
			// Note the intermediary account that was created.
			intermediaryAcc := types.SuperfluidIntermediaryAccount{}
			if tc.superfluidDelegated {
				err = superfluidKeeper.SuperfluidDelegate(ctx, poolJoinAcc.String(), lockID, valAddr.String())
				suite.Require().NoError(err)
				intermediaryAccConnection := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, lockID)
				intermediaryAcc = superfluidKeeper.GetIntermediaryAccount(ctx, intermediaryAccConnection)
			}

			// Superfluid undelegate the lock if the test case requires it.
			if tc.superfluidUndelegating {
				err = superfluidKeeper.SuperfluidUndelegate(ctx, poolJoinAcc.String(), lockID)
				suite.Require().NoError(err)
			}

			// Unlock the lock if the test case requires it.
			if tc.unlocking {
				// If lock was superfluid staked, we can't unlock via `BeginUnlock`,
				// we need to unlock lock via `SuperfluidUnbondLock`
				if tc.superfluidUndelegating {
					err = superfluidKeeper.SuperfluidUnbondLock(ctx, lockID, poolJoinAcc.String())
					suite.Require().NoError(err)
				} else {
					lock, err := lockupKeeper.GetLockByID(ctx, lockID)
					suite.Require().NoError(err)
					err = lockupKeeper.BeginUnlock(ctx, lockID, lock.Coins)
					suite.Require().NoError(err)

					// add time to current time to test lock end time
					ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Hour * 24))
				}
			}

			lock, err := lockupKeeper.GetLockByID(ctx, lockID)
			suite.Require().NoError(err)

			// Run the unlock and migrate logic.
			amount0, amount1, _, poolIdLeaving, poolIdEntering, frozenUntil, err := superfluidKeeper.UnlockAndMigrate(ctx, poolJoinAcc, lockID, poolShareOut)
			suite.Require().NoError(err)

			suite.AssertEventEmitted(ctx, gammtypes.TypeEvtPoolExited, 1)

			// Check that position now exists
			minTick, maxTick := cl.GetMinAndMaxTicksFromExponentAtPriceOne(clPool.GetPrecisionFactorAtPriceOne())
			position, err := suite.App.ConcentratedLiquidityKeeper.GetPosition(ctx, poolIdEntering, poolJoinAcc, minTick, maxTick, frozenUntil)
			suite.Require().NoError(err)
			suite.Require().NotNil(position)

			// Expect the poolIdLeaving to be the balancer pool id
			// Expect the poolIdEntering to be the concentrated liquidity pool id
			suite.Require().Equal(balancerPooId, poolIdLeaving)
			suite.Require().Equal(clPool.GetId(), poolIdEntering)

			// exitPool has rounding difference.
			// We test if correct amt has been exited and frozen by comparing with rounding tolerance.
			defaultErrorTolerance := osmomath.ErrTolerance{
				AdditiveTolerance: sdk.NewDec(2),
				RoundingDir:       osmomath.RoundDown,
			}
			suite.Require().Equal(0, defaultErrorTolerance.Compare(joinPoolAmt.AmountOf("foo").ToDec().Mul(tc.percentOfSharesToMigrate).RoundInt(), amount0))
			suite.Require().Equal(0, defaultErrorTolerance.Compare(joinPoolAmt.AmountOf("stake").ToDec().Mul(tc.percentOfSharesToMigrate).RoundInt(), amount1))

			// Check if the lock was deleted.
			_, err = lockupKeeper.GetLockByID(ctx, lockID)
			suite.Require().Error(err)

			// Additional checks if the lock was superfluid staked.
			if tc.superfluidDelegated {
				// Check if migration deleted intermediary account connection.
				addr := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, lockID)
				suite.Require().Equal(addr.String(), "")

				// Check if migration deleted synthetic lockup.
				_, err = lockupKeeper.GetSyntheticLockup(ctx, lockID, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr.String()))
				suite.Require().Error(err)

				// Check if delegation has reduced from intermediary account.
				delegation, found := stakingKeeper.GetDelegation(ctx, intermediaryAcc.GetAccAddress(), valAddr)
				suite.Require().False(found, "expected no delegation, found delegation w/ %d shares", delegation.Shares)
			}
		})
	}
}
