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
			err = types.ContractHookOutOfGasError{GasLimit: types.ContractHookGasLimit}
		}
	}()

	cosmwasmAddress := k.GetPoolHookContract(ctx, poolId, actionPrefix)
	if cosmwasmAddress != "" {
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
		childCtx := ctx.WithGasMeter(sdk.NewGasMeter(types.ContractHookGasLimit))
		_, err = k.contractKeeper.Sudo(childCtx.WithEventManager(em), cwAddr, msgBz)
		if err != nil {
			return err
		}

		// Consume gas used for calling contract to the parent ctx
		ctx.GasMeter().ConsumeGas(childCtx.GasMeter().GasConsumed(), "Track CL action contract call gas")
	}

	return nil
}

// --- Store helpers ---

// GetPoolHookPrefixStore returns the substore for a specific pool ID where hook-related data is stored.
func (k Keeper) GetPoolHookPrefixStore(ctx sdk.Context, poolID uint64) sdk.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetPoolPrefixStoreKey(poolID))
}

// GetPoolHookContract returns the contract address linked to the passed in action for a specific pool ID.
// For instance, if poolId is `1` and actionPrefix is "beforeSwap", this will return the contract address
// corresponding to the beforeSwap hook on pool 1.
func (k Keeper) GetPoolHookContract(ctx sdk.Context, poolId uint64, actionPrefix string) string {
	store := k.GetPoolHookPrefixStore(ctx, poolId)

	bz := store.Get([]byte(actionPrefix))
	if bz == nil {
		return ""
	}

	return string(bz)
}

// nolint: unused
// setPoolHookContract sets the contract address linked to the passed in hook for a specific pool ID.
func (k Keeper) setPoolHookContract(ctx sdk.Context, poolID uint64, actionPrefix string, cosmwasmAddress string) error {
	store := k.GetPoolHookPrefixStore(ctx, poolID)

	// If cosmwasm address is nil, treat this as a delete operation for the stored address.
	if cosmwasmAddress == "" {
		store.Delete([]byte(actionPrefix))
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
