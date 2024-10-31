// This file implements the poolmanagertypes.PoolModule interface
package cosmwasmpool

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/cosmwasm/msg"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/types"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/events"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"

	"github.com/osmosis-labs/osmosis/osmoutils/cosmwasm"
)

var (
	emptyCoins = sdk.NewCoins()
)

// It converts the given pool to a CosmWasmPool, instantiates the Wasm contract using the contract keeper,
// and then sets the contract address in the CosmWasmPool model before storing it.
// The method returns an error if the pool conversion, contract instantiation, or storage process fails.
//
// Parameters:
// - ctx: The SDK context.
// - pool: The pool interface to be initialized.
// - creatorAddress: The address of the creator of the pool.
//
// Returns:
// - error:
// * if the pool conversion, contract instantiation, or storage process fails.
// * if the code id is not whitelisted by governance.
// - otherwise, nil.
func (k Keeper) InitializePool(ctx sdk.Context, pool poolmanagertypes.PoolI, creatorAddress sdk.AccAddress) error {
	// Convert the pool to CosmWasmPool
	cosmwasmPool, err := k.asCosmwasmPool(pool)
	if err != nil {
		return err
	}

	// Check if the code id is whitelisted.
	codeId := cosmwasmPool.GetCodeId()
	if !k.isWhitelisted(ctx, codeId) {
		return types.CodeIdNotWhitelistedError{CodeId: codeId}
	}

	k.WhitelistCodeId(ctx, codeId)

	cosmwasmpoolModuleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)

	// Instantiate the wasm contract
	contractAddress, _, err := k.contractKeeper.Instantiate(ctx, codeId, cosmwasmpoolModuleAddr, cosmwasmpoolModuleAddr, cosmwasmPool.GetInstantiateMsg(), types.ModuleName, emptyCoins)
	if err != nil {
		return err
	}

	// Store the address in pool model
	cosmwasmPool.SetContractAddress(contractAddress.String())

	// Store the pool model
	k.SetPool(ctx, cosmwasmPool)

	return nil
}

// GetPool retrieves a pool model with the specified pool ID from the store.
// The method returns the pool interface of the corresponding pool model if found, and an error if not found.
//
// Parameters:
// - ctx: The SDK context.
// - poolId: The unique identifier of the pool.
//
// Returns:
// - poolmanagertypes.PoolI: The pool interface of the corresponding pool model, if found.
// - error: An error if the pool model is not found; otherwise, nil.
func (k Keeper) GetPool(ctx sdk.Context, poolId uint64) (poolmanagertypes.PoolI, error) {
	cwPool, err := k.GetPoolById(ctx, poolId)
	if err != nil {
		return nil, err
	}
	return cwPool, nil
}

// GetPools retrieves all pool objects stored in the keeper.
//
// It fetches values from the store associated with the PoolsKey prefix. For each value retrieved,
// it attempts to unmarshal the value into a Pool object. If this operation succeeds,
// the Pool object is added to the returned slice. If an error occurs during unmarshalling,
// the function will return immediately with the encountered error.
//
// Parameters:
// - ctx: The current SDK Context used to access the store.
//
// Returns:
//   - A slice of PoolI interfaces if the operation is successful. Each element in the slice
//     represents a pool that was stored in the keeper.
//   - An error if unmarshalling fails for any of the values fetched from the store.
//     In this case, the slice of PoolI interfaces will be nil.
func (k Keeper) GetPools(ctx sdk.Context) ([]poolmanagertypes.PoolI, error) {
	return osmoutils.GatherValuesFromStorePrefix(
		ctx.KVStore(k.storeKey), types.PoolsKey, func(value []byte) (poolmanagertypes.PoolI, error) {
			pool := model.Pool{}
			err := k.cdc.Unmarshal(value, &pool)
			if err != nil {
				return nil, err
			}
			return &pool, nil
		},
	)
}

// GetPoolsSerializable retrieves all pool objects stored in the keeper.
// Because the Pool struct has a non-serializable wasmKeeper field, this method
// utilizes the CosmWasmPool struct directly instead, which allows it to be serialized
// in import/export genesis.
func (k Keeper) GetPoolsSerializable(ctx sdk.Context) ([]poolmanagertypes.PoolI, error) {
	return osmoutils.GatherValuesFromStorePrefix(
		ctx.KVStore(k.storeKey), types.PoolsKey, func(value []byte) (poolmanagertypes.PoolI, error) {
			pool := model.CosmWasmPool{}
			err := k.cdc.Unmarshal(value, &pool)
			if err != nil {
				return nil, err
			}
			return &pool, nil
		},
	)
}

// GetPoolsWithWasmKeeper behaves the same as GetPools, but it also sets the WasmKeeper field of the pool.
func (k Keeper) GetPoolsWithWasmKeeper(ctx sdk.Context) ([]poolmanagertypes.PoolI, error) {
	return osmoutils.GatherValuesFromStorePrefix(
		ctx.KVStore(k.storeKey), types.PoolsKey, func(value []byte) (poolmanagertypes.PoolI, error) {
			pool := model.CosmWasmPool{}
			err := k.cdc.Unmarshal(value, &pool)
			if err != nil {
				return nil, err
			}
			return &model.Pool{
				CosmWasmPool: pool,
				WasmKeeper:   k.wasmKeeper,
			}, nil
		},
	)
}

// GetPoolDenoms retrieves the list of asset denoms in a CosmWasm-based liquidity pool given its ID.
//
// Parameters:
// - ctx: The context of the query request.
// - poolId: The unique identifier of the CosmWasm-based liquidity pool.
//
// Returns:
// - denoms: A slice of strings representing the asset denoms in the liquidity pool.
// - err: An error if the pool cannot be found or if the CosmWasm query fails.
func (k Keeper) GetPoolDenoms(ctx sdk.Context, poolId uint64) (denoms []string, err error) {
	cosmwasmPool, err := k.GetPoolById(ctx, poolId)
	if err != nil {
		return nil, err
	}

	liquidity := cosmwasmPool.GetTotalPoolLiquidity(ctx)

	denoms = make([]string, 0, liquidity.Len())
	for _, coin := range liquidity {
		denoms = append(denoms, coin.Denom)
	}

	return denoms, nil
}

// CalculateSpotPrice calculates the spot price of a pair of assets in a CosmWasm-based liquidity pool.
//
// Parameters:
// - ctx: The context of the query request.
// - poolId: The unique identifier of the CosmWasm-based liquidity pool.
// - quoteAssetDenom: The denom of the quote asset in the trading pair.
// - baseAssetDenom: The denom of the base asset in the trading pair.
//
// Returns:
// - price: The spot price of the trading pair in the specified liquidity pool.
// - err: An error if the pool cannot be found or if the spot price calculation fails.
func (k Keeper) CalculateSpotPrice(
	ctx sdk.Context,
	poolId uint64,
	quoteAssetDenom string,
	baseAssetDenom string,
) (price osmomath.BigDec, err error) {
	cosmwasmPool, err := k.GetPoolById(ctx, poolId)
	if err != nil {
		return osmomath.BigDec{}, err
	}

	spotPriceBigDec, err := cosmwasmPool.SpotPrice(ctx, quoteAssetDenom, baseAssetDenom)
	if err != nil {
		return osmomath.BigDec{}, err
	}
	// Truncation is acceptable here since the only reason cosmwasmPool returns a BigDec
	// is to maintain compatibility with the `PoolI.SpotPrice` API.
	return spotPriceBigDec, nil
}

// SwapExactAmountIn performs a swap operation with a specified input amount in a CosmWasm-based liquidity pool.
//
// Parameters:
// - ctx: The context of the operation.
// - sender: The address of the account initiating the swap.
// - pool: The liquidity pool in which the swap occurs.
// - tokenIn: The input token (asset) to be swapped.
// - tokenOutDenom: The denom of the output token (asset) to be received.
// - tokenOutMinAmount: The minimum amount of the output token to be received.
// - swapFee: The fee associated with the swap operation.
//
// Returns:
// - osmomath.Int: The actual amount of the output token received after the swap.
// - error: An error if the swap operation fails or if the pool conversion fails.
func (k Keeper) SwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool poolmanagertypes.PoolI,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount osmomath.Int,
	swapFee osmomath.Dec,
) (osmomath.Int, error) {
	cosmwasmPool, err := k.asCosmwasmPool(pool)
	if err != nil {
		return osmomath.Int{}, err
	}

	// Send token in from sender to the pool
	// We do this because sudo message does not support sending coins from the sender
	// However, note that the contract sends the token back to the sender after the swap
	// As a result, we do not need to worry about sending it back here.
	if err := k.bankKeeper.SendCoins(ctx, sender, sdk.MustAccAddressFromBech32(cosmwasmPool.GetContractAddress()), sdk.NewCoins(tokenIn)); err != nil {
		return osmomath.Int{}, err
	}

	request := msg.NewSwapExactAmountInSudoMsg(sender.String(), tokenIn, tokenOutDenom, tokenOutMinAmount, swapFee)
	response, err := cosmwasm.Sudo[msg.SwapExactAmountInSudoMsg, msg.SwapExactAmountInSudoMsgResponse](ctx, k.contractKeeper, cosmwasmPool.GetContractAddress(), request)
	if err != nil {
		return osmomath.Int{}, err
	}

	// Emit swap event. Note that we emit these at the layer of each pool module rather than the poolmanager module
	// since poolmanager has many swap wrapper APIs that we would need to consider.
	// Search for references to this function to see where else it is used.
	// Each new pool module will have to emit this event separately
	events.EmitSwapEvent(ctx, sender, pool.GetId(), sdk.Coins{tokenIn}, sdk.Coins{sdk.Coin{Denom: tokenOutDenom, Amount: response.TokenOutAmount}})

	return response.TokenOutAmount, nil
}

// CalcOutAmtGivenIn calculates the output amount of a token given the input token amount in a CosmWasm-based liquidity pool.
//
// Parameters:
// - ctx: The context of the operation.
// - poolI: The liquidity pool to perform the calculation on.
// - tokenIn: The input token (asset) to be used in the calculation.
// - tokenOutDenom: The denom of the output token (asset) to be received.
// - swapFee: The fee associated with the swap operation.
//
// Returns:
// - sdk.Coin: The calculated output token amount.
// - error: An error if the calculation fails or if the pool conversion fails.
func (k Keeper) CalcOutAmtGivenIn(
	ctx sdk.Context,
	poolI poolmanagertypes.PoolI,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	swapFee osmomath.Dec,
) (tokenOut sdk.Coin, err error) {
	cosmwasmPool, err := k.asCosmwasmPool(poolI)
	if err != nil {
		return sdk.Coin{}, err
	}

	request := msg.NewCalcOutAmtGivenInRequest(tokenIn, tokenOutDenom, swapFee)
	response, err := cosmwasm.Query[msg.CalcOutAmtGivenInRequest, msg.CalcOutAmtGivenInResponse](ctx, k.wasmKeeper, cosmwasmPool.GetContractAddress(), request)
	if err != nil {
		return sdk.Coin{}, err
	}

	return response.TokenOut, nil
}

// SwapExactAmountOut performs a swap operation with a specified output amount in a CosmWasm-based liquidity pool.
//
// Parameters:
// - ctx: The context of the operation.
// - sender: The address of the account initiating the swap.
// - pool: The liquidity pool in which the swap occurs.
// - tokenInDenom: The denom of the input token (asset) to be swapped.
// - tokenInMaxAmount: The maximum amount of the input token allowed to be swapped.
// - tokenOut: The output token (asset) to be received.
// - swapFee: The fee associated with the swap operation.
//
// Returns:
// - osmomath.Int: The actual amount of the input token used in the swap.
// - error: An error if the swap operation fails or if the pool conversion fails.
func (k Keeper) SwapExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool poolmanagertypes.PoolI,
	tokenInDenom string,
	tokenInMaxAmount osmomath.Int,
	tokenOut sdk.Coin,
	swapFee osmomath.Dec,
) (tokenInAmount osmomath.Int, err error) {
	cosmwasmPool, err := k.asCosmwasmPool(pool)
	if err != nil {
		return osmomath.Int{}, err
	}

	contractAddr := sdk.MustAccAddressFromBech32(cosmwasmPool.GetContractAddress())

	// Send token in max amount from sender to the pool
	// We do this because sudo message does not support sending coins from the sender
	// And we need to send the max amount because we do not know how much the contract will use.
	if err := k.bankKeeper.SendCoins(ctx, sender, contractAddr, sdk.NewCoins(sdk.NewCoin(tokenInDenom, tokenInMaxAmount))); err != nil {
		return osmomath.Int{}, err
	}

	// Note that the contract sends the token out back to the sender after the swap
	// As a result, we do not need to worry about sending token out here.
	request := msg.NewSwapExactAmountOutSudoMsg(sender.String(), tokenInDenom, tokenOut, tokenInMaxAmount, swapFee)
	response, err := cosmwasm.Sudo[msg.SwapExactAmountOutSudoMsg, msg.SwapExactAmountOutSudoMsgResponse](ctx, k.contractKeeper, cosmwasmPool.GetContractAddress(), request)
	if err != nil {
		return osmomath.Int{}, err
	}

	tokenInExcessiveAmount := tokenInMaxAmount.Sub(response.TokenInAmount)

	// Do not send any coins if excessive amount is zero

	// required amount should be less than or equal to max amount
	if tokenInExcessiveAmount.IsNegative() {
		return osmomath.Int{}, types.NegativeExcessiveTokenInAmountError{
			TokenInMaxAmount:       tokenInMaxAmount,
			TokenInRequiredAmount:  response.TokenInAmount,
			TokenInExcessiveAmount: tokenInExcessiveAmount,
		}
	}

	// Send excessibe token in from pool back to sender
	if tokenInExcessiveAmount.IsPositive() {
		if err := k.bankKeeper.SendCoins(ctx, contractAddr, sender, sdk.NewCoins(sdk.NewCoin(tokenInDenom, tokenInExcessiveAmount))); err != nil {
			return osmomath.Int{}, err
		}
	}

	// Emit swap event. Note that we emit these at the layer of each pool module rather than the poolmanager module
	// since poolmanager has many swap wrapper APIs that we would need to consider.
	// Search for references to this function to see where else it is used.
	// Each new pool module will have to emit this event separately
	events.EmitSwapEvent(ctx, sender, pool.GetId(), sdk.Coins{sdk.Coin{Denom: tokenInDenom, Amount: response.TokenInAmount}}, sdk.Coins{tokenOut})

	return response.TokenInAmount, nil
}

// CalcInAmtGivenOut calculates the input amount of a token required to get the desired output token amount in a CosmWasm-based liquidity pool.
//
// Parameters:
// - ctx: The context of the operation.
// - poolI: The liquidity pool to perform the calculation on.
// - tokenOut: The desired output token (asset) amount.
// - tokenInDenom: The denom of the input token (asset) to be used in the calculation.
// - swapFee: The fee associated with the swap operation.
//
// Returns:
// - sdk.Coin: The calculated input token amount.
// - error: An error if the calculation fails or if the pool conversion fails.
func (k Keeper) CalcInAmtGivenOut(
	ctx sdk.Context,
	poolI poolmanagertypes.PoolI,
	tokenOut sdk.Coin,
	tokenInDenom string,
	swapFee osmomath.Dec,
) (tokenIn sdk.Coin, err error) {
	cosmwasmPool, err := k.asCosmwasmPool(poolI)
	if err != nil {
		return sdk.Coin{}, err
	}

	request := msg.NewCalcInAmtGivenOutRequest(tokenInDenom, tokenOut, swapFee)
	response, err := cosmwasm.Query[msg.CalcInAmtGivenOutRequest, msg.CalcInAmtGivenOutResponse](ctx, k.wasmKeeper, cosmwasmPool.GetContractAddress(), request)
	if err != nil {
		return sdk.Coin{}, err
	}

	return response.TokenIn, nil
}

// GetTotalPoolLiquidity retrieves the total liquidity of a specific pool identified by poolId.
//
// Parameters:
// - ctx: The current SDK Context used for executing store operations.
// - poolId: The unique identifier of the pool whose total liquidity is to be fetched.
//
// Returns:
//   - the total liquidity of the specified pool,
//     if the operations are successful.
//   - An error if the pool retrieval operation fails. In this case, an empty sdk.Coins object will be returned.
func (k Keeper) GetTotalPoolLiquidity(ctx sdk.Context, poolId uint64) (sdk.Coins, error) {
	pool, err := k.GetPoolById(ctx, poolId)
	if err != nil {
		return sdk.Coins{}, err
	}
	return pool.GetTotalPoolLiquidity(ctx), nil
}

// GetTotalLiquidity retrieves the total liquidity of all cw pools.
func (k Keeper) GetTotalLiquidity(ctx sdk.Context) (sdk.Coins, error) {
	totalLiquidity := sdk.Coins{}
	pools, err := k.GetPoolsWithWasmKeeper(ctx)
	if err != nil {
		return sdk.Coins{}, err
	}
	for _, poolI := range pools {
		cosmwasmPool, ok := poolI.(types.CosmWasmExtension)
		if !ok {
			return nil, types.InvalidPoolTypeError{
				ActualPool: poolI,
			}
		}
		totalPoolLiquidity := cosmwasmPool.GetTotalPoolLiquidity(ctx)
		// We range over the coins and add them one at a time because GetTotalPoolLiquidity
		// doesn't always return a sorted list of coins.
		for _, coin := range totalPoolLiquidity {
			totalLiquidity = totalLiquidity.Add(coin)
		}
	}
	return totalLiquidity, nil
}
