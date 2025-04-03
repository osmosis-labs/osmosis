package keeper_test

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	gammmigration "github.com/osmosis-labs/osmosis/v27/x/gamm/types/migration"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
)

var (
	STAKE            = "stake"
	DefaultAmt0      = osmomath.NewInt(1000000)
	DefaultCoin0     = sdk.NewCoin(STAKE, DefaultAmt0)
	USDC             = "usdc"
	DefaultAmt1      = osmomath.NewInt(5000000000)
	DefaultCoin1     = sdk.NewCoin(USDC, DefaultAmt1)
	DefaultCoins     = sdk.NewCoins(DefaultCoin0, DefaultCoin1)
	DefaultLowerTick = int64(30545000)
	DefaultUpperTick = int64(31500000)
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
		percentOfSharesToMigrate osmomath.Dec
		minExitCoins             sdk.Coins
		expectedError            error
	}
	testCases := map[string]sendTest{
		"lock that is not superfluid delegated, not unlocking": {
			// migrateNonSuperfluidLockBalancerToConcentrated
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("0.9"),
			expectedError:            types.MigratePartialSharesError{SharesToMigrate: "45000000000000000000", SharesInLock: "50000000000000000000"},
		},
		"lock that is not superfluid delegated, unlocking": {
			// migrateNonSuperfluidLockBalancerToConcentrated
			unlocking:                true,
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("0.6"),
			expectedError:            types.MigratePartialSharesError{SharesToMigrate: "30000000000000000000", SharesInLock: "50000000000000000000"},
		},
		"lock that is superfluid delegated, not unlocking (full shares)": {
			// migrateSuperfluidBondedBalancerToConcentrated
			superfluidDelegated:      true,
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
		},
		"lock that is superfluid undelegating, not unlocking (full shares)": {
			// migrateSuperfluidUnbondingBalancerToConcentrated
			superfluidDelegated:      true,
			superfluidUndelegating:   true,
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
		},
		"lock that is superfluid undelegating, unlocking (full shares)": {
			// migrateSuperfluidUnbondingBalancerToConcentrated
			superfluidDelegated:      true,
			superfluidUndelegating:   true,
			unlocking:                true,
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
		},
		"no lock (partial shares)": {
			// MigrateUnlockedPositionFromBalancerToConcentrated
			noLock:                   true,
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("0.3"),
		},
		"no lock (full shares)": {
			// MigrateUnlockedPositionFromBalancerToConcentrated
			noLock:                   true,
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
		},
		"error: non-existent lock": {
			overwriteLockId:          true,
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			expectedError:            errorsmod.Wrap(lockuptypes.ErrLockupNotFound, fmt.Sprintf("lock with ID %d does not exist", 5)),
		},
		"error: lock that is not superfluid delegated, not unlocking, min exit coins more than being exited": {
			// migrateNonSuperfluidLockBalancerToConcentrated
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			minExitCoins:             sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(5000)), sdk.NewCoin("stake", osmomath.NewInt(5000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
		"error: lock that is not superfluid delegated, unlocking, min exit coins more than being exited": {
			// migrateNonSuperfluidLockBalancerToConcentrated
			unlocking:                true,
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			minExitCoins:             sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(5000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
		"error: lock that is superfluid delegated, not unlocking (full shares), min exit coins more than being exited": {
			// migrateSuperfluidBondedBalancerToConcentrated
			superfluidDelegated:      true,
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			minExitCoins:             sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(10000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
		"error: lock that is superfluid undelegating, not unlocking, min exit coins more than being exited": {
			// migrateSuperfluidUnbondingBalancerToConcentrated
			superfluidDelegated:      true,
			superfluidUndelegating:   true,
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			minExitCoins:             sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(40000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
		"lock that is superfluid undelegating, unlocking, min exit coins more than being exited": {
			// migrateSuperfluidUnbondingBalancerToConcentrated
			superfluidDelegated:      true,
			superfluidUndelegating:   true,
			unlocking:                true,
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			minExitCoins:             sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(40000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)
			superfluidKeeper := s.App.SuperfluidKeeper
			lockupKeeper := s.App.LockupKeeper
			stakingKeeper := s.App.StakingKeeper

			// We bundle all migration setup into a single function to avoid repeating the same code for each test case.
			joinPoolAmt, balancerIntermediaryAcc, balancerLock, _, poolJoinAcc, balancerPooId, clPoolId, balancerPoolShareOut, valAddr := s.SetupMigrationTest(s.Ctx, tc.superfluidDelegated, tc.superfluidUndelegating, tc.unlocking, tc.noLock, tc.percentOfSharesToMigrate)
			originalGammLockId := balancerLock.GetID()

			// Depending on the test case, we attempt to migrate a subset of the balancer LP tokens we originally created.
			coinsToMigrate := balancerPoolShareOut
			coinsToMigrate.Amount = coinsToMigrate.Amount.ToLegacyDec().Mul(tc.percentOfSharesToMigrate).RoundInt()

			// Modify migration inputs if necessary
			if tc.overwriteLockId {
				originalGammLockId = originalGammLockId + 1
			}

			balancerDelegationPre, _ := stakingKeeper.GetDelegation(s.Ctx, balancerIntermediaryAcc.GetAccAddress(), valAddr)

			// Run the migration logic.
			positionData, migratedPools, concentratedLockId, err := superfluidKeeper.RouteLockedBalancerToConcentratedMigration(s.Ctx, poolJoinAcc, int64(originalGammLockId), coinsToMigrate, tc.minExitCoins)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectedError)
				return
			}
			s.Require().NoError(err)
			s.AssertEventEmitted(s.Ctx, gammtypes.TypeEvtPoolExited, 1)

			s.ValidateMigrateResult(
				positionData.ID, balancerPooId, migratedPools.LeavingID, clPoolId, migratedPools.EnteringID,
				tc.percentOfSharesToMigrate, positionData.Liquidity,
				*balancerLock,
				joinPoolAmt,
				balancerPoolShareOut, coinsToMigrate,
				positionData.Amount0, positionData.Amount1,
			)

			// If the lock was superfluid delegated:
			if tc.superfluidDelegated && !tc.superfluidUndelegating {
				if tc.percentOfSharesToMigrate.Equal(osmomath.OneDec()) {
					// If we migrated all the shares:

					// The intermediary account connection to the old gamm lock should be deleted.
					addr := superfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, originalGammLockId)
					s.Require().Equal(addr.String(), "")

					// The synthetic lockup should be deleted.
					_, err = lockupKeeper.GetSyntheticLockup(s.Ctx, originalGammLockId, keeper.StakingSyntheticDenom(balancerLock.Coins[0].Denom, valAddr.String()))
					s.Require().Error(err)

					// The delegation from the balancer intermediary account holder should not exist.
					delegation, error := stakingKeeper.GetDelegation(s.Ctx, balancerIntermediaryAcc.GetAccAddress(), valAddr)
					s.Require().Error(error, "expected error, found delegation w/ %d shares", delegation.Shares)

					// Check that the original gamm lockup is deleted.
					_, err := s.App.LockupKeeper.GetLockByID(s.Ctx, originalGammLockId)
					s.Require().Error(err)
				} else if tc.percentOfSharesToMigrate.LT(osmomath.OneDec()) {
					// If we migrated part of the shares:
					// The intermediary account connection to the old gamm lock should still be present.
					addr := superfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, originalGammLockId)
					s.Require().Equal(balancerIntermediaryAcc.GetAccAddress().String(), addr.String())

					// Check if migration deleted synthetic lockup.
					_, err = lockupKeeper.GetSyntheticLockup(s.Ctx, originalGammLockId, keeper.StakingSyntheticDenom(balancerLock.Coins[0].Denom, valAddr.String()))
					s.Require().NoError(err)

					// The delegation from the balancer intermediary account holder should still exist.
					delegation, err := stakingKeeper.GetDelegation(s.Ctx, balancerIntermediaryAcc.GetAccAddress(), valAddr)
					s.Require().NoError(err, "expected delegation, got error instead")
					s.Require().Equal(balancerDelegationPre.Shares.Sub(balancerDelegationPre.Shares.Mul(tc.percentOfSharesToMigrate)).RoundInt().String(), delegation.Shares.RoundInt().String(), "expected %d shares, found %d shares", balancerDelegationPre.Shares.Mul(tc.percentOfSharesToMigrate).RoundInt().String(), delegation.Shares.String())

					// Check what is remaining in the original gamm lock.
					lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, originalGammLockId)
					s.Require().NoError(err)
					s.Require().Equal(balancerPoolShareOut.Amount.Sub(coinsToMigrate.Amount).String(), lock.Coins[0].Amount.String(), "expected %s shares, found %s shares", lock.Coins[0].Amount.String(), balancerPoolShareOut.Amount.Sub(coinsToMigrate.Amount).String())
				}
				// Check the new superfluid staked amount.
				clIntermediaryAcc := superfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, concentratedLockId)
				delegation, err := stakingKeeper.GetDelegation(s.Ctx, clIntermediaryAcc, valAddr)
				s.Require().NoError(err, "expected delegation, got error instead")
				s.Require().Equal(balancerDelegationPre.Shares.Mul(tc.percentOfSharesToMigrate).RoundInt().Sub(osmomath.OneInt()).String(), delegation.Shares.RoundInt().String(), "expected %d shares, found %d shares", balancerDelegationPre.Shares.Mul(tc.percentOfSharesToMigrate).RoundInt().String(), delegation.Shares.String())
			}

			// If the lock was superfluid undelegating:
			if tc.superfluidDelegated && tc.superfluidUndelegating {
				// Regardless oh how many shares we migrated:

				// The intermediary account connection to the old gamm lock should be deleted.
				addr := superfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, originalGammLockId)
				s.Require().Equal(addr.String(), "")

				// The synthetic lockup should be deleted.
				_, err = lockupKeeper.GetSyntheticLockup(s.Ctx, originalGammLockId, keeper.StakingSyntheticDenom(balancerLock.Coins[0].Denom, valAddr.String()))
				s.Require().Error(err)

				// The delegation from the intermediary account holder does not exist.
				delegation, err := stakingKeeper.GetDelegation(s.Ctx, balancerIntermediaryAcc.GetAccAddress(), valAddr)
				s.Require().Error(err, "expected error, found delegation w/ %d shares", delegation.Shares)
			}

			// Run slashing logic if the test case involves locks and check if the new and old locks are slashed.
			if !tc.noLock {
				slashExpected := tc.superfluidDelegated || tc.superfluidUndelegating
				s.SlashAndValidateResult(s.Ctx, originalGammLockId, concentratedLockId, migratedPools.EnteringID, tc.percentOfSharesToMigrate, valAddr, *balancerLock, slashExpected)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMigrateSuperfluidBondedBalancerToConcentrated() {
	defaultJoinTime := s.Ctx.BlockTime()
	type sendTest struct {
		overwriteValidatorAddress bool
		overwriteLockId           bool
		percentOfSharesToMigrate  osmomath.Dec
		tokenOutMins              sdk.Coins
		expectedError             error
	}
	testCases := map[string]sendTest{
		"lock that is superfluid delegated, not unlocking (full shares)": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
		},
		"error: migrate more shares than lock has": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1.1"),
			expectedError:            types.MigrateMoreSharesThanLockHasError{SharesToMigrate: "55000000000000000000", SharesInLock: "50000000000000000000"},
		},
		"error: invalid validator address": {
			overwriteValidatorAddress: true,
			percentOfSharesToMigrate:  osmomath.MustNewDecFromStr("1"),
			expectedError:             errors.New("decoding bech32 failed: invalid checksum"),
		},
		"error: non-existent lock ID": {
			overwriteLockId:          true,
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			expectedError:            lockuptypes.ErrLockupNotFound,
		},
		"error: lock that is superfluid delegated, not unlocking (full shares), token out mins is more than exit coins": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			tokenOutMins:             sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(100000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)
			superfluidKeeper := s.App.SuperfluidKeeper
			lockupKeeper := s.App.LockupKeeper
			stakingKeeper := s.App.StakingKeeper

			// We bundle all migration setup into a single function to avoid repeating the same code for each test case.
			joinPoolAmt, balancerIntermediaryAcc, balancerLock, _, poolJoinAcc, balancerPooId, clPoolId, balancerPoolShareOut, valAddr := s.SetupMigrationTest(s.Ctx, true, false, false, false, tc.percentOfSharesToMigrate)
			originalGammLockId := balancerLock.GetID()

			// Depending on the test case, we attempt to migrate a subset of the balancer LP tokens we originally created.
			coinsToMigrate := balancerPoolShareOut
			coinsToMigrate.Amount = coinsToMigrate.Amount.ToLegacyDec().Mul(tc.percentOfSharesToMigrate).RoundInt()

			// GetMigrationType is called via the migration message router and is always run prior to the migration itself.
			// We use it here just to retrieve the synthetic lock before the migration.
			synthLockBeforeMigration, migrationType, err := superfluidKeeper.GetMigrationType(s.Ctx, int64(originalGammLockId))
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

			balancerDelegationPre, _ := stakingKeeper.GetDelegation(s.Ctx, balancerIntermediaryAcc.GetAccAddress(), valAddr)

			// System under test.
			positionData, concentratedLockId, migratedPools, err := superfluidKeeper.MigrateSuperfluidBondedBalancerToConcentrated(s.Ctx, poolJoinAcc, originalGammLockId, coinsToMigrate, synthLockBeforeMigration.SynthDenom, tc.tokenOutMins)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}
			s.Require().NoError(err)
			s.AssertEventEmitted(s.Ctx, gammtypes.TypeEvtPoolExited, 1)

			s.ValidateMigrateResult(
				positionData.ID, balancerPooId, migratedPools.LeavingID, clPoolId, migratedPools.EnteringID,
				tc.percentOfSharesToMigrate, positionData.Liquidity,
				*balancerLock,
				joinPoolAmt,
				balancerPoolShareOut, coinsToMigrate,
				positionData.Amount0, positionData.Amount1,
			)

			if tc.percentOfSharesToMigrate.Equal(osmomath.OneDec()) {
				// If we migrated all the shares:

				// The intermediary account connection to the old gamm lock should be deleted.
				addr := superfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, originalGammLockId)
				s.Require().Equal(addr.String(), "")

				// The synthetic lockup should be deleted.
				_, err = lockupKeeper.GetSyntheticLockup(s.Ctx, originalGammLockId, keeper.StakingSyntheticDenom(balancerLock.Coins[0].Denom, valAddr.String()))
				s.Require().Error(err)

				// The delegation from the intermediary account holder should not exist.
				delegation, err := stakingKeeper.GetDelegation(s.Ctx, balancerIntermediaryAcc.GetAccAddress(), valAddr)
				s.Require().Error(err, "expected error, found delegation w/ %d shares", delegation.Shares)

				// Check that the original gamm lockup is deleted.
				_, err = s.App.LockupKeeper.GetLockByID(s.Ctx, originalGammLockId)
				s.Require().Error(err)
			} else if tc.percentOfSharesToMigrate.LT(osmomath.OneDec()) {
				// If we migrated part of the shares:
				// The intermediary account connection to the old gamm lock should still be present.
				addr := superfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, originalGammLockId)
				s.Require().Equal(balancerIntermediaryAcc.GetAccAddress().String(), addr.String())

				// Confirm that migration did not delete synthetic lockup.
				gammSynthLock, err := lockupKeeper.GetSyntheticLockup(s.Ctx, originalGammLockId, keeper.StakingSyntheticDenom(balancerLock.Coins[0].Denom, valAddr.String()))
				s.Require().NoError(err)

				s.Require().Equal(originalGammLockId, gammSynthLock.UnderlyingLockId)

				// The delegation from the intermediary account holder should still exist.
				_, err = stakingKeeper.GetDelegation(s.Ctx, balancerIntermediaryAcc.GetAccAddress(), valAddr)
				s.Require().NoError(err, "expected delegation, got error instead")

				// Check what is remaining in the original gamm lock.
				lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, originalGammLockId)
				s.Require().NoError(err)
				s.Require().Equal(balancerPoolShareOut.Amount.Sub(coinsToMigrate.Amount).String(), lock.Coins[0].Amount.String(), "expected %s shares, found %s shares", lock.Coins[0].Amount.String(), balancerPoolShareOut.Amount.Sub(coinsToMigrate.Amount).String())
			}
			// Check the new superfluid staked amount.
			clIntermediaryAcc := superfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, concentratedLockId)
			delegation, err := stakingKeeper.GetDelegation(s.Ctx, clIntermediaryAcc, valAddr)
			s.Require().NoError(err, "expected delegation, got error instead")
			s.Require().Equal(balancerDelegationPre.Shares.Mul(tc.percentOfSharesToMigrate).RoundInt().Sub(osmomath.OneInt()).String(), delegation.Shares.RoundInt().String(), "expected %d shares, found %d shares", balancerDelegationPre.Shares.Mul(tc.percentOfSharesToMigrate).RoundInt().String(), delegation.Shares.String())

			// Check if the new intermediary account connection was created.
			newConcentratedIntermediaryAccount := superfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, concentratedLockId)
			s.Require().NotEqual(newConcentratedIntermediaryAccount.String(), "")

			// Check newly created concentrated lock.
			concentratedLock, err := lockupKeeper.GetLockByID(s.Ctx, concentratedLockId)
			s.Require().NoError(err)

			s.Require().Equal(positionData.Liquidity.TruncateInt().String(), concentratedLock.Coins[0].Amount.String(), "expected %s shares, found %s shares", coinsToMigrate.Amount.String(), concentratedLock.Coins[0].Amount.String())
			s.Require().Equal(balancerLock.Duration, concentratedLock.Duration)
			s.Require().Equal(balancerLock.EndTime, concentratedLock.EndTime)

			// Check if the new synthetic bonded lockup was created.
			clSynthLock, err := lockupKeeper.GetSyntheticLockup(s.Ctx, concentratedLockId, keeper.StakingSyntheticDenom(concentratedLock.Coins[0].Denom, valAddr.String()))
			s.Require().NoError(err)

			s.Require().Equal(concentratedLockId, clSynthLock.UnderlyingLockId)

			// Run slashing logic and check if the new and old locks are slashed.
			s.SlashAndValidateResult(s.Ctx, originalGammLockId, concentratedLockId, clPoolId, tc.percentOfSharesToMigrate, valAddr, *balancerLock, true)
		})
	}
}

func (s *KeeperTestSuite) TestMigrateSuperfluidUnbondingBalancerToConcentrated() {
	defaultJoinTime := s.Ctx.BlockTime()
	type sendTest struct {
		unlocking                 bool
		overwriteValidatorAddress bool
		percentOfSharesToMigrate  osmomath.Dec
		tokenOutMins              sdk.Coins
		expectedError             error
	}
	testCases := map[string]sendTest{
		"lock that is superfluid undelegating, not unlocking (full shares)": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
		},
		"lock that is superfluid undelegating, unlocking (full shares)": {
			unlocking:                true,
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
		},
		"error: invalid validator address": {
			overwriteValidatorAddress: true,
			percentOfSharesToMigrate:  osmomath.MustNewDecFromStr("1"),
			expectedError:             errors.New("decoding bech32 failed: invalid checksum"),
		},
		"error: lock that is superfluid undelegating, not unlocking (full shares), token out mins is more than exit coins": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			tokenOutMins:             sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(100000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)
			superfluidKeeper := s.App.SuperfluidKeeper
			lockupKeeper := s.App.LockupKeeper

			// We bundle all migration setup into a single function to avoid repeating the same code for each test case.
			joinPoolAmt, _, balancerLock, _, poolJoinAcc, balancerPooId, clPoolId, balancerPoolShareOut, valAddr := s.SetupMigrationTest(s.Ctx, true, true, tc.unlocking, false, tc.percentOfSharesToMigrate)
			originalGammLockId := balancerLock.GetID()

			// Since we are testing unbonding migration, we let some time pass (24 hours) since unbonding to simulate a more realistic environment.
			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Hour * 24))

			// Depending on the test case, we attempt to migrate a subset of the balancer LP tokens we originally created.
			coinsToMigrate := balancerPoolShareOut
			coinsToMigrate.Amount = coinsToMigrate.Amount.ToLegacyDec().Mul(tc.percentOfSharesToMigrate).RoundInt()

			// GetMigrationType is called via the migration message router and is always run prior to the migration itself
			synthLockBeforeMigration, migrationType, err := superfluidKeeper.GetMigrationType(s.Ctx, int64(originalGammLockId))
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
			positionData, concentratedLockId, migratedPools, err := superfluidKeeper.MigrateSuperfluidUnbondingBalancerToConcentrated(s.Ctx, poolJoinAcc, originalGammLockId, coinsToMigrate, synthLockBeforeMigration.SynthDenom, tc.tokenOutMins)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}
			s.Require().NoError(err)
			s.AssertEventEmitted(s.Ctx, gammtypes.TypeEvtPoolExited, 1)

			s.ValidateMigrateResult(
				positionData.ID, balancerPooId, migratedPools.LeavingID, clPoolId, migratedPools.EnteringID,
				tc.percentOfSharesToMigrate, positionData.Liquidity,
				*balancerLock,
				joinPoolAmt,
				balancerPoolShareOut, coinsToMigrate,
				positionData.Amount0, positionData.Amount1,
			)

			if tc.percentOfSharesToMigrate.Equal(osmomath.OneDec()) {
				// If we migrated all the shares:

				// The synthetic lockup should be deleted.
				_, err = lockupKeeper.GetSyntheticLockup(s.Ctx, originalGammLockId, keeper.UnstakingSyntheticDenom(balancerLock.Coins[0].Denom, valAddr.String()))
				s.Require().Error(err)
			} else if tc.percentOfSharesToMigrate.LT(osmomath.OneDec()) {
				// If we migrated part of the shares:

				// The synthetic lockup should not be deleted.
				gammSynthLock, err := lockupKeeper.GetSyntheticLockup(s.Ctx, originalGammLockId, keeper.UnstakingSyntheticDenom(balancerLock.Coins[0].Denom, valAddr.String()))
				s.Require().NoError(err)

				s.Require().Equal(originalGammLockId, gammSynthLock.UnderlyingLockId)
			}

			// Check newly created concentrated lock.
			concentratedLock, err := lockupKeeper.GetLockByID(s.Ctx, concentratedLockId)
			s.Require().NoError(err)
			s.Require().Equal(positionData.Liquidity.TruncateInt().String(), concentratedLock.Coins[0].Amount.String(), "expected %s shares, found %s shares", coinsToMigrate.Amount.String(), concentratedLock.Coins[0].Amount.String())
			// If the original lock was not unlocking, then the new lock should have the same duration and end time (regardless of the fact that an hour has passed since the lock began superfluid unbonding).
			expectedConcentratedLockDuration := balancerLock.Duration
			expectedConcentratedLockEndTime := s.Ctx.BlockTime().Add(balancerLock.Duration)
			if tc.unlocking {
				// If the original lock was unlocking, then the new lock should have a duration of 24 hours less than the original lock, and an end time of 24 hours less than the original lock.
				expectedConcentratedLockDuration = expectedConcentratedLockDuration - time.Hour*24
				expectedConcentratedLockEndTime = expectedConcentratedLockEndTime.Add(-time.Hour * 24)
			}
			s.Require().Equal(expectedConcentratedLockDuration, concentratedLock.Duration)
			s.Require().Equal(expectedConcentratedLockEndTime, concentratedLock.EndTime)

			// Check if the new synthetic unbonding lockup was created.
			clSynthLock, err := lockupKeeper.GetSyntheticLockup(s.Ctx, concentratedLockId, keeper.UnstakingSyntheticDenom(concentratedLock.Coins[0].Denom, valAddr.String()))
			s.Require().NoError(err)
			s.Require().Equal(concentratedLockId, clSynthLock.UnderlyingLockId)
			s.Require().NoError(err)
			s.Require().Equal(concentratedLock.Duration, clSynthLock.Duration)
			s.Require().Equal(concentratedLock.EndTime, clSynthLock.EndTime)

			// Run slashing logic and check if the new and old locks are slashed.
			s.SlashAndValidateResult(s.Ctx, originalGammLockId, concentratedLockId, clPoolId, tc.percentOfSharesToMigrate, valAddr, *balancerLock, true)
		})
	}
}

func (s *KeeperTestSuite) TestMigrateNonSuperfluidLockBalancerToConcentrated() {
	defaultJoinTime := s.Ctx.BlockTime()
	type sendTest struct {
		unlocking                bool
		percentOfSharesToMigrate osmomath.Dec
		tokenOutMins             sdk.Coins
		expectedError            error
	}
	testCases := map[string]sendTest{
		"lock that is not superfluid delegated, not unlocking (full shares)": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
		},
		"lock that is not superfluid delegated, unlocking (full shares)": {
			unlocking:                true,
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
		},
		"error: lock that is not superfluid delegated, not unlocking (full shares), token out mins is more than exit coins": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			tokenOutMins:             sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(10000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)
			superfluidKeeper := s.App.SuperfluidKeeper
			lockupKeeper := s.App.LockupKeeper

			// We bundle all migration setup into a single function to avoid repeating the same code for each test case.
			joinPoolAmt, _, balancerLock, _, poolJoinAcc, balancerPooId, clPoolId, balancerPoolShareOut, valAddr := s.SetupMigrationTest(s.Ctx, false, false, tc.unlocking, false, tc.percentOfSharesToMigrate)
			originalGammLockId := balancerLock.GetID()

			// Depending on the test case, we attempt to migrate a subset of the balancer LP tokens we originally created.
			coinsToMigrate := balancerPoolShareOut
			coinsToMigrate.Amount = coinsToMigrate.Amount.ToLegacyDec().Mul(tc.percentOfSharesToMigrate).RoundInt()

			// GetMigrationType is called via the migration message router and is always run prior to the migration itself
			synthLockBeforeMigration, migrationType, err := superfluidKeeper.GetMigrationType(s.Ctx, int64(originalGammLockId))
			s.Require().NoError(err)
			s.Require().Equal((lockuptypes.SyntheticLock{}), synthLockBeforeMigration)
			s.Require().Equal(migrationType, keeper.NonSuperfluid)

			// System under test.
			positionData, concentratedLockId, migratedPools, err := superfluidKeeper.MigrateNonSuperfluidLockBalancerToConcentrated(s.Ctx, poolJoinAcc, originalGammLockId, coinsToMigrate, tc.tokenOutMins)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}
			s.Require().NoError(err)
			s.AssertEventEmitted(s.Ctx, gammtypes.TypeEvtPoolExited, 1)

			s.ValidateMigrateResult(
				positionData.ID, balancerPooId, migratedPools.LeavingID, clPoolId, migratedPools.EnteringID,
				tc.percentOfSharesToMigrate, positionData.Liquidity,
				*balancerLock,
				joinPoolAmt,
				balancerPoolShareOut, coinsToMigrate,
				positionData.Amount0, positionData.Amount1,
			)

			// Check newly created concentrated lock.
			concentratedLock, err := lockupKeeper.GetLockByID(s.Ctx, concentratedLockId)
			s.Require().NoError(err)
			s.Require().Equal(positionData.Liquidity.TruncateInt().String(), concentratedLock.Coins[0].Amount.String(), "expected %s shares, found %s shares", coinsToMigrate.Amount.String(), concentratedLock.Coins[0].Amount.String())
			s.Require().Equal(balancerLock.Duration, concentratedLock.Duration)
			s.Require().Equal(s.Ctx.BlockTime().Add(balancerLock.Duration), concentratedLock.EndTime)

			// Run slashing logic and check if the new and old locks are not slashed.
			s.SlashAndValidateResult(s.Ctx, originalGammLockId, concentratedLockId, clPoolId, tc.percentOfSharesToMigrate, valAddr, *balancerLock, false)
		})
	}
}

func (s *KeeperTestSuite) TestMigrateUnlockedPositionFromBalancerToConcentrated() {
	defaultJoinTime := s.Ctx.BlockTime()
	type sendTest struct {
		percentOfSharesToMigrate osmomath.Dec
		tokenOutMins             sdk.Coins
		expectedError            error
	}
	testCases := map[string]sendTest{
		"no lock (full shares)": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
		},
		"no lock (partial shares)": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("0.9"),
		},
		"no lock (more shares than own)": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1.1"),
			expectedError:            errors.New("insufficient funds"),
		},
		"no lock (no shares)": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("0"),
			expectedError:            errorsmod.Wrapf(gammtypes.ErrInvalidMathApprox, "Trying to exit a negative amount of shares"),
		},
		"error: no lock (full shares), token out mins is more than exit coins": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			tokenOutMins:             sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(10000))),
			expectedError:            gammtypes.ErrLimitMinAmount,
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)
			superfluidKeeper := s.App.SuperfluidKeeper
			gammKeeper := s.App.GAMMKeeper

			// We bundle all migration setup into a single function to avoid repeating the same code for each test case.
			joinPoolAmt, _, balancerLock, _, poolJoinAcc, balancerPooId, clPoolId, balancerPoolShareOut, _ := s.SetupMigrationTest(s.Ctx, false, false, false, true, tc.percentOfSharesToMigrate)
			s.Require().Equal(uint64(0), balancerLock.GetID())

			// Depending on the test case, we attempt to migrate a subset of the balancer LP tokens we originally created.
			coinsToMigrate := balancerPoolShareOut
			coinsToMigrate.Amount = coinsToMigrate.Amount.ToLegacyDec().Mul(tc.percentOfSharesToMigrate).RoundInt()

			// GetMigrationType is called via the migration message router and is always run prior to the migration itself
			synthLockBeforeMigration, migrationType, err := superfluidKeeper.GetMigrationType(s.Ctx, 0)
			s.Require().NoError(err)
			s.Require().Equal((lockuptypes.SyntheticLock{}), synthLockBeforeMigration)
			s.Require().Equal(migrationType, keeper.Unlocked)

			// System under test.
			positionData, migratedPools, err := gammKeeper.MigrateUnlockedPositionFromBalancerToConcentrated(s.Ctx, poolJoinAcc, coinsToMigrate, tc.tokenOutMins)
			if tc.expectedError != nil {
				s.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}
			s.Require().NoError(err)
			s.AssertEventEmitted(s.Ctx, gammtypes.TypeEvtPoolExited, 1)

			s.ValidateMigrateResult(
				positionData.ID, balancerPooId, migratedPools.LeavingID, clPoolId, migratedPools.EnteringID,
				tc.percentOfSharesToMigrate, positionData.Liquidity,
				*balancerLock,
				joinPoolAmt,
				balancerPoolShareOut, coinsToMigrate,
				positionData.Amount0, positionData.Amount1,
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
		overwriteSender           bool
		overwriteSharesDenomValue string
		overwriteLockId           bool
		percentOfSharesToMigrate  osmomath.Dec
		expectedError             error
	}
	testCases := map[string]sendTest{
		"lock that is not superfluid delegated, not unlocking (full shares)": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			isSuperfluidDelegated:    false,
			isSuperfluidUndelegating: false,
		},
		"lock that is not superfluid delegated, not unlocking (partial shares)": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("0.9"),
			isSuperfluidDelegated:    false,
			isSuperfluidUndelegating: false,
		},
		"lock that is not superfluid delegated, unlocking (full shares)": {
			unlocking:                true,
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			isSuperfluidDelegated:    false,
			isSuperfluidUndelegating: false,
		},
		"lock that is superfluid undelegating, not unlocking (full shares)": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			isSuperfluidDelegated:    true,
			isSuperfluidUndelegating: true,
		},
		"lock that is superfluid undelegating, unlocking (full shares)": {
			unlocking:                true,
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			isSuperfluidDelegated:    true,
			isSuperfluidUndelegating: true,
		},
		"lock that is superfluid delegated, not unlocking (full shares)": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			isSuperfluidDelegated:    true,
		},
		"error: denom prefix error": {
			overwriteSharesDenomValue: "cl/pool/2",
			percentOfSharesToMigrate:  osmomath.MustNewDecFromStr("1"),
			expectedError:             types.SharesToMigrateDenomPrefixError{Denom: "cl/pool/2", ExpectedDenomPrefix: gammtypes.GAMMTokenPrefix},
		},
		"error: no canonical link": {
			overwriteSharesDenomValue: "gamm/pool/2",
			percentOfSharesToMigrate:  osmomath.MustNewDecFromStr("1"),
			expectedError:             gammtypes.ConcentratedPoolMigrationLinkNotFoundError{PoolIdLeaving: 2},
		},
		"error: wrong sender": {
			overwriteSender:          true,
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			expectedError:            lockuptypes.ErrNotLockOwner,
		},
		"error: wrong lock ID": {
			overwriteLockId:          true,
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
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
			coinsToMigrate.Amount = coinsToMigrate.Amount.ToLegacyDec().Mul(tc.percentOfSharesToMigrate).RoundInt()

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
			migratedPools, preMigrationLock, remainingLockTime, err := superfluidKeeper.ValidateMigration(ctx, poolJoinAcc, originalGammLockId, coinsToMigrate)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(migratedPools.LeavingID, balancerPooId)
			s.Require().Equal(migratedPools.EnteringID, clPoolId)
			s.Require().Equal(preMigrationLock.GetID(), originalGammLockId)
			s.Require().Equal(preMigrationLock.GetCoins(), sdk.NewCoins(balancerPoolShareOut))
			s.Require().Equal(preMigrationLock.GetDuration(), remainingLockTime)
		})
	}
}

func (s *KeeperTestSuite) TestForceUnlockAndExitBalancerPool() {
	defaultJoinTime := s.Ctx.BlockTime()
	type sendTest struct {
		overwritePreMigrationLock bool
		overwriteShares           bool
		overwritePool             bool
		overwritePoolId           bool
		exitCoinsLengthIsTwo      bool
		percentOfSharesToMigrate  osmomath.Dec
		tokenOutMins              sdk.Coins
		expectedError             error
	}
	testCases := map[string]sendTest{
		"happy path (full shares)": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
		},
		"happy path (partial shares)": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("0.4"),
			expectedError:            types.MigratePartialSharesError{SharesToMigrate: "20000000000000000000", SharesInLock: "50000000000000000000"},
		},
		"attempt to leave a pool that has more than two denoms": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			overwritePool:            true,
			exitCoinsLengthIsTwo:     false,
		},
		"error: lock does not exist": {
			percentOfSharesToMigrate:  osmomath.MustNewDecFromStr("1"),
			overwritePreMigrationLock: true,
			expectedError:             lockuptypes.ErrLockupNotFound,
		},
		"error: attempt to migrate more than lock has": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			overwriteShares:          true,
			expectedError:            types.MigrateMoreSharesThanLockHasError{SharesToMigrate: "50000000000000000001", SharesInLock: "50000000000000000000"},
		},
		"error: attempt to leave a pool that does not exist": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			overwritePoolId:          true,
			expectedError:            fmt.Errorf("pool with ID %d does not exist", 2),
		},
		"error: attempt to leave a pool that has more than two denoms with exitCoinsLengthIsTwo true": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			overwritePool:            true,
			exitCoinsLengthIsTwo:     true,
			expectedError:            types.TwoTokenBalancerPoolError{NumberOfTokens: 4},
		},
		"error: happy path (full shares), token out mins is more than exit coins": {
			percentOfSharesToMigrate: osmomath.MustNewDecFromStr("1"),
			tokenOutMins:             sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(100000))),
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
				err := testutil.FundAccount(ctx, bankKeeper, acc, defaultAcctFunds)
				s.Require().NoError(err)
			}

			// Create a balancer pool of "stake" and "foo".
			msg := balancer.NewMsgCreateBalancerPool(poolCreateAcc, balancer.PoolParams{
				SwapFee: osmomath.NewDecWithPrec(1, 2),
				ExitFee: osmomath.NewDec(0),
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

			sharesToMigrate := balancerPoolShareOut.Amount.ToLegacyDec().Mul(tc.percentOfSharesToMigrate).TruncateInt()
			coinsToMigrate := sdk.NewCoin(balancerPoolDenom, sharesToMigrate)

			// The unbonding duration is the same as the staking module's unbonding duration.
			stakingParams, err := stakingKeeper.GetParams(ctx)
			s.Require().NoError(err)
			unbondingDuration := stakingParams.UnbondingTime

			// Lock the LP tokens for the duration of the unbonding period.
			originalGammLockId := s.LockTokens(poolJoinAcc, sdk.NewCoins(balancerPoolShareOut), unbondingDuration)

			lock, err := lockupKeeper.GetLockByID(ctx, originalGammLockId)
			s.Require().NoError(err)

			if tc.overwritePreMigrationLock {
				lock.ID = lock.ID + 1
			}

			if tc.overwriteShares {
				coinsToMigrate.Amount = lock.Coins[0].Amount.Add(osmomath.NewInt(1))
			}

			if tc.overwritePool {
				multiCoinBalancerPoolId := s.PrepareBalancerPool()
				balancerPooId = multiCoinBalancerPoolId
				shareAmt := osmomath.MustNewDecFromStr("50000000000000000000").TruncateInt()
				newShares := sdk.NewCoin(fmt.Sprintf("gamm/pool/%d", multiCoinBalancerPoolId), shareAmt)
				s.FundAcc(poolJoinAcc, sdk.NewCoins(newShares))
			}

			if tc.overwritePoolId {
				balancerPooId = balancerPooId + 1
			}

			// System under test
			exitCoins, err := superfluidKeeper.ForceUnlockAndExitBalancerPool(ctx, poolJoinAcc, balancerPooId, lock, coinsToMigrate, tc.tokenOutMins, tc.exitCoinsLengthIsTwo)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}
			s.Require().NoError(err)

			defaultErrorTolerance := osmomath.ErrTolerance{
				AdditiveTolerance: osmomath.NewDec(1),
				RoundingDir:       osmomath.RoundDown,
			}

			if tc.percentOfSharesToMigrate.Equal(osmomath.OneDec()) {
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

			if tc.exitCoinsLengthIsTwo {
				for _, coin := range exitCoins {
					// Check that the exit coin is the same amount that we joined with (with one unit rounding down)
					osmoassert.Equal(s.T(), defaultErrorTolerance, tokensIn.AmountOf(coin.Denom).ToLegacyDec().Mul(tc.percentOfSharesToMigrate).RoundInt(), coin.Amount)
				}
			}
		})
	}
}

func (s *KeeperTestSuite) SetupMigrationTest(ctx sdk.Context, superfluidDelegated, superfluidUndelegating, unlocking, noLock bool, percentOfSharesToMigrate osmomath.Dec) (joinPoolAmt sdk.Coins, balancerIntermediaryAcc types.SuperfluidIntermediaryAccount, balancerLock *lockuptypes.PeriodLock, poolCreateAcc, poolJoinAcc sdk.AccAddress, balancerPooId, clPoolId uint64, balancerPoolShareOut sdk.Coin, valAddr sdk.ValAddress) { //nolint:revive // TODO: refactor this function
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
		err := testutil.FundAccount(ctx, bankKeeper, acc, defaultAcctFunds)
		s.Require().NoError(err)
	}

	// Set up a single validator.
	valAddr = s.SetupValidator(stakingtypes.Bonded)

	// Create a balancer pool of "stake" and "foo".
	msg := balancer.NewMsgCreateBalancerPool(poolCreateAcc, balancer.PoolParams{
		SwapFee: osmomath.NewDecWithPrec(1, 2),
		ExitFee: osmomath.NewDec(0),
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
	joinPoolAmt, _ = balanceBeforeJoin.SafeSub(balanceAfterJoin...)

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
	clPool := s.PrepareCustomConcentratedPool(poolCreateAcc, defaultPoolAssets[0].Token.Denom, defaultPoolAssets[1].Token.Denom, 1, osmomath.ZeroDec())
	clPoolId = clPool.GetId()

	// Add a gov sanctioned link between the balancer and concentrated liquidity pool.
	migrationRecord := gammmigration.MigrationRecords{BalancerToConcentratedPoolLinks: []gammmigration.BalancerToConcentratedPoolLink{
		{BalancerPoolId: balancerPooId, ClPoolId: clPoolId},
	}}
	err = gammKeeper.OverwriteMigrationRecords(ctx, migrationRecord)
	s.Require().NoError(err)

	// The unbonding duration is the same as the staking module's unbonding duration.
	stakingParams, err := stakingKeeper.GetParams(ctx)
	unbondingDuration := stakingParams.UnbondingTime

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

func (s *KeeperTestSuite) SlashAndValidateResult(ctx sdk.Context, gammLockId, concentratedLockId, poolIdEntering uint64, percentOfSharesToMigrate osmomath.Dec, valAddr sdk.ValAddress, balancerLock lockuptypes.PeriodLock, expectSlash bool) {
	// Retrieve the concentrated lock and gamm lock prior to slashing.
	concentratedLockPreSlash, err := s.App.LockupKeeper.GetLockByID(ctx, concentratedLockId)
	s.Require().NoError(err)
	gammLockPreSlash, err := s.App.LockupKeeper.GetLockByID(ctx, gammLockId)
	if percentOfSharesToMigrate.LT(osmomath.OneDec()) {
		s.Require().NoError(err)
	} else {
		s.Require().Error(err)
	}

	// Slash the validator.
	slashFactor := osmomath.NewDecWithPrec(5, 2)
	s.App.SuperfluidKeeper.SlashLockupsForValidatorSlash(
		ctx,
		valAddr,
		slashFactor)

	// Retrieve the concentrated lock and gamm lock after slashing.
	concentratedLockPostSlash, err := s.App.LockupKeeper.GetLockByID(ctx, concentratedLockId)
	s.Require().NoError(err)
	gammLockPostSlash, err := s.App.LockupKeeper.GetLockByID(ctx, gammLockId)
	if percentOfSharesToMigrate.LT(osmomath.OneDec()) {
		s.Require().NoError(err)
	} else {
		s.Require().Error(err)
	}

	// Check if the concentrated lock was slashed.
	clDenom := cltypes.GetConcentratedLockupDenomFromPoolId(poolIdEntering)
	slashAmtCL := concentratedLockPreSlash.Coins.AmountOf(clDenom).ToLegacyDec().Mul(slashFactor).TruncateInt()
	if !expectSlash {
		slashAmtCL = osmomath.ZeroInt()
	}
	s.Require().Equal(concentratedLockPreSlash.Coins.AmountOf(clDenom).Sub(slashAmtCL).String(), concentratedLockPostSlash.Coins.AmountOf(clDenom).String())

	// Check if the gamm lock was slashed.
	// We only check if the gamm lock was slashed if the lock was not migrated entirely.
	// Otherwise, there would be no newly created gamm lock to check.
	if percentOfSharesToMigrate.LT(osmomath.OneDec()) {
		gammDenom := balancerLock.Coins[0].Denom
		slashAmtGamm := gammLockPreSlash.Coins.AmountOf(gammDenom).ToLegacyDec().Mul(slashFactor).TruncateInt()
		if !expectSlash {
			slashAmtGamm = osmomath.ZeroInt()
		}
		s.Require().Equal(gammLockPreSlash.Coins.AmountOf(gammDenom).Sub(slashAmtGamm).String(), gammLockPostSlash.Coins.AmountOf(gammDenom).String())
	}
}

// TODO add user balance pre swap and then add to result

func (s *KeeperTestSuite) ValidateMigrateResult(
	positionId, balancerPooId, poolIdLeaving, clPoolId, poolIdEntering uint64,
	percentOfSharesToMigrate, liquidityMigrated osmomath.Dec,
	balancerLock lockuptypes.PeriodLock,
	joinPoolAmt sdk.Coins,
	balancerPoolShareOut, coinsToMigrate sdk.Coin,
	amount0, amount1 osmomath.Int,
) {
	// Check that the concentrated liquidity and join time match what we expect
	position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionId)
	s.Require().NoError(err)
	s.Require().Equal(liquidityMigrated, position.Liquidity)
	s.Require().Equal(s.Ctx.BlockTime(), position.JoinTime)

	// Expect the poolIdLeaving to be the balancer pool id
	// Expect the poolIdEntering to be the concentrated liquidity pool id
	s.Require().Equal(balancerPooId, poolIdLeaving)
	s.Require().Equal(clPoolId, poolIdEntering)

	// exitPool has rounding difference.
	// We test if correct amt has been exited and frozen by comparing with rounding tolerance.
	defaultErrorTolerance := osmomath.ErrTolerance{
		AdditiveTolerance: osmomath.NewDec(2),
		RoundingDir:       osmomath.RoundDown,
	}
	osmoassert.Equal(s.T(), defaultErrorTolerance, joinPoolAmt.AmountOf(defaultPoolAssets[0].Token.Denom).ToLegacyDec().Mul(percentOfSharesToMigrate).RoundInt(), amount0)
	osmoassert.Equal(s.T(), defaultErrorTolerance, joinPoolAmt.AmountOf(defaultPoolAssets[1].Token.Denom).ToLegacyDec().Mul(percentOfSharesToMigrate).RoundInt(), amount1)
}

type Positions struct {
	numAccounts                     int
	numBondedSuperfluid             int
	numUnbondingSuperfluidLocked    int
	numUnbondingSuperfluidUnlocking int
	numVanillaLockLocked            int
	numVanillaLockUnlocking         int
	numNoLock                       int
}

type positionInfo struct {
	joinPoolCoins sdk.Coins
	coin          sdk.Coin
	shares        osmomath.Int
	lockId        uint64
}

type PositionType int

const (
	BondedSuperfluid PositionType = iota
	UnbondingSuperfluidLocked
	UnbondingSuperfluidUnlocking
	VanillaLockLocked
	VanillaLockUnlocking
	NoLock
)

// TestFunctional_VaryingPositions_Migrations proves that, given a set of balancer pool positions of various types, the migration process works as expected.
// By "works as expected", we mean that after we migrate a position, the funds that:
// - get moved to the new cl positions
// - get left behind in the balancer pool
// - get sent back to the user
// all add up to the amount that the user originally had in the balancer pool.
//
// This test also asserts this same invariant at the very end, to ensure that all coins the accounts were funded with are accounted for.
func (s *KeeperTestSuite) TestFunctional_VaryingPositions_Migrations() {
	for i := 0; i < 5; i++ {
		seed := time.Now().UnixNano() + int64(i)
		s.Run(fmt.Sprintf("seed %d", seed), func() {
			r := rand.New(rand.NewSource(seed))

			// Generate random value from 0 to 50 for each position field
			// This is how many positions of each type we will create and migrate
			maxValue := 33
			numBondedSuperfluid := r.Intn(maxValue)
			numUnbondingSuperfluidLocked := r.Intn(maxValue)
			numUnbondingSuperfluidUnlocking := r.Intn(maxValue)
			numVanillaLockLocked := r.Intn(maxValue)
			numVanillaLockUnlocking := r.Intn(maxValue)
			numNoLock := r.Intn(maxValue)

			// Find the largest numPosition value and set numAccounts to be one greater than the largest position value
			// The first account is used to create pools and the rest are used to create positions
			largestPositionValue := osmoutils.Max(numBondedSuperfluid, numUnbondingSuperfluidLocked, numUnbondingSuperfluidUnlocking, numVanillaLockLocked, numVanillaLockUnlocking, numNoLock)
			largestPositionValueInt, ok := largestPositionValue.(int)
			s.Require().True(ok)
			numAccounts := largestPositionValueInt + 1

			positions := Positions{
				numAccounts:                     numAccounts,
				numBondedSuperfluid:             numBondedSuperfluid,
				numUnbondingSuperfluidLocked:    numUnbondingSuperfluidLocked,
				numUnbondingSuperfluidUnlocking: numUnbondingSuperfluidUnlocking,
				numVanillaLockLocked:            numVanillaLockLocked,
				numVanillaLockUnlocking:         numVanillaLockUnlocking,
				numNoLock:                       numNoLock,
			}

			s.SetupTest()
			s.TestAccs = apptesting.CreateRandomAccounts(positions.numAccounts)

			// Create a balancer pool (includes staking denom to be superfluid compatible).
			balancerPoolId := s.PrepareBalancerPoolWithCoins(DefaultCoins...)
			balancerPoolShareDenom := fmt.Sprintf("gamm/pool/%d", balancerPoolId)
			stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
			unbondingDuration := stakingParams.UnbondingTime

			positionInfos := make([][]positionInfo, 6)
			maxDivisor := 10

			// Create Balancer pool with multiple positions
			// Positions should be spread between:
			// - Bonded superfluid
			// - Unbonded superfluid (locked)
			// - Unbonded superfluid (unlocking)
			// - Vanilla lock (locked)
			// - Vanilla lock (unlocking)
			// - No lock
			//
			// After each position is created, we track the position info in a slice of slices.
			// This allows us to easily iterate through all positions and test migration for each position later.
			//
			// Note: you may notice we add one nanosecond to the lock duration between position types. This is so
			// we don't reuse the same lock, since under the hood if an account tries to lock twice with the same
			// lock duration, an "add to lock" call will happen rather than creating a new lock.

			totalFundsForPositionCreation := sdk.NewCoins()
			createPositions := func(posType PositionType, numPositions int, lockDurationFn func(int) time.Duration, superfluidDelegate bool, callbackFn func(int, positionInfo)) {
				for i := 0; i < numPositions; i++ {
					index := i + 1                              // Skip the first account, which is used for pool creation
					divisor := int64(rand.Intn(maxDivisor)) + 1 // Randomly generate a divisor between 1 and maxDivisor
					coin0Amt := DefaultAmt0.QuoRaw(divisor)     // Divide amount0 by the divisor to add some entropy to the position creation
					coin1Amt := DefaultAmt1.QuoRaw(divisor)     // Divide amount1 by the divisor to add some entropy to the position creation
					positionCoins := sdk.NewCoins(sdk.NewCoin(DefaultCoin0.Denom, coin0Amt), sdk.NewCoin(DefaultCoin1.Denom, coin1Amt))
					s.FundAcc(s.TestAccs[index], positionCoins)
					totalFundsForPositionCreation = totalFundsForPositionCreation.Add(positionCoins...) // Track total funds used for position creation, to be used by invariant checks later
					posInfoInternal := s.createBalancerPosition(s.TestAccs[index], balancerPoolId, lockDurationFn(i), balancerPoolShareDenom, positionCoins, superfluidDelegate)
					positionInfos[posType] = append(positionInfos[posType], posInfoInternal) // Track position info for invariant checks later
					callbackFn(index, posInfoInternal)
				}
			}

			createPositions(BondedSuperfluid, positions.numBondedSuperfluid, func(int) time.Duration { return unbondingDuration }, true, func(index int, posInfoInternal positionInfo) {
			})

			createPositions(UnbondingSuperfluidLocked, positions.numUnbondingSuperfluidLocked, func(int) time.Duration { return unbondingDuration + time.Nanosecond }, true, func(index int, posInfoInternal positionInfo) {
				err := s.App.SuperfluidKeeper.SuperfluidUndelegate(s.Ctx, s.TestAccs[index].String(), posInfoInternal.lockId)
				s.Require().NoError(err)
			})

			createPositions(UnbondingSuperfluidUnlocking, positions.numUnbondingSuperfluidUnlocking, func(int) time.Duration { return unbondingDuration + time.Nanosecond*2 }, true, func(index int, posInfoInternal positionInfo) {
				err := s.App.SuperfluidKeeper.SuperfluidUndelegate(s.Ctx, s.TestAccs[index].String(), posInfoInternal.lockId)
				s.Require().NoError(err)
				err = s.App.SuperfluidKeeper.SuperfluidUnbondLock(s.Ctx, posInfoInternal.lockId, s.TestAccs[index].String())
				s.Require().NoError(err)
			})

			createPositions(VanillaLockLocked, positions.numVanillaLockLocked, func(int) time.Duration { return unbondingDuration + time.Nanosecond*3 }, false, func(index int, posInfoInternal positionInfo) {
			})

			createPositions(VanillaLockUnlocking, positions.numVanillaLockUnlocking, func(int) time.Duration { return unbondingDuration + time.Nanosecond*4 }, false, func(index int, posInfoInternal positionInfo) {
				lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, posInfoInternal.lockId)
				s.Require().NoError(err)
				_, err = s.App.LockupKeeper.BeginUnlock(s.Ctx, posInfoInternal.lockId, lock.Coins)
				s.Require().NoError(err)
			})

			createPositions(NoLock, positions.numNoLock, func(int) time.Duration { return time.Duration(0) }, false, func(index int, posInfoInternal positionInfo) {
			})

			// Some funds might not have been completely used when creating the above positions.
			// We note them here and use them when tracking invariants at the very end.
			unusedPositionCreationFunds := s.calculateUnusedPositionCreationFunds(positions.numAccounts, DefaultCoin0.Denom, DefaultCoin1.Denom)

			// Create CL pool
			clPool := s.PrepareConcentratedPoolWithCoins(DefaultCoin0.Denom, DefaultCoin1.Denom)
			clPoolId := clPool.GetId()

			// Match the spot price of the CL pool to the balancer pool by creating a position with the same ratio.
			balancerPool, err := s.App.GAMMKeeper.GetCFMMPool(s.Ctx, balancerPoolId)
			s.Require().NoError(err)
			balancerSpotPrice, err := balancerPool.SpotPrice(s.Ctx, DefaultCoin1.Denom, DefaultCoin0.Denom)
			s.Require().NoError(err)
			// balancerSpotPrice truncation is acceptable because all CFMM pools only allow 18 decimals.
			// The reason why BigDec is returned is to maintain compatibility with the generalized `PoolI.SpotPrice`API.
			s.CreateFullRangePosition(clPool, sdk.NewCoins(sdk.NewCoin(DefaultCoin0.Denom, osmomath.NewInt(100000000)), sdk.NewCoin(DefaultCoin1.Denom, osmomath.NewDec(100000000).Mul(balancerSpotPrice.Dec()).TruncateInt())))

			// Add a gov sanctioned link between the balancer and concentrated liquidity pool.
			migrationRecord := gammmigration.MigrationRecords{BalancerToConcentratedPoolLinks: []gammmigration.BalancerToConcentratedPoolLink{
				{BalancerPoolId: balancerPoolId, ClPoolId: clPoolId},
			}}
			err = s.App.GAMMKeeper.OverwriteMigrationRecords(s.Ctx, migrationRecord)
			s.Require().NoError(err)

			// Register the CL denom as superfluid.
			clPoolDenom := cltypes.GetConcentratedLockupDenomFromPoolId(clPoolId)
			err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{
				Denom:     clPoolDenom,
				AssetType: types.SuperfluidAssetTypeConcentratedShare,
			})
			s.Require().NoError(err)

			// All the following values will be tracked as we migrate each position. We will check them against the invariants at the end.
			totalAmount0Migrated := osmomath.ZeroInt()
			totalAmount1Migrated := osmomath.ZeroInt()
			totalSentBackToOwnersAmount0 := osmomath.ZeroInt()
			totalSentBackToOwnersAmount1 := osmomath.ZeroInt()
			totalBalancerPoolFundsLeftBehindAmount0 := osmomath.ZeroInt()
			totalBalancerPoolFundsLeftBehindAmount1 := osmomath.ZeroInt()

			// Migrate all the positions.
			// We will check certain invariants after each individual migration.
			for _, positionInfo := range positionInfos {
				for i, posInfo := range positionInfo {
					balancerPool, err = s.App.GAMMKeeper.GetCFMMPool(s.Ctx, balancerPoolId)
					s.Require().NoError(err)

					// Note owner and balancer pool balances before migration.
					preClaimOwnerBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, s.TestAccs[i+1])
					preClaimBalancerPoolBalance := balancerPool.GetTotalPoolLiquidity(s.Ctx)

					// Run the migration.
					positionData, _, _, err := s.App.SuperfluidKeeper.RouteLockedBalancerToConcentratedMigration(s.Ctx, s.TestAccs[i+1], int64(posInfo.lockId), posInfo.coin, sdk.Coins{})
					s.Require().NoError(err)

					// Note how much of amount0 and amount1 was actually created in the CL pool from the migration.
					clJoinPoolAmt := sdk.NewCoins(sdk.NewCoin(clPool.GetToken0(), positionData.Amount0), sdk.NewCoin(clPool.GetToken1(), positionData.Amount1))

					// Note owner and balancer pool balances after migration.
					balancerPool, err = s.App.GAMMKeeper.GetCFMMPool(s.Ctx, balancerPoolId)
					s.Require().NoError(err)
					postClaimOwnerBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, s.TestAccs[i+1])
					postClaimBalancerPoolBalance := balancerPool.GetTotalPoolLiquidity(s.Ctx)

					// Check that the diff between the initial balancer position and the new CL position is equal to the amount that was left behind in the balancer pool and the amount that was sent back to the owner.
					for i := range posInfo.joinPoolCoins {
						balancerToClBalanceDelta := posInfo.joinPoolCoins[i].Amount.Sub(clJoinPoolAmt.AmountOf(posInfo.joinPoolCoins[i].Denom))
						userBalanceDelta := postClaimOwnerBalance.AmountOf(posInfo.joinPoolCoins[i].Denom).Sub(preClaimOwnerBalance.AmountOf(posInfo.joinPoolCoins[i].Denom))
						balancerPoolDelta := posInfo.joinPoolCoins[i].Amount.Sub(preClaimBalancerPoolBalance.AmountOf(posInfo.joinPoolCoins[i].Denom).Sub(postClaimBalancerPoolBalance.AmountOf(posInfo.joinPoolCoins[i].Denom)))
						if i == 0 {
							totalBalancerPoolFundsLeftBehindAmount0 = totalBalancerPoolFundsLeftBehindAmount0.Add(balancerPoolDelta)
						} else {
							totalBalancerPoolFundsLeftBehindAmount1 = totalBalancerPoolFundsLeftBehindAmount1.Add(balancerPoolDelta)
						}
						s.Require().Equal(balancerToClBalanceDelta.String(), userBalanceDelta.Add(balancerPoolDelta).String())
					}

					// Add to the total amounts that were migrated and sent back to the owners.
					totalAmount0Migrated = totalAmount0Migrated.Add(positionData.Amount0)
					totalAmount1Migrated = totalAmount1Migrated.Add(positionData.Amount1)
					totalSentBackToOwnersAmount0 = totalSentBackToOwnersAmount0.Add(postClaimOwnerBalance.AmountOf(DefaultCoin0.Denom).Sub(preClaimOwnerBalance.AmountOf(DefaultCoin0.Denom)))
					totalSentBackToOwnersAmount1 = totalSentBackToOwnersAmount1.Add(postClaimOwnerBalance.AmountOf(DefaultCoin1.Denom).Sub(preClaimOwnerBalance.AmountOf(DefaultCoin1.Denom)))
				}
			}

			// Check that we have account for all the funds that were used to create the positions.
			amount0AccountFor := totalAmount0Migrated.Add(totalSentBackToOwnersAmount0).Add(totalBalancerPoolFundsLeftBehindAmount0).Add(unusedPositionCreationFunds.AmountOf(DefaultCoin0.Denom))
			amount1AccountFor := totalAmount1Migrated.Add(totalSentBackToOwnersAmount1).Add(totalBalancerPoolFundsLeftBehindAmount1).Add(unusedPositionCreationFunds.AmountOf(DefaultCoin1.Denom))
			s.Require().Equal(totalFundsForPositionCreation.AmountOf(DefaultCoin0.Denom), amount0AccountFor)
			s.Require().Equal(totalFundsForPositionCreation.AmountOf(DefaultCoin1.Denom), amount1AccountFor)
		})
	}
}

// createBalancerPosition creates a position in a Balancer pool for a given account with optional superfluid delegation.
//
// The function joins the Balancer pool with the specified account, using coins with varying amounts based on the divisor `i`.
// It retrieves the number of shares received (`sharesOut`) from the `JoinSwapExactAmountIn` function.
//
// If `unbondingDuration` is non-zero, the function locks the obtained shares using `LockTokensNoFund` and assigns the lock ID to `lockId`.
//
// If `superfluidDelegate` is true, the function delegates the obtained shares to the default val module using `SuperfluidDelegateToDefaultVal`.
//
// The function returns a `positionInfo` struct with the created position's information.
func (s *KeeperTestSuite) createBalancerPosition(acc sdk.AccAddress, balancerPoolId uint64, unbondingDuration time.Duration, balancerPoolShareDenom string, coins sdk.Coins, superfluidDelegate bool) positionInfo {
	sharesOut, err := s.App.GAMMKeeper.JoinSwapExactAmountIn(s.Ctx, acc, balancerPoolId, coins, osmomath.OneInt())
	s.Require().NoError(err)
	shareCoins := sdk.NewCoins(sdk.NewCoin(balancerPoolShareDenom, sharesOut))

	lockId := uint64(0)
	// Lock tokens if a duration is specified.
	if unbondingDuration != 0 {
		lockId = s.LockTokensNoFund(acc, shareCoins, unbondingDuration)
	}

	// Superfluid delegate if specified.
	if superfluidDelegate {
		err = s.SuperfluidDelegateToDefaultVal(acc, balancerPoolId, lockId)
		s.Require().NoError(err)
	}

	sharesOutCoin := sdk.NewCoin(balancerPoolShareDenom, sharesOut)

	posInfoInternal := positionInfo{
		joinPoolCoins: coins,
		coin:          sharesOutCoin,
		shares:        sharesOut,
		lockId:        lockId,
	}
	return posInfoInternal
}

// calculateUnusedPositionCreationFunds calculates the unused position creation funds for a given number of accounts and positions without locks.
//
// The function iterates over the accounts starting from index 1 and retrieves their balances for the specified coin denominations.
// It aggregates the unused position creation funds by adding the coin amounts to the `unusedPositionCreationFunds` variable.
//
// Parameters:
// - numAccounts: The total number of accounts.
// - numNoLock: The number of positions without locks.
// - coin0Denom: The denomination of the first coin.
// - coin1Denom: The denomination of the second coin.
//
// Returns:
// - sdk.Coins: The total unused position creation funds as a `sdk.Coins` object.
func (s *KeeperTestSuite) calculateUnusedPositionCreationFunds(numAccounts int, coin0Denom, coin1Denom string) sdk.Coins {
	unusedPositionCreationFunds := sdk.Coins{}
	for i := 1; i < numAccounts; i++ {
		balances := s.App.BankKeeper.GetAllBalances(s.Ctx, s.TestAccs[i])
		unusedPositionCreationFunds = unusedPositionCreationFunds.Add(sdk.NewCoin(coin0Denom, balances.AmountOf(coin0Denom)))
		unusedPositionCreationFunds = unusedPositionCreationFunds.Add(sdk.NewCoin(coin1Denom, balances.AmountOf(coin1Denom)))
	}
	return unusedPositionCreationFunds
}
