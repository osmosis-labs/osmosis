package keeper

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v32/x/gamm/pool-models/balancer"
)

// ReplaceBalancerPoolDenomUnsafe renames oldDenom to newDenom in the assets of a balancer pool, and in
// the weights of a smooth weight change if one is set, keeping each list sorted by denom. It is only
// for upgrade handlers.
//
// It moves no funds. The caller must already have exchanged the pool's whole oldDenom balance for the
// same amount of newDenom: the pool's recorded amounts and weights are kept as they are, so the swap
// must be exactly 1:1.
func (k Keeper) ReplaceBalancerPoolDenomUnsafe(ctx sdk.Context, poolId uint64, oldDenom, newDenom string) error {
	if oldDenom == "" || newDenom == "" || oldDenom == newDenom {
		return fmt.Errorf("invalid denom replacement %q -> %q", oldDenom, newDenom)
	}
	pool, err := k.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		return err
	}
	balancerPool, ok := pool.(*balancer.Pool)
	if !ok {
		return fmt.Errorf("pool %d is %T, not a balancer pool", poolId, pool)
	}
	if err := replaceSortedAssetDenom(balancerPool.PoolAssets, oldDenom, newDenom); err != nil {
		return fmt.Errorf("pool %d assets: %w", poolId, err)
	}
	if change := balancerPool.PoolParams.SmoothWeightChangeParams; change != nil {
		if err := replaceSortedAssetDenom(change.InitialPoolWeights, oldDenom, newDenom); err != nil {
			return fmt.Errorf("pool %d initial weights: %w", poolId, err)
		}
		if err := replaceSortedAssetDenom(change.TargetPoolWeights, oldDenom, newDenom); err != nil {
			return fmt.Errorf("pool %d target weights: %w", poolId, err)
		}
	}
	return k.setPool(ctx, balancerPool)
}

// replaceSortedAssetDenom renames oldDenom to newDenom in assets, which must hold oldDenom and not
// newDenom, and sorts assets by denom.
func replaceSortedAssetDenom(assets []balancer.PoolAsset, oldDenom, newDenom string) error {
	replaced := -1
	for i, asset := range assets {
		switch asset.Token.Denom {
		case newDenom:
			return fmt.Errorf("already holds %s", newDenom)
		case oldDenom:
			replaced = i
		}
	}
	if replaced < 0 {
		return fmt.Errorf("does not hold %s", oldDenom)
	}
	assets[replaced].Token.Denom = newDenom
	sort.Slice(assets, func(i, j int) bool { return assets[i].Token.Denom < assets[j].Token.Denom })
	return nil
}
