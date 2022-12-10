package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v13/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v13/x/protorev/types"
)

type TradeInfo struct {
	InputDenom  string
	OutputDenom string
	SwapFee     sdk.Dec
	Pool        gammtypes.CFMMPoolI
}

// BuildRoutes builds all of the possible arbitrage routes given the given tokenIn, tokenOut and poolId that were used in the swap
func (k Keeper) BuildRoutes(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64) [][]TradeInfo {
	routes := make([][]TradeInfo, 0)

	// Check hot routes if enabled
	if tokenPairRoutes, err := k.BuildTokenPairRoutes(ctx, tokenIn, tokenOut, poolId); err == nil {
		routes = append(routes, tokenPairRoutes...)
	}

	// Append an osmo route if one exists
	if osmoRoute, err := k.BuildOsmoRoute(ctx, tokenIn, tokenOut, poolId); err == nil {
		routes = append(routes, osmoRoute)
	}

	// Append an atom route if one exists
	if atomRoute, err := k.BuildAtomRoute(ctx, tokenIn, tokenOut, poolId); err == nil {
		routes = append(routes, atomRoute)
	}

	return routes
}

// BuildTokenPairRoutes builds all of the possible arbitrage routes from the hot routes given the tokenIn, tokenOut and poolId that were used in the swap
func (k Keeper) BuildTokenPairRoutes(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64) ([][]TradeInfo, error) {
	// Get all of the routes from the store that match the given tokenIn and tokenOut
	tokenPairArbRoutes, err := k.GetTokenPairArbRoutes(ctx, tokenIn, tokenOut)
	if err != nil {
		return [][]TradeInfo{}, err
	}

	// Iterate through all of the routes and build hot routes
	routes := make([][]TradeInfo, 0)
	for _, route := range tokenPairArbRoutes.ArbRoutes {
		newRoute := make([]TradeInfo, 0)

		var newTrade TradeInfo
		for _, trade := range route.Trades {
			// 0 is a placeholder for swaps that should be entered into the hot route
			if trade.Pool == 0 {
				pool, err := k.GetAndCheckPool(ctx, poolId)
				if err != nil {
					return [][]TradeInfo{}, err
				}

				newTrade = TradeInfo{
					InputDenom:  tokenOut,
					OutputDenom: tokenIn,
					SwapFee:     pool.GetSwapFee(ctx),
					Pool:        pool,
				}
			} else {
				pool, err := k.GetAndCheckPool(ctx, trade.Pool)
				if err != nil {
					return [][]TradeInfo{}, err
				}

				newTrade = TradeInfo{
					InputDenom:  trade.TokenIn,
					OutputDenom: trade.TokenOut,
					SwapFee:     pool.GetSwapFee(ctx),
					Pool:        pool,
				}
			}

			newRoute = append(newRoute, newTrade)
		}

		// Only append if the arbitrage denom is valid
		// In denom and out denom must be the same
		// In denom must be uosmo or atom
		// There must be exactly three hops in the route
		if len(newRoute) == 3 && (newRoute[0].InputDenom == types.AtomDenomination || newRoute[0].InputDenom == types.OsmosisDenomination) && (newRoute[0].InputDenom == newRoute[2].OutputDenom) {
			routes = append(routes, newRoute)
		}
	}

	return routes, nil
}

// BuildOsmoRoute builds a cyclic arbitrage route that starts and ends with osmo given the tokenIn, tokenOut and poolId that were used in the swap
func (k Keeper) BuildOsmoRoute(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64) ([]TradeInfo, error) {
	// Creating the first trade in the arb
	entryPoolId, err := k.GetOsmoPool(ctx, tokenOut)
	if err != nil {
		return []TradeInfo{}, err
	}
	entryPool, err := k.GetAndCheckPool(ctx, entryPoolId)
	if err != nil {
		return []TradeInfo{}, err
	}
	entryTrade := TradeInfo{
		InputDenom:  types.OsmosisDenomination,
		OutputDenom: tokenOut,
		SwapFee:     entryPool.GetSwapFee(ctx),
		Pool:        entryPool,
	}

	// Creating the second trade in the arb
	middlePool, err := k.GetAndCheckPool(ctx, poolId)
	if err != nil {
		return []TradeInfo{}, err
	}
	middleTrade := TradeInfo{
		InputDenom:  tokenOut,
		OutputDenom: tokenIn,
		SwapFee:     middlePool.GetSwapFee(ctx),
		Pool:        middlePool,
	}

	// Creating the third trade in the arb
	exitPoolId, err := k.GetOsmoPool(ctx, tokenIn)
	if err != nil {
		return []TradeInfo{}, err
	}
	exitPool, err := k.GetAndCheckPool(ctx, exitPoolId)
	if err != nil {
		return []TradeInfo{}, err
	}
	exitTrade := TradeInfo{
		InputDenom:  tokenIn,
		OutputDenom: types.OsmosisDenomination,
		SwapFee:     exitPool.GetSwapFee(ctx),
		Pool:        exitPool,
	}

	return []TradeInfo{entryTrade, middleTrade, exitTrade}, nil
}

// BuildAtomRoute builds a cyclic arbitrage route that starts and ends with atom given the tokenIn, tokenOut and poolId that were used in the swap
func (k Keeper) BuildAtomRoute(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64) ([]TradeInfo, error) {
	// Creating the first trade in the arb
	entryPoolId, err := k.GetAtomPool(ctx, tokenOut)
	if err != nil {
		return []TradeInfo{}, err
	}
	entryPool, err := k.GetAndCheckPool(ctx, entryPoolId)
	if err != nil {
		return []TradeInfo{}, err
	}
	entryTrade := TradeInfo{
		InputDenom:  types.AtomDenomination,
		OutputDenom: tokenOut,
		SwapFee:     entryPool.GetSwapFee(ctx),
		Pool:        entryPool,
	}

	// Creating the second trade in the arb
	middlePool, err := k.GetAndCheckPool(ctx, poolId)
	if err != nil {
		return []TradeInfo{}, err
	}
	middleTrade := TradeInfo{
		InputDenom:  tokenOut,
		OutputDenom: tokenIn,
		SwapFee:     middlePool.GetSwapFee(ctx),
		Pool:        middlePool,
	}

	// Creating the third trade in the arb
	exitPoolId, err := k.GetAtomPool(ctx, tokenIn)
	if err != nil {
		return []TradeInfo{}, err
	}
	exitPool, err := k.GetAndCheckPool(ctx, exitPoolId)
	if err != nil {
		return []TradeInfo{}, err
	}
	exitTrade := TradeInfo{
		InputDenom:  tokenIn,
		OutputDenom: types.AtomDenomination,
		SwapFee:     exitPool.GetSwapFee(ctx),
		Pool:        exitPool,
	}

	return []TradeInfo{entryTrade, middleTrade, exitTrade}, nil
}

// GetAndCheckPool retrieves the pool from the x/gamm module given a poolId and ensures that the pool can be traded on
func (k Keeper) GetAndCheckPool(ctx sdk.Context, poolId uint64) (gammtypes.CFMMPoolI, error) {
	pool, err := k.gammKeeper.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		return pool, err
	}
	if !pool.IsActive(ctx) {
		return pool, fmt.Errorf("pool %d is not active", poolId)
	}
	return pool, nil
}
