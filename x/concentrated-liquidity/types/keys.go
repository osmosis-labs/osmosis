package types

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "concentratedliquidity"
	RouterKey  = ModuleName

	StoreKey     = ModuleName
	KeySeparator = "|"

	uint64ByteSize = 8
	uintBase       = 10

	ClTokenPrefix = "cl/pool"
)

// Key prefixes
var (
	TickPrefix                   = []byte{0x01}
	PositionPrefix               = []byte{0x02}
	PoolPrefix                   = []byte{0x03}
	IncentivePrefix              = []byte{0x04}
	PositionIdPrefix             = []byte{0x08}
	PoolPositionPrefix           = []byte{0x09}
	FeePositionAccumulatorPrefix = []byte{0x0A}
	PoolFeeAccumulatorPrefix     = []byte{0x0B}
	UptimeAccumulatorPrefix      = []byte{0x0C}
	PositionToLockPrefix         = []byte{0x0D}
	PoolIdForLiquidityPrefix     = []byte{0x0E}
	BalancerFullRangePrefix      = []byte{0x0F}
	LockToPositionPrefix         = []byte{0x10}
	ConcentratedLockPrefix       = []byte{0x11}

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

// PositionId<>LockId and LockId<>PositionId Prefix Keys
func PositionIdForLockIdKeys(positionId, lockId uint64) (positionIdToLockIdKey []byte, lockIdToPositionIdKeu []byte) {
	positionIdToLockIdKey = []byte(fmt.Sprintf("%s%s%d", PositionToLockPrefix, KeySeparator, positionId))
	lockIdToPositionIdKeu = []byte(fmt.Sprintf("%s%s%d", LockToPositionPrefix, KeySeparator, lockId))
	return positionIdToLockIdKey, lockIdToPositionIdKeu
}

// PositionToLockPrefix Prefix Keys

func KeyPositionIdForLock(positionId uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%d", PositionToLockPrefix, KeySeparator, positionId))
}

// LockToPositionPrefix Prefix Keys

func KeyLockIdForPositionId(lockId uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%d", LockToPositionPrefix, KeySeparator, lockId))
}

// PoolIdForLiquidity Prefix Keys

func KeyPoolIdForLiquidity(poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%d", PoolIdForLiquidityPrefix, KeySeparator, poolId))
}

// PositionId Prefix Keys

func KeyPositionId(positionId uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%d", PositionIdPrefix, KeySeparator, positionId))
}

// Position Prefix Keys

func KeyAddressPoolIdPositionId(addr sdk.AccAddress, poolId uint64, positionId uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%x%s%d%s%d", PositionPrefix, KeySeparator, addr.Bytes(), KeySeparator, poolId, KeySeparator, positionId))
}

func KeyAddressAndPoolId(addr sdk.AccAddress, poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%x%s%d", PositionPrefix, KeySeparator, addr.Bytes(), KeySeparator, poolId))
}

func KeyUserPositions(addr sdk.AccAddress) []byte {
	return []byte(fmt.Sprintf("%s%s%x", PositionPrefix, KeySeparator, addr.Bytes()))
}

// Pool Position Prefix Keys
// Used to map a pool id to a position id

func KeyPoolPositionPositionId(poolId uint64, positionId uint64) []byte {
	poolIdBz := sdk.Uint64ToBigEndian(poolId)
	positionIdBz := sdk.Uint64ToBigEndian(positionId)
	key := make([]byte, 0, len(PoolPositionPrefix)+uint64ByteSize+uint64ByteSize+len(KeySeparator))
	key = append(key, PoolPositionPrefix...)
	key = append(key, poolIdBz...)
	key = append(key, KeySeparator...)
	key = append(key, positionIdBz...)
	return key
}

func KeyPoolPosition(poolId uint64) []byte {
	poolIdBz := sdk.Uint64ToBigEndian(poolId)
	key := make([]byte, 0, len(PoolPositionPrefix)+uint64ByteSize)
	key = append(key, PoolPositionPrefix...)
	key = append(key, poolIdBz...)
	return key
}

// Pool Prefix Keys
// Used to map a pool id to a pool struct

func KeyPool(poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%d", PoolPrefix, poolId))
}

// Incentive Prefix Keys

func KeyIncentiveRecord(poolId uint64, minUptimeIndex int, denom string, addr sdk.AccAddress) []byte {
	return []byte(fmt.Sprintf("%s%s%d%s%d%s%s%s%s", IncentivePrefix, KeySeparator, poolId, KeySeparator, minUptimeIndex, KeySeparator, denom, KeySeparator, addr))
}

func KeyUptimeIncentiveRecords(poolId uint64, minUptimeIndex int) []byte {
	return []byte(fmt.Sprintf("%s%s%d%s%d", IncentivePrefix, KeySeparator, poolId, KeySeparator, minUptimeIndex))
}

func KeyPoolIncentiveRecords(poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%d", IncentivePrefix, KeySeparator, poolId))
}

// Fee Accumulator Prefix Keys

func KeyFeePositionAccumulator(positionId uint64) string {
	return strings.Join([]string{string(FeePositionAccumulatorPrefix), strconv.FormatUint(positionId, 10)}, KeySeparator)
}

func KeyFeePoolAccumulator(poolId uint64) string {
	poolIdStr := strconv.FormatUint(poolId, uintBase)
	return strings.Join([]string{string(PoolFeeAccumulatorPrefix), poolIdStr}, "/")
}

// Uptme Accumulator Prefix Keys

func KeyUptimeAccumulator(poolId uint64, uptimeIndex uint64) string {
	poolIdStr := strconv.FormatUint(poolId, uintBase)
	uptimeIndexStr := strconv.FormatUint(uptimeIndex, uintBase)
	return strings.Join([]string{string(UptimeAccumulatorPrefix), poolIdStr, uptimeIndexStr}, "/")
}

// Balancer Full Range Prefix Keys

func KeyBalancerFullRange(clPoolId, balancerPoolId, uptimeIndex uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%d%s%d%s%d", BalancerFullRangePrefix, KeySeparator, clPoolId, KeySeparator, balancerPoolId, KeySeparator, uptimeIndex))
}

// Helper Functions
func MustGetPoolIdFromShareDenom(denom string) uint64 {
	if !strings.HasPrefix(denom, ClTokenPrefix) {
		panic("denom does not start with the cl token prefix")
	}
	parts := strings.Split(denom, "/")
	if len(parts) != 3 {
		panic("cl token denom does not have the correct number of parts")
	}
	poolIdStr := parts[2]
	poolId, err := strconv.Atoi(poolIdStr)
	if err != nil {
		panic(err)
	}
	return uint64(poolId)
}

func GetPositionIdFromShareDenom(denom string) (uint64, error) {
	if !strings.HasPrefix(denom, ClTokenPrefix) {
		return uint64(0), fmt.Errorf("denom does not start with the cl token prefix")
	}
	parts := strings.Split(denom, "/")
	if len(parts) != 4 {
		return uint64(0), fmt.Errorf("cl token denom does not have the correct number of parts")
	}
	poolIdStr := parts[3]
	poolId, err := strconv.Atoi(poolIdStr)
	if err != nil {
		return uint64(0), err
	}
	return uint64(poolId), nil
}
