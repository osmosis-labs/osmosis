package keeper

import (
	"strings"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v26/x/callback/types"
)

// GetAllCallbacks lists all the pending callbacks
func (k Keeper) GetAllCallbacks(ctx sdk.Context) (callbacks []types.Callback, err error) {
	err = k.Callbacks.Walk(ctx, nil, func(key collections.Triple[int64, []byte, uint64], value types.Callback) (bool, error) {
		callbacks = append(callbacks, value)
		return false, nil
	})
	return callbacks, err
}

// GetCallbacksByHeight returns the callbacks registered for the given height
func (k Keeper) GetCallbacksByHeight(ctx sdk.Context, height int64) (callbacks []*types.Callback, err error) {
	rng := collections.NewPrefixedTripleRange[int64, []byte, uint64](height)
	err = k.Callbacks.Walk(ctx, rng, func(key collections.Triple[int64, []byte, uint64], value types.Callback) (bool, error) {
		callbacks = append(callbacks, &value)
		return false, nil
	})
	return callbacks, err
}

// IterateCallbacksByHeight iterates over callbacks for registered for the given height and executes them
func (k Keeper) IterateCallbacksByHeight(ctx sdk.Context, height int64, exec func(types.Callback) bool) {
	rng := collections.NewPrefixedTripleRange[int64, []byte, uint64](height)
	_ = k.Callbacks.Walk(ctx, rng, func(key collections.Triple[int64, []byte, uint64], value types.Callback) (bool, error) {
		exec(value)
		return false, nil
	})
}

// ExistsCallback returns true if the callback exists for height with same contract address and same job id
func (k Keeper) ExistsCallback(ctx sdk.Context, height int64, contractAddr string, jobID uint64) (bool, error) {
	contractAddress, err := sdk.AccAddressFromBech32(contractAddr)
	if err != nil {
		return false, err
	}
	return k.Callbacks.Has(ctx, collections.Join3(height, contractAddress.Bytes(), jobID))
}

// GetCallback returns the callback given the height, contract address and job id
func (k Keeper) GetCallback(ctx sdk.Context, height int64, contractAddr string, jobID uint64) (types.Callback, error) {
	contractAddress, err := sdk.AccAddressFromBech32(contractAddr)
	if err != nil {
		return types.Callback{}, err
	}

	return k.Callbacks.Get(ctx, collections.Join3(height, contractAddress.Bytes(), jobID))
}

// DeleteCallback deletes a callback given the height, contract address and job id
func (k Keeper) DeleteCallback(ctx sdk.Context, sender string, callback types.Callback) error {
	contractAddress, err := sdk.AccAddressFromBech32(callback.ContractAddress)
	if err != nil {
		return err
	}
	// If callback delete is requested by someone who is not authorized, return error
	if !isAuthorizedToModify(ctx, k, contractAddress, sender) {
		return types.ErrUnauthorized
	}
	return k.Callbacks.Remove(ctx, collections.Join3(callback.CallbackHeight, contractAddress.Bytes(), callback.JobId))
}

// SaveCallback saves a callback given the height, contract address and job id and callback data
func (k Keeper) SaveCallback(ctx sdk.Context, callback types.Callback) error {
	contractAddress, err := sdk.AccAddressFromBech32(callback.ContractAddress)
	if err != nil {
		return err
	}
	// If contract with given address does not exist, return error
	if !k.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		return types.ErrContractNotFound
	}
	// If callback is requested by someone which is not authorized, return error
	if !isAuthorizedToModify(ctx, k, contractAddress, callback.ReservedBy) {
		return types.ErrUnauthorized
	}
	// If a callback with same job id exists at same height, return error
	exists, err := k.ExistsCallback(ctx, callback.CallbackHeight, contractAddress.String(), callback.JobId)
	if err != nil {
		return err
	}
	if exists {
		return types.ErrCallbackExists
	}
	// If callback is requested for height in the past or present, return error
	if callback.CallbackHeight <= ctx.BlockHeight() {
		return types.ErrCallbackHeightNotInFuture
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	// If callback is requested for height which is too far in the future, return error
	maxFutureReservationHeight := ctx.BlockHeight() + int64(params.MaxFutureReservationLimit)
	if callback.CallbackHeight > maxFutureReservationHeight {
		return types.ErrCallbackHeightTooFarInFuture
	}
	// If there are already too many callbacks registered in a given block, return error
	callbacksForBlock, err := k.GetCallbacksByHeight(ctx, callback.CallbackHeight)
	if err != nil {
		return err
	}
	if len(callbacksForBlock) >= int(params.MaxBlockReservationLimit) {
		return types.ErrBlockFilled
	}

	// Setting the callback gas limit to the module param CallbackGasLimit.
	// This is to ensure that if the param value is decreased in the future, before the callback is executed,
	// it does not fail with "out of gas" error. it wouldnt be fair for the contract to err out of the callback
	// if it wouldnt have been expected to at the time of registration
	params, err = k.GetParams(ctx)
	if err != nil {
		return err
	}
	callback.MaxGasLimit = params.CallbackGasLimit

	return k.Callbacks.Set(ctx, collections.Join3(callback.CallbackHeight, contractAddress.Bytes(), callback.JobId), callback)
}

// nolint: gosimple
func isAuthorizedToModify(ctx sdk.Context, k Keeper, contractAddress sdk.AccAddress, sender string) bool {
	if k.bankKeeper.BlockedAddr(sdk.MustAccAddressFromBech32(sender)) { // Blocked addresses cannot create/delete callbacks as we cant refund to these addresses. And they are module accounts anyway
		return false
	}

	if strings.EqualFold(sender, contractAddress.String()) { // A contract can modify its own callbacks
		return true
	}

	contractInfo := k.wasmKeeper.GetContractInfo(ctx, contractAddress)
	if strings.EqualFold(sender, contractInfo.Admin) { // Admin of the contract can modify its callbacks
		return true
	}

	return false
}
