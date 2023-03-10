package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/osmosis-labs/osmosis/osmoutils"
)

const (
	ModuleName = "concentratedliquidity"
	RouterKey  = ModuleName

	StoreKey     = ModuleName
	KeySeparator = "|"
)

// Key prefixes
var (
	TickPrefix      = []byte{0x01}
	PositionPrefix  = []byte{0x02}
	PoolPrefix      = []byte{0x03}
	IncentivePrefix = []byte{0x04}
)

// TickIndexToBytes converts a tick index to a byte slice. Negative tick indexes
// are prefixed with 0x00 a byte and positive tick indexes are prefixed with a
// 0x01 byte. We do this because big endian byte encoding does not give us in
// order iteration in state due to the tick index values being signed integers.
func TickIndexToBytes(tickIndex int64) []byte {
	key := make([]byte, 9)
	if tickIndex < 0 {
		copy(key[1:], sdk.Uint64ToBigEndian(uint64(tickIndex)))
	} else {
		copy(key[:1], []byte{0x01})
		copy(key[1:], sdk.Uint64ToBigEndian(uint64(tickIndex)))
	}

	return key
}

// TickIndexFromBytes converts an encoded tick index to an int64 value. It returns
// an error if the encoded tick has invalid length.
func TickIndexFromBytes(bz []byte) (int64, error) {
	if len(bz) != 9 {
		return 0, fmt.Errorf("invalid encoded tick index length; expected: 9, got: %d", len(bz))
	}

	return int64(sdk.BigEndianToUint64(bz[1:])), nil
}

// KeyTick returns a key for storing a TickInfo object.
func KeyTick(poolId uint64, tickIndex int64) []byte {
	key := KeyTickPrefix(poolId)
	key = append(key, TickIndexToBytes(tickIndex)...)
	return key
}

// KeyTickPrefix constructs a key prefix for storing a TickInfo object.
func KeyTickPrefix(poolId uint64) []byte {
	var key []byte
	key = append(key, TickPrefix...)
	key = append(key, sdk.Uint64ToBigEndian(poolId)...)
	return key
}

// KeyFullPosition uses pool Id, owner, lower tick, upper tick, joinTime and freezeDuration for keys
func KeyFullPosition(poolId uint64, addr sdk.AccAddress, lowerTick, upperTick int64, joinTime time.Time, freezeDuration time.Duration) []byte {
	joinTimeKey := osmoutils.FormatTimeString(joinTime)
	addrKey := address.MustLengthPrefix(addr.Bytes())
	return []byte(fmt.Sprintf("%s%s%s%s%d%s%d%s%d%s%s%s%d", PositionPrefix, KeySeparator, addrKey, KeySeparator, poolId, KeySeparator, lowerTick, KeySeparator, upperTick, KeySeparator, joinTimeKey, KeySeparator, uint64(freezeDuration)))
}

// KeyPosition uses pool Id, owner, lower tick and upper tick for keys
func KeyPosition(poolId uint64, addr sdk.AccAddress, lowerTick, upperTick int64) []byte {
	addrKey := address.MustLengthPrefix(addr.Bytes())
	return []byte(fmt.Sprintf("%s%s%s%s%d%s%d%s%d", PositionPrefix, KeySeparator, addrKey, KeySeparator, poolId, KeySeparator, lowerTick, KeySeparator, upperTick))
}

func KeyAddressAndPoolId(addr sdk.AccAddress, poolId uint64) []byte {
	addrKey := address.MustLengthPrefix(addr.Bytes())
	return []byte(fmt.Sprintf("%s%s%s%s%d", PositionPrefix, KeySeparator, addrKey, KeySeparator, poolId))
}

func KeyUserPositions(addr sdk.AccAddress) []byte {
	addrKey := address.MustLengthPrefix(addr.Bytes())
	return []byte(fmt.Sprintf("%s%s%s", PositionPrefix, KeySeparator, addrKey))
}

func KeyPool(poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%d", PoolPrefix, poolId))
}

func KeyIncentiveRecord(poolId uint64, denom string, minUptime time.Duration) []byte {
	return []byte(fmt.Sprintf("%s%s%d%s%s%s%d", IncentivePrefix, KeySeparator, poolId, KeySeparator, denom, KeySeparator, uint64(minUptime)))
}

func KeyPoolIncentiveRecords(poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%d", IncentivePrefix, KeySeparator, poolId))
}
