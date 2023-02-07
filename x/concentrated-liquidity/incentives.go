package concentrated_liquidity

import (
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
)

const (
	uptimeAccumPrefix = "uptime"
)

// createUptimeAccumulators creates accumulator objects in store for each supported uptime for the given poolId.
// The accumulators are initialized with the default (zero) values.
func (k Keeper) createUptimeAccumulators(ctx sdk.Context, poolId uint64) error {
	for uptimeIndex := range types.SupportedUptimes {
		err := accum.MakeAccumulator(ctx.KVStore(k.storeKey), getUptimeAccumulatorName(poolId, uint64(uptimeIndex)))
		if err != nil {
			return err
		}
	}

	return nil
}

func getUptimeAccumulatorName(poolId uint64, uptimeIndex uint64) string {
	poolIdStr := strconv.FormatUint(poolId, uintBase)
	uptimeIndexStr := strconv.FormatUint(uptimeIndex, uintBase)
	return strings.Join([]string{uptimeAccumPrefix, poolIdStr, uptimeIndexStr}, "/")
}

// nolint: unused
// getUptimeAccumulators gets the uptime accumulator objects for the given poolId
// Returns error if accumulator for the given poolId does not exist.
func (k Keeper) getUptimeAccumulators(ctx sdk.Context, poolId uint64) ([]accum.AccumulatorObject, error) {
	accums := make([]accum.AccumulatorObject, len(types.SupportedUptimes))
	for uptimeIndex := range types.SupportedUptimes {
		acc, err := accum.GetAccumulator(ctx.KVStore(k.storeKey), getUptimeAccumulatorName(poolId, uint64(uptimeIndex)))
		if err != nil {
			return []accum.AccumulatorObject{}, err
		}

		accums[uptimeIndex] = acc
	}

	return accums, nil
}
