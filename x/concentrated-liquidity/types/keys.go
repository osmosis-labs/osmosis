package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	ModuleName = "concentratedliquidity"
	RouterKey  = ModuleName

	StoreKey = ModuleName
)

// Key prefixes
var (
	TickPrefix     = []byte{0x01}
	PositionPrefix = []byte{0x02}
	PoolPrefix     = []byte{0x03}
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

// KeyPosition uses pool Id, owner, lower tick and upper tick for keys
func KeyPosition(poolId uint64, addr sdk.AccAddress, lowerTick, upperTick int64, isIncentivized bool) []byte {
	var key []byte
	key = append(key, PositionPrefix...)
	key = append(key, address.MustLengthPrefix(addr)...)
	key = append(key, sdk.Uint64ToBigEndian(uint64(lowerTick))...)
	key = append(key, sdk.Uint64ToBigEndian(uint64(upperTick))...)
	key = append(key, []byte(fmt.Sprintf("%t", isIncentivized))...)
	return key
}

func KeyPool(poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%d", PoolPrefix, poolId))
}
