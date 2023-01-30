package concentrated_liquidity

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
)

// getAllPositionsWithVaryingFreezeTimes returns multiple positions indexed by poolId, addr, lowerTick, upperTick with varying freeze times.
func (k Keeper) getAllPositionsWithVaryingFreezeTimes(ctx sdk.Context, poolId uint64, addr sdk.AccAddress, lowerTick, upperTick int64) ([]types.Position, error) {
	return osmoutils.GatherValuesFromStorePrefix(ctx.KVStore(k.storeKey), types.KeyPosition(poolId, addr, lowerTick, upperTick), ParsePositionFromBz)
}

func ParsePositionFromBz(bz []byte) (position types.Position, err error) {
	if len(bz) == 0 {
		return types.Position{}, errors.New("position not found")
	}
	err = proto.Unmarshal(bz, &position)
	return position, err
}
