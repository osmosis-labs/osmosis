package keeper_test

import (
	"fmt"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	cltypes "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v16/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v16/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v16/x/superfluid/types"
)

// We test migrating in the following circumstances:
// 1. Migrating lock that is not superfluid delegated, not unlocking.
// 2. Migrating lock that is not superfluid delegated, unlocking.
// 3. Migrating lock that is superfluid delegated, not unlocking.
// 4. Migrating lock that is superfluid undelegating, not unlocking.
// 5. Migrating lock that is superfluid undelegating, unlocking.
func (s *KeeperTestSuite) TestRouteLockedBalancerToConcentratedMigration() {
	defaultJoinTime := s.Ctx.BlockTime()
	type sendTest struct {
		superfluidDelegated      bool
		superfluidUndelegating   bool
		unlocking                bool
		noLock                   bool
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
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.4"),
		},
		"lock that is superfluid undelegating, not unlocking (full shares)": {
			// migrateSuperfluidUnbondingBalancerToConcentrated
			superfluidDelegated:      true,
			superfluidUndelegating:   true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
		},
		"lock that is superfluid undelegating, not unlocking (partial shares)": {
			// migrateSuperfluidUnbondingBalancerToConcentrated
			superfluidDelegated:      true,
			superfluidUndelegating:   true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.4"),
		},
		"lock that is superfluid undelegating, unlocking (full shares)": {
			// migrateSuperfluidUnbondingBalancerToConcentrated
			superfluidDelegated:      true,
			superfluidUndelegating:   true,
			unlocking:                true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
		},
		"lock that is superfluid undelegating, unlocking (partial shares)": {
			// migrateSuperfluidUnbondingBalancerToConcentrated
			superfluidDelegated:      true,
			superfluidUndelegating:   true,
			unlocking:                true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.3"),
		},
		"no lock (partial shares)": {
			// MigrateUnlockedPositionFromBalancerToConcentrated
			noLock:                   true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.3"),
		},
		"no lock (full shares)": {
			// MigrateUnlockedPositionFromBalancerToConcentrated
			noLock:                   true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
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
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)
			ctx := s.Ctx
			superfluidKeeper := s.App.SuperfluidKeeper
			lockupKeeper := s.App.LockupKeeper
			stakingKeeper := s.App.StakingKeeper

			// We bundle all migration setup into a single function to avoid repeating the same code for each test case.
			joinPoolAmt, balancerIntermediaryAcc, balancerLock, _, poolJoinAcc, balancerPooId, clPoolId, balancerPoolShareOut, valAddr := s.SetupMigrationTest(ctx, tc.superfluidDelegated, tc.superfluidUndelegating, tc.unlocking, tc.noLock, tc.percentOfSharesToMigrate)
			originalGammLockId := balancerLock.GetID()

			// Depending on the test case, we attempt to migrate a subset of the balancer LP tokens we originally created.
			coinsToMigrate := balancerPoolShareOut
			coinsToMigrate.Amount = coinsToMigrate.Amount.ToDec().Mul(tc.percentOfSharesToMigrate).RoundInt()

			// Modify migration inputs if necessary
			if tc.overwriteLockId {
				originalGammLockId = originalGammLockId + 1
			}

			balancerDelegationPre, _ := stakingKeeper.GetDelegation(ctx, balancerIntermediaryAcc.GetAccAddress(), valAddr)

			// Run the migration logic.
			positionId, amount0, amount1, liquidityMigrated, joinTime, poolIdLeaving, poolIdEntering, concentratedLockId, err := superfluidKeeper.RouteLockedBalancerToConcentratedMigration(ctx, poolJoinAcc, originalGammLockId, coinsToMigrate, tc.minExitCoins)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectedError)
				return
			}
			s.Require().NoError(err)
			s.AssertEventEmitted(ctx, gammtypes.TypeEvtPoolExited, 1)

			s.ValidateMigrateResult(
				ctx,
				positionId, balancerPooId, poolIdLeaving, clPoolId, poolIdEntering,
				tc.percentOfSharesToMigrate, liquidityMigrated,
				joinTime,
				*balancerLock,
				joinPoolAmt,
				balancerPoolShareOut, coinsToMigrate,
				amount0, amount1,
			)

			// If the lock was superfluid delegated:
			if tc.superfluidDelegated && !tc.superfluidUndelegating {
				if tc.percentOfSharesToMigrate.Equal(sdk.OneDec()) {
					// If we migrated all the shares:

					// The intermediary account connection to the old gamm lock should be deleted.
					addr := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, originalGammLockId)
					s.Require().Equal(addr.String(), "")

					// The synthetic lockup should be deleted.
					_, err = lockupKeeper.GetSyntheticLockup(ctx, originalGammLockId, keeper.StakingSyntheticDenom(balancerLock.Coins[0].Denom, valAddr.String()))
					s.Require().Error(err)

					// The delegation from the balancer intermediary account holder should not exist.
					delegation, found := stakingKeeper.GetDelegation(ctx, balancerIntermediaryAcc.GetAccAddress(), valAddr)
					s.Require().False(found, "expected no delegation, found delegation w/ %d shares", delegation.Shares)

					// Check that the original gamm lockup is deleted.
					_, err := s.App.LockupKeeper.GetLockByID(ctx, originalGammLockId)
					s.Require().Error(err)
				} else if tc.percentOfSharesToMigrate.LT(sdk.OneDec()) {
					// If we migrated part of the shares:
					// The intermediary account connection to the old gamm lock should still be present.
					addr := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, originalGammLockId)
					s.Require().Equal(balancerIntermediaryAcc.GetAccAddress().String(), addr.String())

					// Check if migration deleted synthetic lockup.
					_, err = lockupKeeper.GetSyntheticLockup(ctx, originalGammLockId, keeper.StakingSyntheticDenom(balancerLock.Coins[0].Denom, valAddr.String()))
					s.Require().NoError(err)

					// The delegation from the balancer intermediary account holder should still exist.
					delegation, found := stakingKeeper.GetDelegation(ctx, balancerIntermediaryAcc.GetAccAddress(), valAddr)
					s.Require().True(found, "expected delegation, found delegation no delegation")
					s.Require().Equal(balancerDelegationPre.Shares.Sub(balancerDelegationPre.Shares.Mul(tc.percentOfSharesToMigrate)).RoundInt().String(), delegation.Shares.RoundInt().String(), "expected %d shares, found %d shares", balancerDelegationPre.Shares.Mul(tc.percentOfSharesToMigrate).RoundInt().String(), delegation.Shares.String())

					// Check what is remaining in the original gamm lock.
					lock, err := s.App.LockupKeeper.GetLockByID(ctx, originalGammLockId)
					s.Require().NoError(err)
					s.Require().Equal(balancerPoolShareOut.Amount.Sub(coinsToMigrate.Amount).String(), lock.Coins[0].Amount.String(), "expected %s shares, found %s shares", lock.Coins[0].Amount.String(), balancerPoolShareOut.Amount.Sub(coinsToMigrate.Amount).String())
				}
				// Check the new superfluid staked amount.
				clIntermediaryAcc := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, concentratedLockId)
				delegation, found := stakingKeeper.GetDelegation(ctx, clIntermediaryAcc, valAddr)
				s.Require().True(found, "expected delegation, found delegation no delegation")
				s.Require().Equal(balancerDelegationPre.Shares.Mul(tc.percentOfSharesToMigrate).RoundInt().Sub(sdk.OneInt()).String(), delegation.Shares.RoundInt().String(), "expected %d shares, found %d shares", balancerDelegationPre.Shares.Mul(tc.percentOfSharesToMigrate).RoundInt().String(), delegation.Shares.String())
			}

			// If the lock was superfluid undelegating:
			if tc.superfluidDelegated && tc.superfluidUndelegating {
				// Regardless oh how many shares we migrated:

				// The intermediary account connection to the old gamm lock should be deleted.
				addr := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, originalGammLockId)
				s.Require().Equal(addr.String(), "")

				// The synthetic lockup should be deleted.
				_, err = lockupKeeper.GetSyntheticLockup(ctx, originalGammLockId, keeper.StakingSyntheticDenom(balancerLock.Coins[0].Denom, valAddr.String()))
				s.Require().Error(err)

				// The delegation from the intermediary account holder does not exist.
				delegation, found := stakingKeeper.GetDelegation(ctx, balancerIntermediaryAcc.GetAccAddress(), valAddr)
				s.Require().False(found, "expected no delegation, found delegation w/ %d shares", delegation.Shares)
			}

			// Run slashing logic if the test case involves locks and check if the new and old locks are slashed.
			if !tc.noLock {
				slashExpected := tc.superfluidDelegated || tc.superfluidUndelegating
				s.SlashAndValidateResult(ctx, originalGammLockId, concentratedLockId, poolIdEntering, tc.percentOfSharesToMigrate, valAddr, *balancerLock, slashExpected)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMigrateSuperfluidBondedBalancerToConcentrated() {
	defaultJoinTime := s.Ctx.BlockTime()
	type sendTest struct {
		overwriteValidatorAddress bool
		overwriteLockId           bool
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
		"error: migrate more shares than lock has": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1.1"),
			expectedError:            types.MigrateMoreSharesThanLockHasError{SharesToMigrate: "55000000000000000000", SharesInLock: "50000000000000000000"},
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
		"error: lock that is superfluid delegated, not unlocking (full shares), token out mins is more than exit coins": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			tokenOutMins:             sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(100000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
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

			// We bundle all migration setup into a single function to avoid repeating the same code for each test case.
			joinPoolAmt, balancerIntermediaryAcc, balancerLock, _, poolJoinAcc, balancerPooId, clPoolId, balancerPoolShareOut, valAddr := s.SetupMigrationTest(ctx, true, false, false, false, tc.percentOfSharesToMigrate)
			originalGammLockId := balancerLock.GetID()

			// Depending on the test case, we attempt to migrate a subset of the balancer LP tokens we originally created.
			coinsToMigrate := balancerPoolShareOut
			coinsToMigrate.Amount = coinsToMigrate.Amount.ToDec().Mul(tc.percentOfSharesToMigrate).RoundInt()

			// RouteMigration is called via the migration message router and is always run prior to the migration itself.
			// We use it here just to retrieve the synthetic lock before the migration.
			synthLockBeforeMigration, migrationType, err := superfluidKeeper.RouteMigration(ctx, poolJoinAcc, originalGammLockId, coinsToMigrate)
			s.Require().NoError(err)
			s.Require().Equal(migrationType, keeper.SuperfluidBonded)

			// Modify migration inputs if necessary

			if tc.overwriteValidatorAddress {
				synthDenomParts := strings.Split(synthLockBeforeMigration.SynthDenom, "/")
				synthDenomParts[4] = "osmovaloper1n69ghlk6404gzxtmtq0w7ma59n9vd9ed9dplg" // invalid, too short
				newSynthDenom := strings.Join(synthDenomParts, "/")
				synthLockBeforeMigration.SynthDenom = newSynthDenom
			}

			if tc.overwriteLockId {
				originalGammLockId = originalGammLockId + 1
			}

			balancerDelegationPre, _ := stakingKeeper.GetDelegation(ctx, balancerIntermediaryAcc.GetAccAddress(), valAddr)

			// System under test.
			positionId, amount0, amount1, liquidityMigrated, joinTime, concentratedLockId, poolIdLeaving, poolIdEntering, err := superfluidKeeper.MigrateSuperfluidBondedBalancerToConcentrated(ctx, poolJoinAcc, originalGammLockId, coinsToMigrate, synthLockBeforeMigration.SynthDenom, tc.tokenOutMins)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}
			s.Require().NoError(err)
			s.AssertEventEmitted(ctx, gammtypes.TypeEvtPoolExited, 1)

			s.ValidateMigrateResult(
				ctx,
				positionId, balancerPooId, poolIdLeaving, clPoolId, poolIdEntering,
				tc.percentOfSharesToMigrate, liquidityMigrated,
				joinTime,
				*balancerLock,
				joinPoolAmt,
				balancerPoolShareOut, coinsToMigrate,
				amount0, amount1,
			)

			if tc.percentOfSharesToMigrate.Equal(sdk.OneDec()) {
				// If we migrated all the shares:

				// The intermediary account connection to the old gamm lock should be deleted.
				addr := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, originalGammLockId)
				s.Require().Equal(addr.String(), "")

				// The synthetic lockup should be deleted.
				_, err = lockupKeeper.GetSyntheticLockup(ctx, originalGammLockId, keeper.StakingSyntheticDenom(balancerLock.Coins[0].Denom, valAddr.String()))
				s.Require().Error(err)

				// The delegation from the intermediary account holder should not exist.
				delegation, found := stakingKeeper.GetDelegation(ctx, balancerIntermediaryAcc.GetAccAddress(), valAddr)
				s.Require().False(found, "expected no delegation, found delegation w/ %d shares", delegation.Shares)

				// Check that the original gamm lockup is deleted.
				_, err := s.App.LockupKeeper.GetLockByID(ctx, originalGammLockId)
				s.Require().Error(err)
			} else if tc.percentOfSharesToMigrate.LT(sdk.OneDec()) {
				// If we migrated part of the shares:
				// The intermediary account connection to the old gamm lock should still be present.
				addr := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, originalGammLockId)
				s.Require().Equal(balancerIntermediaryAcc.GetAccAddress().String(), addr.String())

				// Confirm that migration did not delete synthetic lockup.
				gammSynthLock, err := lockupKeeper.GetSyntheticLockup(ctx, originalGammLockId, keeper.StakingSyntheticDenom(balancerLock.Coins[0].Denom, valAddr.String()))
				s.Require().NoError(err)

				s.Require().Equal(originalGammLockId, gammSynthLock.UnderlyingLockId)

				// The delegation from the intermediary account holder should still exist.
				_, found := stakingKeeper.GetDelegation(ctx, balancerIntermediaryAcc.GetAccAddress(), valAddr)
				s.Require().True(found, "expected delegation, found delegation no delegation")

				// Check what is remaining in the original gamm lock.
				lock, err := s.App.LockupKeeper.GetLockByID(ctx, originalGammLockId)
				s.Require().NoError(err)
				s.Require().Equal(balancerPoolShareOut.Amount.Sub(coinsToMigrate.Amount).String(), lock.Coins[0].Amount.String(), "expected %s shares, found %s shares", lock.Coins[0].Amount.String(), balancerPoolShareOut.Amount.Sub(coinsToMigrate.Amount).String())
			}
			// Check the new superfluid staked amount.
			clIntermediaryAcc := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, concentratedLockId)
			delegation, found := stakingKeeper.GetDelegation(ctx, clIntermediaryAcc, valAddr)
			s.Require().True(found, "expected delegation, found delegation no delegation")
			s.Require().Equal(balancerDelegationPre.Shares.Mul(tc.percentOfSharesToMigrate).RoundInt().Sub(sdk.OneInt()).String(), delegation.Shares.RoundInt().String(), "expected %d shares, found %d shares", balancerDelegationPre.Shares.Mul(tc.percentOfSharesToMigrate).RoundInt().String(), delegation.Shares.String())

			// Check if the new intermediary account connection was created.
			newConcentratedIntermediaryAccount := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, concentratedLockId)
			s.Require().NotEqual(newConcentratedIntermediaryAccount.String(), "")

			// Check newly created concentrated lock.
			concentratedLock, err := lockupKeeper.GetLockByID(ctx, concentratedLockId)
			s.Require().NoError(err)
			s.Require().Equal(liquidityMigrated.TruncateInt().String(), concentratedLock.Coins[0].Amount.String(), "expected %s shares, found %s shares", coinsToMigrate.Amount.String(), concentratedLock.Coins[0].Amount.String())
			s.Require().Equal(balancerLock.Duration, concentratedLock.Duration)
			s.Require().Equal(balancerLock.EndTime, concentratedLock.EndTime)

			// Check if the new synthetic bonded lockup was created.
			clSynthLock, err := lockupKeeper.GetSyntheticLockup(ctx, concentratedLockId, keeper.StakingSyntheticDenom(concentratedLock.Coins[0].Denom, valAddr.String()))
			s.Require().NoError(err)

			s.Require().Equal(concentratedLockId, clSynthLock.UnderlyingLockId)

			// Run slashing logic and check if the new and old locks are slashed.
			s.SlashAndValidateResult(ctx, originalGammLockId, concentratedLockId, clPoolId, tc.percentOfSharesToMigrate, valAddr, *balancerLock, true)
		})
	}
}

func (s *KeeperTestSuite) TestMigrateSuperfluidUnbondingBalancerToConcentrated() {
	defaultJoinTime := s.Ctx.BlockTime()
	type sendTest struct {
		unlocking                 bool
		overwriteValidatorAddress bool
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
		"error: lock that is superfluid undelegating, not unlocking (full shares), token out mins is more than exit coins": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			tokenOutMins:             sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(100000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)
			ctx := s.Ctx
			superfluidKeeper := s.App.SuperfluidKeeper
			lockupKeeper := s.App.LockupKeeper

			// We bundle all migration setup into a single function to avoid repeating the same code for each test case.
			joinPoolAmt, _, balancerLock, _, poolJoinAcc, balancerPooId, clPoolId, balancerPoolShareOut, valAddr := s.SetupMigrationTest(ctx, true, true, tc.unlocking, false, tc.percentOfSharesToMigrate)
			originalGammLockId := balancerLock.GetID()

			// Depending on the test case, we attempt to migrate a subset of the balancer LP tokens we originally created.
			coinsToMigrate := balancerPoolShareOut
			coinsToMigrate.Amount = coinsToMigrate.Amount.ToDec().Mul(tc.percentOfSharesToMigrate).RoundInt()

			// RouteMigration is called via the migration message router and is always run prior to the migration itself
			synthLockBeforeMigration, migrationType, err := superfluidKeeper.RouteMigration(ctx, poolJoinAcc, originalGammLockId, coinsToMigrate)
			s.Require().NoError(err)
			s.Require().Equal(migrationType, keeper.SuperfluidUnbonding)

			// Modify migration inputs if necessary

			if tc.overwriteValidatorAddress {
				synthDenomParts := strings.Split(synthLockBeforeMigration.SynthDenom, "/")
				synthDenomParts[4] = "osmovaloper1n69ghlk6404gzxtmtq0w7ma59n9vd9ed9dplg" // invalid, too short
				newSynthDenom := strings.Join(synthDenomParts, "/")
				synthLockBeforeMigration.SynthDenom = newSynthDenom
			}

			// System under test.
			positionId, amount0, amount1, liquidityMigrated, joinTime, concentratedLockId, poolIdLeaving, poolIdEntering, err := superfluidKeeper.MigrateSuperfluidUnbondingBalancerToConcentrated(ctx, poolJoinAcc, originalGammLockId, coinsToMigrate, synthLockBeforeMigration.SynthDenom, tc.tokenOutMins)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}
			s.Require().NoError(err)
			s.AssertEventEmitted(ctx, gammtypes.TypeEvtPoolExited, 1)

			s.ValidateMigrateResult(
				ctx,
				positionId, balancerPooId, poolIdLeaving, clPoolId, poolIdEntering,
				tc.percentOfSharesToMigrate, liquidityMigrated,
				joinTime,
				*balancerLock,
				joinPoolAmt,
				balancerPoolShareOut, coinsToMigrate,
				amount0, amount1,
			)

			if tc.percentOfSharesToMigrate.Equal(sdk.OneDec()) {
				// If we migrated all the shares:

				// The synthetic lockup should be deleted.
				_, err = lockupKeeper.GetSyntheticLockup(ctx, originalGammLockId, keeper.UnstakingSyntheticDenom(balancerLock.Coins[0].Denom, valAddr.String()))
				s.Require().Error(err)
			} else if tc.percentOfSharesToMigrate.LT(sdk.OneDec()) {
				// If we migrated part of the shares:

				// The synthetic lockup should not be deleted.
				gammSynthLock, err := lockupKeeper.GetSyntheticLockup(ctx, originalGammLockId, keeper.UnstakingSyntheticDenom(balancerLock.Coins[0].Denom, valAddr.String()))
				s.Require().NoError(err)

				s.Require().Equal(originalGammLockId, gammSynthLock.UnderlyingLockId)
			}

			// Check newly created concentrated lock.
			concentratedLock, err := lockupKeeper.GetLockByID(ctx, concentratedLockId)
			s.Require().NoError(err)
			s.Require().Equal(liquidityMigrated.TruncateInt().String(), concentratedLock.Coins[0].Amount.String(), "expected %s shares, found %s shares", coinsToMigrate.Amount.String(), concentratedLock.Coins[0].Amount.String())
			s.Require().Equal(balancerLock.Duration, concentratedLock.Duration)
			s.Require().Equal(s.Ctx.BlockTime().Add(balancerLock.Duration), concentratedLock.EndTime)

			// Check if the new synthetic unbonding lockup was created.
			clSynthLock, err := lockupKeeper.GetSyntheticLockup(ctx, concentratedLockId, keeper.UnstakingSyntheticDenom(concentratedLock.Coins[0].Denom, valAddr.String()))
			s.Require().NoError(err)

			s.Require().Equal(concentratedLockId, clSynthLock.UnderlyingLockId)

			// Run slashing logic and check if the new and old locks are slashed.
			s.SlashAndValidateResult(ctx, originalGammLockId, concentratedLockId, clPoolId, tc.percentOfSharesToMigrate, valAddr, *balancerLock, true)
		})
	}
}

func (s *KeeperTestSuite) TestMigrateNonSuperfluidLockBalancerToConcentrated() {
	defaultJoinTime := s.Ctx.BlockTime()
	type sendTest struct {
		unlocking                bool
		percentOfSharesToMigrate sdk.Dec
		tokenOutMins             sdk.Coins
		expectedError            error
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
		"error: lock that is not superfluid delegated, not unlocking (full shares), token out mins is more than exit coins": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			tokenOutMins:             sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)
			ctx := s.Ctx
			superfluidKeeper := s.App.SuperfluidKeeper
			lockupKeeper := s.App.LockupKeeper

			// We bundle all migration setup into a single function to avoid repeating the same code for each test case.
			joinPoolAmt, _, balancerLock, _, poolJoinAcc, balancerPooId, clPoolId, balancerPoolShareOut, valAddr := s.SetupMigrationTest(ctx, false, false, tc.unlocking, false, tc.percentOfSharesToMigrate)
			originalGammLockId := balancerLock.GetID()

			// Depending on the test case, we attempt to migrate a subset of the balancer LP tokens we originally created.
			coinsToMigrate := balancerPoolShareOut
			coinsToMigrate.Amount = coinsToMigrate.Amount.ToDec().Mul(tc.percentOfSharesToMigrate).RoundInt()

			// RouteMigration is called via the migration message router and is always run prior to the migration itself
			synthLockBeforeMigration, migrationType, err := superfluidKeeper.RouteMigration(ctx, poolJoinAcc, originalGammLockId, coinsToMigrate)
			s.Require().NoError(err)
			s.Require().Equal((lockuptypes.SyntheticLock{}), synthLockBeforeMigration)
			s.Require().Equal(migrationType, keeper.NonSuperfluid)

			// System under test.
			positionId, amount0, amount1, liquidityMigrated, joinTime, concentratedLockId, poolIdLeaving, poolIdEntering, err := superfluidKeeper.MigrateNonSuperfluidLockBalancerToConcentrated(ctx, poolJoinAcc, originalGammLockId, coinsToMigrate, tc.tokenOutMins)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}
			s.Require().NoError(err)
			s.AssertEventEmitted(ctx, gammtypes.TypeEvtPoolExited, 1)

			s.ValidateMigrateResult(
				ctx,
				positionId, balancerPooId, poolIdLeaving, clPoolId, poolIdEntering,
				tc.percentOfSharesToMigrate, liquidityMigrated,
				joinTime,
				*balancerLock,
				joinPoolAmt,
				balancerPoolShareOut, coinsToMigrate,
				amount0, amount1,
			)

			// Check newly created concentrated lock.
			concentratedLock, err := lockupKeeper.GetLockByID(ctx, concentratedLockId)
			s.Require().NoError(err)
			s.Require().Equal(liquidityMigrated.TruncateInt().String(), concentratedLock.Coins[0].Amount.String(), "expected %s shares, found %s shares", coinsToMigrate.Amount.String(), concentratedLock.Coins[0].Amount.String())
			s.Require().Equal(balancerLock.Duration, concentratedLock.Duration)
			s.Require().Equal(s.Ctx.BlockTime().Add(balancerLock.Duration), concentratedLock.EndTime)

			// Run slashing logic and check if the new and old locks are not slashed.
			s.SlashAndValidateResult(ctx, originalGammLockId, concentratedLockId, clPoolId, tc.percentOfSharesToMigrate, valAddr, *balancerLock, false)
		})
	}
}

func (s *KeeperTestSuite) TestMigrateUnlockedPositionFromBalancerToConcentrated() {
	defaultJoinTime := s.Ctx.BlockTime()
	type sendTest struct {
		unlocking                bool
		percentOfSharesToMigrate sdk.Dec
		tokenOutMins             sdk.Coins
		expectedError            error
	}
	testCases := map[string]sendTest{
		"no lock (full shares)": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
		},
		"no lock (partial shares)": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.9"),
		},
		"no lock (more shares than own)": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1.1"),
			expectedError:            fmt.Errorf("insufficient funds"),
		},
		"no lock (no shares)": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0"),
			expectedError:            errorsmod.Wrapf(gammtypes.ErrInvalidMathApprox, "Trying to exit a negative amount of shares"),
		},
		"error: no lock (full shares), token out mins is more than exit coins": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			tokenOutMins:             sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)
			ctx := s.Ctx
			superfluidKeeper := s.App.SuperfluidKeeper
			gammKeeper := s.App.GAMMKeeper

			// We bundle all migration setup into a single function to avoid repeating the same code for each test case.
			joinPoolAmt, _, balancerLock, _, poolJoinAcc, balancerPooId, clPoolId, balancerPoolShareOut, _ := s.SetupMigrationTest(ctx, false, false, false, true, tc.percentOfSharesToMigrate)
			s.Require().Equal(uint64(0), balancerLock.GetID())

			// Depending on the test case, we attempt to migrate a subset of the balancer LP tokens we originally created.
			coinsToMigrate := balancerPoolShareOut
			coinsToMigrate.Amount = coinsToMigrate.Amount.ToDec().Mul(tc.percentOfSharesToMigrate).RoundInt()

			// RouteMigration is called via the migration message router and is always run prior to the migration itself
			synthLockBeforeMigration, migrationType, err := superfluidKeeper.RouteMigration(ctx, poolJoinAcc, 0, coinsToMigrate)
			s.Require().NoError(err)
			s.Require().Equal((lockuptypes.SyntheticLock{}), synthLockBeforeMigration)
			s.Require().Equal(migrationType, keeper.Unlocked)

			// System under test.
			positionId, amount0, amount1, liquidityMigrated, joinTime, poolIdLeaving, poolIdEntering, err := gammKeeper.MigrateUnlockedPositionFromBalancerToConcentrated(ctx, poolJoinAcc, coinsToMigrate, tc.tokenOutMins)
			if tc.expectedError != nil {
				s.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}
			s.Require().NoError(err)
			s.AssertEventEmitted(ctx, gammtypes.TypeEvtPoolExited, 1)

			s.ValidateMigrateResult(
				ctx,
				positionId, balancerPooId, poolIdLeaving, clPoolId, poolIdEntering,
				tc.percentOfSharesToMigrate, liquidityMigrated,
				joinTime,
				*balancerLock,
				joinPoolAmt,
				balancerPoolShareOut, coinsToMigrate,
				amount0, amount1,
			)
		})
	}
}

func (s *KeeperTestSuite) TestValidateMigration() {
	defaultJoinTime := s.Ctx.BlockTime()
	type sendTest struct {
		isSuperfluidDelegated     bool
		isSuperfluidUndelegating  bool
		unlocking                 bool
		overwritePreMigrationLock bool
		overwriteSender           bool
		overwriteSharesDenomValue string
		overwriteLockId           bool
		percentOfSharesToMigrate  sdk.Dec
		tokenOutMins              sdk.Coins
		expectedError             error
	}
	testCases := map[string]sendTest{
		"lock that is not superfluid delegated, not unlocking (full shares)": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			isSuperfluidDelegated:    false,
			isSuperfluidUndelegating: false,
		},
		"lock that is not superfluid delegated, not unlocking (partial shares)": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.9"),
			isSuperfluidDelegated:    false,
			isSuperfluidUndelegating: false,
		},
		"lock that is not superfluid delegated, unlocking (full shares)": {
			unlocking:                true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			isSuperfluidDelegated:    false,
			isSuperfluidUndelegating: false,
		},
		"lock that is not superfluid delegated, unlocking (partial shares)": {
			unlocking:                true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.6"),
			isSuperfluidDelegated:    false,
			isSuperfluidUndelegating: false,
		},
		"lock that is superfluid undelegating, not unlocking (full shares)": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			isSuperfluidDelegated:    true,
			isSuperfluidUndelegating: true,
		},
		"lock that is superfluid undelegating, not unlocking (partial shares)": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.5"),
			isSuperfluidDelegated:    true,
			isSuperfluidUndelegating: true,
		},
		"lock that is superfluid undelegating, unlocking (full shares)": {
			unlocking:                true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			isSuperfluidDelegated:    true,
			isSuperfluidUndelegating: true,
		},
		"lock that is superfluid undelegating, unlocking (partial shares)": {
			unlocking:                true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.3"),
			isSuperfluidDelegated:    true,
			isSuperfluidUndelegating: true,
		},
		"lock that is superfluid delegated, not unlocking (full shares)": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			isSuperfluidDelegated:    true,
		},
		"lock that is superfluid delegated, not unlocking (partial shares)": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("0.5"),
			isSuperfluidDelegated:    true,
		},
		"error: denom prefix error": {
			overwriteSharesDenomValue: "cl/pool/2",
			percentOfSharesToMigrate:  sdk.MustNewDecFromStr("1"),
			expectedError:             types.SharesToMigrateDenomPrefixError{Denom: "cl/pool/2", ExpectedDenomPrefix: gammtypes.GAMMTokenPrefix},
		},
		"error: no canonical link": {
			overwriteSharesDenomValue: "gamm/pool/2",
			percentOfSharesToMigrate:  sdk.MustNewDecFromStr("1"),
			expectedError:             gammtypes.ConcentratedPoolMigrationLinkNotFoundError{PoolIdLeaving: 2},
		},
		"error: wrong sender": {
			overwriteSender:          true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			expectedError:            lockuptypes.ErrNotLockOwner,
		},
		"error: wrong lock ID": {
			overwriteLockId:          true,
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			expectedError:            lockuptypes.ErrLockupNotFound,
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)
			ctx := s.Ctx
			superfluidKeeper := s.App.SuperfluidKeeper

			// We bundle all migration setup into a single function to avoid repeating the same code for each test case.
			_, _, balancerLock, _, poolJoinAcc, balancerPooId, clPoolId, balancerPoolShareOut, _ := s.SetupMigrationTest(ctx, tc.isSuperfluidDelegated, tc.isSuperfluidUndelegating, tc.unlocking, false, tc.percentOfSharesToMigrate)
			originalGammLockId := balancerLock.GetID()

			// Depending on the test case, we attempt to migrate a subset of the balancer LP tokens we originally created.
			coinsToMigrate := balancerPoolShareOut
			coinsToMigrate.Amount = coinsToMigrate.Amount.ToDec().Mul(tc.percentOfSharesToMigrate).RoundInt()

			// Modify migration inputs if necessary
			if tc.overwriteSender {
				poolJoinAcc = s.TestAccs[0]
			}

			if tc.overwriteLockId {
				originalGammLockId = originalGammLockId + 10
			}

			if tc.overwriteSharesDenomValue != "" {
				coinsToMigrate.Denom = tc.overwriteSharesDenomValue
			}

			// System under test.
			poolIdLeaving, poolIdEntering, preMigrationLock, remainingLockTime, err := superfluidKeeper.ValidateMigration(ctx, poolJoinAcc, originalGammLockId, coinsToMigrate)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(poolIdLeaving, balancerPooId)
			s.Require().Equal(poolIdEntering, clPoolId)
			s.Require().Equal(preMigrationLock.GetID(), originalGammLockId)
			s.Require().Equal(preMigrationLock.GetCoins(), sdk.NewCoins(balancerPoolShareOut))
			s.Require().Equal(preMigrationLock.GetDuration(), remainingLockTime)
		})
	}
}

func (s *KeeperTestSuite) TestValidateSharesToMigrateUnlockAndExitBalancerPool() {
	defaultJoinTime := s.Ctx.BlockTime()
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
			expectedError:            types.MigrateMoreSharesThanLockHasError{SharesToMigrate: "50000000000000000001", SharesInLock: "50000000000000000000"},
		},
		"error: attempt to leave a pool that does not exist": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			overwritePoolId:          true,
			expectedError:            fmt.Errorf("pool with ID %d does not exist", 2),
		},
		"error: attempt to leave a pool that has more than two denoms": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			overwritePool:            true,
			expectedError:            types.TwoTokenBalancerPoolError{NumberOfTokens: 4},
		},
		"error: happy path (full shares), token out mins is more than exit coins": {
			percentOfSharesToMigrate: sdk.MustNewDecFromStr("1"),
			tokenOutMins:             sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(100000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)
			ctx := s.Ctx
			superfluidKeeper := s.App.SuperfluidKeeper
			lockupKeeper := s.App.LockupKeeper
			poolmanagerKeeper := s.App.PoolManagerKeeper
			bankKeeper := s.App.BankKeeper
			gammKeeper := s.App.GAMMKeeper
			stakingKeeper := s.App.StakingKeeper

			// Generate and fund two accounts.
			// Account 1 will be the account that creates the pool.
			// Account 2 will be the account that joins the pool.
			delAddrs := CreateRandomAccounts(2)
			poolCreateAcc := delAddrs[0]
			poolJoinAcc := delAddrs[1]
			for _, acc := range delAddrs {
				err := simapp.FundAccount(bankKeeper, ctx, acc, defaultAcctFunds)
				s.Require().NoError(err)
			}

			// Create a balancer pool of "stake" and "foo".
			msg := balancer.NewMsgCreateBalancerPool(poolCreateAcc, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDec(0),
			}, defaultPoolAssets, defaultFutureGovernor)
			balancerPooId, err := poolmanagerKeeper.CreatePool(ctx, msg)
			s.Require().NoError(err)

			// Join the balancer pool.
			tokensIn, _, err := gammKeeper.JoinPoolNoSwap(ctx, poolJoinAcc, balancerPooId, gammtypes.OneShare.MulRaw(50), sdk.Coins{})
			s.Require().NoError(err)

			// Determine the balancer pool's LP token denomination.
			balancerPoolDenom := gammtypes.GetPoolShareDenom(balancerPooId)

			// Note how much of the balancer pool's LP token the account that joined the pool has.
			balancerPoolShareOut := bankKeeper.GetBalance(ctx, poolJoinAcc, balancerPoolDenom)

			sharesToMigrate := balancerPoolShareOut.Amount.ToDec().Mul(tc.percentOfSharesToMigrate).TruncateInt()
			coinsToMigrate := sdk.NewCoin(balancerPoolDenom, sharesToMigrate)

			// The unbonding duration is the same as the staking module's unbonding duration.
			unbondingDuration := stakingKeeper.GetParams(ctx).UnbondingTime

			// Lock the LP tokens for the duration of the unbonding period.
			originalGammLockId := s.LockTokens(poolJoinAcc, sdk.NewCoins(balancerPoolShareOut), unbondingDuration)

			lock, err := lockupKeeper.GetLockByID(ctx, originalGammLockId)
			s.Require().NoError(err)

			if tc.overwritePreMigrationLock {
				lock.ID = lock.ID + 1
			}

			if tc.overwriteShares {
				coinsToMigrate.Amount = lock.Coins[0].Amount.Add(sdk.NewInt(1))
			}

			if tc.overwritePool {
				multiCoinBalancerPoolId := s.PrepareBalancerPool()
				balancerPooId = multiCoinBalancerPoolId
				shareAmt := sdk.MustNewDecFromStr("50000000000000000000").TruncateInt()
				newShares := sdk.NewCoin(fmt.Sprintf("gamm/pool/%d", multiCoinBalancerPoolId), shareAmt)
				s.FundAcc(poolJoinAcc, sdk.NewCoins(newShares))
			}

			if tc.overwritePoolId {
				balancerPooId = balancerPooId + 1
			}

			// System under test
			exitCoins, err := superfluidKeeper.ValidateSharesToMigrateUnlockAndExitBalancerPool(ctx, poolJoinAcc, balancerPooId, lock, coinsToMigrate, tc.tokenOutMins, lock.Duration)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}
			s.Require().NoError(err)

			defaultErrorTolerance := osmomath.ErrTolerance{
				AdditiveTolerance: sdk.NewDec(1),
				RoundingDir:       osmomath.RoundDown,
			}

			if tc.percentOfSharesToMigrate.Equal(sdk.OneDec()) {
				// If all of the shares were migrated, the original lock should be deleted
				_, err := lockupKeeper.GetLockByID(ctx, originalGammLockId)
				s.Require().Error(err)
			} else {
				// If only a portion of the shares were migrated, the original lock should still exist (with the remaining shares)
				lock, err := lockupKeeper.GetLockByID(ctx, originalGammLockId)
				s.Require().NoError(err)
				expectedSharesStillInOldLock := balancerPoolShareOut.Amount.Sub(sharesToMigrate)
				s.Require().Equal(expectedSharesStillInOldLock.String(), lock.Coins[0].Amount.String())
			}

			for _, coin := range exitCoins {
				// Check that the exit coin is the same amount that we joined with (with one unit rounding down)
				s.Require().Equal(0, defaultErrorTolerance.Compare(tokensIn.AmountOf(coin.Denom).ToDec().Mul(tc.percentOfSharesToMigrate).RoundInt(), coin.Amount))
			}
		})
	}
}

func (s *KeeperTestSuite) SetupMigrationTest(ctx sdk.Context, superfluidDelegated, superfluidUndelegating, unlocking, noLock bool, percentOfSharesToMigrate sdk.Dec) (joinPoolAmt sdk.Coins, balancerIntermediaryAcc types.SuperfluidIntermediaryAccount, balancerLock *lockuptypes.PeriodLock, poolCreateAcc, poolJoinAcc sdk.AccAddress, balancerPooId, clPoolId uint64, balancerPoolShareOut sdk.Coin, valAddr sdk.ValAddress) {
	bankKeeper := s.App.BankKeeper
	gammKeeper := s.App.GAMMKeeper
	superfluidKeeper := s.App.SuperfluidKeeper
	lockupKeeper := s.App.LockupKeeper
	stakingKeeper := s.App.StakingKeeper
	poolmanagerKeeper := s.App.PoolManagerKeeper

	fullRangeCoins := sdk.NewCoins(defaultPoolAssets[0].Token, defaultPoolAssets[1].Token)

	// Generate and fund two accounts.
	// Account 1 will be the account that creates the pool.
	// Account 2 will be the account that joins the pool.
	delAddrs := CreateRandomAccounts(2)
	poolCreateAcc = delAddrs[0]
	poolJoinAcc = delAddrs[1]
	for _, acc := range delAddrs {
		err := simapp.FundAccount(bankKeeper, ctx, acc, defaultAcctFunds)
		s.Require().NoError(err)
	}

	// Set up a single validator.
	valAddr = s.SetupValidator(stakingtypes.Bonded)

	// Create a balancer pool of "stake" and "foo".
	msg := balancer.NewMsgCreateBalancerPool(poolCreateAcc, balancer.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.NewDec(0),
	}, defaultPoolAssets, defaultFutureGovernor)
	balancerPooId, err := poolmanagerKeeper.CreatePool(ctx, msg)
	s.Require().NoError(err)

	// Join the balancer pool.
	// Note the account balance before and after joining the pool.
	balanceBeforeJoin := bankKeeper.GetAllBalances(ctx, poolJoinAcc)
	_, _, err = gammKeeper.JoinPoolNoSwap(ctx, poolJoinAcc, balancerPooId, gammtypes.OneShare.MulRaw(50), sdk.Coins{})
	s.Require().NoError(err)
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
	s.Require().NoError(err)

	// Note how much of the balancer pool's LP token the account that joined the pool has.
	balancerPoolShareOut = bankKeeper.GetBalance(ctx, poolJoinAcc, balancerPoolDenom)

	// Create a cl pool with the same underlying assets as the balancer pool.
	clPool := s.PrepareCustomConcentratedPool(poolCreateAcc, defaultPoolAssets[0].Token.Denom, defaultPoolAssets[1].Token.Denom, 1, sdk.ZeroDec())
	clPoolId = clPool.GetId()

	// Add a gov sanctioned link between the balancer and concentrated liquidity pool.
	migrationRecord := gammtypes.MigrationRecords{BalancerToConcentratedPoolLinks: []gammtypes.BalancerToConcentratedPoolLink{
		{BalancerPoolId: balancerPooId, ClPoolId: clPoolId},
	}}
	gammKeeper.OverwriteMigrationRecordsAndRedirectDistrRecords(ctx, migrationRecord)

	// The unbonding duration is the same as the staking module's unbonding duration.
	unbondingDuration := stakingKeeper.GetParams(ctx).UnbondingTime

	// Lock the LP tokens for the duration of the unbonding period.
	originalGammLockId := uint64(0)
	if !noLock {
		originalGammLockId = s.LockTokens(poolJoinAcc, sdk.NewCoins(balancerPoolShareOut), unbondingDuration)
	}

	// Superfluid delegate the balancer lock if the test case requires it.
	// Note the intermediary account that was created.
	if superfluidDelegated {
		err = superfluidKeeper.SuperfluidDelegate(ctx, poolJoinAcc.String(), originalGammLockId, valAddr.String())
		s.Require().NoError(err)
		intermediaryAccConnection := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, originalGammLockId)
		balancerIntermediaryAcc = superfluidKeeper.GetIntermediaryAccount(ctx, intermediaryAccConnection)
	}

	// Superfluid undelegate the lock if the test case requires it.
	if superfluidUndelegating {
		err = superfluidKeeper.SuperfluidUndelegate(ctx, poolJoinAcc.String(), originalGammLockId)
		s.Require().NoError(err)
	}

	// Unlock the balancer lock if the test case requires it.
	if unlocking {
		// If lock was superfluid staked, we can't unlock via `BeginUnlock`,
		// we need to unlock lock via `SuperfluidUnbondLock`
		if superfluidUndelegating {
			err = superfluidKeeper.SuperfluidUnbondLock(ctx, originalGammLockId, poolJoinAcc.String())
			s.Require().NoError(err)
		} else {
			lock, err := lockupKeeper.GetLockByID(ctx, originalGammLockId)
			s.Require().NoError(err)
			_, err = lockupKeeper.BeginUnlock(ctx, originalGammLockId, lock.Coins)
			s.Require().NoError(err)
		}
	}

	balancerLock = &lockuptypes.PeriodLock{}
	if !noLock {
		balancerLock, err = lockupKeeper.GetLockByID(ctx, originalGammLockId)
		s.Require().NoError(err)
	}

	// Create a full range position in the concentrated liquidity pool.
	// This is to have a spot price and liquidity value to work off when migrating.
	s.CreateFullRangePosition(clPool, fullRangeCoins)

	// Register the CL full range LP tokens as a superfluid asset.
	clPoolDenom := cltypes.GetConcentratedLockupDenomFromPoolId(clPoolId)
	err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{
		Denom:     clPoolDenom,
		AssetType: types.SuperfluidAssetTypeConcentratedShare,
	})

	s.Require().NoError(err)
	return joinPoolAmt, balancerIntermediaryAcc, balancerLock, poolCreateAcc, poolJoinAcc, balancerPooId, clPoolId, balancerPoolShareOut, valAddr
}

func (s *KeeperTestSuite) SlashAndValidateResult(ctx sdk.Context, gammLockId, concentratedLockId, poolIdEntering uint64, percentOfSharesToMigrate sdk.Dec, valAddr sdk.ValAddress, balancerLock lockuptypes.PeriodLock, expectSlash bool) {
	// Retrieve the concentrated lock and gamm lock prior to slashing.
	concentratedLockPreSlash, err := s.App.LockupKeeper.GetLockByID(s.Ctx, concentratedLockId)
	s.Require().NoError(err)
	gammLockPreSlash, err := s.App.LockupKeeper.GetLockByID(s.Ctx, gammLockId)
	if percentOfSharesToMigrate.LT(sdk.OneDec()) {
		s.Require().NoError(err)
	} else {
		s.Require().Error(err)
	}

	// Slash the validator.
	slashFactor := sdk.NewDecWithPrec(5, 2)
	s.App.SuperfluidKeeper.SlashLockupsForValidatorSlash(
		s.Ctx,
		valAddr,
		s.Ctx.BlockHeight(),
		slashFactor)

	// Retrieve the concentrated lock and gamm lock after slashing.
	concentratedLockPostSlash, err := s.App.LockupKeeper.GetLockByID(s.Ctx, concentratedLockId)
	s.Require().NoError(err)
	gammLockPostSlash, err := s.App.LockupKeeper.GetLockByID(s.Ctx, gammLockId)
	if percentOfSharesToMigrate.LT(sdk.OneDec()) {
		s.Require().NoError(err)
	} else {
		s.Require().Error(err)
	}

	// Check if the concentrated lock was slashed.
	clDenom := cltypes.GetConcentratedLockupDenomFromPoolId(poolIdEntering)
	slashAmtCL := concentratedLockPreSlash.Coins.AmountOf(clDenom).ToDec().Mul(slashFactor).TruncateInt()
	if !expectSlash {
		slashAmtCL = sdk.ZeroInt()
	}
	s.Require().Equal(concentratedLockPreSlash.Coins.AmountOf(clDenom).Sub(slashAmtCL).String(), concentratedLockPostSlash.Coins.AmountOf(clDenom).String())

	// Check if the gamm lock was slashed.
	// We only check if the gamm lock was slashed if the lock was not migrated entirely.
	// Otherwise, there would be no newly created gamm lock to check.
	if percentOfSharesToMigrate.LT(sdk.OneDec()) {
		gammDenom := balancerLock.Coins[0].Denom
		slashAmtGamm := gammLockPreSlash.Coins.AmountOf(gammDenom).ToDec().Mul(slashFactor).TruncateInt()
		if !expectSlash {
			slashAmtGamm = sdk.ZeroInt()
		}
		s.Require().Equal(gammLockPreSlash.Coins.AmountOf(gammDenom).Sub(slashAmtGamm).String(), gammLockPostSlash.Coins.AmountOf(gammDenom).String())
	}
}

func (s *KeeperTestSuite) ValidateMigrateResult(
	ctx sdk.Context,
	positionId, balancerPooId, poolIdLeaving, clPoolId, poolIdEntering uint64,
	percentOfSharesToMigrate, liquidityMigrated sdk.Dec,
	joinTime time.Time,
	balancerLock lockuptypes.PeriodLock,
	joinPoolAmt sdk.Coins,
	balancerPoolShareOut, coinsToMigrate sdk.Coin,
	amount0, amount1 sdk.Int,
) {
	// Check that the concentrated liquidity and join time match what we expect
	position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(ctx, positionId)
	s.Require().NoError(err)
	s.Require().Equal(liquidityMigrated, position.Liquidity)
	s.Require().Equal(joinTime, position.JoinTime)

	// Expect the poolIdLeaving to be the balancer pool id
	// Expect the poolIdEntering to be the concentrated liquidity pool id
	s.Require().Equal(balancerPooId, poolIdLeaving)
	s.Require().Equal(clPoolId, poolIdEntering)

	// exitPool has rounding difference.
	// We test if correct amt has been exited and frozen by comparing with rounding tolerance.
	defaultErrorTolerance := osmomath.ErrTolerance{
		AdditiveTolerance: sdk.NewDec(2),
		RoundingDir:       osmomath.RoundDown,
	}
	s.Require().Equal(0, defaultErrorTolerance.Compare(joinPoolAmt.AmountOf(defaultPoolAssets[0].Token.Denom).ToDec().Mul(percentOfSharesToMigrate).RoundInt(), amount0))
	s.Require().Equal(0, defaultErrorTolerance.Compare(joinPoolAmt.AmountOf(defaultPoolAssets[1].Token.Denom).ToDec().Mul(percentOfSharesToMigrate).RoundInt(), amount1))
}
