package concentrated_pool

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	db "github.com/tendermint/tm-db"

	"github.com/osmosis-labs/osmosis/v12/osmomath"
	"github.com/osmosis-labs/osmosis/v12/osmoutils"
	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

// NextInitializedTick returns the next initialized tick index based on the
// current or provided tick index. If no initialized tick exists, <0, false>
// will be returned. The zeroForOne argument indicates if we need to find the next
// initialized tick to the left or right of the current tick index, where true
// indicates searching to the left.
func (p Pool) NextInitializedTick(ctx sdk.Context, poolTickKVStore sdk.KVStore, poolId uint64, tickIndex int64, zeroForOne bool) (next int64, initialized bool) {
	prefixStore := poolTickKVStore

	var startKey []byte
	if !zeroForOne {
		startKey = types.TickIndexToBytes(tickIndex)
	} else {
		// When looking to the left of the current tick, we need to evaluate the
		// current tick as well. The end cursor for reverse iteration is non-inclusive
		// so must add one and handle overflow.
		startKey = types.TickIndexToBytes(osmomath.Max(tickIndex, tickIndex+1))
	}

	var iter db.Iterator
	if !zeroForOne {
		iter = prefixStore.Iterator(startKey, nil)
	} else {
		iter = prefixStore.ReverseIterator(nil, startKey)
	}

	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		// Since, we constructed our prefix store with <TickPrefix | poolID>, the
		// key is the encoding of a tick index.
		tick, err := types.TickIndexFromBytes(iter.Key())
		if err != nil {
			panic(fmt.Errorf("invalid tick index (%s): %v", string(iter.Key()), err))
		}

		if !zeroForOne && tick > tickIndex {
			return tick, true
		}
		if zeroForOne && tick <= tickIndex {
			return tick, true
		}
	}

	return 0, false
}

func (p Pool) GetTickInfo(ctx sdk.Context, poolTickKVStore sdk.KVStore, poolId uint64, tickIndex int64) (tickInfo TickInfo, err error) {
	tickStruct := TickInfo{}
	key := types.KeyTick(poolId, tickIndex)

	found, err := osmoutils.GetIfFound(poolTickKVStore, key, &tickStruct)
	// return 0 values if key has not been initialized
	if !found {
		return TickInfo{LiquidityGross: sdk.ZeroInt(), LiquidityNet: sdk.ZeroInt()}, err
	}
	if err != nil {
		return tickStruct, err
	}

	return tickStruct, nil
}

func (p Pool) crossTick(ctx sdk.Context, poolTickKVStore sdk.KVStore, poolId uint64, tickIndex int64) (liquidityDelta sdk.Int, err error) {
	tickInfo, err := p.GetTickInfo(ctx, poolTickKVStore, poolId, tickIndex)
	if err != nil {
		return sdk.Int{}, err
	}

	return tickInfo.LiquidityNet, nil
}
