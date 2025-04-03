package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	epochstypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
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

// UpdatePools first deletes all of the pools paired with any base denom in the store and then adds the highest liquidity pools that match to the store
func (k Keeper) UpdatePools(ctx sdk.Context) error {
	// baseDenomPools maps each base denom to a map of the highest liquidity pools paired with that base denom
	// ex. {osmo -> {atom : 100, weth : 200}}
	baseDenomPools := make(map[string]map[string]LiquidityPoolStruct)
	baseDenoms, err := k.GetAllBaseDenoms(ctx)
	if err != nil {
		return err
	}

	// Delete any pools that currently exist in the store + initialize baseDenomPools
	for _, baseDenom := range baseDenoms {
		k.DeleteAllPoolsForBaseDenom(ctx, baseDenom.Denom)
		baseDenomPools[baseDenom.Denom] = make(map[string]LiquidityPoolStruct)
	}

	// Update baseDenomPools with the highest liquidity pools
	if err := k.UpdateHighestLiquidityPools(ctx, baseDenomPools); err != nil {
		return err
	}

	// Update the pools in the store
	for baseDenom, pools := range baseDenomPools {
		for denom, pool := range pools {
			k.SetPoolForDenomPair(ctx, baseDenom, denom, pool.PoolId)
		}
	}

	return nil
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
