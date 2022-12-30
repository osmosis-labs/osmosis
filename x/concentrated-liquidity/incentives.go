package concentrated_liquidity

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/sumtree"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/model"
)

// Initializes incentives accumulator
// Also initializes sumtree to track jointimes
// Intended to only have 1-2 per pool (one for intenral/OSMO incentives, one for external ones)
// Each incentive token, when added, should specify emission rate per second and uptime requirement for claiming
//
// Should test this to ensure we can get all pools before specific jointime etc. from sumtree
// Also means this should be run at pool creation so sumtree is properly populated
func (k Keeper) initIncentivesForPool(ctx sdk.Context, poolID uint64) {
	// Create internal accumulator & store in pool's IntIncAccum field
	
	// Create external accumulator & store in pool's ExtIncAccum field
	
	// Create sumtree & initialize.
	// 
	// Since running functions on non-existant sumtrees is actually well-defined behavior,
	// this initialization process might not even be necessary and can potentially be abstracted
	// into a separate `accumulationStore` function as in lockup module.
	//
	// Note that this tree can be (and should be) pulled on the fly using
	// an equivalent of lockup module's `accumulationStore` function instead
	// of trying to store it statically in state (since it rebalances).
	//
	// Reference: lock.go's impl w/ new store using custom prefix + key=10
	
	// Every position update should update this tree (e.g. regenerate tree & Increase() or Decrease() by amt on the relevant key)
	// If we are doing by jointime, this might mean decrease node @ key = original jointime by full position amount and increase node @
	// key = curBlockTime by full position amount
}

// Creates an incentive for the passed in denom in either the internal or external accumulator
// TODO: figure out how to handle attack where someone dusts with low rate for a denom & blocks that
// denom from being used as an incentive at a higher/lower rate
//
// Could potentially allow for multiple incentives of same denom type using incID if they are in diff accums
/* func createIncentive(incID, denom, isInternal, rate, uptime_requirement, duration, start_time, initAmount) { should scale gas w/ num tokens } */

// loads up denom's incentive bucket with amount
/* func addToIncentive(incID, pool, denom, amount) */

// TODO: move to types folder
var KeyPrefixJoinTimeAccumulation = []byte("join-time-sumtree")
var KeyPrefixTimestamp = []byte("timestamp")

func (k Keeper) accumulationStore(ctx sdk.Context, poolID uint64) sumtree.Tree {
	return sumtree.NewTree(prefix.NewStore(ctx.KVStore(k.storeKey), accumulationStorePrefix(poolID)), 10)
}

// Internal fn to generate store prefixes for pool sumtree store
// TODO: move to store.go file along with other store helpers
func accumulationStorePrefix(poolID uint64) (res []byte) {
	// Does it make sense to take len(string(poolID)) here to represent capacity in bytes?
	capacity := len(KeyPrefixJoinTimeAccumulation) + len(fmt.Sprint(poolID)) + 1
	res = make([]byte, len(KeyPrefixJoinTimeAccumulation), capacity)
	copy(res, KeyPrefixJoinTimeAccumulation)
	res = append(res, []byte("pool"+fmt.Sprint(poolID)+"/")...)
	return
}

// Copied from lockup's `func getTimeKey`
func accumulationTimeKey(timestamp time.Time) (res []byte) {
	timeBz := sdk.FormatTimeBytes(timestamp)
	timeBzL := len(timeBz)
	prefixL := len(KeyPrefixTimestamp)

	bz := make([]byte, prefixL+8+timeBzL)

	// copy the prefix
	copy(bz[:prefixL], KeyPrefixTimestamp)

	// copy the encoded time bytes length
	copy(bz[prefixL:prefixL+8], sdk.Uint64ToBigEndian(uint64(timeBzL)))

	// copy the encoded time bytes
	copy(bz[prefixL+8:prefixL+8+timeBzL], timeBz)
	return bz
}

func (k Keeper) updateLiquidityTree(ctx sdk.Context, poolId uint64, position *model.Position, liquidityBefore sdk.Dec, liquidityAfter sdk.Dec) {
	// Sumtree updates
	// Clear old place in position tree
	// TODO: move this into UpdateLiquidityTree helper
	k.accumulationStore(ctx, poolId).Decrease(accumulationTimeKey(position.JoinTime), liquidityBefore)

	// Update position's JoinTime to = curBlocktime
	// convert to int64 (gives seconds) then use with time.Unix(...) that converts back to Time
	// also consider duration
	position.JoinTime = ctx.BlockTime()

	// Add new position time to position tree
	k.accumulationStore(ctx, poolId).Decrease(accumulationTimeKey(position.JoinTime), liquidityAfter)
}
