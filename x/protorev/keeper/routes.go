package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v13/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v13/x/protorev/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// BuildRoutes builds all of the possible arbitrage routes given the tokenIn, tokenOut and poolId that were used in the swap
func (k Keeper) BuildRoutes(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64) []swaproutertypes.SwapAmountInRoutes {
	routes := make([]swaproutertypes.SwapAmountInRoutes, 0)

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
func (k Keeper) BuildTokenPairRoutes(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64) ([]swaproutertypes.SwapAmountInRoutes, error) {
	// Get all of the routes from the store that match the given tokenIn and tokenOut
	tokenPairArbRoutes, err := k.GetTokenPairArbRoutes(ctx, tokenIn, tokenOut)
	if err != nil {
		return []swaproutertypes.SwapAmountInRoutes{}, err
	}

	// Iterate through all of the routes and build hot routes
	routes := make([]swaproutertypes.SwapAmountInRoutes, 0)
	for _, route := range tokenPairArbRoutes.ArbRoutes {
		newRoute, err := k.BuildHotRoute(ctx, route, tokenIn, tokenOut, poolId)
		if err == nil {
			routes = append(routes, newRoute)
		}
	}

	return routes, nil
}

// BuildHotRoute constructs a cyclic arbitrage route given a hot route from the store and information about the swap that should be placed
// in the hot route.
func (k Keeper) BuildHotRoute(ctx sdk.Context, route *types.Route, tokenIn, tokenOut string, poolId uint64) (swaproutertypes.SwapAmountInRoutes, error) {
	newRoute := make(swaproutertypes.SwapAmountInRoutes, 0)

	for _, trade := range route.Trades {
		// 0 is a placeholder for pools swapped on that should be entered into the hot route
		if trade.Pool == 0 {
			newRoute = append(newRoute, swaproutertypes.SwapAmountInRoute{
				PoolId:        poolId,
				TokenOutDenom: trade.TokenOut,
			})
		} else {
			newRoute = append(newRoute, swaproutertypes.SwapAmountInRoute{
				PoolId:        trade.Pool,
				TokenOutDenom: trade.TokenOut,
			})
		}
	}

	// Check that the hot route is valid
	if err := k.CheckValidHotRoute(ctx, newRoute); err != nil {
		return swaproutertypes.SwapAmountInRoutes{}, err
	}
	return newRoute, nil
}

// CheckValidHotRoute checks if the cyclic arbitrage route that was built using the hot routes method is correct. Much of the stateless
// validation achieves the desired checks, however, we also check that the route is traversing pools that
// are active.
func (k Keeper) CheckValidHotRoute(ctx sdk.Context, route swaproutertypes.SwapAmountInRoutes) error {
	if route.Length() != 3 {
		return fmt.Errorf("invalid hot route length")
	}

	// Ensure that all of the pools in the route exist and are active
	for _, poolId := range route.PoolIds() {
		_, err := k.GetAndCheckPool(ctx, poolId)
		if err != nil {
			return err
		}
	}

	return nil
}

// BuildOsmoRoute builds a cyclic arbitrage route that starts and ends with osmo given the tokenIn, tokenOut and poolId that were used in the swap
func (k Keeper) BuildOsmoRoute(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64) (swaproutertypes.SwapAmountInRoutes, error) {
	return k.BuildRoute(ctx, types.OsmosisDenomination, tokenIn, tokenOut, poolId, k.GetOsmoPool)
}

// BuildAtomRoute builds a cyclic arbitrage route that starts and ends with atom given the tokenIn, tokenOut and poolId that were used in the swap
func (k Keeper) BuildAtomRoute(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64) (swaproutertypes.SwapAmountInRoutes, error) {
	return k.BuildRoute(ctx, types.AtomDenomination, tokenIn, tokenOut, poolId, k.GetAtomPool)
}

// BuildRoute constructs a cyclic arbitrage route that is starts/ends with swapDenom (atom or osmo) given the swap (tokenIn, tokenOut, poolId), and
// a function that can get the poolId from the store given a (token, swapDenom) pair.
func (k Keeper) BuildRoute(ctx sdk.Context, swapDenom, tokenIn, tokenOut string, poolId uint64, getPoolIDFromStore func(sdk.Context, string) (uint64, error)) (swaproutertypes.SwapAmountInRoutes, error) {
	// Creating the first trade in the arb
	entryPoolId, err := getPoolIDFromStore(ctx, tokenOut)
	if err != nil {
		return swaproutertypes.SwapAmountInRoutes{}, err
	}

	// Check that the pool exists and is active
	_, err = k.GetAndCheckPool(ctx, entryPoolId)
	if err != nil {
		return swaproutertypes.SwapAmountInRoutes{}, err
	}
	// Create the first swap for the MultiHopSwap Route
	entryRoute := swaproutertypes.SwapAmountInRoute{
		PoolId:        entryPoolId,
		TokenOutDenom: tokenOut,
	}

	// Creating the second trade in the arb
	_, err = k.GetAndCheckPool(ctx, poolId)
	if err != nil {
		return swaproutertypes.SwapAmountInRoutes{}, err
	}
	middleRoute := swaproutertypes.SwapAmountInRoute{
		PoolId:        poolId,
		TokenOutDenom: tokenIn,
	}

	// Creating the third trade in the arb
	exitPoolId, err := getPoolIDFromStore(ctx, tokenIn)
	if err != nil {
		return swaproutertypes.SwapAmountInRoutes{}, err
	}
	_, err = k.GetAndCheckPool(ctx, exitPoolId)
	if err != nil {
		return swaproutertypes.SwapAmountInRoutes{}, err
	}
	exitRoute := swaproutertypes.SwapAmountInRoute{
		PoolId:        exitPoolId,
		TokenOutDenom: swapDenom,
	}

	return swaproutertypes.SwapAmountInRoutes{entryRoute, middleRoute, exitRoute}, nil
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
