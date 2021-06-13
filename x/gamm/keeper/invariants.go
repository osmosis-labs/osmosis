package keeper

// DONTCOVER

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/gamm/types"
)
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

func PoolProductConstantInvariant(keeper Keeper) sdk.Invariant {
	constants := make(map[uint64]sdk.Dec)

	return func(ctx sdk.Context) (string, bool) {
		pools, err := keeper.GetPools(ctx)
		if err != nil {
			return sdk.FormatInvariant(types.ModuleName, "pool-product-constant",
				fmt.Sprintf("\tgamm pool retrieval failed")), true
		}

		for _, pool := range pools {
			constant, ok := constants[pool.GetId()]
			if !ok {
				constants[pool.GetId()] = pool.PoolProductConstant()
				continue
			}
			if !constant.Equal(pool.PoolProductConstant()) {
				return sdk.FormatInvariant(types.ModuleName, "pool-product-constant",
					fmt.Sprintf("\tgamm pool id %d\n\t product constant changed", pool.GetId())), true
			}
		}

		return sdk.FormatInvariant(types.ModuleName, "pool-product-constant",
			fmt.Sprintf("\tgamm all pool product constant preserved\n")), false
	}
}
