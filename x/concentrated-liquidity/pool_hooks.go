package concentrated_liquidity

import (
	"encoding/json"

	"cosmossdk.io/store"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	types "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

// --- Pool Hooks ---

type msgBuilderFn func(poolId uint64) ([]byte, error)

// BeforeCreatePosition is a hook that is called before a position is created.
func (k Keeper) BeforeCreatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, tokensProvided sdk.Coins, amount0Min osmomath.Int, amount1Min osmomath.Int, lowerTick int64, upperTick int64) error {
	msgBuilderFn := func(poolId uint64) ([]byte, error) {
		msg := types.BeforeCreatePositionMsg{PoolId: poolId, Owner: owner, TokensProvided: osmoutils.CWCoinsFromSDKCoins(tokensProvided), Amount0Min: amount0Min, Amount1Min: amount1Min, LowerTick: lowerTick, UpperTick: upperTick}
		return json.Marshal(types.BeforeCreatePositionSudoMsg{BeforeCreatePosition: msg})
	}
	return k.callPoolActionListener(ctx, msgBuilderFn, poolId, types.BeforeActionPrefix(types.CreatePositionPrefix))
}

// AfterCreatePosition is a hook that is called after a position is created.
func (k Keeper) AfterCreatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, tokensProvided sdk.Coins, amount0Min osmomath.Int, amount1Min osmomath.Int, lowerTick int64, upperTick int64) error {
	msgBuilderFn := func(poolId uint64) ([]byte, error) {
		msg := types.AfterCreatePositionMsg{PoolId: poolId, Owner: owner, TokensProvided: osmoutils.CWCoinsFromSDKCoins(tokensProvided), Amount0Min: amount0Min, Amount1Min: amount1Min, LowerTick: lowerTick, UpperTick: upperTick}
		return json.Marshal(types.AfterCreatePositionSudoMsg{AfterCreatePosition: msg})
	}
	return k.callPoolActionListener(ctx, msgBuilderFn, poolId, types.AfterActionPrefix(types.CreatePositionPrefix))
}

// BeforeWithdrawPosition is a hook that is called before liquidity is withdrawn from a position.
func (k Keeper) BeforeWithdrawPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, positionId uint64, amountToWithdraw osmomath.Dec) error {
	msgBuilderFn := func(poolId uint64) ([]byte, error) {
		msg := types.BeforeWithdrawPositionMsg{PoolId: poolId, Owner: owner, PositionId: positionId, AmountToWithdraw: amountToWithdraw}
		return json.Marshal(types.BeforeWithdrawPositionSudoMsg{BeforeWithdrawPosition: msg})
	}
	return k.callPoolActionListener(ctx, msgBuilderFn, poolId, types.BeforeActionPrefix(types.WithdrawPositionPrefix))
}

// AfterWithdrawPosition is a hook that is called after liquidity is withdrawn from a position.
func (k Keeper) AfterWithdrawPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, positionId uint64, amountToWithdraw osmomath.Dec) error {
	msgBuilderFn := func(poolId uint64) ([]byte, error) {
		msg := types.AfterWithdrawPositionMsg{PoolId: poolId, Owner: owner, PositionId: positionId, AmountToWithdraw: amountToWithdraw}
		return json.Marshal(types.AfterWithdrawPositionSudoMsg{AfterWithdrawPosition: msg})
	}
	return k.callPoolActionListener(ctx, msgBuilderFn, poolId, types.AfterActionPrefix(types.WithdrawPositionPrefix))
}

// BeforeSwapExactAmountIn is a hook that is called before a swap is executed (exact amount in).
func (k Keeper) BeforeSwapExactAmountIn(ctx sdk.Context, poolId uint64, sender sdk.AccAddress, tokenIn sdk.Coin, tokenOutDenom string, tokenOutMinAmount osmomath.Int, spreadFactor osmomath.Dec) error {
	msgBuilderFn := func(poolId uint64) ([]byte, error) {
		msg := types.BeforeSwapExactAmountInMsg{PoolId: poolId, Sender: sender, TokenIn: osmoutils.CWCoinFromSDKCoin(tokenIn), TokenOutDenom: tokenOutDenom, TokenOutMinAmount: tokenOutMinAmount, SpreadFactor: spreadFactor}
		return json.Marshal(types.BeforeSwapExactAmountInSudoMsg{BeforeSwapExactAmountIn: msg})
	}
	return k.callPoolActionListener(ctx, msgBuilderFn, poolId, types.BeforeActionPrefix(types.SwapExactAmountInPrefix))
}

// AfterSwapExactAmountIn is a hook that is called after a swap is executed (exact amount in).
func (k Keeper) AfterSwapExactAmountIn(ctx sdk.Context, poolId uint64, sender sdk.AccAddress, tokenIn sdk.Coin, tokenOutDenom string, tokenOutMinAmount osmomath.Int, spreadFactor osmomath.Dec) error {
	msgBuilderFn := func(poolId uint64) ([]byte, error) {
		msg := types.AfterSwapExactAmountInMsg{PoolId: poolId, Sender: sender, TokenIn: osmoutils.CWCoinFromSDKCoin(tokenIn), TokenOutDenom: tokenOutDenom, TokenOutMinAmount: tokenOutMinAmount, SpreadFactor: spreadFactor}
		return json.Marshal(types.AfterSwapExactAmountInSudoMsg{AfterSwapExactAmountIn: msg})
	}
	return k.callPoolActionListener(ctx, msgBuilderFn, poolId, types.AfterActionPrefix(types.SwapExactAmountInPrefix))
}

// BeforeSwapExactAmountOut is a hook that is called before a swap is executed (exact amount out).
func (k Keeper) BeforeSwapExactAmountOut(ctx sdk.Context, poolId uint64, sender sdk.AccAddress, tokenInDenom string, tokenInMaxAmount osmomath.Int, tokenOut sdk.Coin, spreadFactor osmomath.Dec) error {
	msgBuilderFn := func(poolId uint64) ([]byte, error) {
		msg := types.BeforeSwapExactAmountOutMsg{PoolId: poolId, Sender: sender, TokenInDenom: tokenInDenom, TokenInMaxAmount: tokenInMaxAmount, TokenOut: osmoutils.CWCoinFromSDKCoin(tokenOut), SpreadFactor: spreadFactor}
		return json.Marshal(types.BeforeSwapExactAmountOutSudoMsg{BeforeSwapExactAmountOut: msg})
	}
	return k.callPoolActionListener(ctx, msgBuilderFn, poolId, types.BeforeActionPrefix(types.SwapExactAmountOutPrefix))
}

// AfterSwapExactAmountOut is a hook that is called after a swap is executed (exact amount out).
func (k Keeper) AfterSwapExactAmountOut(ctx sdk.Context, poolId uint64, sender sdk.AccAddress, tokenInDenom string, tokenInMaxAmount osmomath.Int, tokenOut sdk.Coin, spreadFactor osmomath.Dec) error {
	msgBuilderFn := func(poolId uint64) ([]byte, error) {
		msg := types.AfterSwapExactAmountOutMsg{PoolId: poolId, Sender: sender, TokenInDenom: tokenInDenom, TokenInMaxAmount: tokenInMaxAmount, TokenOut: osmoutils.CWCoinFromSDKCoin(tokenOut), SpreadFactor: spreadFactor}
		return json.Marshal(types.AfterSwapExactAmountOutSudoMsg{AfterSwapExactAmountOut: msg})
	}
	return k.callPoolActionListener(ctx, msgBuilderFn, poolId, types.AfterActionPrefix(types.SwapExactAmountOutPrefix))
}

// callPoolActionListener processes and dispatches the passed in message to the contract corresponding to the hook
// defined by the given pool ID and action prefix (e.g. pool Id: 1, action prefix: "beforeSwap").
//
// This function returns an error if the contract address in state is invalid (should be impossible) or if the contract execution fails.
//
// Since it is possible for this function to be triggered in begin block code, we need to directly meter its execution and set a limit.
// If no contract is linked to the hook, this function is a no-op.
func (k Keeper) callPoolActionListener(ctx sdk.Context, msgBuilderFn msgBuilderFn, poolId uint64, actionPrefix string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = types.ContractHookOutOfGasError{GasLimit: k.GetParams(ctx).HookGasLimit}
		}
	}()

	cosmwasmAddress := k.getPoolHookContract(ctx, poolId, actionPrefix)
	if cosmwasmAddress == "" {
		return nil
	}

	msgBz, err := msgBuilderFn(poolId)
	if err != nil {
		return err
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
	childCtx := ctx.WithGasMeter(storetypes.NewGasMeter(k.GetParams(ctx).HookGasLimit))
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
func (k Keeper) getPoolHookPrefixStore(ctx sdk.Context, poolID uint64) store.KVStore {
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

	validActionPrefixes := types.GetAllActionPrefixes()
	if !osmoutils.Contains(validActionPrefixes, actionPrefix) {
		return types.InvalidActionPrefixError{ActionPrefix: actionPrefix, ValidActions: validActionPrefixes}
	}

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
func deletePoolHookContract(store store.KVStore, actionPrefix string) {
	store.Delete([]byte(actionPrefix))
}
