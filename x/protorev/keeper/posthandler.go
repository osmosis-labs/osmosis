package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v14/x/poolmanager/types"
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
	cacheCtx = cacheCtx.WithGasMeter(sdk.NewInfiniteGasMeter())

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
	if enabled, err := k.GetProtoRevEnabled(ctx); err != nil || !enabled {
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
func (k Keeper) ProtoRevTrade(ctx sdk.Context, swappedPools []SwapToBackrun) error {
	// Get the total number of pool points that can be consumed in this transaction
	maxPoolPoints, err := k.CalcMaxPoolPointsForTx(ctx)
	if err != nil {
		return err
	}

	// Iterate and build arbitrage routes for each pool that was swapped on
	for index := 0; index < len(swappedPools) && *maxPoolPoints > 0; index++ {
		// Build the routes for the pool that was swapped on
		routes := k.BuildRoutes(ctx, swappedPools[index].TokenInDenom, swappedPools[index].TokenOutDenom, swappedPools[index].PoolId, maxPoolPoints)

		// Find optimal input amounts for routes
		maxProfitInputCoin, maxProfitAmount, optimalRoute := k.IterateRoutes(ctx, routes)

		// The error that returns here is particularly focused on the minting/burning of coins, and the execution of the MultiHopSwapExactAmountIn.
		if maxProfitAmount.GT(sdk.ZeroInt()) {
			if err := k.ExecuteTrade(ctx, optimalRoute, maxProfitInputCoin); err != nil {
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
		if swap, ok := msg.(*poolmanagertypes.MsgSwapExactAmountIn); ok {
			for _, route := range swap.Routes {
				swappedPools = append(swappedPools, SwapToBackrun{
					PoolId:        route.PoolId,
					TokenOutDenom: route.TokenOutDenom,
					TokenInDenom:  swap.TokenIn.Denom})
			}
		} else if swap, ok := msg.(*poolmanagertypes.MsgSwapExactAmountOut); ok {
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
