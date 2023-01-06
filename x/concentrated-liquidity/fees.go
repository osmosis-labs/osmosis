package concentrated_liquidity

import (
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
)

const (
	feeAccumPrefix        = "fee"
	feeAccumNameSeparator = "/"
	uintBase              = 10
)

// createFeeAccumulator creates an accumulator object in the store using the given poolId
func (k Keeper) createFeeAccumulator(ctx sdk.Context, poolId uint64) error {
	err := accum.MakeAccumulator(ctx.KVStore(k.storeKey), getFeeAccumulatorName(poolId))
	if err != nil {
		return err
	}
	return nil
}

// nolint: unused
// getFeeAccumulator gets the fee accumulator object using the given poolOd
// returns error if accumulator for the given poolId does not exist.
func (k Keeper) getFeeAccumulator(ctx sdk.Context, poolId uint64) (accum.AccumulatorObject, error) {
	acc, err := accum.GetAccumulator(ctx.KVStore(k.storeKey), getFeeAccumulatorName(poolId))
	if err != nil {
		return accum.AccumulatorObject{}, err
	}

	return acc, nil
}

// getFeeAccumulatorName ƒormats the given poolID and returns the fee accumulator name
func getFeeAccumulatorName(poolId uint64) string {
	poolIdStr := strconv.FormatUint(poolId, uintBase)
	return strings.Join([]string{feeAccumPrefix, poolIdStr}, "/")
}
