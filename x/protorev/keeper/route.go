package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/protorev/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// BuildRoutes builds all of the possible routes given the given tokenIn and tokenOut
func (k Keeper) BuildRoutes(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64) [][]uint64 {
	routes := make([][]uint64, 0)

	// Append all of the search routes if they exist
	if routes, err := k.BuildSearcherRoutes(ctx, tokenIn, tokenOut, poolId); err == nil {
		routes = append(routes, routes...)
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

// BuildSearcherRoutes builds all of the possible routes given the given tokenIn and tokenOut from the
// SearcherRoutes store
func (k Keeper) BuildSearcherRoutes(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64) ([][]uint64, error) {
	routes := make([][]uint64, 0)

	// Get all of the routes from the store that match the given tokenIn and tokenOut
	searcherRoutes, err := k.GetSearcherRoutes(ctx, tokenIn, tokenOut)

	if err != nil {
		return [][]uint64{}, err
	}

	// Iterate through all of the routes and find the ones that match the given tokenIn and tokenOut
	for _, route := range searcherRoutes.Routes {
		newRoute := make([]uint64, 0)

		// Every searcher route must place a 0 place holder for where the current swap should be placed in the route
		for _, hopId := range route.Pools {
			if hopId == 0 {
				newRoute = append(newRoute, poolId)
			} else {
				newRoute = append(newRoute, hopId)
			}
		}

		routes = append(routes, newRoute)
	}

	return routes, nil
}

// BuildOsmoRoute builds a route from the given tokenIn to the given tokenOut, using the given poolId as the middle hop
func (k Keeper) BuildOsmoRoute(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64) ([]uint64, error) {
	// Ignore the swap when there is a Omso/Atom swap
	if types.CheckPerfectMatch(tokenIn, tokenOut) {
		return []uint64{}, sdkerrors.Wrapf(ErrNoOsmoRoute, "denom: %s", tokenIn)
	}

	// Ignore when the swapped value contains Atom
	if _, matched := types.CheckMatch(tokenIn, tokenOut, types.OsmosisDenomination); matched {
		return []uint64{}, sdkerrors.Wrapf(ErrNoOsmoRoute, "denom: %s", tokenIn)
	}

	// Cyclic arbitrage will always be profitable in the opposite direction of what is given
	entryPool, err := k.GetOsmoPool(ctx, tokenOut)
	if err != nil {
		return []uint64{}, sdkerrors.Wrapf(ErrNoOsmoRoute, "denom: %s", tokenIn)
	}

	exitPool, err := k.GetOsmoPool(ctx, tokenIn)
	if err != nil {
		return []uint64{}, sdkerrors.Wrapf(ErrNoOsmoRoute, "denom: %s", tokenOut)
	}

	return []uint64{entryPool, poolId, exitPool}, nil
}

// BuildAtomRoute builds a route from the given tokenIn to the given tokenOut, using the given poolId as the middle hop
func (k Keeper) BuildAtomRoute(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64) ([]uint64, error) {
	// Ignore the swap when there is a Omso/Atom swap
	if types.CheckPerfectMatch(tokenIn, tokenOut) {
		return []uint64{}, sdkerrors.Wrapf(ErrNoAtomRoute, "denom: %s", tokenIn)
	}

	// Ignore when the swapped value contains Atom
	if _, matched := types.CheckMatch(tokenIn, tokenOut, types.AtomDenomination); matched {
		return []uint64{}, sdkerrors.Wrapf(ErrNoAtomRoute, "denom: %s", tokenIn)
	}

	// Cyclic arbitrage will always be profitable in the opposite direction of what is given
	entryPool, err := k.GetAtomPool(ctx, tokenOut)
	if err != nil {
		return []uint64{}, sdkerrors.Wrapf(ErrNoAtomRoute, "denom: %s", tokenIn)
	}

	exitPool, err := k.GetAtomPool(ctx, tokenIn)
	if err != nil {
		return []uint64{}, sdkerrors.Wrapf(ErrNoAtomRoute, "denom: %s", tokenOut)
	}

	return []uint64{entryPool, poolId, exitPool}, nil
}
