package keeper

import (
	"errors"
	"testing"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v32/tests/epochdata"
	pooltypes "github.com/osmosis-labs/osmosis/v32/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v32/x/protorev/types"
)

type routeWriteStore struct {
	storetypes.KVStore
	sets, deletes int
}

func (s *routeWriteStore) Set(key, value []byte) { s.sets++; s.KVStore.Set(key, value) }
func (s *routeWriteStore) Delete(key []byte)     { s.deletes++; s.KVStore.Delete(key) }

type routeWriteMultiStore struct {
	storetypes.MultiStore
	key   storetypes.StoreKey
	store *routeWriteStore
}

func (m routeWriteMultiStore) GetKVStore(key storetypes.StoreKey) storetypes.KVStore {
	if key == m.key {
		return m.store
	}
	return m.MultiStore.GetKVStore(key)
}

func routeSyncContext(tb testing.TB) (Keeper, sdk.Context, *routeWriteStore) {
	tb.Helper()
	key := storetypes.NewKVStoreKey(types.StoreKey)
	tc := testutil.DefaultContextWithDB(tb, key, storetypes.NewTransientStoreKey(types.TransientStoreKey))
	counted := &routeWriteStore{KVStore: tc.CMS.GetKVStore(key)}
	ctx := tc.Ctx.WithMultiStore(routeWriteMultiStore{MultiStore: tc.CMS, key: key, store: counted})
	return Keeper{storeKey: key}, ctx, counted
}

func TestSyncHighestLiquidityPools(t *testing.T) {
	k, ctx, writes := routeSyncContext(t)
	k.SetPoolForDenomPair(ctx, "uosmo", "unchanged", 1)
	k.SetPoolForDenomPair(ctx, "uosmo", "changed", 2)
	k.SetPoolForDenomPair(ctx, "uosmo", "removed", 3)
	k.SetPoolForDenomPair(ctx, "uosmo/other", "removed", 4)
	desired := map[string]LiquidityPoolStruct{
		"unchanged": {PoolId: 1}, "changed": {PoolId: 20}, "added": {PoolId: 30},
	}
	writes.sets, writes.deletes = 0, 0
	k.syncHighestLiquidityPools(ctx, "uosmo", desired)
	require.Equal(t, 2, writes.sets)
	require.Equal(t, 1, writes.deletes)
	for denom, pool := range desired {
		actual, err := k.GetPoolForDenomPair(ctx, "uosmo", denom)
		require.NoError(t, err)
		require.Equal(t, pool.PoolId, actual)
	}
	_, err := k.GetPoolForDenomPair(ctx, "uosmo", "removed")
	require.Error(t, err)
	actual, err := k.GetPoolForDenomPair(ctx, "uosmo/other", "removed")
	require.NoError(t, err)
	require.Equal(t, uint64(4), actual)
	writes.sets, writes.deletes = 0, 0
	k.syncHighestLiquidityPools(ctx, "uosmo", desired)
	require.Zero(t, writes.sets, "unchanged routes must not be rewritten")
	require.Zero(t, writes.deletes)
	k.syncHighestLiquidityPools(ctx, "uosmo", nil)
	require.Equal(t, len(desired), writes.deletes)
}

type routeTestPool struct {
	pooltypes.PoolI
	id uint64
}

func (p routeTestPool) GetId() uint64             { return p.id }
func (p routeTestPool) IsActive(sdk.Context) bool { return true }

type routeTestManager struct {
	types.PoolManagerKeeper
	liquidity map[uint64]sdk.Coins
	fail      bool
}

func (m routeTestManager) AllPools(sdk.Context) ([]pooltypes.PoolI, error) {
	return []pooltypes.PoolI{routeTestPool{id: 1}, routeTestPool{id: 2}}, nil
}

func (m routeTestManager) GetTotalPoolLiquidity(_ sdk.Context, id uint64) (sdk.Coins, error) {
	if m.fail {
		return nil, errors.New("liquidity unavailable")
	}
	return m.liquidity[id], nil
}

func TestUpdatePoolsPreservesTieBreakingAndRoutesOnError(t *testing.T) {
	k, ctx, writes := routeSyncContext(t)
	require.NoError(t, k.SetBaseDenoms(ctx, []types.BaseDenom{{Denom: "uosmo", StepSize: osmomath.OneInt()}}))
	coins := sdk.NewCoins(sdk.NewInt64Coin("uosmo", 100), sdk.NewInt64Coin("uatom", 100))
	k.poolmanagerKeeper = routeTestManager{liquidity: map[uint64]sdk.Coins{1: coins, 2: coins}}
	k.SetPoolForDenomPair(ctx, "uosmo", "uatom", 2)
	require.NoError(t, k.UpdatePools(ctx))
	id, err := k.GetPoolForDenomPair(ctx, "uosmo", "uatom")
	require.NoError(t, err)
	require.Equal(t, uint64(1), id, "equal liquidity keeps the first pool in ID order")
	writes.sets, writes.deletes = 0, 0
	k.poolmanagerKeeper = routeTestManager{fail: true}
	require.Error(t, k.UpdatePools(ctx))
	id, err = k.GetPoolForDenomPair(ctx, "uosmo", "uatom")
	require.NoError(t, err)
	require.Equal(t, uint64(1), id)
	require.Zero(t, writes.sets)
	require.Zero(t, writes.deletes)
}

// rewritePoolRoutesLegacy reproduces the persistent writes from before the
// optimization, without any later route-cache revision bookkeeping.
func rewritePoolRoutesLegacy(k Keeper, ctx sdk.Context, base string, routes map[string]LiquidityPoolStruct) {
	key := append(types.KeyPrefixDenomPairToPool, types.GetKeyPrefixDenomPairToPool(base, "")...)
	k.DeleteAllEntriesForKeyPrefix(ctx, key)
	for denom, pool := range routes {
		store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDenomPairToPool)
		store.Set(types.GetKeyPrefixDenomPairToPool(base, denom), sdk.Uint64ToBigEndian(pool.PoolId))
	}
}

// This isolates cache merge costs for unchanged observed routes. It excludes
// IAVL commits, pool scans, contract queries, and swap execution.
func BenchmarkEpochRoutePersistence(b *testing.B) {
	workload := epochdata.Load(b)
	for _, delta := range []bool{false, true} {
		name := "rewrite_all"
		if delta {
			name = "deltas"
		}
		b.Run(name, func(b *testing.B) {
			k, ctx, _ := routeSyncContext(b)
			selected := make(map[string]map[string]LiquidityPoolStruct)
			for _, route := range workload.Routes {
				if selected[route.Base] == nil {
					selected[route.Base] = make(map[string]LiquidityPoolStruct)
				}
				selected[route.Base][route.Quote] = LiquidityPoolStruct{PoolId: route.PoolID}
				k.SetPoolForDenomPair(ctx, route.Base, route.Quote, route.PoolID)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cacheCtx, write := ctx.CacheContext()
				for base, routes := range selected {
					if delta {
						k.syncHighestLiquidityPools(cacheCtx, base, routes)
					} else {
						rewritePoolRoutesLegacy(k, cacheCtx, base, routes)
					}
				}
				write()
			}
		})
	}
}

func committedRouteContext(tb testing.TB) (Keeper, testutil.TestContext) {
	tb.Helper()
	key := storetypes.NewKVStoreKey(types.StoreKey)
	tkey := storetypes.NewTransientStoreKey(types.TransientStoreKey)
	return Keeper{storeKey: key, transientKey: tkey}, testutil.DefaultContextWithDB(tb, key, tkey)
}

func seedRecordedRoutes(tb testing.TB, k Keeper, ctx sdk.Context) map[string]map[string]LiquidityPoolStruct {
	tb.Helper()
	selected := make(map[string]map[string]LiquidityPoolStruct)
	for _, route := range epochdata.Load(tb).Routes {
		if selected[route.Base] == nil {
			selected[route.Base] = make(map[string]LiquidityPoolStruct)
		}
		selected[route.Base][route.Quote] = LiquidityPoolStruct{PoolId: route.PoolID}
		k.SetPoolForDenomPair(ctx, route.Base, route.Quote, route.PoolID)
	}
	return selected
}

func TestRouteRefreshRequiresCoordinatedUpgrade(t *testing.T) {
	legacy, old := committedRouteContext(t)
	optimized, next := committedRouteContext(t)
	selected := seedRecordedRoutes(t, legacy, old.Ctx)
	seedRecordedRoutes(t, optimized, next.Ctx)
	require.Equal(t, old.CMS.Commit().Hash, next.CMS.Commit().Hash)

	oldCtx, oldWrite := old.Ctx.CacheContext()
	newCtx, newWrite := next.Ctx.CacheContext()
	for base, routes := range selected {
		rewritePoolRoutesLegacy(legacy, oldCtx, base, routes)
		optimized.syncHighestLiquidityPools(newCtx, base, routes)
	}
	oldWrite()
	newWrite()
	for base, routes := range selected {
		for denom, pool := range routes {
			oldID, err := legacy.GetPoolForDenomPair(old.Ctx, base, denom)
			require.NoError(t, err)
			newID, err := optimized.GetPoolForDenomPair(next.Ctx, base, denom)
			require.NoError(t, err)
			require.Equal(t, pool.PoolId, oldID)
			require.Equal(t, oldID, newID)
		}
	}
	require.NotEqual(t, old.CMS.Commit().Hash, next.CMS.Commit().Hash,
		"equal route values can have different IAVL node versions; a coordinated upgrade is required")
}

// BenchmarkEpochRouteCommit adds actual IAVL commits to the unchanged-route
// workload. It still excludes the rest of the daily epoch.
func BenchmarkEpochRouteCommit(b *testing.B) {
	for _, delta := range []bool{false, true} {
		name := "rewrite_all"
		if delta {
			name = "deltas"
		}
		b.Run(name, func(b *testing.B) {
			k, tc := committedRouteContext(b)
			selected := seedRecordedRoutes(b, k, tc.Ctx)
			tc.CMS.Commit()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ctx, write := tc.Ctx.CacheContext()
				for base, routes := range selected {
					if delta {
						k.syncHighestLiquidityPools(ctx, base, routes)
					} else {
						rewritePoolRoutesLegacy(k, ctx, base, routes)
					}
				}
				write()
				tc.CMS.Commit()
			}
		})
	}
}
