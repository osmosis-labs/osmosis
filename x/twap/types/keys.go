package types

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
)

const (
	ModuleName = "twap"

	StoreKey          = ModuleName
	TransientStoreKey = "transient_" + ModuleName // this is silly we have to do this
	RouterKey         = ModuleName

	QuerierRoute = ModuleName
	// Contract: Coin denoms cannot contain this character
	KeySeparator = "|"
)

var (
	PruningStateKey                    = []byte{0x01}
	mostRecentTWAPsNoSeparator         = "recent_twap"
	historicalTWAPTimeIndexNoSeparator = "historical_time_index"
	historicalTWAPPoolIndexNoSeparator = "historical_pool_index"

	// We do key management to let us easily meet the goals of (AKA minimal iteration):
	// * Get most recent twap for a (pool id, asset 1, asset 2) with no iteration
	// * Get all records for all pools, within a given time range
	// * Get all records for a (pool id, asset 1, asset 2), within a given time range

	// format is just pool id | denom1 | denom2
	// made for getting most recent key
	mostRecentTWAPsPrefix = mostRecentTWAPsNoSeparator + KeySeparator
	// format is time | pool id | denom1 | denom2
	// made for efficiently deleting records by time in pruning
	HistoricalTWAPTimeIndexPrefix = historicalTWAPTimeIndexNoSeparator + KeySeparator
	// format is pool id | denom1 | denom2 | time
	// made for efficiently getting records given (pool id, denom1, denom2) and time bounds
	HistoricalTWAPPoolIndexPrefix = historicalTWAPPoolIndexNoSeparator + KeySeparator
)

// TODO: make utility command to automatically interlace separators

func FormatKeyPoolTwapRecords(poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%d", HistoricalTWAPPoolIndexPrefix, poolId))
}

func FormatMostRecentTWAPKey(poolId uint64, denom1, denom2 string) []byte {
	poolIdS := osmoutils.FormatFixedLengthU64(poolId)
	return []byte(fmt.Sprintf("%s%s%s%s%s%s", mostRecentTWAPsPrefix, poolIdS, KeySeparator, denom1, KeySeparator, denom2))
}

// TODO: Replace historical management with ORM, we currently accept 2x write amplification right now.
func FormatHistoricalTimeIndexTWAPKey(accumulatorWriteTime time.Time, poolId uint64, denom1, denom2 string) []byte {
	var buffer bytes.Buffer
	timeS := osmoutils.FormatTimeString(accumulatorWriteTime)
	fmt.Fprintf(&buffer, "%s%s%s%d%s%s%s%s", HistoricalTWAPTimeIndexPrefix, timeS, KeySeparator, poolId, KeySeparator, denom1, KeySeparator, denom2)
	return buffer.Bytes()
}

func FormatHistoricalPoolIndexTWAPKey(poolId uint64, denom1, denom2 string, accumulatorWriteTime time.Time) []byte {
	timeS := osmoutils.FormatTimeString(accumulatorWriteTime)
	return FormatHistoricalPoolIndexTWAPKeyFromStrTime(poolId, denom1, denom2, timeS)
}

func FormatHistoricalPoolIndexTWAPKeyFromStrTime(poolId uint64, denom1, denom2 string, accumulatorWriteTimeString string) []byte {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "%s%d%s%s%s%s%s%s", HistoricalTWAPPoolIndexPrefix, poolId, KeySeparator, denom1, KeySeparator, denom2, KeySeparator, accumulatorWriteTimeString)
	return buffer.Bytes()
}

// returns timeString, poolIdString, denom1, denom2, error
// nolint: revive
func ParseFieldsFromHistoricalTimeKey(bz []byte) (string, uint64, string, string, error) {
	split := bytes.Split(bz, []byte(KeySeparator))
	if len(split) != 5 {
		return "", 0, "", "", errors.New("invalid key")
	}
	timeS := string(split[1])
	poolId, err := strconv.Atoi(string(split[2]))
	if err != nil {
		return "", 0, "", "", err
	}
	denom1 := string(split[3])
	denom2 := string(split[4])
	return timeS, uint64(poolId), denom1, denom2, err
}

func FormatHistoricalPoolIndexTimePrefix(poolId uint64, denom1, denom2 string) []byte {
	return []byte(fmt.Sprintf("%s%d%s%s%s%s%s", HistoricalTWAPPoolIndexPrefix, poolId, KeySeparator, denom1, KeySeparator, denom2, KeySeparator))
}

func FormatHistoricalPoolIndexTimeSuffix(poolId uint64, denom1, denom2 string, accumulatorWriteTime time.Time) []byte {
	timeS := osmoutils.FormatTimeString(accumulatorWriteTime)
	// . acts as a suffix for lexicographical orderings
	return []byte(fmt.Sprintf("%s%d%s%s%s%s%s%s.", HistoricalTWAPPoolIndexPrefix, poolId, KeySeparator, denom1, KeySeparator, denom2, KeySeparator, timeS))
}

// GetAllMostRecentTwapsForPool returns all of the most recent twap records for a pool id.
// if the pool id doesn't exist, then this returns a blank list.
func GetAllMostRecentTwapsForPool(store sdk.KVStore, poolId uint64) ([]TwapRecord, error) {
	poolIdS := osmoutils.FormatFixedLengthU64(poolId)
	poolIdPlusOneS := osmoutils.FormatFixedLengthU64(poolId + 1)
	startPrefix := fmt.Sprintf("%s%s%s", mostRecentTWAPsPrefix, poolIdS, KeySeparator)
	endPrefix := fmt.Sprintf("%s%s%s", mostRecentTWAPsPrefix, poolIdPlusOneS, KeySeparator)
	return osmoutils.GatherValuesFromStore(store, []byte(startPrefix), []byte(endPrefix), ParseTwapFromBz)
}

func GetMostRecentTwapForPool(store sdk.KVStore, poolId uint64, denom1, denom2 string) (TwapRecord, error) {
	key := FormatMostRecentTWAPKey(poolId, denom1, denom2)
	bz := store.Get(key)
	return ParseTwapFromBz(bz)
}

func ParseTwapFromBz(bz []byte) (twap TwapRecord, err error) {
	if len(bz) == 0 {
		return TwapRecord{}, errors.New("twap not found")
	}
	err = proto.Unmarshal(bz, &twap)
	if twap.GeometricTwapAccumulator.IsNil() {
		twap.GeometricTwapAccumulator = osmomath.ZeroDec()
	}
	return twap, err
}
