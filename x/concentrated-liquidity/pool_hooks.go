package concentrated_liquidity

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/osmosis-labs/osmosis/v20/x/concentrated-liquidity/types"
)

// nolint: unused
// callPoolActionListener processes and dispatches the passed in message to the contract corresponding to the hook
// defined by the given pool ID and action prefix (e.g. pool Id: 1, action prefix: "beforeSwap").
//
// This function returns an error if the contract address in state is invalid (should be impossible) or if the contract execution fails.
//
// Since it is possible for this function to be triggered in begin block code, we need to directly meter its execution and set a limit.
// If no contract is linked to the hook, this function is a no-op.
func (k Keeper) callPoolActionListener(ctx sdk.Context, msgBz []byte, poolId uint64, actionPrefix string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = types.ContractHookOutOfGasError{GasLimit: k.GetParams(ctx).HookGasLimit}
		}
	}()

	cosmwasmAddress := k.getPoolHookContract(ctx, poolId, actionPrefix)
	if cosmwasmAddress == "" {
		return nil
	}

	cwAddr, err := sdk.AccAddressFromBech32(cosmwasmAddress)
	if err != nil {
		return err
	}

	em := sdk.NewEventManager()

	// Since it is possible for this hook to be triggered in begin block code, we need to
	// directly meter its execution and set a limit. See comments on `ContractHookGasLimit`
	// for details on how the specific limit was chosen.
	//
	// We ensure this limit only applies to this call by creating a child context with a gas
	// limit and then metering the gas used in parent context once the operation is completed.
	childCtx := ctx.WithGasMeter(sdk.NewGasMeter(k.GetParams(ctx).HookGasLimit))
	_, err = k.contractKeeper.Sudo(childCtx.WithEventManager(em), cwAddr, msgBz)
	if err != nil {
		return err
	}

	// Consume gas used for calling contract to the parent ctx
	ctx.GasMeter().ConsumeGas(childCtx.GasMeter().GasConsumed(), "Track CL action contract call gas")

	return nil
}

// --- Store helpers ---

// nolint: unused
// getPoolHookPrefixStore returns the substore for a specific pool ID where hook-related data is stored.
func (k Keeper) getPoolHookPrefixStore(ctx sdk.Context, poolID uint64) sdk.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetPoolPrefixStoreKey(poolID))
}

// nolint: unused
// getPoolHookContract returns the contract address linked to the passed in action for a specific pool ID.
// For instance, if poolId is `1` and actionPrefix is "beforeSwap", this will return the contract address
// corresponding to the beforeSwap hook on pool 1.
func (k Keeper) getPoolHookContract(ctx sdk.Context, poolId uint64, actionPrefix string) string {
	store := k.getPoolHookPrefixStore(ctx, poolId)

	bz := store.Get([]byte(actionPrefix))
	if bz == nil {
		return ""
	}

	return string(bz)
}

// nolint: unused
// setPoolHookContract sets the contract address linked to the passed in hook for a specific pool ID.
// Passing in an empty string for `cosmwasmAddress` will be interpreted as a deletion for the contract associated
// with the given poolId and actionPrefix.
// Attempting to delete a non-existent contract in state will simply be a no-op.
func (k Keeper) setPoolHookContract(ctx sdk.Context, poolID uint64, actionPrefix string, cosmwasmAddress string) error {
	store := k.getPoolHookPrefixStore(ctx, poolID)

	// If cosmwasm address is nil, treat this as a delete operation for the stored address.
	if cosmwasmAddress == "" {
		deletePoolHookContract(store, actionPrefix)
		return nil
	}

	// Verify that the cosmwasm address is valid bech32 that can be converted to AccAddress.
	_, err := sdk.AccAddressFromBech32(cosmwasmAddress)
	if err != nil {
		return err
	}

	store.Set([]byte(actionPrefix), []byte(cosmwasmAddress))

	return nil
}

// nolint: unused
// deletePoolHookContract deletes the pool hook contract corresponding to the given action prefix from the passed in store.
// It takes in a store directly instead of ctx and pool ID to avoid doing another read (to fetch pool hook prefix store) for
// an abstraction that was primarily added for code readability reasons.
func deletePoolHookContract(store sdk.KVStore, actionPrefix string) {
	store.Delete([]byte(actionPrefix))
}
