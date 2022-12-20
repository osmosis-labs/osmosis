package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v13/x/gamm/types"
)

type SwapToBackrun struct {
	PoolId        uint64
	TokenOutDenom string
	TokenInDenom  string
}

type ProtoRevDecorator struct {
	ProtoRevKeeper Keeper
}

func NewProtoRevDecorator(protoRevDecorator Keeper) ProtoRevDecorator {
	return ProtoRevDecorator{
		ProtoRevKeeper: protoRevDecorator,
	}
}

// This posthandler will first check if there were any swaps in the tx. If so, collect all of the pools, build three
// pool routes for cyclic arbitrage, and then execute the optimal route if it exists.
func (protoRevDec ProtoRevDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// Create a cache context to execute the posthandler such that
	// 1. If there is an error, then the cache context is discarded
	// 2. If there is no error, then the cache context is written to the main context
	cacheCtx, write := ctx.CacheContext()

	txGasWanted := cacheCtx.GasMeter().Limit()
	// Ignore cases where limit is 0 (edge case for genUtil init)
	if txGasWanted == 0 {
		return next(ctx, tx, simulate)
	}

	// Change the gas meter to be an infinite gas meter which allows the entire posthandler to work without consuming any gas
	cacheCtx = cacheCtx.WithGasMeter(sdk.NewInfiniteGasMeter())

	// Only execute the protoRev module if it is enabled
	if enabled, err := protoRevDec.ProtoRevKeeper.GetProtoRevEnabled(cacheCtx); err != nil || !enabled {
		return next(ctx, tx, simulate)
	}

	// Get the max number of pools to iterate through
	maxPoolsToIterate, err := protoRevDec.ProtoRevKeeper.GetMaxPools(cacheCtx)
	if err != nil {
		return next(ctx, tx, simulate)
	}

	// Find routes for every single pool that was swapped on (up to maxPoolsToIterate pools per tx)
	swappedPools := ExtractSwappedPools(tx)
	tradeErr := error(nil)
	for index, swap := range swappedPools {
		if uint64(index) >= maxPoolsToIterate {
			break
		}
		// If there was is an error executing the trade, break and set tradeErr
		if err := protoRevDec.ProtoRevKeeper.ProtoRevTrade(cacheCtx, swap); err != nil {
			tradeErr = err
			break
		}
	}
	// If there was no error, write the cache context to the main context
	if tradeErr == nil {
		write()
		ctx.EventManager().EmitEvents(cacheCtx.EventManager().Events())
	}

	return next(ctx, tx, simulate)
}

// ProtoRevTrade wraps around the build routes, iterate routes, and execute trade functionality to execute an cyclic arbitrage trade
// if it exists. It returns an error if there was an error executing the trade and a boolean if the trade was executed.
func (k Keeper) ProtoRevTrade(ctx sdk.Context, swap SwapToBackrun) error {
	// Build the routes for the swap
	routes := k.BuildRoutes(ctx, swap.TokenInDenom, swap.TokenOutDenom, swap.PoolId)

	if len(routes) != 0 {
		// Find optimal input amounts for routes
		maxProfitInputCoin, maxProfitAmount, optimalRoute := k.IterateRoutes(ctx, routes)

		// The error that returns here is particularly focused on the minting/burning of coins, and the execution of the MultiHopSwapExactAmountIn.
		if maxProfitAmount.GT(sdk.ZeroInt()) {
			if err := k.ExecuteTrade(ctx, optimalRoute, maxProfitInputCoin, swap.PoolId); err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

// ExtractSwappedPools checks if there were any swaps made on pools and if so returns a list of all the pools that were
// swapped on and metadata about the swap
func ExtractSwappedPools(tx sdk.Tx) []SwapToBackrun {
	swappedPools := make([]SwapToBackrun, 0)

	// Extract only swaps types and the swapped pools from the tx
	for _, msg := range tx.GetMsgs() {
		if swap, ok := msg.(*gammtypes.MsgSwapExactAmountIn); ok {
			for _, route := range swap.Routes {
				swappedPools = append(swappedPools, SwapToBackrun{
					PoolId:        route.PoolId,
					TokenOutDenom: route.TokenOutDenom,
					TokenInDenom:  swap.TokenIn.Denom})
			}
		} else if swap, ok := msg.(*gammtypes.MsgSwapExactAmountOut); ok {
			for _, route := range swap.Routes {
				swappedPools = append(swappedPools, SwapToBackrun{
					PoolId:        route.PoolId,
					TokenOutDenom: swap.TokenOut.Denom,
					TokenInDenom:  route.TokenInDenom})
			}
		}
	}

	return swappedPools
}
