package types

import (
	"errors"
	fmt "fmt"
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
var historicalTWAPsPrefix = "historical_twap" + KeySeparator

// TODO: make utility command to automatically interlace separators

func FormatMostRecentTWAPKey(poolId uint64, denom1 string, denom2 string) []byte {
	return []byte(fmt.Sprintf("%s%s%d%s%s%s%s", mostRecentTWAPsPrefix, KeySeparator, poolId, KeySeparator, denom1, KeySeparator, denom2))
}

func FormatHistoricalTWAPKey(poolId uint64, accumulatorWriteTime time.Time, denom1 string, denom2 string) []byte {
	return []byte(fmt.Sprintf("%s%s%d%s%s%s%s%s%s", historicalTWAPsPrefix, KeySeparator, poolId, KeySeparator, accumulatorWriteTime, KeySeparator, denom1, KeySeparator, denom2))
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
