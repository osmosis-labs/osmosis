package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
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
	upperGasLimitMeter := sdk.NewGasMeter(sdk.Gas(50_000_000))
	cacheCtx = cacheCtx.WithGasMeter(upperGasLimitMeter)

	// Check if the protorev posthandler can be executed
	if err := protoRevDec.ProtoRevKeeper.AnteHandleCheck(cacheCtx); err != nil {
		return next(ctx, tx, simulate)
	}

	// Extract all of the pools that were swapped in the tx
	swappedPools := protoRevDec.ProtoRevKeeper.ExtractSwappedPools(cacheCtx)
	if len(swappedPools) == 0 {
		return next(ctx, tx, simulate)
	}

	// Attempt to execute arbitrage trades
	if err := protoRevDec.ProtoRevKeeper.ProtoRevTrade(cacheCtx, swappedPools); err == nil {
		write()
		ctx.EventManager().EmitEvents(cacheCtx.EventManager().Events())
	} else {
		ctx.Logger().Error("ProtoRevTrade failed with error: " + err.Error())
	}

	// Delete swaps to backrun for next transaction without consuming gas
	// from the current transaction's gas meter, but instead from a new gas meter with 50mil gas.
	// 50 mil gas was chosen as an arbitrary large number to ensure deletion does not run out of gas.
	protoRevDec.ProtoRevKeeper.DeleteSwapsToBackrun(ctx.WithGasMeter(sdk.NewGasMeter(sdk.Gas(50_000_000))))

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
		if maxProfitAmount.GT(osmomath.ZeroInt()) {
			if err := k.ExecuteTrade(ctx, optimalRoute, maxProfitInputCoin, pool, remainingTxPoolPoints, remainingBlockPoolPoints); err != nil {
				return err
			}
		}
	}

	return nil
}

// ExtractSwappedPools checks if there were any swaps made on pools and if so returns a list of all the pools that were
// swapped on and metadata about the swap
func (k Keeper) ExtractSwappedPools(ctx sdk.Context) []SwapToBackrun {
	swappedPools := make([]SwapToBackrun, 0)

	swapsToBackrun, err := k.GetSwapsToBackrun(ctx)
	if err != nil {
		return swappedPools
	}

	for _, swap := range swapsToBackrun.Trades {
		swappedPools = append(swappedPools, SwapToBackrun{
			PoolId:        swap.Pool,
			TokenInDenom:  swap.TokenIn,
			TokenOutDenom: swap.TokenOut,
		})
	}

	return swappedPools
}
