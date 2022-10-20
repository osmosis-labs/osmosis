package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	ModuleName = "concentrated-liquidity"

	StoreKey = ModuleName
)

// Key prefixes
var (
	TickPrefix     = []byte{0x01}
	PositionPrefix = []byte{0x02}
)

// KeyTick returns a key for storing a TickInfo object.
func KeyTick(poolId uint64, tickIndex int64) []byte {
	key := KeyTickPrefix(poolId)
	key = append(key, sdk.Uint64ToBigEndian(uint64(tickIndex))...)
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
func KeyPosition(poolId uint64, addr sdk.AccAddress, lowerTick, upperTick int64) []byte {
	var key []byte
	key = append(key, PositionPrefix...)
	key = append(key, address.MustLengthPrefix(addr)...)
	key = append(key, sdk.Uint64ToBigEndian(uint64(lowerTick))...)
	key = append(key, sdk.Uint64ToBigEndian(uint64(upperTick))...)
	return key
}
