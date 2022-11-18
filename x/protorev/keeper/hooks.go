package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/osmosis-labs/osmosis/v12/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v12/x/protorev/types"
)

type EpochHooks struct {
	k Keeper
}

// Struct used to track the pool with the highest liquidity
type LiquidityPoolStruct struct {
	Liquidity sdk.Int
	PoolId    uint64
}

var (
	_ epochstypes.EpochHooks = EpochHooks{}
)

func (k Keeper) EpochHooks() epochstypes.EpochHooks {
	return EpochHooks{k}
}

///////////////////////////////////////////////////////

// BeforeEpochStart is the epoch start hook.
func (h EpochHooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

// AfterEpochEnd is the epoch end hook.
func (h EpochHooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	// Get the highest liquidity pools from the store
	osmoPools, atomPools, err := h.k.GetHighestLiquidityPools(ctx)
	if err != nil {
		return err
	}

	// Update the pools in the store
	for token, poolInfo := range osmoPools {
		h.k.SetOsmoPool(ctx, token, poolInfo.PoolId)
	}
	for token, poolInfo := range atomPools {
		h.k.SetAtomPool(ctx, token, poolInfo.PoolId)
	}

	return nil
}

// getHighestLiquidityPools returns the highest liquidity pools for both Osmo and Atom pairs and Osmo/Atom
func (k Keeper) GetHighestLiquidityPools(ctx sdk.Context) (map[string]LiquidityPoolStruct, map[string]LiquidityPoolStruct, error) {
	pools, err := k.gammKeeper.GetPoolsAndPoke(ctx)
	if err != nil {
		return nil, nil, err
	}

	osmoPools := make(map[string]LiquidityPoolStruct)
	atomPools := make(map[string]LiquidityPoolStruct)

	// Iterate through all pools and find valid matches
	for _, pool := range pools {
		coins := pool.GetTotalPoolLiquidity(ctx)

		// Pool must be active and the number of coins must be 2
		if pool.IsActive(ctx) && len(coins) == 2 {
			tokenA := coins[0]
			tokenB := coins[1]

			newPool := LiquidityPoolStruct{
				PoolId:    pool.GetId(),
				Liquidity: tokenA.Amount.Mul(tokenB.Amount),
			}

			// Check if there is a match with osmo
			if otherDenom, match := types.CheckMatch(tokenA.Denom, tokenB.Denom, types.OsmosisDenomination); match {
				if currPool, ok := osmoPools[otherDenom]; !ok {
					osmoPools[otherDenom] = newPool
				} else {
					if newPool.Liquidity.GT(currPool.Liquidity) {
						osmoPools[otherDenom] = newPool
					}
				}
			}

			// Check if there is a match with atom
			if otherDenom, match := types.CheckMatch(tokenA.Denom, tokenB.Denom, types.AtomDenomination); match {
				if currPool, ok := atomPools[otherDenom]; !ok {
					atomPools[otherDenom] = newPool
				} else {
					if newPool.Liquidity.GT(currPool.Liquidity) {
						atomPools[otherDenom] = newPool
					}
				}
			}
		}
	}

	return osmoPools, atomPools, nil
}
