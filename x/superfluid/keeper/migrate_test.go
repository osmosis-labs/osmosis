package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v15/x/superfluid/types"
)

// We test migrating in the following circumstances:
// 1. Migrating lock that is not superfluid delegated, not unlocking.
// 2. Migrating lock that is not superfluid delegated, unlocking.
// 3. Migrating lock that is superfluid delegated, not unlocking.
// 4. Migrating lock that is superfluid undelegating, not unlocking.
// 5. Migrating lock that is superfluid undelegating, unlocking.
func (suite *KeeperTestSuite) TestMigrateLockedPositionFromBalancerToConcentrated() {
	defaultJoinTime := suite.Ctx.BlockTime()
	type sendTest struct {
		superfluidDelegated            bool
		superfluidUndelegating         bool
		unlocking                      bool
		overwriteLockId                bool
		multiAssetLock                 bool
		clLiquidityLock                bool
		noInitialConcentratedSpotPrice bool
		percentOfSharesToMigrate       sdk.Dec
		expectedError                  error
	}
	testCases := map[string]sendTest{
		"lock that is not superfluid delegated, not unlocking": {
			// migrateNonSuperfluidLockBalancerToConcentrated
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.9"),
		},
		"lock that is not superfluid delegated, unlocking": {
			// migrateNonSuperfluidLockBalancerToConcentrated
			unlocking:                true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.6"),
		},
		"lock that is superfluid delegated, not unlocking (full shares)": {
			// migrateSuperfluidBondedBalancerToConcentrated
			superfluidDelegated:      true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
		},
		"lock that is superfluid delegated, not unlocking (partial shares)": {
			// migrateSuperfluidBondedBalancerToConcentrated
			superfluidDelegated:      true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.5"),
		},
		"lock that is superfluid undelegating, not unlocking": {
			// migrateSuperfluidUnbondingBalancerToConcentrated
			superfluidDelegated:      true,
			superfluidUndelegating:   true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.5"),
		},
		"lock that is superfluid undelegating, unlocking": {
			// migrateSuperfluidUnbondingBalancerToConcentrated
			superfluidDelegated:      true,
			superfluidUndelegating:   true,
			unlocking:                true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.3"),
		},
		"error: non-existent lock": {
			overwriteLockId:          true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			expectedError:            sdkerrors.Wrap(lockuptypes.ErrLockupNotFound, fmt.Sprintf("lock with ID %d does not exist", 5)),
		},
		"error: multi-asset lock": {
			multiAssetLock:           true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			expectedError:            superfluidtypes.ErrMultipleCoinsLockupNotSupported,
		},
		"error: lock with cl assets": {
			clLiquidityLock:          true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			expectedError:            superfluidtypes.ErrLockUnpoolNotAllowed,
		},
		"error: cannot add sf asset without spot price": {
			noInitialConcentratedSpotPrice: true,
			percentOfSharesToMigrate:       sdk.MustNewDecFromStr("1"),
			expectedError:                  fmt.Errorf("panic occurred during execution"),
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			suite.SetupTest()
			suite.Ctx = suite.Ctx.WithBlockTime(defaultJoinTime)
			ctx := suite.Ctx
			bankKeeper := suite.App.BankKeeper
			gammKeeper := suite.App.GAMMKeeper
			superfluidKeeper := suite.App.SuperfluidKeeper
			lockupKeeper := suite.App.LockupKeeper
			stakingKeeper := suite.App.StakingKeeper
			poolmanagerKeeper := suite.App.PoolManagerKeeper

			fullRangeCoins := sdk.NewCoins(defaultPoolAssets[0].Token, defaultPoolAssets[1].Token)

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

			// Determine the balancer pool's LP token denomination.
			balancerPoolDenom := gammtypes.GetPoolShareDenom(balancerPooId)

			// Register the balancer pool's LP token as a superfluid asset
			err = superfluidKeeper.AddNewSuperfluidAsset(ctx, types.SuperfluidAsset{
				Denom:     balancerPoolDenom,
				AssetType: types.SuperfluidAssetTypeLPShare,
			})
			suite.Require().NoError(err)

			// Note how much of the balancer pool's LP token the account that joined the pool has.
			balancerPoolShareOut := bankKeeper.GetBalance(ctx, poolJoinAcc, balancerPoolDenom)

			// Create a cl pool with the same underlying assets as the balancer pool.
			clPool := suite.PrepareCustomConcentratedPool(poolCreateAcc, defaultPoolAssets[0].Token.Denom, defaultPoolAssets[1].Token.Denom, 1, sdk.ZeroDec())
			clPoolId := clPool.GetId()

			// Add a gov sanctioned link between the balancer and concentrated liquidity pool.
			migrationRecord := gammtypes.MigrationRecords{BalancerToConcentratedPoolLinks: []gammtypes.BalancerToConcentratedPoolLink{
				{BalancerPoolId: balancerPooId, ClPoolId: clPoolId},
			}}
			gammKeeper.OverwriteMigrationRecords(ctx, migrationRecord)

			// The unbonding duration is the same as the staking module's unbonding duration.
			unbondingDuration := stakingKeeper.GetParams(ctx).UnbondingTime

			// Lock the LP tokens for the duration of the unbonding period.
			originalGammLockId := suite.LockTokens(poolJoinAcc, sdk.NewCoins(balancerPoolShareOut), unbondingDuration)

			// Superfluid delegate the balancer lock if the test case requires it.
			// Note the intermediary account that was created.
			balancerIntermediaryAcc := types.SuperfluidIntermediaryAccount{}
			if tc.superfluidDelegated {
				err = superfluidKeeper.SuperfluidDelegate(ctx, poolJoinAcc.String(), originalGammLockId, valAddr.String())
				suite.Require().NoError(err)
				intermediaryAccConnection := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, originalGammLockId)
				balancerIntermediaryAcc = superfluidKeeper.GetIntermediaryAccount(ctx, intermediaryAccConnection)
			}

			// Superfluid undelegate the lock if the test case requires it.
			if tc.superfluidUndelegating {
				err = superfluidKeeper.SuperfluidUndelegate(ctx, poolJoinAcc.String(), originalGammLockId)
				suite.Require().NoError(err)
			}

			// Unlock the balancer lock if the test case requires it.
			if tc.unlocking {
				// If lock was superfluid staked, we can't unlock via `BeginUnlock`,
				// we need to unlock lock via `SuperfluidUnbondLock`
				if tc.superfluidUndelegating {
					err = superfluidKeeper.SuperfluidUnbondLock(ctx, originalGammLockId, poolJoinAcc.String())
					suite.Require().NoError(err)
				} else {
					lock, err := lockupKeeper.GetLockByID(ctx, originalGammLockId)
					suite.Require().NoError(err)
					_, err = lockupKeeper.BeginUnlock(ctx, originalGammLockId, lock.Coins)
					suite.Require().NoError(err)
				}
			}

			balancerLock, err := lockupKeeper.GetLockByID(ctx, originalGammLockId)
			suite.Require().NoError(err)

			// Depending on the test case, we attempt to migrate a subset of the balancer LP tokens we originally created.
			coinsToMigrate := balancerPoolShareOut
			coinsToMigrate.Amount = coinsToMigrate.Amount.ToDec().Mul(tc.percentOfSharesToMigrate).RoundInt()

			if !tc.noInitialConcentratedSpotPrice {
				// Create a full range position in the concentrated liquidity pool.
				// This is to have a spot price and liquidity value to work off when migrating.
				suite.CreateFullRangePosition(clPool, fullRangeCoins)
			}

			// Register the CL full range LP tokens as a superfluid asset.
			clPoolDenom := cltypes.GetConcentratedLockupDenomFromPoolId(clPoolId)
			err = suite.App.SuperfluidKeeper.AddNewSuperfluidAsset(suite.Ctx, types.SuperfluidAsset{
				Denom:     clPoolDenom,
				AssetType: types.SuperfluidAssetTypeConcentratedShare,
			})
			if tc.noInitialConcentratedSpotPrice {
				// If we didn't create a full range position, we expect an error since no spot price exits to determine the osmo equivalent.
				suite.Require().Error(err)
				suite.Require().Equal(tc.expectedError.Error(), err.Error())
				return
			} else {
				suite.Require().NoError(err)
			}

			if tc.overwriteLockId {
				originalGammLockId = 5
			}

			if tc.multiAssetLock {
				originalGammLockId = suite.LockTokens(poolJoinAcc, sdk.NewCoins(balancerPoolShareOut, sdk.NewCoin("foo", sdk.NewInt(100))), unbondingDuration)
			}

			if tc.clLiquidityLock {
				clCoin := sdk.NewCoin(clPoolDenom, sdk.NewInt(100))
				suite.FundAcc(poolJoinAcc, sdk.NewCoins(clCoin))
				originalGammLockId = suite.LockTokens(poolJoinAcc, sdk.NewCoins(clCoin), unbondingDuration)
			}

			// Run the migration logic.
			positionId, amount0, amount1, _, _, poolIdLeaving, poolIdEntering, newGammLockId, concentratedLockId, err := superfluidKeeper.MigrateLockedPositionFromBalancerToConcentrated(ctx, poolJoinAcc, originalGammLockId, coinsToMigrate)
			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expectedError)
				return
			}
			suite.Require().NoError(err)
			suite.AssertEventEmitted(ctx, gammtypes.TypeEvtPoolExited, 1)

			newGammLock, err := lockupKeeper.GetLockByID(ctx, newGammLockId)
			if tc.percentOfSharesToMigrate.LT(sdk.OneDec()) {
				// If we migrated a subset of the balancer LP tokens, we expect the new gamm lock to have a the same end time.
				suite.Require().NoError(err)
				suite.Require().Equal(balancerLock.EndTime, newGammLock.EndTime)
			} else {
				// If we migrated all of the balancer LP tokens, we expect no new gamm lock to be created.
				suite.Require().Error(err)
				suite.Require().Nil(newGammLock)
			}

			// Check that the concentrated liquidity position now exists
			position, err := suite.App.ConcentratedLiquidityKeeper.GetPositionLiquidity(ctx, positionId)
			suite.Require().NoError(err)
			suite.Require().NotNil(position)

			// Expect the poolIdLeaving to be the balancer pool id
			// Expect the poolIdEntering to be the concentrated liquidity pool id
			suite.Require().Equal(balancerPooId, poolIdLeaving)
			suite.Require().Equal(clPoolId, poolIdEntering)

			// exitPool has rounding difference.
			// We test if correct amt has been exited and frozen by comparing with rounding tolerance.
			defaultErrorTolerance := osmomath.ErrTolerance{
				AdditiveTolerance: sdk.NewDec(2),
				RoundingDir:       osmomath.RoundDown,
			}
			suite.Require().Equal(0, defaultErrorTolerance.Compare(joinPoolAmt.AmountOf(defaultPoolAssets[0].Token.Denom).ToDec().Mul(tc.percentOfSharesToMigrate).RoundInt(), amount0))
			suite.Require().Equal(0, defaultErrorTolerance.Compare(joinPoolAmt.AmountOf(defaultPoolAssets[1].Token.Denom).ToDec().Mul(tc.percentOfSharesToMigrate).RoundInt(), amount1))

			// Check if the original gamm lock was deleted.
			_, err = lockupKeeper.GetLockByID(ctx, originalGammLockId)
			suite.Require().Error(err)

			// If we didn't migrate the entire gamm lock, we expect a new gamm lock to be created with the remaining lock time and coins associated with it.
			if tc.percentOfSharesToMigrate.LT(sdk.OneDec()) {
				// Check if the new gamm lock was created.
				newGammLock, err := lockupKeeper.GetLockByID(ctx, newGammLockId)
				suite.Require().NoError(err)
				// The new gamm lock should have the same owner and end time.
				// The new gamm lock should have the difference in coins between the original lock and the coins migrated.
				suite.Require().Equal(sdk.NewCoins(balancerPoolShareOut.Sub(coinsToMigrate)).String(), newGammLock.Coins.String())
				suite.Require().Equal(balancerLock.Owner, newGammLock.Owner)
				suite.Require().Equal(balancerLock.EndTime.String(), newGammLock.EndTime.String())
				// If original gamm lock was unlocking, the new gamm lock should also be unlocking.
				if balancerLock.IsUnlocking() {
					suite.Require().True(newGammLock.IsUnlocking())
				}
			} else {
				suite.Require().Equal(uint64(0), newGammLockId)
			}

			// Additional checks if the orignial gamm lock was superfluid staked.
			if tc.superfluidDelegated {
				// Check if migration deleted intermediary account connection.
				addr := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, originalGammLockId)
				suite.Require().Equal(addr.String(), "")

				// Check if migration deleted synthetic lockup.
				_, err = lockupKeeper.GetSyntheticLockup(ctx, originalGammLockId, keeper.StakingSyntheticDenom(balancerLock.Coins[0].Denom, valAddr.String()))
				suite.Require().Error(err)

				// If a new gamm position was not created and restaked, check if delegation has reduced from intermediary account.
				if tc.percentOfSharesToMigrate.Equal(sdk.OneDec()) {
					delegation, found := stakingKeeper.GetDelegation(ctx, balancerIntermediaryAcc.GetAccAddress(), valAddr)
					suite.Require().False(found, "expected no delegation, found delegation w/ %d shares", delegation.Shares)
				}
			}

			// Run slashing logic if the test case is superfluid staked or superfluid undelegating and check if the new and old locks are slashed.
			if tc.superfluidDelegated || tc.superfluidUndelegating {
				// Retrieve the concentrated lock and gamm lock prior to slashing.
				concentratedLockPreSlash, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, concentratedLockId)
				suite.Require().NoError(err)
				gammLockPreSlash, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, newGammLockId)
				if err != nil && newGammLockId != 0 {
					suite.Require().NoError(err)
				}

				// Slash the validator.
				slashFactor := sdk.NewDecWithPrec(5, 2)
				suite.App.SuperfluidKeeper.SlashLockupsForValidatorSlash(
					suite.Ctx,
					valAddr,
					suite.Ctx.BlockHeight(),
					slashFactor)

				// Retrieve the concentrated lock and gamm lock after slashing.
				concentratedLockPostSlash, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, concentratedLockId)
				suite.Require().NoError(err)
				gammLockPostSlash, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, newGammLockId)
				if err != nil && newGammLockId != 0 {
					suite.Require().NoError(err)
				}

				// Check if the concentrated lock was slashed.
				clDenom := cltypes.GetConcentratedLockupDenomFromPoolId(poolIdEntering)
				slashAmtCL := concentratedLockPreSlash.Coins.AmountOf(clDenom).ToDec().Mul(slashFactor).TruncateInt()
				suite.Require().Equal(concentratedLockPreSlash.Coins.AmountOf(clDenom).Sub(slashAmtCL).String(), concentratedLockPostSlash.Coins.AmountOf(clDenom).String())

				// Check if the gamm lock was slashed.
				// We only check if the gamm lock was slashed if the lock was not migrated entirely.
				// Otherwise, there would be no newly created gamm lock to check.
				if tc.percentOfSharesToMigrate.LT(sdk.OneDec()) {
					gammDenom := balancerLock.Coins[0].Denom
					slashAmtGamm := gammLockPreSlash.Coins.AmountOf(gammDenom).ToDec().Mul(slashFactor).TruncateInt()
					suite.Require().Equal(gammLockPreSlash.Coins.AmountOf(gammDenom).Sub(slashAmtGamm).String(), gammLockPostSlash.Coins.AmountOf(gammDenom).String())
				}
			}
		})
	}
}
