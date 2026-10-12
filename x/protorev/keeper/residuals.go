package keeper

import (
	"strings"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v31/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v31/x/protorev/types"
)

// BurnResidualBalances swaps protorev module balances that are not base denoms into OSMO and burns
// the OSMO by sending it to the null address. These balances are not arb profits (arbs always settle
// in a base denom); they come from direct sends or from denoms that were removed from the base denom
// list. They are burned directly and do not go through the developer fee split.
//
// Candidates are drawn from the highest liquidity pool store rather than from the module balance, so
// arbitrary tokens sent to the module account cannot grow the work done here. At most
// MaxResidualSwapsPerEpoch swaps are attempted per call, and the starting candidate rotates with
// block height so a fixed set of failing or dust balances cannot starve the rest.
//
// Each swap and burn runs in its own cache context; a denom that fails to swap is logged and left in
// the module account for a later epoch. Returns the total amount of OSMO burned.
func (k Keeper) BurnResidualBalances(ctx sdk.Context, baseDenoms []types.BaseDenom) sdk.Coin {
	burned := sdk.NewCoin(types.OsmosisDenomination, osmomath.ZeroInt())

	candidates := k.residualSwapCandidates(ctx, baseDenoms)
	if len(candidates) == 0 {
		return burned
	}

	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	start := int(ctx.BlockHeight() % int64(len(candidates)))
	attempts := 0

	for i := 0; i < len(candidates) && attempts < types.MaxResidualSwapsPerEpoch; i++ {
		denom := candidates[(start+i)%len(candidates)]
		balance := k.bankKeeper.GetBalance(ctx, moduleAddr, denom)
		if !balance.IsPositive() {
			continue
		}

		route, found := k.residualSwapRoute(ctx, baseDenoms, denom)
		if !found {
			continue
		}
		attempts++

		amountOut := osmomath.ZeroInt()
		err := osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			// No slippage bound, matching the txfees non-native fee swaps. Balances are drained
			// daily, so each swap is expected to be small relative to the pool.
			out, err := k.poolmanagerKeeper.RouteExactAmountInNoTakerFee(cacheCtx, moduleAddr, route, balance, osmomath.ZeroInt())
			if err != nil {
				return err
			}
			amountOut = out
			osmoOut := sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, out))
			return k.bankKeeper.SendCoinsFromModuleToAccount(cacheCtx, types.ModuleName, types.DefaultNullAddress, osmoOut)
		})
		if err != nil {
			ctx.Logger().Error("protorev: failed to swap residual balance to OSMO", "coin", balance.String(), "error", err)
			continue
		}
		burned = burned.AddAmount(amountOut)
	}

	if burned.IsPositive() {
		ctx.Logger().Info("protorev: burned residual balances", "osmo_burned", burned.String(), "swaps_attempted", attempts)
	}
	return burned
}

// residualSwapCandidates returns the unique non-base denoms that are paired with a current base denom
// in the highest liquidity pool store, in store key order.
func (k Keeper) residualSwapCandidates(ctx sdk.Context, baseDenoms []types.BaseDenom) []string {
	isBaseDenom := make(map[string]bool, len(baseDenoms))
	for _, baseDenom := range baseDenoms {
		isBaseDenom[baseDenom.Denom] = true
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDenomPairToPool)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyPrefixDenomPairToPool)
	defer iterator.Close()

	seen := make(map[string]bool)
	var candidates []string
	for ; iterator.Valid(); iterator.Next() {
		pair := string(iterator.Key()[len(types.KeyPrefixDenomPairToPool):])
		baseDenom, matchDenom, ok := strings.Cut(pair, "|")
		// Entries for denoms no longer in the base denom list are stale and not used for routing.
		if !ok || !isBaseDenom[baseDenom] || isBaseDenom[matchDenom] || seen[matchDenom] {
			continue
		}
		seen[matchDenom] = true
		candidates = append(candidates, matchDenom)
	}
	return candidates
}

// residualSwapRoute returns a route from denom to OSMO: a direct OSMO pool if one is stored, otherwise
// a two hop route through the first base denom (in priority order) that has pools for both legs.
func (k Keeper) residualSwapRoute(ctx sdk.Context, baseDenoms []types.BaseDenom, denom string) ([]poolmanagertypes.SwapAmountInRoute, bool) {
	if poolId, err := k.GetPoolForDenomPair(ctx, types.OsmosisDenomination, denom); err == nil {
		return []poolmanagertypes.SwapAmountInRoute{{PoolId: poolId, TokenOutDenom: types.OsmosisDenomination}}, true
	}

	for _, baseDenom := range baseDenoms {
		if baseDenom.Denom == types.OsmosisDenomination {
			continue
		}
		firstPoolId, err := k.GetPoolForDenomPair(ctx, baseDenom.Denom, denom)
		if err != nil {
			continue
		}
		secondPoolId, err := k.GetPoolForDenomPairNoOrder(ctx, types.OsmosisDenomination, baseDenom.Denom)
		if err != nil {
			continue
		}
		return []poolmanagertypes.SwapAmountInRoute{
			{PoolId: firstPoolId, TokenOutDenom: baseDenom.Denom},
			{PoolId: secondPoolId, TokenOutDenom: types.OsmosisDenomination},
		}, true
	}
	return nil, false
}
