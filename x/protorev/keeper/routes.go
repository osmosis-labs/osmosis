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

type Route struct {
	Trades []TradeInfo
}

// BuildRoutes builds all of the possible arbitrage routes given the tokenIn, tokenOut and poolId that were used in the swap
func (k Keeper) BuildRoutes(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64) []Route {
	routes := make([]Route, 0)

	// Append hot routes if they exist
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
func (k Keeper) BuildTokenPairRoutes(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64) ([]Route, error) {
	// Get all of the routes from the store that match the given tokenIn and tokenOut
	tokenPairArbRoutes, err := k.GetTokenPairArbRoutes(ctx, tokenIn, tokenOut)
	if err != nil {
		return []Route{}, err
	}

	// Iterate through all of the routes and build hot routes
	routes := make([]Route, 0)
	for _, route := range tokenPairArbRoutes.ArbRoutes {
		newRoute, err := k.BuildTradeInfoHotRoute(ctx, route, tokenIn, tokenOut, poolId)
		if err == nil {
			routes = append(routes, newRoute)
		}
	}

	return routes, nil
}

// BuildTradeInfoHotRoute constructs a cyclic arbitrage route given a hot route from the store and information about the swap that should be placed
// in the hot route.
func (k Keeper) BuildTradeInfoHotRoute(ctx sdk.Context, route *types.Route, tokenIn, tokenOut string, poolId uint64) (Route, error) {
	newRoute := Route{Trades: make([]TradeInfo, len(route.Trades))}

	for index, trade := range route.Trades {
		var newTrade TradeInfo
		// 0 is a placeholder for swaps that should be entered into the hot route
		if trade.Pool == 0 {
			pool, err := k.GetAndCheckPool(ctx, poolId)
			if err != nil {
				return Route{}, err
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
				return Route{}, err
			}

			newTrade = TradeInfo{
				InputDenom:  trade.TokenIn,
				OutputDenom: trade.TokenOut,
				SwapFee:     pool.GetSwapFee(ctx),
				Pool:        pool,
			}
		}

		newRoute.Trades[index] = newTrade
	}

	if err := k.CheckValidHotRoute(newRoute); err != nil {
		return Route{}, err
	}
	return newRoute, nil
}

// CheckValidHotRoute checks if the cyclic arbitrage route that was built using the hot routes method is correct. The criteria for a valid hot route is that
// the in denom and out denom must be the same, in denom must be uosmo or atom, and there must be exactly three hops in the route
func (k Keeper) CheckValidHotRoute(route Route) error {
	if len(route.Trades) != 3 {
		return fmt.Errorf("invalid hot route length")
	}

	if route.Trades[0].InputDenom != route.Trades[2].OutputDenom {
		return fmt.Errorf("invalid hot route in and out denoms. in: %s, out: %s", route.Trades[0].InputDenom, route.Trades[2].OutputDenom)
	}

	if route.Trades[0].InputDenom != types.OsmosisDenomination && route.Trades[0].InputDenom != types.AtomDenomination {
		return fmt.Errorf("invalid hot route in denom")
	}

	return nil
}

// BuildOsmoRoute builds a cyclic arbitrage route that starts and ends with osmo given the tokenIn, tokenOut and poolId that were used in the swap
func (k Keeper) BuildOsmoRoute(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64) (Route, error) {
	return k.BuildTradeInfoRoute(ctx, types.OsmosisDenomination, tokenIn, tokenOut, poolId, k.GetOsmoPool)
}

// BuildAtomRoute builds a cyclic arbitrage route that starts and ends with atom given the tokenIn, tokenOut and poolId that were used in the swap
func (k Keeper) BuildAtomRoute(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64) (Route, error) {
	return k.BuildTradeInfoRoute(ctx, types.AtomDenomination, tokenIn, tokenOut, poolId, k.GetAtomPool)
}

// BuildTradeInfoRoute constructs a cyclic arbitrage route that is starts/ends with swapDenom (atom or osmo) given the swap (tokenIn, tokenOut, poolId), and
// a function that can get the poolId from the store given a (token, swapDenom) pair.
func (k Keeper) BuildTradeInfoRoute(ctx sdk.Context, swapDenom, tokenIn, tokenOut string, poolId uint64, getPoolIDFromStore func(sdk.Context, string) (uint64, error)) (Route, error) {
	// Creating the first trade in the arb
	entryPoolId, err := getPoolIDFromStore(ctx, tokenOut)
	if err != nil {
		return Route{}, err
	}
	entryPool, err := k.GetAndCheckPool(ctx, entryPoolId)
	if err != nil {
		return Route{}, err
	}
	entryTrade := TradeInfo{
		InputDenom:  swapDenom,
		OutputDenom: tokenOut,
		SwapFee:     entryPool.GetSwapFee(ctx),
		Pool:        entryPool,
	}

	// Creating the second trade in the arb
	middlePool, err := k.GetAndCheckPool(ctx, poolId)
	if err != nil {
		return Route{}, err
	}
	middleTrade := TradeInfo{
		InputDenom:  tokenOut,
		OutputDenom: tokenIn,
		SwapFee:     middlePool.GetSwapFee(ctx),
		Pool:        middlePool,
	}

	// Creating the third trade in the arb
	exitPoolId, err := getPoolIDFromStore(ctx, tokenIn)
	if err != nil {
		return Route{}, err
	}
	exitPool, err := k.GetAndCheckPool(ctx, exitPoolId)
	if err != nil {
		return Route{}, err
	}
	exitTrade := TradeInfo{
		InputDenom:  tokenIn,
		OutputDenom: swapDenom,
		SwapFee:     exitPool.GetSwapFee(ctx),
		Pool:        exitPool,
	}

	return Route{Trades: []TradeInfo{entryTrade, middleTrade, exitTrade}}, nil
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
