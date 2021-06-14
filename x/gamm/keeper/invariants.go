package keeper

// DONTCOVER

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/gamm/types"
)

var powPrecision, _ = sdk.NewDecFromStr("0.00000001")

const poolBalanceInvariantName = "pool-account-balance-equals-expected"
// RegisterInvariants registers all governance invariants
func RegisterInvariants(ir sdk.InvariantRegistry, keeper Keeper, bk types.BankKeeper) {
	ir.RegisterRoute(types.ModuleName, poolBalanceInvariantName, PoolAccountInvariant(keeper, bk))
	ir.RegisterRoute(types.ModuleName, "pool-total-weight", PoolTotalWeightInvariant(keeper, bk))
	ir.RegisterRoute(types.ModuleName, "pool-product-constant", PoolProductConstantInvariant(keeper))
	// ir.RegisterRoute(types.ModuleName, "spot-price", SpotPriceInvariant(keeper, bk))
}

// AllInvariants runs all invariants of the gamm module
func AllInvariants(keeper Keeper, bk types.BankKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg, broke := PoolAccountInvariant(keeper, bk)(ctx)
		if broke {
			return msg, broke
		}
		msg, broke = PoolProductConstantInvariant(keeper)(ctx)
		if broke {
			return msg, broke
		}
		return PoolTotalWeightInvariant(keeper, bk)(ctx)
	}
}

// PoolAccountInvariant checks that the pool account balance reflects the sum of
// pool assets
func PoolAccountInvariant(keeper Keeper, bk types.BankKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		pools, err := keeper.GetPools(ctx)
		if err != nil {
			return sdk.FormatInvariant(types.ModuleName, poolBalanceInvariantName,
				fmt.Sprintf("\tgamm pool retrieval failed")), true
		}

		for _, pool := range pools {
			assetCoins := types.PoolAssetsCoins(pool.GetAllPoolAssets())
			accCoins := bk.GetAllBalances(ctx, pool.GetAddress())
			if !assetCoins.IsEqual(accCoins) {
				return sdk.FormatInvariant(types.ModuleName, poolBalanceInvariantName,
					fmt.Sprintf("\tgamm pool id %d\n\tasset coins: %s\n\taccount coins: %s\n",
					pool.GetId(), assetCoins, accCoins)), true
			}
		}

		return sdk.FormatInvariant(types.ModuleName, poolBalanceInvariantName,
			fmt.Sprintf("\tgamm all pool asset coins and account coins match\n")), false
	}
}

// PoolTotalWeightInvariant checks that the pool total weight reflect the sum of
// pool weights
func PoolTotalWeightInvariant(keeper Keeper, bk types.BankKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		pools, err := keeper.GetPools(ctx)
		if err != nil {
			return sdk.FormatInvariant(types.ModuleName, "pool-total-weight",
				fmt.Sprintf("\tgamm pool retrieval failed")), true
		}

		for _, pool := range pools {
			totalWeight := sdk.ZeroInt()
			for _, asset := range pool.GetAllPoolAssets() {
				totalWeight = totalWeight.Add(asset.Weight)
			}
			if !totalWeight.Equal(pool.GetTotalWeight()) {
				return sdk.FormatInvariant(types.ModuleName, "pool-total-weight",
					fmt.Sprintf("\tgamm pool id %d\n\tcalculated weight sum %s\n\tpool total weight: %s\n",
					pool.GetId(), totalWeight, pool.GetTotalWeight())), true
			}
		}

		return sdk.FormatInvariant(types.ModuleName, "pool-total-weight",
			fmt.Sprintf("\tgamm all pool calculated and stored total weight match\n")), false
	}
}

func genericPow(base, exp sdk.Dec) sdk.Dec {
	if !base.GTE(sdk.OneDec().MulInt64(2)) {
		return pow(base, exp)
	}
	return powApprox(sdk.OneDec().Quo(base), exp.Neg(), powPrecision)
}

func constantChange(p1, p2 types.PoolI) sdk.Dec {
	product := sdk.OneDec()
	totalWeight := p1.GetTotalWeight()
	assets1 := p1.GetAllPoolAssets()
	for _, asset1 := range assets1 {
		asset2, err := p2.GetPoolAsset(asset1.Token.Denom)
		if err != nil {
			panic(err)
		}
		w := asset1.Weight.ToDec().Quo(totalWeight.ToDec())
		ratio := asset1.Token.Amount.ToDec().Quo(asset2.Token.Amount.ToDec())
		power := genericPow(ratio, w)
		product = product.Mul(power)
	}

	return product
}

var (
	errorMargin, _ = sdk.NewDecFromStr("0.0001") // 0.01%
)

func PoolProductConstantInvariant(keeper Keeper) sdk.Invariant {
	pools := make(map[uint64]types.PoolI)

	return func(ctx sdk.Context) (string, bool) {
		newpools, err := keeper.GetPools(ctx)
		if err != nil {
			return sdk.FormatInvariant(types.ModuleName, "pool-product-constant",
				fmt.Sprintf("\tgamm pool retrieval failed")), true
		}

		for _, pool := range newpools {
			oldpool, ok := pools[pool.GetId()]
			if !ok {
				pools[pool.GetId()] = pool
				continue
			}

			change := constantChange(oldpool, pool)
			if !(sdk.OneDec().Sub(errorMargin).LT(change) && change.LT(sdk.OneDec().Add(errorMargin))) {
				return sdk.FormatInvariant(types.ModuleName, "pool-product-constant",
				fmt.Sprintf("\tgamm pool id %d product constant changed\n\tdelta: %s\n", pool.GetId(), change.String())), true
			}

			pools[pool.GetId()] = pool
		}

		return sdk.FormatInvariant(types.ModuleName, "pool-product-constant",
			fmt.Sprintf("\tgamm all pool product constant preserved\n")), false
	}
}
