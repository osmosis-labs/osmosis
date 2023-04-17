package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
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

// This posthandler will first check if there were any swaps in the tx. If so, collect all of the pools, build routes for cyclic arbitrage,
// and then execute the optimal route if it exists.
func (protoRevDec ProtoRevDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if ctx.IsCheckTx() {
		return next(ctx, tx, simulate)
	}

	// Create a cache context to execute the posthandler such that
	// 1. If there is an error, then the cache context is discarded
	// 2. If there is no error, then the cache context is written to the main context with no gas consumed
	cacheCtx, write := ctx.CacheContext()
	// CacheCtx's by default _share_ their gas meter with the parent.
	// In our case, the cache ctx is given a new gas meter instance entirely,
	// so gas usage is not counted towards tx gas usage.
	//
	// 50M is chosen as a large enough number to ensure that the posthandler will not run out of gas,
	// but will eventually terminate in event of an accidental infinite loop with some gas usage.
	cacheCtx = cacheCtx.WithGasMeter(sdk.NewGasMeter(sdk.Gas(50_000_000)))

	// Check if the protorev posthandler can be executed
	if err := protoRevDec.ProtoRevKeeper.AnteHandleCheck(cacheCtx); err != nil {
		return next(ctx, tx, simulate)
	}

	// Extract all of the pools that were swapped in the tx
	swappedPools := ExtractSwappedPools(tx)
	if len(swappedPools) == 0 {
		return next(ctx, tx, simulate)
	}

	// Attempt to execute arbitrage trades
	if err := protoRevDec.ProtoRevKeeper.ProtoRevTrade(cacheCtx, swappedPools); err == nil {
		write()
		ctx.EventManager().EmitEvents(cacheCtx.EventManager().Events())
	} else {
		ctx.Logger().Error("ProtoRevTrade failed with error", err)
	}

	return next(ctx, tx, simulate)
}

// AnteHandleCheck checks if the module is enabled and if the number of routes to be processed per block has been reached.
func (k Keeper) AnteHandleCheck(ctx sdk.Context) error {
	// Only execute the posthandler if the module is enabled
	if !k.GetProtoRevEnabled(ctx) {
		return fmt.Errorf("protorev is not enabled")
	}

	latestBlockHeight, err := k.GetLatestBlockHeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block height")
	}

	currentRouteCount, err := k.GetPointCountForBlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current pool point count")
	}

	maxRouteCount, err := k.GetMaxPointsPerBlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to get max pool points per block")
	}

	// Only execute the posthandler if the number of routes to be processed per block has not been reached
	blockHeight := uint64(ctx.BlockHeight())
	if blockHeight == latestBlockHeight {
		if currentRouteCount >= maxRouteCount {
			return fmt.Errorf("max pool points for the current block has been reached")
		}
	} else {
		// Reset the current pool point count
		k.SetPointCountForBlock(ctx, 0)
		k.SetLatestBlockHeight(ctx, blockHeight)
	}

	return nil
}

// ProtoRevTrade wraps around the build routes, iterate routes, and execute trade functionality to execute cyclic arbitrage trades
// if they exist. It returns an error if there was an issue executing any single trade.
func (k Keeper) ProtoRevTrade(ctx sdk.Context, swappedPools []SwapToBackrun) (err error) {
	// recover from panic
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Protorev failed due to internal reason: %v", r)
		}
	}()

	// Get the total number of pool points that can be consumed in this transaction
	remainingTxPoolPoints, remainingBlockPoolPoints, err := k.GetRemainingPoolPoints(ctx)
	if err != nil {
		return err
	}

	// Iterate and build arbitrage routes for each pool that was swapped on
	for _, pool := range swappedPools {
		// Build the routes for the pool that was swapped on
		routes := k.BuildRoutes(ctx, pool.TokenInDenom, pool.TokenOutDenom, pool.PoolId)

		// Find optimal route (input coin, profit, route) for the given routes
		maxProfitInputCoin, maxProfitAmount, optimalRoute := k.IterateRoutes(ctx, routes, &remainingTxPoolPoints, &remainingBlockPoolPoints)

		// The error that returns here is particularly focused on the minting/burning of coins, and the execution of the MultiHopSwapExactAmountIn.
		if maxProfitAmount.GT(sdk.ZeroInt()) {
			if err := k.ExecuteTrade(ctx, optimalRoute, maxProfitInputCoin, pool, remainingTxPoolPoints, remainingBlockPoolPoints); err != nil {
				return err
			}
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
		switch msg := msg.(type) {
		case *poolmanagertypes.MsgSwapExactAmountIn:
			swappedPools = append(swappedPools, extractSwapInPools(msg.Routes, msg.TokenIn.Denom)...)
		case *poolmanagertypes.MsgSwapExactAmountOut:
			swappedPools = append(swappedPools, extractSwapOutPools(msg.Routes, msg.TokenOut.Denom)...)
		case *gammtypes.MsgSwapExactAmountIn:
			swappedPools = append(swappedPools, extractSwapInPools(msg.Routes, msg.TokenIn.Denom)...)
		case *gammtypes.MsgSwapExactAmountOut:
			swappedPools = append(swappedPools, extractSwapOutPools(msg.Routes, msg.TokenOut.Denom)...)
		}
	}

	return swappedPools
}

// extractSwapInPools extracts the pools that were swapped on for a MsgSwapExactAmountIn
func extractSwapInPools(routes []poolmanagertypes.SwapAmountInRoute, tokenInDenom string) []SwapToBackrun {
	swappedPools := make([]SwapToBackrun, 0)

	prevTokenIn := tokenInDenom
	for _, route := range routes {
		swappedPools = append(swappedPools, SwapToBackrun{
			PoolId:        route.PoolId,
			TokenOutDenom: route.TokenOutDenom,
			TokenInDenom:  prevTokenIn})

		prevTokenIn = route.TokenOutDenom
	}

	return swappedPools
}

// extractSwapOutPools extracts the pools that were swapped on for a MsgSwapExactAmountOut
func extractSwapOutPools(routes []poolmanagertypes.SwapAmountOutRoute, tokenOutDenom string) []SwapToBackrun {
	swappedPools := make([]SwapToBackrun, 0)

	prevTokenOut := tokenOutDenom
	for i := len(routes) - 1; i >= 0; i-- {
		route := routes[i]
		swappedPools = append(swappedPools, SwapToBackrun{
			PoolId:        route.PoolId,
			TokenOutDenom: prevTokenOut,
			TokenInDenom:  route.TokenInDenom})

		prevTokenOut = route.TokenInDenom
	}

	return swappedPools
}
