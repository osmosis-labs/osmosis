package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/osmosis-labs/osmosis/x/gamm/types"
)

func (k Keeper) GetOrCreatePoolTwapHistory(ctx sdk.Context, poolId uint64, queryTime time.Time) (types.PoolTwapHistory, error) {
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

	bz := store.Get(poolTwapKey)
	poolTwap := types.PoolTwapHistory{}
	k.cdc.MustUnmarshalBinaryBare(bz, &poolTwap)

	return poolTwap, nil
}

func (k Keeper) GetPoolTwapHistory(ctx sdk.Context, poolId uint64, queryTime time.Time) (types.PoolTwapHistory, bool) {
	store := ctx.KVStore(k.storeKey)
	poolTwapKey := k.GetPoolTwapKey(ctx, poolId, queryTime)

	// returns false if pool twap history has not existed before
	if len(poolTwapKey) == 0 {
		return types.PoolTwapHistory{}, false
	}

	bz := store.Get(poolTwapKey)
	poolTwap := types.PoolTwapHistory{}
	k.cdc.MustUnmarshalBinaryBare(bz, &poolTwap)

	return poolTwap, true
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

// UpdatePoolTwap update pool twap with token(s) that have changed
func (k Keeper) UpdatePoolTwap(ctx sdk.Context, poolId uint64, tokens ...string) (err error) {
	currentTime := ctx.BlockTime()
	recentPoolTwap, err := k.GetOrCreatePoolTwapHistory(ctx, poolId, currentTime)
	if err != nil {
		return err
	}

	if len(tokens) > 2 {
		return fmt.Errorf("tokens should be two or less")
	}
	currentTimeElapsedDuration := currentTime.Sub(recentPoolTwap.TimeStamp)
	// division between integer, since minimum unit for duration is seconds
	currentTimeElapsed := currentTimeElapsedDuration.Nanoseconds() / 1_000_000_000

	// iterate through the array of spot prices,
	// updating all spot prices that are related to the changed token
	for i, twapPair := range recentPoolTwap.TwapPairs {
		// if any of the spot price pairs in twap history are related to the swapped tokens,
		// update the spot price cumulative
		if contains(tokens, twapPair.TokenIn, twapPair.TokenOut) {
			changedSpotPrice, err := k.CalculateSpotPrice(ctx, poolId, twapPair.TokenIn, twapPair.TokenOut)
			if err != nil {
				return err
			}
			recentPoolTwap.TwapPairs[i].PriceCumulative = recentPoolTwap.TwapPairs[i].PriceCumulative.Add(changedSpotPrice.Mul(sdk.NewDec(currentTimeElapsed)))
			recentPoolTwap.TwapPairs[i].SpotPrice = changedSpotPrice
		} else {
			recentPoolTwap.TwapPairs[i].PriceCumulative = recentPoolTwap.TwapPairs[i].PriceCumulative.Add(recentPoolTwap.TwapPairs[i].SpotPrice.Mul(sdk.NewDec(currentTimeElapsed)))
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

	// the most recent value in iterator points to the most recently added pool twap key
	if iterator.Valid() {
		poolTwapKey = iterator.Key()
	}

	return poolTwapKey
}

// GetRecentPoolTwapSpotPrice gets the most recent spot price of token pair in a specific duration
func (k Keeper) GetRecentPoolTwapSpotPrice(
	ctx sdk.Context,
	poolId uint64,
	tokenInDenom string,
	tokenOutDenom string,
	duration time.Duration,
) (sdk.Dec, error) {
	if duration < 1 {
		return sdk.Dec{}, fmt.Errorf("duration should be more than 1 second")
	}

	currentTime := ctx.BlockTime()
	// a second is added when querying pool twap history as it is queries current time exclusive
	currentTimeAdjacentPoolTwap, exists := k.GetPoolTwapHistory(ctx, poolId, currentTime.Add(time.Second))
	if !exists {
		return sdk.Dec{}, fmt.Errorf("pool twap history prior to current time does not exist")
	}

	// TODO: use duration.Seconds?
	desiredTime := ctx.BlockTime().Add(-duration * time.Second)
	desiredTimeAdjacentPoolTwap, exists := k.GetPoolTwapHistory(ctx, poolId, desiredTime.Add(time.Second))
	if !exists {
		return sdk.Dec{}, fmt.Errorf("pool twap history prior to time before duration does not exist")
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

	// if no previous pool twap records for desiredTime has been found, omit error
	if currentTimeAdjacentTime.Before(desiredTimeAdjacentTime) {
		return sdk.Dec{}, fmt.Errorf("twap uncalculatable for the given duration")
	}

	desiredTimeElapsed := desiredTime.Sub(desiredTimeAdjacentTime).Nanoseconds() / 1_000_000_000
	desiredTimePriceCumulative := desiredTimeAdjacentPriceCumulative.Add(desiredTimeAdjacentSpotPrice.Mul(sdk.NewDec(desiredTimeElapsed)))

	var currentTimeElapsed int64

	// if two poolTwaps are pointing to the same twap history, currentTimePriceCumulative should be calculated
	// using values of desiredTimePriceCumulative
	if currentTimeAdjacentTime == desiredTimeAdjacentTime {
		currentTimeElapsed = currentTime.Sub(desiredTime).Nanoseconds() / 1_000_000_000
	} else {
		currentTimeElapsed = currentTime.Sub(currentTimeAdjacentTime).Nanoseconds() / 1_000_000_000
	}
	currentTimePriceCumulative := currentTimeAdjacentPriceCumulative.Add(currentTimeAdjacentSpotPrice.Mul(sdk.NewDec(currentTimeElapsed)))

	// if adjacent time pool twap and current pool twap points to the same pool twap history,
	// no additional calculation needs to be done to get priceCumulative
	if currentTimeElapsed < 1 {
		currentTimePriceCumulative = currentTimeAdjacentPriceCumulative
	}
	if desiredTimeElapsed < 1 {
		desiredTimePriceCumulative = desiredTimeAdjacentPriceCumulative
	}

	// twap = (priceCumulative2 - priceCumulative1) / (timeStamp1 - timeStamp2)
	priceCumulativeDifference := currentTimePriceCumulative.Sub(desiredTimePriceCumulative)
	timeDifference := sdk.NewDec(currentTime.Sub(desiredTime).Nanoseconds()).QuoInt(sdk.NewInt(1_000_000_000))

	twap := priceCumulativeDifference.Quo(timeDifference)

	return twap, nil
}

// SetNextTwapHistoryDeleteIndex sets the next twa history for deletion
func (k Keeper) SetNextTwapHistoryDeleteIndex(ctx sdk.Context, index uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&gogotypes.UInt64Value{Value: index})
	store.Set(types.KeyNextTwapHistoryDeleteIndex, bz)
}

// GetNextTwapHistoryDeleteIndex returns the next pool number to be deleted
func (k Keeper) GetNextTwapHistoryDeleteIndex(ctx sdk.Context) uint64 {
	var nextTwapHistoryIndex uint64
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyNextTwapHistoryDeleteIndex)
	if bz == nil {
		// initialize the next pool twap to be deleted
		nextTwapHistoryIndex = 1
	} else {
		val := gogotypes.UInt64Value{}

		err := k.cdc.UnmarshalBinaryBare(bz, &val)
		if err != nil {
			panic(err)
		}
		nextTwapHistoryIndex = val.GetValue()
	}

	return nextTwapHistoryIndex
}

// DeleteTwapHisotry deletes any twap history records that are beyond currentTime + keepDuration
func (k Keeper) DeletePoolTwapHistory(ctx sdk.Context, poolId uint64, keepDuration time.Duration) error {
	if keepDuration < 1 {
		return fmt.Errorf("duration should be more than 1 seconds")
	}
	currentTime := ctx.BlockTime()
	store := ctx.KVStore(k.storeKey)

	// keep the closest pool twap history further than currentTime + keepDuration
	lastToKeepPoolTwapHistory, exists := k.GetPoolTwapHistory(ctx, poolId, currentTime.Add(-keepDuration))
	if !exists {
		return nil
	}

	latestPoolTwapHistoryToDelete, exists := k.GetPoolTwapHistory(ctx, poolId, lastToKeepPoolTwapHistory.TimeStamp)
	for exists {
		latestPoolTwapHistoryKeyToDelete := k.GetPoolTwapKey(ctx, poolId, latestPoolTwapHistoryToDelete.TimeStamp.Add(time.Second))
		store.Delete(latestPoolTwapHistoryKeyToDelete)
		latestPoolTwapHistoryToDelete, exists = k.GetPoolTwapHistory(ctx, poolId, latestPoolTwapHistoryToDelete.TimeStamp.Add(time.Second))
	}

	return nil
}

// DeleteTwapHistoryWithParams deletes pool twap history using the interval and duration provided by params
func (k Keeper) DeleteTwapHistoryWithParams(ctx sdk.Context) {
	params := k.GetParams(ctx)

	shouldDelete := (uint64(ctx.BlockHeight()) % params.TwapHistoryDeletionInterval) == 0
	if shouldDelete {
		poolId := k.GetNextTwapHistoryDeleteIndex(ctx)
		for i := uint64(0); i < params.NumTwapHistoryPerDeletion; i++ {
			err := k.DeletePoolTwapHistory(ctx, poolId, params.TwapHistoryKeepDuration)
			if err != nil {
				panic(err)
			}
			poolId += 1
		}
		k.SetNextTwapHistoryDeleteIndex(ctx, poolId)
	}
}

// contains function validates whether specic string(s) are contained in a slice
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
