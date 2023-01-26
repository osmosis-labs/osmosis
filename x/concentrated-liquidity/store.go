package concentrated_liquidity

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/gogo/protobuf/proto"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
)

// // getAllPositionsWithVaryingFreezeTimes returns multiple positions indexed by poolId, addr, lowerTick, upperTick with varying freeze times.
// func (k Keeper) getAllPositionsWithVaryingFreezeTimes(ctx sdk.Context, poolId uint64, addr sdk.AccAddress, lowerTick, upperTick int64) ([]model.Position, error) {
// 	return osmoutils.GatherValuesFromStorePrefix(ctx.KVStore(k.storeKey), types.PositionPrefix, ParsePositionFromBz)
// }

// getAllPositionsWithVaryingFreezeTimes returns multiple positions indexed by poolId, addr, lowerTick, upperTick with varying freeze times.
func (k Keeper) getAllPositionsWithVaryingFreezeTimes(ctx sdk.Context, poolId uint64, addr sdk.AccAddress, lowerTick, upperTick int64) ([]model.Position, error) {
	var key []byte
	// key = append(key, types.PositionPrefix...)
	// key = append(key, sdk.Uint64ToBigEndian(poolId)...)
	//key = []byte(fmt.Sprintf("%s%s%s%s%d%s%d%s%d%s%s", types.PositionPrefix, types.KeySeparator, addrKey, KeySeparator, poolId, KeySeparator, lowerTick, KeySeparator, upperTick, KeySeparator, frozenUntilKey))
	addrKey := address.MustLengthPrefix(addr.Bytes())
	//key = []byte(fmt.Sprintf("%s%s", types.PositionPrefix, types.KeySeparator))
	key = []byte(fmt.Sprintf("%s%s%s%s", types.PositionPrefix, types.KeySeparator, addrKey, types.KeySeparator))
	// positions, err := osmoutils.GatherValuesFromStorePrefix(ctx.KVStore(k.storeKey), key, ParsePositionFromBz)
	// if err != nil {
	// 	return nil, err
	// }
	// var filteredPositions []model.Position
	// for _, position := range positions {
	// 	if position.Owner.Equals(addr) && position.LowerTick == lowerTick && position.UpperTick == upperTick {
	// 		filteredPositions = append(filteredPositions, position)
	// 	}
	// }
	return osmoutils.GatherValuesFromStorePrefix(ctx.KVStore(k.storeKey), key, ParsePositionFromBz)
}

func ParsePositionFromBz(bz []byte) (position model.Position, err error) {
	fmt.Printf("bz: %v \n", bz)
	if len(bz) == 0 {
		return model.Position{}, errors.New("position not found")
	}
	err = proto.Unmarshal(bz, &position)
	return position, err
}
