package keeper

import (
	"bytes"
	"sort"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v32/x/protorev/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

type EpochHooks struct {
	k Keeper
}

// Struct used to track the pool with the highest liquidity
type LiquidityPoolStruct struct {
	Liquidity osmomath.Int
	PoolId    uint64
}

var _ epochstypes.EpochHooks = EpochHooks{}

func (k Keeper) EpochHooks() epochstypes.EpochHooks {
	return EpochHooks{k}
}

// GetModuleName implements types.EpochHooks.
func (EpochHooks) GetModuleName() string {
	return types.ModuleName
}

// BeforeEpochStart is the epoch start hook.
func (h EpochHooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

// AfterEpochEnd is the epoch end hook.
func (h EpochHooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	if h.k.GetProtoRevEnabled(ctx) {
		switch epochIdentifier {
		case "day":
			// Calculate and distribute protorev profits
			err := h.CalculateAndDistributeProfits(ctx)
			if err != nil {
				return err
			}

			// Increment number of days since module genesis to properly calculate developer fees after cyclic arbitrage trades
			if daysSinceGenesis, err := h.k.GetDaysSinceModuleGenesis(ctx); err != nil {
				h.k.SetDaysSinceModuleGenesis(ctx, 1)
			} else {
				h.k.SetDaysSinceModuleGenesis(ctx, daysSinceGenesis+1)
			}

			// Update the pools in the store
			return h.k.UpdatePools(ctx)
		}
	}

	return nil
}

// CalculateAndDistributeProfits is executed after epoch. It gets the current base denom profits and distributes them.
func (h EpochHooks) CalculateAndDistributeProfits(ctx sdk.Context) error {
	// Get the current arb profits (only in base denoms to prevent spam vector)
	profit, err := h.k.CurrentBaseDenomProfits(ctx)
	if err != nil {
		return err
	}

	// Distribute profits to developer account, community pool, and burn osmo
	err = h.k.DistributeProfit(ctx, profit)
	if err != nil {
		return err
	}
	return nil
}

// CurrentBaseDenomProfits retrieves the current balance of the protorev module account and filters for base denoms.
func (k Keeper) CurrentBaseDenomProfits(ctx sdk.Context) (sdk.Coins, error) {
	moduleAcc := k.accountKeeper.GetModuleAddress(types.ModuleName)

	baseDenoms, err := k.GetAllBaseDenoms(ctx)
	if err != nil {
		return nil, err
	}

	// Get the current protorev balance of all denoms
	protorevBalanceAllDenoms := k.bankKeeper.GetAllBalances(ctx, moduleAcc)

	// Filter for base denoms
	var protorevBalanceBaseDenoms sdk.Coins

	for _, baseDenom := range baseDenoms {
		amountOfBaseDenom := protorevBalanceAllDenoms.AmountOf(baseDenom.Denom)
		if !amountOfBaseDenom.IsZero() {
			protorevBalanceBaseDenoms = append(protorevBalanceBaseDenoms, sdk.NewCoin(baseDenom.Denom, amountOfBaseDenom))
		}
	}

	return protorevBalanceBaseDenoms.Sort(), nil
}

// UpdatePools recomputes the highest liquidity pools and only writes routes whose
// selected pool changed. Routes without an eligible pool are removed.
func (k Keeper) UpdatePools(ctx sdk.Context) error {
	// baseDenomPools maps each base denom to a map of the highest liquidity pools paired with that base denom
	// ex. {osmo -> {atom : 100, weth : 200}}
	baseDenomPools := make(map[string]map[string]LiquidityPoolStruct)
	baseDenoms, err := k.GetAllBaseDenoms(ctx)
	if err != nil {
		return err
	}

	// Keep the existing routes until all pool liquidity has been read successfully.
	for _, baseDenom := range baseDenoms {
		baseDenomPools[baseDenom.Denom] = make(map[string]LiquidityPoolStruct)
	}

	// Update baseDenomPools with the highest liquidity pools
	if err := k.UpdateHighestLiquidityPools(ctx, baseDenomPools); err != nil {
		return err
	}

	for _, baseDenom := range baseDenoms {
		k.syncHighestLiquidityPools(ctx, baseDenom.Denom, baseDenomPools[baseDenom.Denom])
	}

	return nil
}

// syncHighestLiquidityPools leaves unchanged route entries untouched. Iterators
// are closed before applying any writes to their key range.
func (k Keeper) syncHighestLiquidityPools(ctx sdk.Context, baseDenom string, pools map[string]LiquidityPoolStruct) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDenomPairToPool)
	store = prefix.NewStore(store, types.GetKeyPrefixDenomPairToPool(baseDenom, ""))
	seen := make(map[string]struct{}, len(pools))
	var obsolete [][]byte
	type update struct {
		denom string
		value []byte
	}
	var updates []update
	iterator := store.Iterator(nil, nil)
	for ; iterator.Valid(); iterator.Next() {
		denom := string(iterator.Key())
		pool, exists := pools[denom]
		if !exists {
			obsolete = append(obsolete, append([]byte(nil), iterator.Key()...))
			continue
		}
		seen[denom] = struct{}{}
		value := sdk.Uint64ToBigEndian(pool.PoolId)
		if !bytes.Equal(iterator.Value(), value) {
			updates = append(updates, update{denom, value})
		}
	}
	iterator.Close()
	for _, key := range obsolete {
		store.Delete(key)
	}
	for _, entry := range updates {
		store.Set([]byte(entry.denom), entry.value)
	}
	var added []string
	for denom := range pools {
		if _, exists := seen[denom]; !exists {
			added = append(added, denom)
		}
	}
	sort.Strings(added)
	for _, denom := range added {
		store.Set([]byte(denom), sdk.Uint64ToBigEndian(pools[denom].PoolId))
	}
}

// UpdateHighestLiquidityPools updates the baseDenomPools map (passed in by reference) with the
// highest liquidity pools for each base denom by iterating through all pools, getting the
// total liquidity for each pool, and updating the highest liquidity pools based upon comparing total liquidity.
func (k Keeper) UpdateHighestLiquidityPools(ctx sdk.Context, baseDenomPools map[string]map[string]LiquidityPoolStruct) error {
	pools, err := k.poolmanagerKeeper.AllPools(ctx)
	if err != nil {
		return err
	}

	for _, pool := range pools {
		coins, err := k.poolmanagerKeeper.GetTotalPoolLiquidity(ctx, pool.GetId())
		if err != nil {
			return err
		}

		// Pool must be active and the number of coins must be 2
		if pool.IsActive(ctx) && len(coins) == 2 {
			tokenA := coins[0]
			tokenB := coins[1]

			newPool := LiquidityPoolStruct{
				PoolId:    pool.GetId(),
				Liquidity: tokenA.Amount.Mul(tokenB.Amount),
			}

			// Update happens both ways to ensure the pools that contain multiple base denoms are properly updated
			if highestLiquidityPools, ok := baseDenomPools[tokenA.Denom]; ok {
				k.compareAndStoreHighestLiquidityPool(tokenB.Denom, highestLiquidityPools, newPool)
			}
			if highestLiquidityPools, ok := baseDenomPools[tokenB.Denom]; ok {
				k.compareAndStoreHighestLiquidityPool(tokenA.Denom, highestLiquidityPools, newPool)
			}
		}
	}

	return nil
}

// compareAndStoreHighestLiquidityPool updates the pool with the highest liquidity for the base denom
func (k Keeper) compareAndStoreHighestLiquidityPool(denom string, pools map[string]LiquidityPoolStruct, newPool LiquidityPoolStruct) {
	if currPool, ok := pools[denom]; !ok {
		pools[denom] = newPool
	} else {
		if newPool.Liquidity.GT(currPool.Liquidity) {
			pools[denom] = newPool
		}
	}
}
