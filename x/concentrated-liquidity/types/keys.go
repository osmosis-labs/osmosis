package types

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
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

	uint64ByteSize = 8
)

// Key prefixes
var (
	TickPrefix      = []byte{0x01}
	PositionPrefix  = []byte{0x02}
	PoolPrefix      = []byte{0x03}
	IncentivePrefix = []byte{0x04}

	// n.b. we negative prefix must be less than the positive prefix for proper iteration
	TickNegativePrefix = []byte{0x05}
	TickPositivePrefix = []byte{0x06}

	KeyNextGlobalPositionId = []byte{0x07}

	// prefix, pool id, sign byte, tick index
	TickKeyLengthBytes = len(TickPrefix) + uint64ByteSize + 1 + uint64ByteSize
)

// TickIndexToBytes converts a tick index to a byte slice. Negative tick indexes
// are prefixed with 0x00 a byte and positive tick indexes are prefixed with a
// 0x01 byte. We do this because big endian byte encoding does not give us in
// order iteration in state due to the tick index values being signed integers.
func TickIndexToBytes(tickIndex int64) []byte {
	key := make([]byte, 9)
	if tickIndex < 0 {
		copy(key[:1], TickNegativePrefix)
		copy(key[1:], sdk.Uint64ToBigEndian(uint64(tickIndex)))
	} else {
		copy(key[:1], TickPositivePrefix)
		copy(key[1:], sdk.Uint64ToBigEndian(uint64(tickIndex)))
	}

	return key
}

// TickIndexFromBytes converts an encoded tick index to an int64 value. It returns
// an error if the encoded tick has invalid length.
func TickIndexFromBytes(bz []byte) (int64, error) {
	if len(bz) != 9 {
		return 0, InvalidTickIndexEncodingError{Length: len(bz)}
	}

	return int64(sdk.BigEndianToUint64(bz[1:])), nil
}

// KeyTick generates a tick key for a given pool and tick index by concatenating
// the tick prefix key (generated using keyTickPrefixByPoolIdPrealloc) with the KeySeparator
// and the tick index bytes. This function is used to create unique keys for ticks
// within a pool.
//
// Parameters:
// - poolId (uint64): The pool id for which the tick key is to be generated.
// - tickIndex (int64): The tick index for which the tick key is to be generated.
//
// Returns:
// - []byte: A byte slice representing the generated tick key.
func KeyTick(poolId uint64, tickIndex int64) []byte {
	// 8 bytes for unsigned pool id and 8 bytes for signed tick index.
	key := keyTickPrefixByPoolIdPrealloc(poolId, TickKeyLengthBytes)
	key = append(key, TickIndexToBytes(tickIndex)...)
	return key
}

// KeyTickPrefixByPoolId generates a tick prefix key for a given pool by calling
// the keyTickPrefixByPoolIdPrealloc function with the appropriate pre-allocated memory size.
// The resulting tick prefix key is used as a base for generating unique tick keys
// within a pool.
//
// Parameters:
// - poolId (uint64): The pool id for which the tick prefix key is to be generated.
//
// Returns:
// - []byte: A byte slice representing the generated tick prefix key.
func KeyTickPrefixByPoolId(poolId uint64) []byte {
	return keyTickPrefixByPoolIdPrealloc(poolId, len(TickPrefix)+uint64ByteSize)
}

// keyTickPrefixByPoolIdPrealloc generates a tick prefix key for a given pool by concatenating
// the TickPrefix, KeySeparator, and the big-endian representation of the pool id.
// The function pre-allocates memory for the resulting key to improve performance.
//
// Parameters:
// - poolId (uint64): The pool id for which the tick prefix key is to be generated.
// - preAllocBytes (int): The number of bytes to pre-allocate for the resulting key.
//
// Returns:
// - []byte: A byte slice representing the generated tick prefix key.
func keyTickPrefixByPoolIdPrealloc(poolId uint64, preAllocBytes int) []byte {
	key := make([]byte, 0, preAllocBytes)
	key = append(key, TickPrefix...)
	key = append(key, sdk.Uint64ToBigEndian(poolId)...)
	return key
}

// KeyFullPosition uses pool Id, owner, lower tick, upper tick, joinTime, freezeDuration, and positionId for keys
func KeyFullPosition(poolId uint64, addr sdk.AccAddress, lowerTick, upperTick int64, joinTime time.Time, freezeDuration time.Duration, positionId uint64) []byte {
	joinTimeKey := osmoutils.FormatTimeString(joinTime)

	var builder strings.Builder
	builder.Grow(len(PositionPrefix) + 9*len(KeySeparator) + len(addr.Bytes())*2 + 5*20) // pre-allocating the buffer

	builder.Write(PositionPrefix)
	builder.WriteString(KeySeparator)
	builder.WriteString(hex.EncodeToString(addr.Bytes()))
	builder.WriteString(KeySeparator)
	builder.WriteString(strconv.FormatUint(poolId, 10))
	builder.WriteString(KeySeparator)
	builder.WriteString(strconv.FormatInt(lowerTick, 10))
	builder.WriteString(KeySeparator)
	builder.WriteString(strconv.FormatInt(upperTick, 10))
	builder.WriteString(KeySeparator)
	builder.WriteString(joinTimeKey)
	builder.WriteString(KeySeparator)
	builder.WriteString(strconv.FormatUint(uint64(freezeDuration), 10))
	builder.WriteString(KeySeparator)
	builder.WriteString(strconv.FormatUint(positionId, 10))

	return []byte(builder.String())
}

// KeyPosition uses pool Id, owner, lower tick and upper tick for keys
func KeyPosition(poolId uint64, addr sdk.AccAddress, lowerTick, upperTick int64) []byte {
	return []byte(fmt.Sprintf("%s%s%x%s%d%s%d%s%d", PositionPrefix, KeySeparator, addr.Bytes(), KeySeparator, poolId, KeySeparator, lowerTick, KeySeparator, upperTick))
}

func KeyAddressAndPoolId(addr sdk.AccAddress, poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%x%s%d", PositionPrefix, KeySeparator, addr.Bytes(), KeySeparator, poolId))
}

func KeyUserPositions(addr sdk.AccAddress) []byte {
	return []byte(fmt.Sprintf("%s%s%x", PositionPrefix, KeySeparator, addr.Bytes()))
}

func KeyPool(poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%d", PoolPrefix, poolId))
}

func KeyIncentiveRecord(poolId uint64, denom string, minUptime time.Duration, addr sdk.AccAddress) []byte {
	addrKey := address.MustLengthPrefix(addr.Bytes())
	return []byte(fmt.Sprintf("%s%s%d%s%s%s%d%s%s", IncentivePrefix, KeySeparator, poolId, KeySeparator, denom, KeySeparator, uint64(minUptime), KeySeparator, addrKey))
}

func KeyPoolIncentiveRecords(poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%d", IncentivePrefix, KeySeparator, poolId))
}
