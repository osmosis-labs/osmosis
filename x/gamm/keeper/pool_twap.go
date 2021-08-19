package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/x/gamm/types"
)

func (k Keeper) GetPoolTwapHistory(ctx sdk.Context, poolId uint64, queryTime time.Time) (types.PoolTwapHistory, error) {
	store := ctx.KVStore(k.storeKey)
	poolTwapKey := k.GetPoolTwapKey(ctx, poolId, queryTime)

	// if twap have not existed before, create new pool twap
	if len(poolTwapKey) == 0 {
		poolTwap, err := k.newPoolTwapHistory(ctx, poolId)
		if err != nil {
			return types.PoolTwapHistory{}, err
		}
		k.SetPoolTwap(ctx, poolTwap)
		return poolTwap, nil

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

func (k Keeper) CreatePoolTwap(ctx sdk.Context, poolId uint64) (err error) {
	poolTwap, err := k.newPoolTwapHistory(ctx, poolId)
	if err != nil {
		return err
	}

	k.SetPoolTwap(ctx, poolTwap)
	return nil
}

func (k Keeper) newPoolTwapHistory(ctx sdk.Context, poolId uint64) (types.PoolTwapHistory, error) {
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
					PriceCumulative: sdk.ZeroDec(),
					SpotPrice:       spotPrice,
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
	return poolTwap, nil
}

// update pool twap with token(s) that have changed
func (k Keeper) UpdatePoolTwap(ctx sdk.Context, poolId uint64, tokens ...string) (err error) {
	currentTime := ctx.BlockTime()
	recentPoolTwap, err := k.GetPoolTwapHistory(ctx, poolId, currentTime)
	if err != nil {
		return err
	}

	if len(tokens) > 3 {
		return fmt.Errorf("tokens should be two or less")
	}

	// iterate through the array of spot prices,
	// updating all spot prices that are related to the changed token
	for i, spotPrice := range recentPoolTwap.TwapPairs {
		currentTimeElapsedDuration := currentTime.Sub(recentPoolTwap.TimeStamp)
		currentTimeElapsed := sdk.NewDec(currentTimeElapsedDuration.Nanoseconds()).QuoInt(sdk.NewInt(100000000))

		// if any of the spot price pairs in twap history are realted to the swapped tokens,
		// update the spot price cumulative
		if contains(tokens, spotPrice.TokenIn, spotPrice.TokenOut) {
			changedSpotPrice, err := k.CalculateSpotPrice(ctx, poolId, spotPrice.TokenIn, spotPrice.TokenOut)
			if err != nil {
				return err
			}
			recentPoolTwap.TwapPairs[i].PriceCumulative = recentPoolTwap.TwapPairs[i].PriceCumulative.Add(changedSpotPrice.Mul(currentTimeElapsed))
			recentPoolTwap.TwapPairs[i].SpotPrice = changedSpotPrice
		} else {
			recentPoolTwap.TwapPairs[i].PriceCumulative = recentPoolTwap.TwapPairs[i].PriceCumulative.Add(recentPoolTwap.TwapPairs[i].SpotPrice.Mul(currentTimeElapsed))
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

func (k Keeper) GetPoolTwapKey(ctx sdk.Context, poolId uint64, queryTime time.Time) []byte {
	store := ctx.KVStore(k.storeKey)
	timestamp := queryTime.Unix()

	iteratorStart := types.GetKeyPoolTwaps(poolId, 0)
	iteratorEnd := types.GetKeyPoolTwaps(poolId, timestamp)

	// use reverse iterator to list in time order
	iterator := store.ReverseIterator(iteratorStart, iteratorEnd)
	defer iterator.Close()
	poolTwapKey := []byte{}

	// the most recent value in iterator points to the most
	// recently added pool twap key
	if iterator.Valid() {
		poolTwapKey = iterator.Key()
	}

	return poolTwapKey
}

// gets the most recent spot price of token pair in a specific duration
func (k Keeper) GetRecentPoolTwapSpotPrice(
	ctx sdk.Context,
	poolId uint64,
	tokenInDenom string,
	tokenOutDenom string,
	duration time.Duration,
) (sdk.Dec, error) {
	currentTime := ctx.BlockTime()
	currentTimeAdjacentPoolTwap, err := k.GetPoolTwapHistory(ctx, poolId, currentTime)
	if err != nil {
		return sdk.Dec{}, err
	}

	desiredTime := ctx.BlockTime().Add(-duration)
	desiredTimeAdjacentPoolTwap, err := k.GetPoolTwapHistory(ctx, poolId, desiredTime)
	if err != nil {
		return sdk.Dec{}, err
	}

	var currentTimeAdjacentPriceCumulative, desiredTimeAdjacentPriceCumulative sdk.Dec
	var currentTimeAdjacentSpotPrice, desiredTimeAdjacentSpotPrice sdk.Dec
	var currentTimeAdjacentTime, desiredTimeAdjacentTime time.Time

	// same index between recentPoolTwap and desiredPoolTwap
	// can be used since they share same order
	for i, twapPair := range currentTimeAdjacentPoolTwap.TwapPairs {
		if twapPair.TokenIn == tokenInDenom && twapPair.TokenOut == tokenOutDenom {
			currentTimeAdjacentPriceCumulative = currentTimeAdjacentPoolTwap.TwapPairs[i].PriceCumulative
			desiredTimeAdjacentPriceCumulative = desiredTimeAdjacentPoolTwap.TwapPairs[i].PriceCumulative

			currentTimeAdjacentSpotPrice = currentTimeAdjacentPoolTwap.TwapPairs[i].SpotPrice
			desiredTimeAdjacentSpotPrice = desiredTimeAdjacentPoolTwap.TwapPairs[i].SpotPrice

			currentTimeAdjacentTime = currentTimeAdjacentPoolTwap.TimeStamp
			desiredTimeAdjacentTime = desiredTimeAdjacentPoolTwap.TimeStamp
		}
	}

	currentTimeElapsedDuration := currentTime.Sub(currentTimeAdjacentTime)
	desiredTimeElapsedDuration := currentTime.Sub(desiredTimeAdjacentTime)

	currentTimeElapsed := sdk.NewDec(currentTimeElapsedDuration.Nanoseconds()).QuoInt(sdk.NewInt(100000000))
	desiredTimeElapsed := sdk.NewDec(desiredTimeElapsedDuration.Nanoseconds()).QuoInt(sdk.NewInt(100000000))

	curentTimePriceCumulative := currentTimeAdjacentPriceCumulative.Add(currentTimeAdjacentSpotPrice.Mul(currentTimeElapsed))
	desiredTimePriceCumulative := desiredTimeAdjacentPriceCumulative.Add(desiredTimeAdjacentSpotPrice.Mul(desiredTimeElapsed))

	// twap calculated using (priceCumulative2 - priceCumulative1) / (timeStamp1 - timeStamp2)
	priceCumulativeDifference := curentTimePriceCumulative.Sub(desiredTimePriceCumulative)
	timeDifference := sdk.NewDec(currentTime.Sub(desiredTime).Nanoseconds()).QuoInt(sdk.NewInt(1000000000))
	twap := priceCumulativeDifference.Quo(timeDifference)

	return twap, nil
}

// function for validating whether specic string(s) are contained in a slice
func contains(s []string, e ...string) bool {
	for _, a := range s {
		for _, b := range e {
			if a == b {
				return true
			}
		}
	}
	return false
}
