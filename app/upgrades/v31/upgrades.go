package v31

import (
	"context"
	"fmt"

	"cosmossdk.io/store/prefix"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v30/app/keepers"
	"github.com/osmosis-labs/osmosis/v30/app/upgrades"
	poolmanager "github.com/osmosis-labs/osmosis/v30/x/poolmanager"
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
func updateTakerFeeDistribution(ctx sdk.Context, poolManagerKeeper *poolmanager.Keeper, accountKeeper *authkeeper.AccountKeeper) error {
	poolManagerParams := poolManagerKeeper.GetParams(ctx)

	// Set community_pool to 0, burn and staking rewards to 70%:30%
	poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.CommunityPool = osmomath.ZeroDec()
	poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.Burn = osmomath.MustNewDecFromStr("0.7")
	poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.StakingRewards = osmomath.MustNewDecFromStr("0.3")

	// Set non-OSMO taker fee distribution: staking_rewards=22.5%, burn=52.5%, community_pool=25%
	poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.StakingRewards = osmomath.MustNewDecFromStr("0.225")
	poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.Burn = osmomath.MustNewDecFromStr("0.525")
	poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.CommunityPool = osmomath.MustNewDecFromStr("0.25")

	poolManagerKeeper.SetParams(ctx, poolManagerParams)

	// Ensure new module account exists for non‑native taker fee burn bucket. Error if it already exists.
	err := osmoutils.CreateModuleAccountByName(ctx, accountKeeper, txfeestypes.TakerFeeBurnName)
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

	// Undelegate all intermediary account positions
	// This must happen first to properly handle the actual staking delegations
	if err := undelegateAllIntermediaryAccounts(ctx, keepers); err != nil {
		return err
	}

	// Delete all synthetic locks
	// This removes the overlay tracking before deleting the mappings
	if err := deleteAllSyntheticLocks(ctx, keepers); err != nil {
		return err
	}

	// Delete all superfluid storage
	// Finally clean up all references and configuration
	if err := deleteAllSuperfluidStorage(ctx, keepers); err != nil {
		return err
	}

	ctx.Logger().Info("Superfluid cleanup completed successfully")
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

	for _, intermediaryAcc := range intermediaryAccounts {
		if err := undelegateSingleIntermediaryAccount(ctx, keepers, intermediaryAcc); err != nil {
			// Log error but continue with other accounts
			ctx.Logger().Error(fmt.Sprintf("Failed to undelegate intermediary account %s: %v",
				intermediaryAcc.GetAccAddress().String(), err))
			continue
		}
	}

	return nil
}

// undelegateSingleIntermediaryAccount handles the undelegation and burning for a single intermediary account.
func undelegateSingleIntermediaryAccount(ctx sdk.Context, keepers *keepers.AppKeepers, intermediaryAcc superfuidtypes.SuperfluidIntermediaryAccount) error {
	// Get validator address
	valAddr, err := sdk.ValAddressFromBech32(intermediaryAcc.ValAddr)
	if err != nil {
		return fmt.Errorf("invalid validator address %s: %w", intermediaryAcc.ValAddr, err)
	}

	// Check if there's a delegation from this intermediary account
	delegation, err := keepers.StakingKeeper.GetDelegation(ctx, intermediaryAcc.GetAccAddress(), valAddr)
	if err != nil {
		// No delegation found, skip
		return nil
	}

	// Get validator to calculate tokens from shares
	validator, err := keepers.StakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		return fmt.Errorf("validator not found for %s: %w", intermediaryAcc.ValAddr, err)
	}

	// Calculate the amount of tokens to undelegate
	tokens := validator.TokensFromShares(delegation.Shares)
	osmoAmount := tokens.RoundInt()

	if !osmoAmount.IsPositive() {
		return nil
	}

	// Validate unbond amount and get shares
	shares, err := keepers.StakingKeeper.ValidateUnbondAmount(ctx, intermediaryAcc.GetAccAddress(), valAddr, osmoAmount)
	if err != nil {
		return fmt.Errorf("failed to validate unbond amount: %w", err)
	}

	// Instant undelegate (bypass normal 21-day unbonding period)
	undelegatedCoins, err := keepers.StakingKeeper.InstantUndelegate(ctx, intermediaryAcc.GetAccAddress(), valAddr, shares)
	if err != nil {
		return fmt.Errorf("failed to instant undelegate: %w", err)
	}

	// Send coins to superfluid module
	err = keepers.BankKeeper.SendCoinsFromAccountToModule(ctx, intermediaryAcc.GetAccAddress(), superfuidtypes.ModuleName, undelegatedCoins)
	if err != nil {
		return fmt.Errorf("failed to send coins to module: %w", err)
	}

	// Burn the coins (these were synthetically minted during superfluid delegation)
	err = keepers.BankKeeper.BurnCoins(ctx, superfuidtypes.ModuleName, undelegatedCoins)
	if err != nil {
		return fmt.Errorf("failed to burn coins: %w", err)
	}

	// Adjust supply offset to maintain proper total supply accounting
	bondDenom, err := keepers.StakingKeeper.BondDenom(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bond denom: %w", err)
	}
	keepers.BankKeeper.AddSupplyOffset(ctx, bondDenom, undelegatedCoins.AmountOf(bondDenom))

	ctx.Logger().Info(fmt.Sprintf("Undelegated and burned %s from intermediary account %s (validator: %s)",
		undelegatedCoins.String(), intermediaryAcc.GetAccAddress().String(), intermediaryAcc.ValAddr))

	return nil
}

// deleteAllSyntheticLocks deletes all synthetic locks associated with superfluid staking
// and unlocks the underlying locks to return gamm tokens to delegators.
// Synthetic locks are overlay locks that track superfluid staking status. There are two types:
// - Staking locks: {denom}/superbonding/{valAddr} - tracks active superfluid stake
// - Unstaking locks: {denom}/superunbonding/{valAddr} - tracks undelegating positions
// Both types must be cleaned up and the underlying locks must be unlocked.
func deleteAllSyntheticLocks(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	connections := keepers.SuperfluidKeeper.GetAllLockIdIntermediaryAccountConnections(ctx)
	ctx.Logger().Info(fmt.Sprintf("Found %d lock-intermediary connections to clean up", len(connections)))

	for _, connection := range connections {
		if err := deleteSyntheticLockAndUnlock(ctx, keepers, connection); err != nil {
			// Log error but continue with other locks
			ctx.Logger().Error(fmt.Sprintf("Failed to delete synthetic locks for lock %d: %v", connection.LockId, err))
			continue
		}
	}

	return nil
}

// deleteSyntheticLockAndUnlock deletes both staking and unstaking synthetic locks for a single lock
// and unlocks the underlying lock to return gamm tokens to the delegator.
func deleteSyntheticLockAndUnlock(ctx sdk.Context, keepers *keepers.AppKeepers, connection superfuidtypes.LockIdIntermediaryAccountConnection) error {
	// Get the lock
	lock, err := keepers.LockupKeeper.GetLockByID(ctx, connection.LockId)
	if err != nil {
		return fmt.Errorf("failed to get lock %d: %w", connection.LockId, err)
	}

	// Get intermediary account
	accAddr, err := sdk.AccAddressFromBech32(connection.IntermediaryAccount)
	if err != nil {
		return fmt.Errorf("invalid intermediary account address %s: %w", connection.IntermediaryAccount, err)
	}

	store := ctx.KVStore(keepers.GetKey(superfuidtypes.StoreKey))
	prefixStore := prefix.NewStore(store, superfuidtypes.KeyPrefixIntermediaryAccount)
	intermediaryAccBytes := prefixStore.Get(accAddr)
	if intermediaryAccBytes == nil {
		return fmt.Errorf("intermediary account not found for lock %d", connection.LockId)
	}

	var intermediaryAcc superfuidtypes.SuperfluidIntermediaryAccount
	err = intermediaryAcc.Unmarshal(intermediaryAccBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal intermediary account: %w", err)
	}

	// Delete both staking and unstaking synthetic locks if they exist
	if len(lock.Coins) > 0 {
		denom := lock.Coins[0].Denom

		// Try to delete staking synthetic lock: {denom}/superbonding/{valAddr}
		stakingSynthDenom := fmt.Sprintf("%s/superbonding/%s", denom, intermediaryAcc.ValAddr)
		if err := keepers.LockupKeeper.DeleteSyntheticLockup(ctx, connection.LockId, stakingSynthDenom); err != nil {
			// Log but don't fail - synthetic lock might not exist
			ctx.Logger().Info(fmt.Sprintf("Could not delete staking synthetic lock for lock %d: %v (may not exist)", connection.LockId, err))
		}

		// Try to delete unstaking synthetic lock: {denom}/superunbonding/{valAddr}
		unstakingSynthDenom := fmt.Sprintf("%s/superunbonding/%s", denom, intermediaryAcc.ValAddr)
		if err := keepers.LockupKeeper.DeleteSyntheticLockup(ctx, connection.LockId, unstakingSynthDenom); err != nil {
			// Log but don't fail - synthetic lock might not exist
			ctx.Logger().Info(fmt.Sprintf("Could not delete unstaking synthetic lock for lock %d: %v (may not exist)", connection.LockId, err))
		}
	}

	// Force unlock the underlying lock to return gamm tokens to the delegator
	err = keepers.LockupKeeper.ForceUnlock(ctx, *lock)
	if err != nil {
		return fmt.Errorf("failed to force unlock lock %d: %w", connection.LockId, err)
	}

	ctx.Logger().Info(fmt.Sprintf("Unlocked lock %d, returned %s to delegator %s",
		lock.ID, lock.Coins.String(), lock.Owner))

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
