package keeper_test

import (
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/types"
)

// We test migrating in the following circumstances:
// 1. Migrating lock that is not superfluid delegated, not unlocking.
// 2. Migrating lock that is not superfluid delegated, unlocking.
// 3. Migrating lock that is superfluid delegated, not unlocking.
// 4. Migrating lock that is superfluid undelegating, not unlocking.
// 5. Migrating lock that is superfluid undelegating, unlocking.
func (suite *KeeperTestSuite) TestRouteLockedBalancerToConcentratedMigration() {
	defaultJoinTime := suite.Ctx.BlockTime()
	type sendTest struct {
		superfluidDelegated      bool
		superfluidUndelegating   bool
		unlocking                bool
		overwriteLockId          bool
		percentOfSharesToMigrate sdk.Dec
		minExitCoins             sdk.Coins
		expectedError            error
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
			expectedError:            errorsmod.Wrap(lockuptypes.ErrLockupNotFound, fmt.Sprintf("lock with ID %d does not exist", 5)),
		},
		"error: lock that is not superfluid delegated, not unlocking, min exit coins more than being exitted": {
			// migrateNonSuperfluidLockBalancerToConcentrated
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.9"),
			minExitCoins:             sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(5000)), sdk.NewCoin("stake", sdk.NewInt(5000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
		"error: lock that is not superfluid delegated, unlocking, min exit coins more than being exitted": {
			// migrateNonSuperfluidLockBalancerToConcentrated
			unlocking:                true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.6"),
			minExitCoins:             sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(4000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
		"error: lock that is superfluid delegated, not unlocking (full shares), min exit coins more than being exitted": {
			// migrateSuperfluidBondedBalancerToConcentrated
			superfluidDelegated:      true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			minExitCoins:             sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
		"error: lock that is superfluid delegated, not unlocking (partial shares, min exit coins more than being exitted": {
			// migrateSuperfluidBondedBalancerToConcentrated
			superfluidDelegated:      true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.5"),
			minExitCoins:             sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(3000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
		"error: lock that is superfluid undelegating, not unlocking, min exit coins more than being exitted": {
			// migrateSuperfluidUnbondingBalancerToConcentrated
			superfluidDelegated:      true,
			superfluidUndelegating:   true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.5"),
			minExitCoins:             sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(40000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
		"lock that is superfluid undelegating, unlocking, min exit coins more than being exitted": {
			// migrateSuperfluidUnbondingBalancerToConcentrated
			superfluidDelegated:      true,
			superfluidUndelegating:   true,
			unlocking:                true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.3"),
			minExitCoins:             sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(40000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
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

			// We bundle all migration setup into a single function to avoid repeating the same code for each test case.
			joinPoolAmt, balancerIntermediaryAcc, balancerLock, _, poolJoinAcc, balancerPooId, clPoolId, balancerPoolShareOut, valAddr := suite.SetupMigrationTest(ctx, tc.superfluidDelegated, tc.superfluidUndelegating, tc.unlocking, tc.percentOfSharesToMigrate)
			originalGammLockId := balancerLock.GetID()

			// Depending on the test case, we attempt to migrate a subset of the balancer LP tokens we originally created.
			coinsToMigrate := balancerPoolShareOut
			coinsToMigrate.Amount = coinsToMigrate.Amount.ToDec().Mul(tc.percentOfSharesToMigrate).RoundInt()

			// Modify migration inputs if necessary
			if tc.overwriteLockId {
				originalGammLockId = originalGammLockId + 1
			}

			// Run the migration logic.
			positionId, amount0, amount1, _, _, poolIdLeaving, poolIdEntering, newGammLockId, concentratedLockId, err := superfluidKeeper.RouteLockedBalancerToConcentratedMigration(ctx, poolJoinAcc, originalGammLockId, coinsToMigrate, tc.minExitCoins)
			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expectedError)
				return
			}
			suite.Require().NoError(err)
			suite.AssertEventEmitted(ctx, gammtypes.TypeEvtPoolExited, 1)

			suite.ValidateMigrateResult(
				ctx,
				positionId, balancerPooId, poolIdLeaving, clPoolId, poolIdEntering, originalGammLockId, newGammLockId,
				tc.percentOfSharesToMigrate,
				*balancerLock,
				joinPoolAmt,
				balancerPoolShareOut, coinsToMigrate,
				amount0, amount1,
			)

			// Additional checks if the original gamm lock was superfluid staked.
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
			slashExpected := tc.superfluidDelegated || tc.superfluidUndelegating
			suite.SlashAndValidateResult(ctx, newGammLockId, concentratedLockId, poolIdEntering, tc.percentOfSharesToMigrate, valAddr, *balancerLock, slashExpected)
		})
	}
}

func (suite *KeeperTestSuite) TestMigrateSuperfluidBondedBalancerToConcentrated() {
	defaultJoinTime := suite.Ctx.BlockTime()
	type sendTest struct {
		overwriteValidatorAddress bool
		overwriteLockId           bool
		overwriteShares           bool
		overwritePool             bool
		percentOfSharesToMigrate  sdk.Dec
		tokenOutMins              sdk.Coins
		expectedError             error
	}
	testCases := map[string]sendTest{
		"lock that is superfluid delegated, not unlocking (full shares)": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
		},
		"lock that is superfluid delegated, not unlocking (partial shares)": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.5"),
		},
		"error: invalid validator address": {
			overwriteValidatorAddress: true,
			percentOfSharesToMigrate:  sdk.MustNewDecFromStr("1"),
			expectedError:             fmt.Errorf("decoding bech32 failed: invalid checksum"),
		},
		"error: non-existent lock ID": {
			overwriteLockId:          true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			expectedError:            lockuptypes.ErrLockupNotFound,
		},
		"error: attempt to migrate more shares than the lock has": {
			overwriteShares:          true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			expectedError:            fmt.Errorf("shares to migrate must be less than or equal to shares in lock"),
		},
		"error: pool has more than two assets": {
			overwritePool:            true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			expectedError:            fmt.Errorf("Balancer pool must have exactly two tokens"),
		},
		"error: lock that is superfluid delegated, not unlocking (full shares), token out mins is more than exit coins": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			tokenOutMins:             sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(100000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
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

			// We bundle all migration setup into a single function to avoid repeating the same code for each test case.
			joinPoolAmt, balancerIntermediaryAcc, balancerLock, _, poolJoinAcc, balancerPooId, clPoolId, balancerPoolShareOut, valAddr := suite.SetupMigrationTest(ctx, true, false, false, tc.percentOfSharesToMigrate)
			originalGammLockId := balancerLock.GetID()

			// Depending on the test case, we attempt to migrate a subset of the balancer LP tokens we originally created.
			coinsToMigrate := balancerPoolShareOut
			coinsToMigrate.Amount = coinsToMigrate.Amount.ToDec().Mul(tc.percentOfSharesToMigrate).RoundInt()

			// PrepareMigration is called via the migration message router and is always run prior to the migration itself
			poolIdLeaving, poolIdEntering, concentratedPool, preMigrationLock, remainingLockTime, synthLockBeforeMigration, _, _, err := superfluidKeeper.PrepareMigration(ctx, poolJoinAcc, originalGammLockId, coinsToMigrate)
			suite.Require().NoError(err)

			// Modify migration inputs if necessary

			if tc.overwriteValidatorAddress {
				synthDenomParts := strings.Split(synthLockBeforeMigration[0].SynthDenom, "/")
				synthDenomParts[4] = "osmovaloper1n69ghlk6404gzxtmtq0w7ma59n9vd9ed9dplg" // invalid, too short
				newSynthDenom := strings.Join(synthDenomParts, "/")
				synthLockBeforeMigration[0].SynthDenom = newSynthDenom
			}

			if tc.overwriteLockId {
				originalGammLockId = originalGammLockId + 1
			}

			if tc.overwriteShares {
				coinsToMigrate.Amount = preMigrationLock.Coins[0].Amount.Add(sdk.NewInt(1))
			}

			if tc.overwritePool {
				multiCoinBalancerPoolId := suite.PrepareBalancerPool()
				poolIdLeaving = multiCoinBalancerPoolId
				shareAmt := sdk.MustNewDecFromStr("50000000000000000000").TruncateInt()
				newShares := sdk.NewCoin(fmt.Sprintf("gamm/pool/%d", multiCoinBalancerPoolId), shareAmt)
				suite.FundAcc(poolJoinAcc, sdk.NewCoins(newShares))
			}

			// System under test.
			positionId, amount0, amount1, _, _, newGammLockId, concentratedLockId, err := superfluidKeeper.MigrateSuperfluidBondedBalancerToConcentrated(ctx, poolJoinAcc, poolIdLeaving, poolIdEntering, preMigrationLock, originalGammLockId, coinsToMigrate, synthLockBeforeMigration[0].SynthDenom, concentratedPool, remainingLockTime, tc.tokenOutMins)
			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}
			suite.Require().NoError(err)
			suite.AssertEventEmitted(ctx, gammtypes.TypeEvtPoolExited, 1)

			suite.ValidateMigrateResult(
				ctx,
				positionId, balancerPooId, poolIdLeaving, clPoolId, poolIdEntering, originalGammLockId, newGammLockId,
				tc.percentOfSharesToMigrate,
				*balancerLock,
				joinPoolAmt,
				balancerPoolShareOut, coinsToMigrate,
				amount0, amount1,
			)

			// Check if migration deleted intermediary account connection.
			originalGammIntermediaryAccount := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, originalGammLockId)
			suite.Require().Equal(originalGammIntermediaryAccount.String(), "")

			// Check if migration deleted synthetic lockup.
			_, err = lockupKeeper.GetSyntheticLockup(ctx, originalGammLockId, keeper.StakingSyntheticDenom(balancerLock.Coins[0].Denom, valAddr.String()))
			suite.Require().Error(err)

			// Check if the new intermediary account connection was created.
			newConcentratedIntermediaryAccount := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, concentratedLockId)
			suite.Require().NotEqual(newConcentratedIntermediaryAccount.String(), "")

			// Check if the new synthetic bonded lockup was created.
			concentratedLock, err := lockupKeeper.GetLockByID(ctx, concentratedLockId)
			suite.Require().NoError(err)
			_, err = lockupKeeper.GetSyntheticLockup(ctx, concentratedLockId, keeper.StakingSyntheticDenom(concentratedLock.Coins[0].Denom, valAddr.String()))
			suite.Require().NoError(err)

			// If a new gamm position was not created and restaked, check if delegation has reduced from intermediary account.
			if tc.percentOfSharesToMigrate.Equal(sdk.OneDec()) {
				delegation, found := stakingKeeper.GetDelegation(ctx, balancerIntermediaryAcc.GetAccAddress(), valAddr)
				suite.Require().False(found, "expected no delegation, found delegation w/ %d shares", delegation.Shares)
			}

			// Run slashing logic and check if the new and old locks are slashed.
			suite.SlashAndValidateResult(ctx, newGammLockId, concentratedLockId, poolIdEntering, tc.percentOfSharesToMigrate, valAddr, *balancerLock, true)
		})
	}
}

func (suite *KeeperTestSuite) TestMigrateSuperfluidUnbondingBalancerToConcentrated() {
	defaultJoinTime := suite.Ctx.BlockTime()
	type sendTest struct {
		unlocking                 bool
		overwriteValidatorAddress bool
		overwritePreMigrationLock bool
		overwriteShares           bool
		overwritePool             bool
		percentOfSharesToMigrate  sdk.Dec
		tokenOutMins              sdk.Coins
		expectedError             error
	}
	testCases := map[string]sendTest{
		"lock that is superfluid undelegating, not unlocking (full shares)": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
		},
		"lock that is superfluid undelegating, not unlocking (partial shares)": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.5"),
		},
		"lock that is superfluid undelegating, unlocking (full shares)": {
			unlocking:                true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
		},
		"lock that is superfluid undelegating, unlocking (partial shares)": {
			unlocking:                true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.3"),
		},
		"error: invalid validator address": {
			overwriteValidatorAddress: true,
			percentOfSharesToMigrate:  sdk.MustNewDecFromStr("1"),
			expectedError:             fmt.Errorf("decoding bech32 failed: invalid checksum"),
		},
		"error: non-existent pre migration lock": {
			overwritePreMigrationLock: true,
			percentOfSharesToMigrate:  sdk.MustNewDecFromStr("1"),
			expectedError:             lockuptypes.ErrLockupNotFound,
		},
		"error: attempt to migrate more shares than the lock has": {
			overwriteShares:          true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			expectedError:            fmt.Errorf("shares to migrate must be less than or equal to shares in lock"),
		},
		"error: pool has more than two assets": {
			overwritePool:            true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			expectedError:            fmt.Errorf("Balancer pool must have exactly two tokens"),
		},
		"error: lock that is superfluid undelegating, not unlocking (full shares), token out mins is more than exit coins": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			tokenOutMins:             sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(100000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			suite.SetupTest()
			suite.Ctx = suite.Ctx.WithBlockTime(defaultJoinTime)
			ctx := suite.Ctx
			superfluidKeeper := suite.App.SuperfluidKeeper
			lockupKeeper := suite.App.LockupKeeper

			// We bundle all migration setup into a single function to avoid repeating the same code for each test case.
			joinPoolAmt, _, balancerLock, _, poolJoinAcc, balancerPooId, clPoolId, balancerPoolShareOut, valAddr := suite.SetupMigrationTest(ctx, true, true, tc.unlocking, tc.percentOfSharesToMigrate)
			originalGammLockId := balancerLock.GetID()

			// Depending on the test case, we attempt to migrate a subset of the balancer LP tokens we originally created.
			coinsToMigrate := balancerPoolShareOut
			coinsToMigrate.Amount = coinsToMigrate.Amount.ToDec().Mul(tc.percentOfSharesToMigrate).RoundInt()

			// PrepareMigration is called via the migration message router and is always run prior to the migration itself
			poolIdLeaving, poolIdEntering, concentratedPool, preMigrationLock, remainingLockTime, synthLockBeforeMigration, _, _, err := superfluidKeeper.PrepareMigration(ctx, poolJoinAcc, originalGammLockId, coinsToMigrate)
			suite.Require().NoError(err)

			// Modify migration inputs if necessary

			if tc.overwriteValidatorAddress {
				synthDenomParts := strings.Split(synthLockBeforeMigration[0].SynthDenom, "/")
				synthDenomParts[4] = "osmovaloper1n69ghlk6404gzxtmtq0w7ma59n9vd9ed9dplg" // invalid, too short
				newSynthDenom := strings.Join(synthDenomParts, "/")
				synthLockBeforeMigration[0].SynthDenom = newSynthDenom
			}

			if tc.overwritePreMigrationLock {
				preMigrationLock.ID = preMigrationLock.ID + 1
			}

			if tc.overwriteShares {
				coinsToMigrate.Amount = preMigrationLock.Coins[0].Amount.Add(sdk.NewInt(1))
			}

			if tc.overwritePool {
				multiCoinBalancerPoolId := suite.PrepareBalancerPool()
				poolIdLeaving = multiCoinBalancerPoolId
				shareAmt := sdk.MustNewDecFromStr("50000000000000000000").TruncateInt()
				newShares := sdk.NewCoin(fmt.Sprintf("gamm/pool/%d", multiCoinBalancerPoolId), shareAmt)
				suite.FundAcc(poolJoinAcc, sdk.NewCoins(newShares))
			}

			// System under test.
			positionId, amount0, amount1, _, _, newGammLockId, concentratedLockId, err := superfluidKeeper.MigrateSuperfluidUnbondingBalancerToConcentrated(ctx, poolJoinAcc, poolIdLeaving, poolIdEntering, preMigrationLock, coinsToMigrate, synthLockBeforeMigration[0].SynthDenom, concentratedPool, remainingLockTime, tc.tokenOutMins)
			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}
			suite.Require().NoError(err)
			suite.AssertEventEmitted(ctx, gammtypes.TypeEvtPoolExited, 1)

			suite.ValidateMigrateResult(
				ctx,
				positionId, balancerPooId, poolIdLeaving, clPoolId, poolIdEntering, originalGammLockId, newGammLockId,
				tc.percentOfSharesToMigrate,
				*balancerLock,
				joinPoolAmt,
				balancerPoolShareOut, coinsToMigrate,
				amount0, amount1,
			)

			// Check if the new synthetic unbonding lockup was created.
			concentratedLock, err := lockupKeeper.GetLockByID(ctx, concentratedLockId)
			suite.Require().NoError(err)
			_, err = lockupKeeper.GetSyntheticLockup(ctx, concentratedLockId, keeper.UnstakingSyntheticDenom(concentratedLock.Coins[0].Denom, valAddr.String()))
			suite.Require().NoError(err)

			// Run slashing logic and check if the new and old locks are slashed.
			suite.SlashAndValidateResult(ctx, newGammLockId, concentratedLockId, poolIdEntering, tc.percentOfSharesToMigrate, valAddr, *balancerLock, true)
		})
	}
}

func (suite *KeeperTestSuite) TestMigrateNonSuperfluidLockBalancerToConcentrated() {
	defaultJoinTime := suite.Ctx.BlockTime()
	type sendTest struct {
		unlocking                 bool
		overwritePreMigrationLock bool
		overwriteShares           bool
		overwritePool             bool
		percentOfSharesToMigrate  sdk.Dec
		tokenOutMins              sdk.Coins
		expectedError             error
	}
	testCases := map[string]sendTest{
		"lock that is not superfluid delegated, not unlocking (full shares)": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
		},
		"lock that is not superfluid delegated, not unlocking (partial shares)": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.9"),
		},
		"lock that is not superfluid delegated, unlocking (full shares)": {
			unlocking:                true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
		},
		"lock that is not superfluid delegated, unlocking (partial shares)": {
			unlocking:                true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.6"),
		},
		"error: non-existent pre migration lock": {
			overwritePreMigrationLock: true,
			percentOfSharesToMigrate:  sdk.MustNewDecFromStr("1"),
			expectedError:             lockuptypes.ErrLockupNotFound,
		},
		"error: attempt to migrate more shares than the lock has": {
			overwriteShares:          true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			expectedError:            fmt.Errorf("shares to migrate must be less than or equal to shares in lock"),
		},
		"error: pool has more than two assets": {
			overwritePool:            true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			expectedError:            fmt.Errorf("Balancer pool must have exactly two tokens"),
		},
		"error: lock that is not superfluid delegated, not unlocking (full shares), token out mins is more than exit coins": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			tokenOutMins:             sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			suite.SetupTest()
			suite.Ctx = suite.Ctx.WithBlockTime(defaultJoinTime)
			ctx := suite.Ctx
			superfluidKeeper := suite.App.SuperfluidKeeper

			// We bundle all migration setup into a single function to avoid repeating the same code for each test case.
			joinPoolAmt, _, balancerLock, _, poolJoinAcc, balancerPooId, clPoolId, balancerPoolShareOut, valAddr := suite.SetupMigrationTest(ctx, false, false, tc.unlocking, tc.percentOfSharesToMigrate)
			originalGammLockId := balancerLock.GetID()

			// Depending on the test case, we attempt to migrate a subset of the balancer LP tokens we originally created.
			coinsToMigrate := balancerPoolShareOut
			coinsToMigrate.Amount = coinsToMigrate.Amount.ToDec().Mul(tc.percentOfSharesToMigrate).RoundInt()

			// PrepareMigration is called via the migration message router and is always run prior to the migration itself
			poolIdLeaving, poolIdEntering, concentratedPool, preMigrationLock, remainingLockTime, _, _, _, err := superfluidKeeper.PrepareMigration(ctx, poolJoinAcc, originalGammLockId, coinsToMigrate)
			suite.Require().NoError(err)

			// Modify migration inputs if necessary

			if tc.overwritePreMigrationLock {
				preMigrationLock.ID = preMigrationLock.ID + 1
			}

			if tc.overwriteShares {
				coinsToMigrate.Amount = preMigrationLock.Coins[0].Amount.Add(sdk.NewInt(1))
			}

			if tc.overwritePool {
				multiCoinBalancerPoolId := suite.PrepareBalancerPool()
				poolIdLeaving = multiCoinBalancerPoolId
				shareAmt := sdk.MustNewDecFromStr("50000000000000000000").TruncateInt()
				newShares := sdk.NewCoin(fmt.Sprintf("gamm/pool/%d", multiCoinBalancerPoolId), shareAmt)
				suite.FundAcc(poolJoinAcc, sdk.NewCoins(newShares))
			}

			// System under test.
			positionId, amount0, amount1, _, _, newGammLockId, concentratedLockId, err := superfluidKeeper.MigrateNonSuperfluidLockBalancerToConcentrated(ctx, poolJoinAcc, poolIdLeaving, poolIdEntering, preMigrationLock, coinsToMigrate, concentratedPool, remainingLockTime, tc.tokenOutMins)
			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}
			suite.Require().NoError(err)
			suite.AssertEventEmitted(ctx, gammtypes.TypeEvtPoolExited, 1)

			suite.ValidateMigrateResult(
				ctx,
				positionId, balancerPooId, poolIdLeaving, clPoolId, poolIdEntering, originalGammLockId, newGammLockId,
				tc.percentOfSharesToMigrate,
				*balancerLock,
				joinPoolAmt,
				balancerPoolShareOut, coinsToMigrate,
				amount0, amount1,
			)

			// Run slashing logic and check if the new and old locks are not slashed.
			suite.SlashAndValidateResult(ctx, newGammLockId, concentratedLockId, poolIdEntering, tc.percentOfSharesToMigrate, valAddr, *balancerLock, false)
		})
	}
}

func (suite *KeeperTestSuite) TestValidateSharesToMigrateUnlockAndExitBalancerPool() {
	defaultJoinTime := suite.Ctx.BlockTime()
	type sendTest struct {
		overwritePreMigrationLock bool
		overwriteShares           bool
		overwritePool             bool
		overwritePoolId           bool
		percentOfSharesToMigrate  sdk.Dec
		tokenOutMins              sdk.Coins
		expectedError             error
	}
	testCases := map[string]sendTest{
		"happy path (full shares)": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
		},
		"happy path (partial shares)": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.4"),
		},
		"error: lock does not exist": {
			percentOfSharesToMigrate:  sdk.MustNewDecFromStr("1"),
			overwritePreMigrationLock: true,
			expectedError:             lockuptypes.ErrLockupNotFound,
		},
		"error: attempt to migrate more than lock has": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			overwriteShares:          true,
			expectedError:            fmt.Errorf("shares to migrate must be less than or equal to shares in lock"),
		},
		"error: attempt to leave a pool that does not exist": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			overwritePoolId:          true,
			expectedError:            fmt.Errorf("pool with ID %d does not exist", 2),
		},
		"error: attempt to leave a pool that has more than two denoms": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			overwritePool:            true,
			expectedError:            fmt.Errorf("Balancer pool must have exactly two tokens"),
		},
		"error: happy path (full shares), token out mins is more than exit coins": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			tokenOutMins:             sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(100000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			suite.SetupTest()
			suite.Ctx = suite.Ctx.WithBlockTime(defaultJoinTime)
			ctx := suite.Ctx
			superfluidKeeper := suite.App.SuperfluidKeeper
			lockupKeeper := suite.App.LockupKeeper
			poolmanagerKeeper := suite.App.PoolManagerKeeper
			bankKeeper := suite.App.BankKeeper
			gammKeeper := suite.App.GAMMKeeper
			stakingKeeper := suite.App.StakingKeeper

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

			// Create a balancer pool of "stake" and "foo".
			msg := balancer.NewMsgCreateBalancerPool(poolCreateAcc, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDec(0),
			}, defaultPoolAssets, defaultFutureGovernor)
			balancerPooId, err := poolmanagerKeeper.CreatePool(ctx, msg)
			suite.Require().NoError(err)

			// Join the balancer pool.
			tokensIn, _, err := gammKeeper.JoinPoolNoSwap(ctx, poolJoinAcc, balancerPooId, gammtypes.OneShare.MulRaw(50), sdk.Coins{})
			suite.Require().NoError(err)

			// Determine the balancer pool's LP token denomination.
			balancerPoolDenom := gammtypes.GetPoolShareDenom(balancerPooId)

			// Note how much of the balancer pool's LP token the account that joined the pool has.
			balancerPoolShareOut := bankKeeper.GetBalance(ctx, poolJoinAcc, balancerPoolDenom)

			sharesToMigrate := balancerPoolShareOut.Amount.ToDec().Mul(tc.percentOfSharesToMigrate).TruncateInt()
			coinsToMigrate := sdk.NewCoin(balancerPoolDenom, sharesToMigrate)

			// The unbonding duration is the same as the staking module's unbonding duration.
			unbondingDuration := stakingKeeper.GetParams(ctx).UnbondingTime

			// Lock the LP tokens for the duration of the unbonding period.
			originalGammLockId := suite.LockTokens(poolJoinAcc, sdk.NewCoins(coinsToMigrate), unbondingDuration)

			lock, err := lockupKeeper.GetLockByID(ctx, originalGammLockId)
			suite.Require().NoError(err)

			if tc.overwritePreMigrationLock {
				lock.ID = lock.ID + 1
			}

			if tc.overwriteShares {
				coinsToMigrate.Amount = lock.Coins[0].Amount.Add(sdk.NewInt(1))
			}

			if tc.overwritePool {
				multiCoinBalancerPoolId := suite.PrepareBalancerPool()
				balancerPooId = multiCoinBalancerPoolId
				shareAmt := sdk.MustNewDecFromStr("50000000000000000000").TruncateInt()
				newShares := sdk.NewCoin(fmt.Sprintf("gamm/pool/%d", multiCoinBalancerPoolId), shareAmt)
				suite.FundAcc(poolJoinAcc, sdk.NewCoins(newShares))
			}

			if tc.overwritePoolId {
				balancerPooId = balancerPooId + 1
			}

			// System under test
			exitCoins, err := superfluidKeeper.ValidateSharesToMigrateUnlockAndExitBalancerPool(ctx, poolJoinAcc, balancerPooId, lock, coinsToMigrate, tc.tokenOutMins)
			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}
			suite.Require().NoError(err)

			defaultErrorTolerance := osmomath.ErrTolerance{
				AdditiveTolerance: sdk.NewDec(1),
				RoundingDir:       osmomath.RoundDown,
			}

			for _, coin := range exitCoins {
				// Check that the exit coin is the same amount that we joined with (with one unit rounding down)
				suite.Require().Equal(0, defaultErrorTolerance.Compare(tokensIn.AmountOf(coin.Denom).ToDec().Mul(tc.percentOfSharesToMigrate).RoundInt(), coin.Amount))
			}
		})
	}
}

func (suite *KeeperTestSuite) SetupMigrationTest(ctx sdk.Context, superfluidDelegated, superfluidUndelegating, unlocking bool, percentOfSharesToMigrate sdk.Dec) (joinPoolAmt sdk.Coins, balancerIntermediaryAcc types.SuperfluidIntermediaryAccount, balancerLock *lockuptypes.PeriodLock, poolCreateAcc, poolJoinAcc sdk.AccAddress, balancerPooId, clPoolId uint64, balancerPoolShareOut sdk.Coin, valAddr sdk.ValAddress) {
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
	poolCreateAcc = delAddrs[0]
	poolJoinAcc = delAddrs[1]
	for _, acc := range delAddrs {
		err := simapp.FundAccount(bankKeeper, ctx, acc, defaultAcctFunds)
		suite.Require().NoError(err)
	}

	// Set up a single validator.
	valAddr = suite.SetupValidator(stakingtypes.Bonded)

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
	joinPoolAmt, _ = balanceBeforeJoin.SafeSub(balanceAfterJoin)

	// Determine the balancer pool's LP token denomination.
	balancerPoolDenom := gammtypes.GetPoolShareDenom(balancerPooId)

	// Register the balancer pool's LP token as a superfluid asset
	err = superfluidKeeper.AddNewSuperfluidAsset(ctx, types.SuperfluidAsset{
		Denom:     balancerPoolDenom,
		AssetType: types.SuperfluidAssetTypeLPShare,
	})
	suite.Require().NoError(err)

	// Note how much of the balancer pool's LP token the account that joined the pool has.
	balancerPoolShareOut = bankKeeper.GetBalance(ctx, poolJoinAcc, balancerPoolDenom)

	// Create a cl pool with the same underlying assets as the balancer pool.
	clPool := suite.PrepareCustomConcentratedPool(poolCreateAcc, defaultPoolAssets[0].Token.Denom, defaultPoolAssets[1].Token.Denom, 1, sdk.ZeroDec())
	clPoolId = clPool.GetId()

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
	if superfluidDelegated {
		err = superfluidKeeper.SuperfluidDelegate(ctx, poolJoinAcc.String(), originalGammLockId, valAddr.String())
		suite.Require().NoError(err)
		intermediaryAccConnection := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, originalGammLockId)
		balancerIntermediaryAcc = superfluidKeeper.GetIntermediaryAccount(ctx, intermediaryAccConnection)
	}

	// Superfluid undelegate the lock if the test case requires it.
	if superfluidUndelegating {
		err = superfluidKeeper.SuperfluidUndelegate(ctx, poolJoinAcc.String(), originalGammLockId)
		suite.Require().NoError(err)
	}

	// Unlock the balancer lock if the test case requires it.
	if unlocking {
		// If lock was superfluid staked, we can't unlock via `BeginUnlock`,
		// we need to unlock lock via `SuperfluidUnbondLock`
		if superfluidUndelegating {
			err = superfluidKeeper.SuperfluidUnbondLock(ctx, originalGammLockId, poolJoinAcc.String())
			suite.Require().NoError(err)
		} else {
			lock, err := lockupKeeper.GetLockByID(ctx, originalGammLockId)
			suite.Require().NoError(err)
			_, err = lockupKeeper.BeginUnlock(ctx, originalGammLockId, lock.Coins)
			suite.Require().NoError(err)
		}
	}

	balancerLock, err = lockupKeeper.GetLockByID(ctx, originalGammLockId)
	suite.Require().NoError(err)

	// Create a full range position in the concentrated liquidity pool.
	// This is to have a spot price and liquidity value to work off when migrating.
	suite.CreateFullRangePosition(clPool, fullRangeCoins)

	// Register the CL full range LP tokens as a superfluid asset.
	clPoolDenom := cltypes.GetConcentratedLockupDenomFromPoolId(clPoolId)
	err = suite.App.SuperfluidKeeper.AddNewSuperfluidAsset(suite.Ctx, types.SuperfluidAsset{
		Denom:     clPoolDenom,
		AssetType: types.SuperfluidAssetTypeConcentratedShare,
	})

	suite.Require().NoError(err)
	return joinPoolAmt, balancerIntermediaryAcc, balancerLock, poolCreateAcc, poolJoinAcc, balancerPooId, clPoolId, balancerPoolShareOut, valAddr
}

func (suite *KeeperTestSuite) SlashAndValidateResult(ctx sdk.Context, newGammLockId, concentratedLockId, poolIdEntering uint64, percentOfSharesToMigrate sdk.Dec, valAddr sdk.ValAddress, balancerLock lockuptypes.PeriodLock, expectSlash bool) {
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
	if !expectSlash {
		slashAmtCL = sdk.ZeroInt()
	}
	suite.Require().Equal(concentratedLockPreSlash.Coins.AmountOf(clDenom).Sub(slashAmtCL).String(), concentratedLockPostSlash.Coins.AmountOf(clDenom).String())

	// Check if the gamm lock was slashed.
	// We only check if the gamm lock was slashed if the lock was not migrated entirely.
	// Otherwise, there would be no newly created gamm lock to check.
	if percentOfSharesToMigrate.LT(sdk.OneDec()) {
		gammDenom := balancerLock.Coins[0].Denom
		slashAmtGamm := gammLockPreSlash.Coins.AmountOf(gammDenom).ToDec().Mul(slashFactor).TruncateInt()
		if !expectSlash {
			slashAmtGamm = sdk.ZeroInt()
		}
		suite.Require().Equal(gammLockPreSlash.Coins.AmountOf(gammDenom).Sub(slashAmtGamm).String(), gammLockPostSlash.Coins.AmountOf(gammDenom).String())
	}
}

func (suite *KeeperTestSuite) ValidateMigrateResult(
	ctx sdk.Context,
	positionId, balancerPooId, poolIdLeaving, clPoolId, poolIdEntering, originalGammLockId, newGammLockId uint64,
	percentOfSharesToMigrate sdk.Dec,
	balancerLock lockuptypes.PeriodLock,
	joinPoolAmt sdk.Coins,
	balancerPoolShareOut, coinsToMigrate sdk.Coin,
	amount0, amount1 sdk.Int,
) {
	lockupKeeper := suite.App.LockupKeeper

	newGammLock, err := lockupKeeper.GetLockByID(ctx, newGammLockId)
	if percentOfSharesToMigrate.LT(sdk.OneDec()) {
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
	suite.Require().Equal(0, defaultErrorTolerance.Compare(joinPoolAmt.AmountOf(defaultPoolAssets[0].Token.Denom).ToDec().Mul(percentOfSharesToMigrate).RoundInt(), amount0))
	suite.Require().Equal(0, defaultErrorTolerance.Compare(joinPoolAmt.AmountOf(defaultPoolAssets[1].Token.Denom).ToDec().Mul(percentOfSharesToMigrate).RoundInt(), amount1))

	// Check if the original gamm lock was deleted.
	_, err = lockupKeeper.GetLockByID(ctx, originalGammLockId)
	suite.Require().Error(err)

	// If we didn't migrate the entire gamm lock, we expect a new gamm lock to be created with the remaining lock time and coins associated with it.
	if percentOfSharesToMigrate.LT(sdk.OneDec()) {
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
}
