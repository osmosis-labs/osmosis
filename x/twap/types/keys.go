package types

import (
	"errors"
	fmt "fmt"
	"strings"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	"github.com/osmosis-labs/osmosis/v12/osmoutils"
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
	mostRecentTWAPsNoSeparator         = "recent_twap"
	historicalTWAPTimeIndexNoSeparator = "historical_time_index"
	historicalTWAPPoolIndexNoSeparator = "historical_pool_index"
	expectedKeySeparatedParts          = 5

	mostRecentTWAPsPrefix = mostRecentTWAPsNoSeparator + KeySeparator
	// keySeparatorPlusOne is used for creating prefixes for the key end in iterators
	// when we want to get all of the keys in a prefix. Since it is one byte larger
	// than the original key separator and the end prefix is exclusive, it is valid
	// for getting all values under the original key separator.
	HistoricalTWAPTimeIndexPrefix = historicalTWAPTimeIndexNoSeparator + KeySeparator
	HistoricalTWAPPoolIndexPrefix = historicalTWAPPoolIndexNoSeparator + KeySeparator
)

// TODO: make utility command to automatically interlace separators

func FormatMostRecentTWAPKey(poolId uint64, denom1, denom2 string) []byte {
	return []byte(fmt.Sprintf("%s%d%s%s%s%s", mostRecentTWAPsPrefix, poolId, KeySeparator, denom1, KeySeparator, denom2))
}

// TODO: Replace historical management with ORM, we currently accept 2x write amplification right now.
func FormatHistoricalTimeIndexTWAPKey(accumulatorWriteTime time.Time, poolId uint64, denom1, denom2 string) []byte {
	timeS := osmoutils.FormatTimeString(accumulatorWriteTime)
	return []byte(fmt.Sprintf("%s%s%s%d%s%s%s%s", HistoricalTWAPTimeIndexPrefix, timeS, KeySeparator, poolId, KeySeparator, denom1, KeySeparator, denom2))
}

func FormatHistoricalPoolIndexTWAPKey(poolId uint64, accumulatorWriteTime time.Time, denom1, denom2 string) []byte {
	timeS := osmoutils.FormatTimeString(accumulatorWriteTime)
	return []byte(fmt.Sprintf("%s%d%s%s%s%s%s%s", HistoricalTWAPPoolIndexPrefix, poolId, KeySeparator, timeS, KeySeparator, denom1, KeySeparator, denom2))
}

func FormatHistoricalPoolIndexTimePrefix(poolId uint64, accumulatorWriteTime time.Time) []byte {
	timeS := osmoutils.FormatTimeString(accumulatorWriteTime)
	return []byte(fmt.Sprintf("%s%d%s%s%s", HistoricalTWAPPoolIndexPrefix, poolId, KeySeparator, timeS, KeySeparator))
}

func ParseTimeFromHistoricalTimeIndexKey(key []byte) (time.Time, error) {
	keyS := string(key)
	s := strings.Split(keyS, KeySeparator)
	if len(s) != expectedKeySeparatedParts {
		return time.Time{}, KeySeparatorLengthError{ExpectedLength: expectedKeySeparatedParts, ActualLength: len(s)}
	}
	if s[0] != historicalTWAPTimeIndexNoSeparator {
		return time.Time{}, UnexpectedSeparatorError{ExpectedSeparator: historicalTWAPPoolIndexNoSeparator, ActualSeparator: s[0]}
	}
	t, err := osmoutils.ParseTimeString(s[1])
	if err != nil {
		return time.Time{}, TimeStringKeyFormatError{Key: keyS, Err: err}
	}
	return t, nil
}

func ParseTimeFromHistoricalPoolIndexKey(key []byte) (time.Time, error) {
	keyS := string(key)
	s := strings.Split(keyS, KeySeparator)
	if len(s) != expectedKeySeparatedParts {
		return time.Time{}, KeySeparatorLengthError{ExpectedLength: expectedKeySeparatedParts, ActualLength: len(s)}
	}
	if s[0] != historicalTWAPPoolIndexNoSeparator {
		return time.Time{}, UnexpectedSeparatorError{ExpectedSeparator: historicalTWAPPoolIndexNoSeparator, ActualSeparator: s[0]}
	}
	// Time is always the third item in correctly formatted pool index keys (as opposed to the second item in time index keys)
	t, err := osmoutils.ParseTimeString(s[2])
	if err != nil {
		return time.Time{}, TimeStringKeyFormatError{Key: keyS, Err: err}
	}
	return t, nil
}

// GetAllMostRecentTwapsForPool returns all of the most recent twap records for a pool id.
// if the pool id doesn't exist, then this returns a blank list.
func GetAllMostRecentTwapsForPool(store sdk.KVStore, poolId uint64) ([]TwapRecord, error) {
	startPrefix := fmt.Sprintf("%s%d%s", mostRecentTWAPsPrefix, poolId, KeySeparator)
	endPrefix := fmt.Sprintf("%s%d%s", mostRecentTWAPsPrefix, poolId+1, KeySeparator)
	return osmoutils.GatherValuesFromStore(store, []byte(startPrefix), []byte(endPrefix), ParseTwapFromBz)
}

func ParseTwapFromBz(bz []byte) (twap TwapRecord, err error) {
	if len(bz) == 0 {
		return TwapRecord{}, errors.New("twap not found")
	}
	err = proto.Unmarshal(bz, &twap)
	return twap, err
}
