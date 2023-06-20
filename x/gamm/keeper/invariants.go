package keeper

// DONTCOVER

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
)

const poolBalanceInvariantName = "pool-account-balance-equals-expected"

// RegisterInvariants registers all gamm invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, keeper Keeper, bk types.BankKeeper) {
	ir.RegisterRoute(types.ModuleName, poolBalanceInvariantName, PoolAccountInvariant(keeper, bk))
}

// AllInvariants runs all invariants of the gamm module
func AllInvariants(keeper Keeper, bk types.BankKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg, broke := PoolAccountInvariant(keeper, bk)(ctx)
		return msg, broke
	}
}

// PoolAccountInvariant checks that the pool account balance reflects the sum of
// pool assets
func PoolAccountInvariant(keeper Keeper, bk types.BankKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		pools, err := keeper.GetPoolsAndPoke(ctx)
		if err != nil {
			return sdk.FormatInvariant(types.ModuleName, poolBalanceInvariantName,
				"\tgamm pool retrieval failed"), true
		}

		for _, pool := range pools {
			expectedCoins := pool.GetTotalPoolLiquidity(ctx)
			actualCoins := bk.GetAllBalances(ctx, pool.GetAddress())
			if !actualCoins.IsAllGTE(expectedCoins) {
				return sdk.FormatInvariant(types.ModuleName, poolBalanceInvariantName,
					fmt.Sprintf("\tgamm pool id %d\n\t pool-expected coins: %s\n\t account coins: %s\n",
						pool.GetId(), expectedCoins, actualCoins)), true
			}
		}

		return sdk.FormatInvariant(types.ModuleName, poolBalanceInvariantName,
			"\tgamm all pool asset coins and account coins match\n"), false
	}
}
