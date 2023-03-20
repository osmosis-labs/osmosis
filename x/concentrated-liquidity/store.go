package concentrated_liquidity

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types/genesis"
)

const (
	positionPrefixNumComponents = 8
	uint64Bytes                 = 8
)

// getAllPositionsWithVaryingFreezeTimes returns multiple positions indexed by poolId, addr, lowerTick, upperTick with varying freeze times.
func (k Keeper) getAllPositionsWithVaryingFreezeTimes(ctx sdk.Context, poolId uint64, addr sdk.AccAddress, lowerTick, upperTick int64) ([]sdk.Dec, error) {
	return osmoutils.GatherValuesFromStorePrefix(ctx.KVStore(k.storeKey), types.KeyPosition(poolId, addr, lowerTick, upperTick), ParseLiquidityFromBz)
}

// getAllPositions gets all CL positions for export genesis.
func (k Keeper) getAllPositions(ctx sdk.Context) ([]model.Position, error) {
	return osmoutils.GatherValuesFromStorePrefixWithKeyParser(ctx.KVStore(k.storeKey), types.PositionPrefix, ParseFullPositionFromBytes)
}

// ParseLiquidityFromBz parses and returns a position's liquidity from a byte array.
// Returns an error if the byte array is empty.
// Returns an error if fails to parse.
func ParseLiquidityFromBz(bz []byte) (sdk.Dec, error) {
	if len(bz) == 0 {
		return sdk.Dec{}, errors.New("position not found")
	}
	liquidityStruct := &sdk.DecProto{}
	err := proto.Unmarshal(bz, liquidityStruct)
	return liquidityStruct.Dec, err
}

// ParseTickFromBz takes a byte slice representing the serialized tick data and
// attempts to parse it into a TickInfo struct using the protobuf Unmarshal function.
// If the byte slice is empty or the unmarshalling fails, an appropriate error is returned.
//
// Parameters:
// - bz ([]byte): A byte slice representing the serialized tick data.
//
// Returns:
// - model.TickInfo: A struct containing the parsed tick information.
// - error: An error if the byte slice is empty or if the unmarshalling fails.
func ParseTickFromBz(bz []byte) (tick model.TickInfo, err error) {
	if len(bz) == 0 {
		return model.TickInfo{}, errors.New("tick not found")
	}
	err = proto.Unmarshal(bz, &tick)
	return tick, err
}

// ParseFullTickFromBytes takes key and value byte slices and attempts to parse
// them into a FullTick struct. If the key or value is not valid, an appropriate
// error is returned. The function expects the key to have three components
// 1. The tick prefix (1 byte)
// 2. The pool id (8 bytes)
// 3. The tick index (1 byte for sign + 8 bytes for unsigned integer)
//
// The function returns a FullTick struct containing the pool id, tick index, and
// tick information.
//
// Parameters:
// - key ([]byte): A byte slice representing the key.
// - value ([]byte): A byte slice representing the value.
//
// Returns:
// - genesis.FullTick: A struct containing the parsed pool id, tick index, and tick information.
// - error: An error if the key or value is not valid or if the parsing fails.
func ParseFullTickFromBytes(key, value []byte) (tick genesis.FullTick, err error) {
	if len(key) == 0 {
		return genesis.FullTick{}, types.ErrKeyNotFound
	}
	if len(value) == 0 {
		return genesis.FullTick{}, types.ValueNotFoundForKeyError{Key: key}
	}

	if len(key) != types.TickKeyLengthBytes {
		return genesis.FullTick{}, types.InvalidTickKeyByteLengthError{Length: len(key)}
	}

	prefix := key[0:len(types.TickPrefix)]
	if !bytes.Equal(types.TickPrefix, prefix) {
		return genesis.FullTick{}, types.InvalidPrefixError{Actual: string(prefix), Expected: string(types.TickPrefix)}
	}

	key = key[len(types.TickPrefix):]

	// We only care about the last 2 components, which are:
	// - pool id
	// - tick index
	poolIdBytes := key[0:uint64Bytes]
	poolId := sdk.BigEndianToUint64(poolIdBytes)

	key = key[uint64Bytes:]

	tickIndex, err := types.TickIndexFromBytes(key)
	if err != nil {
		return genesis.FullTick{}, err
	}

	tickValue, err := ParseTickFromBz(value)
	if err != nil {
		return genesis.FullTick{}, types.ValueParseError{Wrapped: err}
	}

	return genesis.FullTick{
		PoolId:    poolId,
		TickIndex: tickIndex,
		Info:      tickValue,
	}, nil
}

// ParseFullPositionFromBytes parses a full position from key and value bytes.
// Returns a struct containing the pool id, lower tick, upper tick, join time, freeze duration, and liquidity
// associated with the position.
// Returns an error if the key or value is not found.
// Returns an error if fails to parse either.
func ParseFullPositionFromBytes(key, value []byte) (model.Position, error) {
	if len(key) == 0 {
		return model.Position{}, types.ErrKeyNotFound
	}
	if len(value) == 0 {
		return model.Position{}, types.ValueNotFoundForKeyError{Key: key}
	}

	keyStr := string(key)

	// These may include irrelevant parts of the prefix such as the module prefix
	// and position prefix.
	fullPositionKeyComponents := strings.Split(keyStr, types.KeySeparator)

	if len(fullPositionKeyComponents) < positionPrefixNumComponents {
		return model.Position{}, types.InvalidKeyComponentError{
			KeyStr:                keyStr,
			KeySeparator:          types.KeySeparator,
			NumComponentsExpected: positionPrefixNumComponents,
			ComponentsExpectedStr: "position prefix, owner address, pool id, lower tick, upper tick, join time, freeze duration, position id",
		}
	}

	prefix := fullPositionKeyComponents[0]
	if strings.Compare(prefix, string(types.PositionPrefix)) != 0 {
		return model.Position{}, types.InvalidPrefixError{Actual: prefix, Expected: string(types.PositionPrefix)}
	}

	// We only care about the last 6 components, which are:
	// - owner address
	// - pool id
	// - lower tick
	// - upper tick
	// - join time
	// - freeze duration
	address, err := sdk.AccAddressFromHex(fullPositionKeyComponents[1])
	if err != nil {
		return model.Position{}, err
	}

	poolId, err := strconv.ParseUint(fullPositionKeyComponents[2], 10, 64)
	if err != nil {
		return model.Position{}, err
	}

	lowerTick, err := strconv.ParseInt(fullPositionKeyComponents[3], 10, 64)
	if err != nil {
		return model.Position{}, err
	}

	upperTick, err := strconv.ParseInt(fullPositionKeyComponents[4], 10, 64)
	if err != nil {
		return model.Position{}, err
	}

	joinTime, err := osmoutils.ParseTimeString(fullPositionKeyComponents[5])
	if err != nil {
		return model.Position{}, err
	}

	freezeDuration, err := strconv.ParseUint(fullPositionKeyComponents[6], 10, 64)
	if err != nil {
		return model.Position{}, err
	}

	positionId, err := strconv.ParseUint(fullPositionKeyComponents[7], 10, 64)
	if err != nil {
		return model.Position{}, err
	}

	liquidity, err := ParseLiquidityFromBz(value)
	if err != nil {
		return model.Position{}, types.ValueParseError{Wrapped: err}
	}

	return model.Position{
		PositionId:     positionId,
		Address:        address.String(),
		PoolId:         poolId,
		LowerTick:      lowerTick,
		UpperTick:      upperTick,
		Liquidity:      liquidity,
		JoinTime:       joinTime,
		FreezeDuration: time.Duration(freezeDuration),
	}, nil
}

// ParseIncentiveRecordBodyFromBz parses an IncentiveRecord from a byte array.
// Returns a struct containing the denom and min uptime associated with the incentive record.
// Returns an error if the byte array is empty.
// Returns an error if fails to parse.
func ParseIncentiveRecordBodyFromBz(bz []byte) (incentiveRecordBody types.IncentiveRecordBody, err error) {
	if len(bz) == 0 {
		return types.IncentiveRecordBody{}, errors.New("incentive record not found")
	}
	err = proto.Unmarshal(bz, &incentiveRecordBody)
	if err != nil {
		return types.IncentiveRecordBody{}, err
	}

	return incentiveRecordBody, nil
}

// ParseFullIncentiveRecordFromBz parses an incentive record from a byte array.
// Returns a struct containing the state associated with the incentive.
// Returns an error if the byte array is empty.
// Returns an error if fails to parse.
func ParseFullIncentiveRecordFromBz(key []byte, value []byte) (incentiveRecord types.IncentiveRecord, err error) {
	if len(key) == 0 {
		return types.IncentiveRecord{}, types.ErrKeyNotFound
	}
	if len(value) == 0 {
		return types.IncentiveRecord{}, types.ValueNotFoundForKeyError{Key: key}
	}

	keyStr := string(key)

	// These may include irrelevant parts of the prefix such as the module prefix.
	incentiveRecordKeyComponents := strings.Split(keyStr, types.KeySeparator)

	// We only care about the last 4 components, which are:
	// - pool id
	// - incentive denom
	// - min uptime
	// - incentive creator

	relevantIncentiveKeyComponents := incentiveRecordKeyComponents[len(incentiveRecordKeyComponents)-4:]

	incentivePrefix := incentiveRecordKeyComponents[0]
	if incentivePrefix != string(types.IncentivePrefix) {
		return types.IncentiveRecord{}, fmt.Errorf("Wrong incentive prefix, got: %v, required %v", []byte(incentivePrefix), types.IncentivePrefix)
	}

	poolId, err := strconv.ParseUint(relevantIncentiveKeyComponents[0], 10, 64)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	incentiveDenom := relevantIncentiveKeyComponents[1]

	minUptime, err := strconv.ParseUint(relevantIncentiveKeyComponents[2], 10, 64)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	// Note that we skip the first byte since we prefix addresses by length in key
	incentiveCreator := sdk.AccAddress(relevantIncentiveKeyComponents[3][1:])
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	incentiveBody, err := ParseIncentiveRecordBodyFromBz(value)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	return types.IncentiveRecord{
		PoolId:           poolId,
		IncentiveDenom:   incentiveDenom,
		IncentiveCreator: incentiveCreator,
		RemainingAmount:  incentiveBody.RemainingAmount,
		EmissionRate:     incentiveBody.EmissionRate,
		StartTime:        incentiveBody.StartTime,
		MinUptime:        time.Duration(minUptime),
	}, nil
}
