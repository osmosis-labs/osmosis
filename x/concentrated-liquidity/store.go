package concentrated_liquidity

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
)

const (
	positionPrefixNumComponents = 8
	uint64Bytes                 = 8
)

// getAllPositions gets all CL positions for export genesis.
func (k Keeper) getAllPositions(ctx sdk.Context) ([]model.Position, error) {
	return osmoutils.GatherValuesFromStorePrefix(
		ctx.KVStore(k.storeKey), types.PositionIdPrefix, ParsePositionFromBz)
}

// ParsePositionIdFromBz parses and returns a position's id from a byte array.
// Returns an error if the byte array is empty.
// Returns an error if fails to parse.
func ParsePositionIdFromBz(bz []byte) (uint64, error) {
	if len(bz) == 0 {
		return 0, errors.New("position not found when parsing position id")
	}
	return sdk.BigEndianToUint64(bz), nil
}

// ParsePositionFromBz parses and returns a position from a byte array.
// Returns an error if the byte slice is empty.
// Returns an error if fails to unmarshal.
func ParsePositionFromBz(value []byte) (model.Position, error) {
	position := model.Position{}
	err := proto.Unmarshal(value, &position)
	if err != nil {
		return model.Position{}, err
	}
	return position, nil
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
	if err != nil {
		return model.TickInfo{}, err
	}

	return tick, err
}

// ParseTickFromBzAndRemoveUnInitializedUptimeTrackers is an extension of ParseTickFromBz.
// Due to how proto unmarshals empty lists or fields with nil values, we manually iterate over the uptime trackers list
// and remove instances of empty fields.
func ParseTickFromBzAndRemoveUnInitializedUptimeTrackers(bz []byte) (model.TickInfo, error) {
	tick, err := ParseTickFromBz(bz)
	if err != nil {
		return model.TickInfo{}, nil
	}

	newTickInfo := model.TickInfo{}
	newTickInfo.LiquidityGross = tick.LiquidityGross
	newTickInfo.LiquidityNet = tick.LiquidityNet
	newTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal = tick.SpreadRewardGrowthOppositeDirectionOfLastTraversal
	for _, list := range tick.UptimeTrackers.List {
		if list.UptimeGrowthOutside.Len() != 0 {
			newTickInfo.UptimeTrackers = tick.UptimeTrackers
		}
	}

	return newTickInfo, nil
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
	// - min uptime
	// - incentive id

	relevantIncentiveKeyComponents := incentiveRecordKeyComponents[len(incentiveRecordKeyComponents)-4:]

	incentivePrefix := incentiveRecordKeyComponents[0]
	if incentivePrefix != string(types.IncentivePrefix) {
		return types.IncentiveRecord{}, fmt.Errorf("Wrong incentive prefix, got: %v, required %v", []byte(incentivePrefix), types.IncentivePrefix)
	}

	poolId, err := strconv.ParseUint(relevantIncentiveKeyComponents[0], 10, 64)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	minUptimeIndex, err := strconv.ParseUint(relevantIncentiveKeyComponents[1], 10, 64)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	incentiveBody, err := ParseIncentiveRecordBodyFromBz(value)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	incentiveRecordId, err := strconv.ParseUint(relevantIncentiveKeyComponents[2], 10, 64)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	incentiveRecordBody := types.IncentiveRecordBody{
		RemainingCoin: incentiveBody.RemainingCoin,
		EmissionRate:  incentiveBody.EmissionRate,
		StartTime:     incentiveBody.StartTime,
	}

	return types.IncentiveRecord{
		PoolId:              poolId,
		IncentiveRecordBody: incentiveRecordBody,
		MinUptime:           types.SupportedUptimes[minUptimeIndex],
		IncentiveId:         incentiveRecordId,
	}, nil
}
