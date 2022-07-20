package types

import (
	"errors"
	fmt "fmt"
	"strings"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	"github.com/osmosis-labs/osmosis/v10/osmoutils"
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

var AlteredPoolIdsPrefix = []byte{0}

var mostRecentTWAPsPrefix = "recent_twap" + KeySeparator
var historicalTWAPTimeIndexPrefix = "historical_time_index" + KeySeparator
var historicalTWAPPoolIndexPrefix = "historical_pool_index" + KeySeparator

// TODO: make utility command to automatically interlace separators

func FormatMostRecentTWAPKey(poolId uint64, denom1 string, denom2 string) []byte {
	return []byte(fmt.Sprintf("%s%s%d%s%s%s%s", mostRecentTWAPsPrefix, KeySeparator, poolId, KeySeparator, denom1, KeySeparator, denom2))
}

// TODO: Replace historical management with ORM, we currently accept 2x write amplification right now.
func FormatHistoricalTimeIndexTWAPKey(accumulatorWriteTime time.Time, poolId uint64, denom1 string, denom2 string) []byte {
	timeS := osmoutils.FormatTimeString(accumulatorWriteTime)
	return []byte(fmt.Sprintf("%s%s%s%s%d%s%s%s%s", historicalTWAPTimeIndexPrefix, KeySeparator, timeS, KeySeparator, poolId, KeySeparator, denom1, KeySeparator, denom2))
}

func FormatHistoricalPoolIndexTWAPKey(poolId uint64, accumulatorWriteTime time.Time, denom1 string, denom2 string) []byte {
	timeS := osmoutils.FormatTimeString(accumulatorWriteTime)
	return []byte(fmt.Sprintf("%s%s%d%s%s%s%s%s%s", historicalTWAPPoolIndexPrefix, KeySeparator, poolId, KeySeparator, timeS, KeySeparator, denom1, KeySeparator, denom2))
}

func FormatHistoricalPoolIndexTimePrefix(poolId uint64, accumulatorWriteTime time.Time) []byte {
	timeS := osmoutils.FormatTimeString(accumulatorWriteTime)
	return []byte(fmt.Sprintf("%s%s%d%s%s%s", historicalTWAPPoolIndexPrefix, KeySeparator, poolId, KeySeparator, timeS, KeySeparator))
}

func ParseTimeFromHistoricalTimeIndexKey(key []byte) time.Time {
	keyS := string(key)
	s := strings.Split(keyS, KeySeparator)
	if len(s) != 5 || s[0] != historicalTWAPTimeIndexPrefix {
		panic("Called ParseTimeFromHistoricalTimeIndexKey on incorrectly formatted key")
	}
	t, err := osmoutils.ParseTimeString(s[1])
	if err != nil {
		panic("incorrectly formatted time string in key")
	}
	return t
}

func ParseTimeFromHistoricalPoolIndexKey(key []byte) time.Time {
	keyS := string(key)
	s := strings.Split(keyS, KeySeparator)
	if len(s) != 5 || s[0] != historicalTWAPPoolIndexPrefix {
		panic("Called ParseTimeFromHistoricalPoolIndexKey on incorrectly formatted key")
	}
	t, err := osmoutils.ParseTimeString(s[2])
	if err != nil {
		panic("incorrectly formatted time string in key")
	}
	return t
}

func GetAllMostRecentTwapsForPool(store sdk.KVStore, poolId uint64) ([]TwapRecord, error) {
	startPrefix := fmt.Sprintf("%s%s%d%s", mostRecentTWAPsPrefix, KeySeparator, poolId, KeySeparator)
	endPrefix := fmt.Sprintf("%s%s%d%s", mostRecentTWAPsPrefix, KeySeparator, poolId+1, KeySeparator)
	return osmoutils.GatherValuesFromStore(store, []byte(startPrefix), []byte(endPrefix), ParseTwapFromBz)
}

func ParseTwapFromBz(bz []byte) (twap TwapRecord, err error) {
	if len(bz) > 0 {
		return TwapRecord{}, errors.New("twap not found")
	}
	err = proto.Unmarshal(bz, &twap)
	return twap, err
}
