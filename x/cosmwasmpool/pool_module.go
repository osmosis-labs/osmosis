// This file implements the poolmanagertypes.PoolModule interface
package cosmwasmpool

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/cosmwasm"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v15/x/tokenfactory/types"
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
// - error: An error if the pool conversion, contract instantiation, or storage process fails; otherwise, nil.
func (k Keeper) InitializePool(ctx sdk.Context, pool poolmanagertypes.PoolI, creatorAddress sdk.AccAddress) error {
	// Convert the pool to CosmWasmPool
	cosmwasmPool, err := k.convertToCosmwasmPool(pool)
	if err != nil {
		return err
	}

	// Instantiate the wasm contract
	contractAddress, _, err := k.contractKeeper.Instantiate(ctx, cosmwasmPool.GetCodeId(), cosmwasmPool.GetAddress(), cosmwasmPool.GetAddress(), cosmwasmPool.GetInstantiateMsg(), types.ModuleName, emptyCoins)
	if err != nil {
		return err
	}

	// Store the address in pool model
	cosmwasmPool.SetContractAddress(contractAddress.String())

	// Store the pool model
	k.setPool(ctx, cosmwasmPool)

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
	concentratedPool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return nil, err
	}
	return concentratedPool, nil
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
	cosmwasmPool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return nil, err
	}

	request := cosmwasm.GetPoolDenoms{}
	respose, err := cosmwasm.Query[cosmwasm.GetPoolDenoms, cosmwasm.GetPoolDenomsResponse](ctx, k.wasmKeeper, cosmwasmPool.GetContractAddress(), request)
	if err != nil {
		return nil, err
	}

	return respose.PoolDenoms, nil
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
) (price sdk.Dec, err error) {
	cosmwasmPool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.Dec{}, err
	}

	return cosmwasmPool.SpotPrice(ctx, quoteAssetDenom, baseAssetDenom)
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
// - sdk.Int: The actual amount of the output token received after the swap.
// - error: An error if the swap operation fails or if the pool conversion fails.
func (k Keeper) SwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool poolmanagertypes.PoolI,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount sdk.Int,
	swapFee sdk.Dec,
) (sdk.Int, error) {
	cosmwasmPool, err := k.convertToCosmwasmPool(pool)
	if err != nil {
		return sdk.Int{}, err
	}

	request := cosmwasm.NewSwapExactAmountInRequest(sender.String(), tokenIn, tokenOutDenom, tokenOutMinAmount, swapFee)
	response, err := cosmwasm.Sudo[cosmwasm.SwapExactAmountInRequest, cosmwasm.SwapExactAmountInResponse](ctx, k.contractKeeper, cosmwasmPool.GetContractAddress(), request)
	if err != nil {
		return sdk.Int{}, err
	}

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
	swapFee sdk.Dec,
) (tokenOut sdk.Coin, err error) {
	cosmwasmPool, err := k.convertToCosmwasmPool(poolI)
	if err != nil {
		return sdk.Coin{}, err
	}

	request := cosmwasm.NewCalcOutAmtGivenInRequest(tokenIn, tokenOutDenom, swapFee)
	response, err := cosmwasm.Query[cosmwasm.CalcOutAmtGivenInRequest, cosmwasm.CalcOutAmtGivenInResponse](ctx, k.wasmKeeper, cosmwasmPool.GetContractAddress(), request)
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
// - sdk.Int: The actual amount of the input token used in the swap.
// - error: An error if the swap operation fails or if the pool conversion fails.
func (k Keeper) SwapExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool poolmanagertypes.PoolI,
	tokenInDenom string,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
	swapFee sdk.Dec,
) (tokenInAmount sdk.Int, err error) {
	cosmwasmPool, err := k.convertToCosmwasmPool(pool)
	if err != nil {
		return sdk.Int{}, err
	}

	request := cosmwasm.NewSwapExactAmountOutRequest(sender.String(), tokenInDenom, tokenOut, tokenInMaxAmount, swapFee)
	response, err := cosmwasm.Sudo[cosmwasm.SwapExactAmountOutRequest, cosmwasm.SwapExactAmountOutResponse](ctx, k.contractKeeper, cosmwasmPool.GetContractAddress(), request)
	if err != nil {
		return sdk.Int{}, err
	}

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
	swapFee sdk.Dec,
) (tokenIn sdk.Coin, err error) {
	cosmwasmPool, err := k.convertToCosmwasmPool(poolI)
	if err != nil {
		return sdk.Coin{}, err
	}

	request := cosmwasm.NewCalcInAmtGivenOutRequest(tokenInDenom, tokenOut, swapFee)
	response, err := cosmwasm.Query[cosmwasm.CalcInAmtGivenOutRequest, cosmwasm.CalcInAmtGivenOutResponse](ctx, k.wasmKeeper, cosmwasmPool.GetContractAddress(), request)
	if err != nil {
		return sdk.Coin{}, err
	}

	return response.TokenIn, nil
}
