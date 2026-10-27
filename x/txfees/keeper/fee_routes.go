package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	pooltypes "github.com/osmosis-labs/osmosis/v32/x/poolmanager/types"
)

// Keep pair order: ProtoRev prefers the first orientation if both are stored.
type feeDenomPair struct{ first, second string }

type feeResolvedRoute struct {
	route []pooltypes.SwapAmountInRoute
	kind  string
	err   error
}

// feeRouteResolver is owned by one epoch-hook invocation. It caches route
// selection, including misses, but never swap outputs or mutable pool state.
type feeRouteResolver struct {
	k                  Keeper
	routes             map[feeDenomPair]feeResolvedRoute
	loadIntermediaries func(sdk.Context) []string
	intermediaries     []string
	paramsLoaded       bool
	revision           uint64
	initialized        bool
}

func newFeeRouteResolver(k Keeper) *feeRouteResolver {
	return &feeRouteResolver{
		k: k, routes: make(map[feeDenomPair]feeResolvedRoute),
		loadIntermediaries: func(ctx sdk.Context) []string { return k.GetParams(ctx).FeeSwapIntermediaryDenomList },
	}
}

func (r *feeRouteResolver) resolve(ctx sdk.Context, output, input string) ([]pooltypes.SwapAmountInRoute, string, error) {
	// Pool creation or route administration during a swap can change the table.
	// Invalidate successful and failed lookups after such a change.
	revision := r.k.protorevKeeper.GetPoolRoutesRevision(ctx)
	if !r.initialized || revision != r.revision {
		clear(r.routes)
		r.paramsLoaded = false
		r.revision, r.initialized = revision, true
	}
	key := feeDenomPair{output, input}
	if entry, ok := r.routes[key]; ok {
		return entry.route, entry.kind, entry.err
	}
	route, err := r.k.getDirectRoute(ctx, output, input)
	kind := "single_hop"
	if err != nil {
		if !r.paramsLoaded {
			r.intermediaries = r.loadIntermediaries(ctx)
			r.paramsLoaded = true
		}
		route, kind, err = r.k.get2HopRoute(ctx, output, input, r.intermediaries)
	}
	r.routes[key] = feeResolvedRoute{route, kind, err}
	return route, kind, err
}
