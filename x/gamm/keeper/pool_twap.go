package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/x/gamm/types"
)

func (k Keeper) GetPoolTwap(ctx sdk.Context, poolId uint64) (types.PoolTwapHistory, error) {
	store := ctx.KVStore(k.storeKey)
	poolTwapKey := k.GetRecentPoolTwapKey(ctx, poolId)

	if len(poolTwapKey) == 0 {
		return types.PoolTwapHistory{}, fmt.Errorf("pool twap does not exist")
	}
	if !store.Has(poolTwapKey) {
		return types.PoolTwapHistory{}, fmt.Errorf("pool twap with ID %d does not exist", poolId)
	}

	bz := store.Get(poolTwapKey)
	poolTwap := types.PoolTwapHistory{}
	k.cdc.MustUnmarshalBinaryBare(bz, &poolTwap)

	return poolTwap, nil
}

func (k Keeper) SetPoolTwap(ctx sdk.Context, poolTwap types.PoolTwapHistory) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&poolTwap)
	timestamp := ctx.BlockTime().Unix()

	poolTwapKey := types.GetKeyPoolTwaps(poolTwap.PoolId, timestamp)
	store.Set(poolTwapKey, bz)
}

func (k Keeper) newPoolTwap(ctx sdk.Context, poolId uint64) (types.PoolTwapHistory, error) {
	pool, err := k.GetPool(ctx, poolId)
	if err != nil {
		return types.PoolTwapHistory{}, err
	}

	var twapPairs []*types.TwapPair
	// iterate through all assets, creating all possible pairs
	for i, tokenIn := range pool.GetAllPoolAssets() {
		for j, tokenOut := range pool.GetAllPoolAssets() {
			// if it is not the same token, create a twap pair
			if i != j {
				spotPrice, err := k.CalculateSpotPrice(ctx, poolId, tokenIn.Token.Denom, tokenOut.Token.Denom)
				if err != nil {
					return types.PoolTwapHistory{}, err
				}
				twapPair := types.TwapPair{
					TokenIn:         tokenIn.Token.Denom,
					TokenOut:        tokenOut.Token.Denom,
					PriceCumulative: spotPrice,
				}

				twapPairs = append(twapPairs, &twapPair)
			}
		}
	}

	poolTwap := types.PoolTwapHistory{
		TimeStamp: ctx.BlockTime(),
		PoolId:    poolId,
		TwapPairs: twapPairs,
	}
	// fmt.Printf("\n Result of newPoolTwap: %s", poolTwap.String())
	return poolTwap, nil
}

func (k Keeper) CreatePoolTwap(ctx sdk.Context, poolId uint64) (err error) {
	// fmt.Printf("\n CreatePoolTwap Called\n")
	poolTwap, err := k.newPoolTwap(ctx, poolId)
	if err != nil {
		return err
	}
	k.SetPoolTwap(ctx, poolTwap)
	return nil
}

// update pool twap with single token that has changed
func (k Keeper) RecordPoolTwap(ctx sdk.Context, poolId uint64, changedToken string) (err error) {
	recentPoolTwap, err := k.GetPoolTwap(ctx, poolId)
	if err != nil {
		return err
	}
	// iterate through the array of spot prices,
	// updating all spot prices that are related to the changed token
	fmt.Printf("\n Recording spot price for pool: %d", poolId)
	for i, spotPrice := range recentPoolTwap.TwapPairs {
		if changedToken == spotPrice.TokenIn || changedToken == spotPrice.TokenOut {
			fmt.Printf("\n token in: %s", spotPrice.TokenIn)
			fmt.Printf("\n token out: %s", spotPrice.TokenOut)
			changedSpotPrice, err := k.CalculateSpotPrice(ctx, poolId, spotPrice.TokenIn, spotPrice.TokenOut)
			if err != nil {
				return err
			}
			fmt.Printf("\n Calculated spot price: %d", changedSpotPrice)
			recentPoolTwap.TwapPairs[i].PriceCumulative = changedSpotPrice
		}
	}

	poolTwap := types.PoolTwapHistory{
		TimeStamp: ctx.BlockTime(),
		PoolId:    poolId,
		TwapPairs: recentPoolTwap.TwapPairs,
	}

	k.SetPoolTwap(ctx, poolTwap)
	return nil
}

func (k Keeper) GetRecentPoolTwapKey(ctx sdk.Context, poolId uint64) []byte {
	store := ctx.KVStore(k.storeKey)
	timestamp := ctx.BlockTime().Unix()

	iteratorStart := types.GetKeyPoolTwaps(poolId, 0)
	iteratorEnd := types.GetKeyPoolTwaps(poolId, timestamp)

	// use reverse iterator to list in time order
	iterator := store.ReverseIterator(iteratorStart, iteratorEnd)
	defer iterator.Close()
	recentPoolTwapKey := []byte{}

	// the most recent value in iterator points to the most
	// recently added pool twap key
	if iterator.Valid() {
		recentPoolTwapKey = iterator.Key()
	}

	return recentPoolTwapKey
}
