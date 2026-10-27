package keeper

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v32/tests/epochdata"
	pooltypes "github.com/osmosis-labs/osmosis/v32/x/poolmanager/types"
	protorev "github.com/osmosis-labs/osmosis/v32/x/protorev/keeper"
	protorevtypes "github.com/osmosis-labs/osmosis/v32/x/protorev/types"
	"github.com/osmosis-labs/osmosis/v32/x/txfees/types"
)

type countedFeeLookup struct {
	types.ProtorevKeeper
	calls int
}

func (c *countedFeeLookup) GetPoolForDenomPairNoOrder(ctx sdk.Context, a, b string) (uint64, error) {
	c.calls++
	return c.ProtorevKeeper.GetPoolForDenomPairNoOrder(ctx, a, b)
}

func feeRouteContext(tb testing.TB) (Keeper, protorev.Keeper, sdk.Context, *countedFeeLookup) {
	tb.Helper()
	key := storetypes.NewKVStoreKey(protorevtypes.StoreKey)
	tkey := storetypes.NewTransientStoreKey(protorevtypes.TransientStoreKey)
	tc := testutil.DefaultContextWithDB(tb, key, tkey)
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	space := paramtypes.NewSubspace(cdc, codec.NewLegacyAmino(), key, tkey, protorevtypes.ModuleName)
	pr := protorev.NewKeeper(cdc, key, tkey, space, nil, nil, nil, nil, nil, nil, nil)
	lookup := &countedFeeLookup{ProtorevKeeper: pr}
	feeSpace := paramtypes.NewSubspace(cdc, codec.NewLegacyAmino(), key, tkey, types.ModuleName).WithKeyTable(types.ParamKeyTable())
	k := Keeper{protorevKeeper: lookup, paramSpace: feeSpace}
	k.SetParams(tc.Ctx, types.DefaultParams())
	return k, pr, tc.Ctx, lookup
}

func TestFeeRouteResolverInvalidation(t *testing.T) {
	k, pr, ctx, lookups := feeRouteContext(t)
	pr.SetPoolForDenomPair(ctx, "umid", "ufoo", 1)
	pr.SetPoolForDenomPair(ctx, "umid", "uosmo", 2)
	resolver := newFeeRouteResolver(k)
	loads := 0
	resolver.loadIntermediaries = func(sdk.Context) []string { loads++; return []string{"umid"} }
	route, kind, err := resolver.resolve(ctx, "uosmo", "ufoo")
	require.NoError(t, err)
	require.Equal(t, "two_hop_via_umid", kind)
	require.Equal(t, []pooltypes.SwapAmountInRoute{{PoolId: 1, TokenOutDenom: "umid"}, {PoolId: 2, TokenOutDenom: "uosmo"}}, route)
	calls := lookups.calls
	_, _, err = resolver.resolve(ctx, "uosmo", "ufoo")
	require.NoError(t, err)
	require.Equal(t, calls, lookups.calls)
	require.Equal(t, 1, loads)

	_, _, err = resolver.resolve(ctx, "uosmo", "umissing")
	require.Error(t, err)
	calls = lookups.calls
	_, _, err = resolver.resolve(ctx, "uosmo", "umissing")
	require.Error(t, err)
	require.Equal(t, calls, lookups.calls, "missing routes should also be cached")

	// A newly available direct route must supersede the cached two-hop route.
	pr.SetPoolForDenomPair(ctx, "uosmo", "ufoo", 3)
	route, kind, err = resolver.resolve(ctx, "uosmo", "ufoo")
	require.NoError(t, err)
	require.Equal(t, "single_hop", kind)
	require.Equal(t, uint64(3), route[0].PoolId)
	// Newly created routes must also invalidate cached misses.
	pr.SetPoolForDenomPair(ctx, "uosmo", "umissing", 4)
	route, _, err = resolver.resolve(ctx, "uosmo", "umissing")
	require.NoError(t, err)
	require.Equal(t, uint64(4), route[0].PoolId)
	pr.DeleteAllPoolsForBaseDenom(ctx, "uosmo")
	route, kind, err = resolver.resolve(ctx, "uosmo", "ufoo")
	require.NoError(t, err)
	require.Equal(t, "two_hop_via_umid", kind)
	require.Len(t, route, 2)
	require.Equal(t, 2, loads, "intermediaries should be reloaded after invalidation")
}

func TestFeeRouteResolverPreservesPairOrientation(t *testing.T) {
	k, pr, ctx, _ := feeRouteContext(t)
	pr.SetPoolForDenomPair(ctx, "uosmo", "uatom", 1)
	pr.SetPoolForDenomPair(ctx, "uatom", "uosmo", 2)
	resolver := newFeeRouteResolver(k)
	for _, test := range []struct {
		output, input string
		pool          uint64
	}{{"uosmo", "uatom", 1}, {"uatom", "uosmo", 2}} {
		route, _, err := resolver.resolve(ctx, test.output, test.input)
		require.NoError(t, err)
		require.Equal(t, test.pool, route[0].PoolId)
	}
}

func TestFeeRouteResolverTracksCommittedRoutingChanges(t *testing.T) {
	k, pr, ctx, _ := feeRouteContext(t)
	pr.SetPoolForDenomPair(ctx, "uosmo", "uatom", 1)
	resolver := newFeeRouteResolver(k)
	route, _, err := resolver.resolve(ctx, "uosmo", "uatom")
	require.NoError(t, err)
	require.Equal(t, uint64(1), route[0].PoolId)

	// A failed swap's cached writes, including its revision, are discarded.
	failed, _ := ctx.CacheContext()
	pr.SetPoolForDenomPair(failed, "uosmo", "uatom", 2)
	route, _, err = resolver.resolve(ctx, "uosmo", "uatom")
	require.NoError(t, err)
	require.Equal(t, uint64(1), route[0].PoolId)

	// A successful swap merges both the route and its invalidation counter.
	committed, write := ctx.CacheContext()
	pr.SetPoolForDenomPair(committed, "uosmo", "uatom", 3)
	write()
	route, _, err = resolver.resolve(ctx, "uosmo", "uatom")
	require.NoError(t, err)
	require.Equal(t, uint64(3), route[0].PoolId)
}

func uncachedFeeRoute(k Keeper, ctx sdk.Context, request epochdata.Request) ([]pooltypes.SwapAmountInRoute, string, error) {
	route, err := k.getDirectRoute(ctx, request.Output, request.Input)
	if err == nil {
		return route, "single_hop", nil
	}
	return k.get2HopRoute(ctx, request.Output, request.Input, k.GetParams(ctx).FeeSwapIntermediaryDenomList)
}

func TestRecordedFeeRoutesMatchUncached(t *testing.T) {
	workload := epochdata.Load(t)
	require.Equal(t, 199, len(workload.Routes))
	require.Equal(t, 434, len(workload.Requests))
	k, pr, ctx, _ := feeRouteContext(t)
	params := types.DefaultParams()
	params.FeeSwapIntermediaryDenomList = workload.Intermediaries
	k.SetParams(ctx, params)
	for _, route := range workload.Routes {
		pr.SetPoolForDenomPair(ctx, route.Base, route.Quote, route.PoolID)
	}
	resolver := newFeeRouteResolver(k)
	for _, request := range workload.Requests {
		expected, kind, err := uncachedFeeRoute(k, ctx, request)
		require.NoError(t, err)
		actual, cachedKind, err := resolver.resolve(ctx, request.Output, request.Input)
		require.NoError(t, err)
		require.Equal(t, expected, actual)
		require.Equal(t, kind, cachedKind)
	}
}

// Only route resolution is measured. Pool math, swaps, and fee distributions are
// covered by the keeper suite but excluded from this onchain-derived benchmark.
func BenchmarkRecordedEpochFeeRoutes(b *testing.B) {
	workload := epochdata.Load(b)
	for _, cached := range []bool{false, true} {
		name := "uncached"
		if cached {
			name = "cached"
		}
		b.Run(name, func(b *testing.B) {
			k, pr, ctx, lookup := feeRouteContext(b)
			params := types.DefaultParams()
			params.FeeSwapIntermediaryDenomList = workload.Intermediaries
			k.SetParams(ctx, params)
			for _, route := range workload.Routes {
				pr.SetPoolForDenomPair(ctx, route.Base, route.Quote, route.PoolID)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				epochCtx, _ := ctx.CacheContext()
				var resolver *feeRouteResolver
				if cached {
					resolver = newFeeRouteResolver(k)
				}
				for _, request := range workload.Requests {
					var err error
					if cached {
						_, _, err = resolver.resolve(epochCtx, request.Output, request.Input)
					} else {
						_, _, err = uncachedFeeRoute(k, epochCtx, request)
					}
					if err != nil {
						b.Fatal(err)
					}
				}
			}
			b.StopTimer()
			b.ReportMetric(float64(lookup.calls)/float64(b.N), "pool-lookups/epoch")
		})
	}
}
