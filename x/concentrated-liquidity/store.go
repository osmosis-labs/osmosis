package concentrated_liquidity

import (
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
)

// getAllPositionsWithVaryingFreezeTimes returns multiple positions indexed by poolId, addr, lowerTick, upperTick with varying freeze times.
func (k Keeper) getAllPositionsWithVaryingFreezeTimes(ctx sdk.Context, poolId uint64, addr sdk.AccAddress, lowerTick, upperTick int64) ([]model.Position, error) {
	return osmoutils.GatherValuesFromStorePrefix(ctx.KVStore(k.storeKey), types.KeyPosition(poolId, addr, lowerTick, upperTick), ParsePositionFromBz)
}

// ParsePositionFromBz parses a position from a byte array.
// Returns a struct containing the liquidity associated with the position.
// Returns an error if the byte array is empty.
// Returns an error if fails to parse.
func ParsePositionFromBz(bz []byte) (position model.Position, err error) {
	if len(bz) == 0 {
		return model.Position{}, errors.New("position not found")
	}
	err = proto.Unmarshal(bz, &position)
	return position, err
}

// ParseFullPositionFromBytes parses a full position from key and value bytes.
// Returns a struct containing the pool id, lower tick, upper tick, join time, freeze duration, and liquidity
// associated with the position.
// Returns an error if the key or value is not found.
// Returns an error if fails to parse either.
func ParseFullPositionFromBytes(key, value []byte) (types.FullPositionByOwnerResult, error) {
	if len(key) == 0 {
		return types.FullPositionByOwnerResult{}, errors.New("key not found")
	}
	if len(value) == 0 {
		return types.FullPositionByOwnerResult{}, fmt.Errorf("value not found for key (%s)", value)
	}

	keyStr := string(key)

	// These may include irrelevant parts of the prefix such as the module prefix
	// and position prefix.
	fullPositionKeyComponents := strings.Split(keyStr, types.KeySeparator)

	if len(fullPositionKeyComponents) < 6 {
		return types.FullPositionByOwnerResult{}, fmt.Errorf(`invalid position key (%s), must have at least 6 components:
	(position prefix, owner address, pool id, lower tick, upper tick, join time, freeze duration),
	all separated by (%s)`, keyStr, types.KeySeparator)
	}

	// We only care about the last 5 components, which are:
	// - pool id
	// - lower tick
	// - upper tick
	// - join time
	// - freeze duration
	relevantPositionKeyComponents := fullPositionKeyComponents[len(fullPositionKeyComponents)-5:]

	positionPrefix := fullPositionKeyComponents[0]
	if positionPrefix != string(types.PositionPrefix) {
		return types.FullPositionByOwnerResult{}, fmt.Errorf("Wrong position prefix, got: %v, required %v", []byte(positionPrefix), types.PositionPrefix)
	}

	poolId, err := strconv.ParseUint(relevantPositionKeyComponents[0], 10, 64)
	if err != nil {
		return types.FullPositionByOwnerResult{}, err
	}

	lowerTick, err := strconv.ParseInt(relevantPositionKeyComponents[1], 10, 64)
	if err != nil {
		return types.FullPositionByOwnerResult{}, err
	}

	upperTick, err := strconv.ParseInt(relevantPositionKeyComponents[2], 10, 64)
	if err != nil {
		return types.FullPositionByOwnerResult{}, err
	}

	joinTime, err := osmoutils.ParseTimeString(relevantPositionKeyComponents[3])
	if err != nil {
		return types.FullPositionByOwnerResult{}, err
	}

	freezeDuration, err := strconv.ParseUint(relevantPositionKeyComponents[4], 10, 64)
	if err != nil {
		return types.FullPositionByOwnerResult{}, err
	}

	positionValue, err := ParsePositionFromBz(value)
	if err != nil {
		return types.FullPositionByOwnerResult{}, err
	}

	return types.FullPositionByOwnerResult{
		PoolId:         poolId,
		LowerTick:      lowerTick,
		UpperTick:      upperTick,
		JoinTime:       joinTime,
		FreezeDuration: time.Duration(freezeDuration),
		Liquidity:      positionValue.Liquidity,
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
		return types.IncentiveRecord{}, errors.New("key not found")
	}
	if len(value) == 0 {
		return types.IncentiveRecord{}, fmt.Errorf("value not found for key (%s)", value)
	}

	keyStr := string(key)

	// These may include irrelevant parts of the prefix such as the module prefix.
	incentiveRecordKeyComponents := strings.Split(keyStr, types.KeySeparator)

	// We only care about the last 3 components, which are:
	// - pool id
	// - incentive denom
	// - min uptime

	relevantIncentiveKeyComponents := incentiveRecordKeyComponents[len(incentiveRecordKeyComponents)-3:]

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

	incentiveBody, err := ParseIncentiveRecordBodyFromBz(value)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	return types.IncentiveRecord{
		PoolId:          poolId,
		IncentiveDenom:  incentiveDenom,
		RemainingAmount: incentiveBody.RemainingAmount,
		EmissionRate:    incentiveBody.EmissionRate,
		StartTime:       incentiveBody.StartTime,
		MinUptime:       time.Duration(minUptime),
	}, nil
}
