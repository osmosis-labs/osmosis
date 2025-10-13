package v31

import (
	"context"
	"fmt"
	"time"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v30/app/keepers"
	"github.com/osmosis-labs/osmosis/v30/app/upgrades"
	poolmanager "github.com/osmosis-labs/osmosis/v30/x/poolmanager"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v30/x/poolmanager/types"
	superfuidtypes "github.com/osmosis-labs/osmosis/v30/x/superfluid/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v30/x/txfees/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		sdkCtx := sdk.UnwrapSDKContext(ctx)

		err = updateTakerFeeDistribution(sdkCtx, keepers.PoolManagerKeeper, keepers.AccountKeeper)
		if err != nil {
			return nil, err
		}

		// Undelegate all remaining superfluid stake and clean up all superfluid storage
		err = cleanupSuperfluid(sdkCtx, keepers)
		if err != nil {
			return nil, err
		}

		return migrations, nil
	}
}

// updateTakerFeeDistribution updates the community_pool and burn values in the osmo_taker_fee_distribution
// This changes taker fees from being sent to the community pool to being burned instead.
// It also sets up the staking rewards smoothing feature with a smoothing factor of 7.
func updateTakerFeeDistribution(ctx sdk.Context, poolManagerKeeper *poolmanager.Keeper, accountKeeper *authkeeper.AccountKeeper) error {
	// Set OSMO taker fee distribution: community_pool to 0, burn and staking rewards to 70%:30%
	osmoTakerFeeDistribution := poolmanagertypes.TakerFeeDistributionPercentage{
		CommunityPool:  osmomath.ZeroDec(),
		Burn:           osmomath.MustNewDecFromStr("0.7"),
		StakingRewards: osmomath.MustNewDecFromStr("0.3"),
	}
	poolManagerKeeper.SetParam(ctx, poolmanagertypes.KeyOsmoTakerFeeDistribution, osmoTakerFeeDistribution)

	// Set non-OSMO taker fee distribution: staking_rewards=22.5%, burn=52.5%, community_pool=25%
	nonOsmoTakerFeeDistribution := poolmanagertypes.TakerFeeDistributionPercentage{
		StakingRewards: osmomath.MustNewDecFromStr("0.225"),
		Burn:           osmomath.MustNewDecFromStr("0.525"),
		CommunityPool:  osmomath.MustNewDecFromStr("0.25"),
	}
	poolManagerKeeper.SetParam(ctx, poolmanagertypes.KeyNonOsmoTakerFeeDistribution, nonOsmoTakerFeeDistribution)

	// Set daily staking rewards smoothing factor to 7
	// This distributes 1/7th of the staking rewards buffer each day to smooth APR display
	dailyStakingRewardsSmoothingFactor := uint64(7)
	poolManagerKeeper.SetParam(ctx, poolmanagertypes.KeyDailyStakingRewardsSmoothingFactor, dailyStakingRewardsSmoothingFactor)

	// Ensure new module account exists for non‑native taker fee burn bucket. Error if it already exists.
	err := osmoutils.CreateModuleAccountByName(ctx, accountKeeper, txfeestypes.TakerFeeBurnName)
	if err != nil {
		return err
	}

	// Create the staking rewards smoothing buffer module account
	err = osmoutils.CreateModuleAccountByName(ctx, accountKeeper, txfeestypes.TakerFeeStakingRewardsBuffer)
	if err != nil {
		return err
	}

	return nil
}

// cleanupSuperfluid undelegates all remaining superfluid stake and removes all superfluid storage.
// This is the main entry point that orchestrates the complete cleanup in three phases:
// 1. Undelegate all intermediary account positions
// 2. Delete all synthetic locks
// 3. Delete all superfluid storage
func cleanupSuperfluid(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	ctx.Logger().Info("Starting superfluid cleanup: undelegating all positions and removing all state")
	startTime := time.Now()

	// Undelegate all intermediary account positions
	// This must happen first to properly handle the actual staking delegations
	if err := undelegateAllIntermediaryAccounts(ctx, keepers); err != nil {
		return err
	}

	// Delete all synthetic locks
	// This removes the overlay tracking before deleting the mappings
	if err := unlockAllSyntheticLocks(ctx, keepers); err != nil {
		return err
	}

	// Delete all superfluid storage
	// Finally clean up all references and configuration
	if err := deleteAllSuperfluidStorage(ctx, keepers); err != nil {
		return err
	}

	elapsed := time.Since(startTime)

	ctx.Logger().Info("Superfluid cleanup completed successfully, took " + elapsed.String())
	return nil
}

// undelegateAllIntermediaryAccounts undelegates and burns all tokens from intermediary accounts.
// Intermediary accounts are special accounts that hold the actual staking delegations on behalf
// of users' locked LP shares. For each intermediary account:
// 1. Calculate the delegated OSMO amount
// 2. Perform instant undelegation (bypass 21-day unbonding)
// 3. Burn the undelegated OSMO tokens (these were synthetically minted during delegation)
// 4. Adjust the supply offset to maintain proper accounting
//
// Reference: x/superfluid/keeper/stake.go:462-542 (mintOsmoTokensAndDelegate, forceUndelegateAndBurnOsmoTokens)
func undelegateAllIntermediaryAccounts(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	intermediaryAccounts := keepers.SuperfluidKeeper.GetAllIntermediaryAccounts(ctx)
	ctx.Logger().Info(fmt.Sprintf("Found %d intermediary accounts to clean up", len(intermediaryAccounts)))
	totalUndelegatedCoins := sdk.Coins{}

	for _, intermediaryAcc := range intermediaryAccounts {
		undelegatedCoins, err := undelegateSingleIntermediaryAccount(ctx, keepers, intermediaryAcc)
		if err != nil {
			// Log error but continue with other accounts
			ctx.Logger().Error(fmt.Sprintf("Failed to undelegate intermediary account %s: %v",
				intermediaryAcc.GetAccAddress().String(), err))
			continue
		}
		totalUndelegatedCoins = totalUndelegatedCoins.Add(undelegatedCoins...)
	}

	// Validate that denoms contain only uosmo
	bondDenom, err := keepers.StakingKeeper.BondDenom(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bond denom: %w", err)
	}

	denoms := totalUndelegatedCoins.Denoms()
	if len(denoms) > 1 || (len(denoms) == 1 && denoms[0] != bondDenom) {
		return fmt.Errorf("expected only %s denom, but got: %v", bondDenom, denoms)
	}

	totalAmount := totalUndelegatedCoins.AmountOf(bondDenom)
	if totalAmount.GT(totalSuperfluidDelegationAmount) {
		return fmt.Errorf("total undelegated amount %s is greater than expected %s", totalAmount.String(), totalSuperfluidDelegationAmount.String())
	}

	ctx.Logger().Info(fmt.Sprintf("Successfully undelegated and burned %s %s from %d intermediary accounts",
		totalAmount.String(), bondDenom, len(intermediaryAccounts)))

	return nil
}

// undelegateSingleIntermediaryAccount handles the undelegation and burning for a single intermediary account.
func undelegateSingleIntermediaryAccount(ctx sdk.Context, keepers *keepers.AppKeepers, intermediaryAcc superfuidtypes.SuperfluidIntermediaryAccount) (sdk.Coins, error) {
	// Get validator address
	valAddr, err := sdk.ValAddressFromBech32(intermediaryAcc.ValAddr)
	if err != nil {
		return sdk.Coins{}, fmt.Errorf("invalid validator address %s: %w", intermediaryAcc.ValAddr, err)
	}

	// Check if there's a delegation from this intermediary account
	delegation, err := keepers.StakingKeeper.GetDelegation(ctx, intermediaryAcc.GetAccAddress(), valAddr)
	if err != nil {
		// No delegation found, skip
		return sdk.Coins{}, nil
	}

	// Get validator to calculate tokens from shares
	validator, err := keepers.StakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		return sdk.Coins{}, fmt.Errorf("validator not found for %s: %w", intermediaryAcc.ValAddr, err)
	}

	// Calculate the amount of tokens to undelegate
	tokens := validator.TokensFromShares(delegation.Shares)
	osmoAmount := tokens.RoundInt()

	if !osmoAmount.IsPositive() {
		return sdk.Coins{}, nil
	}

	// Validate unbond amount and get shares
	shares, err := keepers.StakingKeeper.ValidateUnbondAmount(ctx, intermediaryAcc.GetAccAddress(), valAddr, osmoAmount)
	if err != nil {
		return sdk.Coins{}, fmt.Errorf("failed to validate unbond amount: %w", err)
	}

	// Instant undelegate (bypass normal 21-day unbonding period)
	undelegatedCoins, err := keepers.StakingKeeper.InstantUndelegate(ctx, intermediaryAcc.GetAccAddress(), valAddr, shares)
	if err != nil {
		return sdk.Coins{}, fmt.Errorf("failed to instant undelegate: %w", err)
	}

	// Send coins to superfluid module
	err = keepers.BankKeeper.SendCoinsFromAccountToModule(ctx, intermediaryAcc.GetAccAddress(), superfuidtypes.ModuleName, undelegatedCoins)
	if err != nil {
		return sdk.Coins{}, fmt.Errorf("failed to send coins to module: %w", err)
	}

	// Burn the coins (these were synthetically minted during superfluid delegation)
	err = keepers.BankKeeper.BurnCoins(ctx, superfuidtypes.ModuleName, undelegatedCoins)
	if err != nil {
		return sdk.Coins{}, fmt.Errorf("failed to burn coins: %w", err)
	}

	// Adjust supply offset to maintain proper total supply accounting
	bondDenom, err := keepers.StakingKeeper.BondDenom(ctx)
	if err != nil {
		return sdk.Coins{}, fmt.Errorf("failed to get bond denom: %w", err)
	}
	keepers.BankKeeper.AddSupplyOffset(ctx, bondDenom, undelegatedCoins.AmountOf(bondDenom))

	ctx.Logger().Info(fmt.Sprintf("Undelegated and burned %s from intermediary account %s (validator: %s)",
		undelegatedCoins.String(), intermediaryAcc.GetAccAddress().String(), intermediaryAcc.ValAddr))

	return undelegatedCoins, nil
}

// unlockAllSyntheticLocks deletes all synthetic locks associated with superfluid staking
// and unlocks the underlying locks to return gamm tokens to delegators.
// Synthetic locks are overlay locks that track superfluid staking status. There are two types:
// - Staking locks: {denom}/superbonding/{valAddr} - tracks active superfluid stake
// - Unstaking locks: {denom}/superunbonding/{valAddr} - tracks undelegating positions
// Both types must be cleaned up and the underlying locks must be unlocked.
func unlockAllSyntheticLocks(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	connections := keepers.SuperfluidKeeper.GetAllLockIdIntermediaryAccountConnections(ctx)
	ctx.Logger().Info(fmt.Sprintf("Found %d lock-intermediary connections to clean up", len(connections)))

	for _, connection := range connections {
		if err := deleteForceUnlockSyntheticLock(ctx, keepers, connection); err != nil {
			// Log error but continue with other locks
			ctx.Logger().Error(fmt.Sprintf("Failed to delete synthetic locks for lock %d: %v", connection.LockId, err))
			continue
		}
	}

	return nil
}

// deleteForceUnlockSyntheticLock unlocks the underlying lock to return gamm tokens to the delegator.
// ForceUnlock automatically handles deleting any associated synthetic locks.
func deleteForceUnlockSyntheticLock(ctx sdk.Context, keepers *keepers.AppKeepers, connection superfuidtypes.LockIdIntermediaryAccountConnection) error {
	// Get the lock
	lock, err := keepers.LockupKeeper.GetLockByID(ctx, connection.LockId)
	if err != nil {
		return fmt.Errorf("failed to get lock %d: %w", connection.LockId, err)
	}

	// Force unlock the underlying lock to return gamm tokens to the delegator
	// Note: ForceUnlock automatically handles deleting any synthetic locks first
	err = keepers.LockupKeeper.ForceUnlock(ctx, *lock)
	if err != nil {
		return fmt.Errorf("failed to force unlock lock %d: %w", connection.LockId, err)
	}

	return nil
}

// deleteAllSuperfluidStorage deletes all superfluid-related storage from the KV store.
// This includes:
// 1. Lock-intermediary account connections (KeyPrefixLockIntermediaryAccAddr = 0x05)
// 2. Intermediary accounts (KeyPrefixIntermediaryAccount = 0x04)
// 3. Superfluid assets (KeyPrefixSuperfluidAsset = 0x01)
// 4. OSMO equivalent multipliers (KeyPrefixTokenMultiplier = 0x03)
// 5. Unpool allowed pools (KeyUnpoolAllowedPools = 0x06)
func deleteAllSuperfluidStorage(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	store := ctx.KVStore(keepers.GetKey(superfuidtypes.StoreKey))

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	keysToDelete := [][]byte{}
	for ; iterator.Valid(); iterator.Next() {
		keysToDelete = append(keysToDelete, iterator.Key())
	}

	for _, key := range keysToDelete {
		store.Delete(key)
	}

	return nil
}
