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

	ConcentratedLiquidityTokenPrefix = "cl/pool"
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
func PositionIdForLockIdKeys(positionId, lockId uint64) (positionIdToLockIdKey []byte, lockIdToPositionIdKey []byte) {
	positionIdToLockIdKey = []byte(fmt.Sprintf("%s%s%d", PositionToLockPrefix, KeySeparator, positionId))
	lockIdToPositionIdKey = []byte(fmt.Sprintf("%s%s%d", LockToPositionPrefix, KeySeparator, lockId))
	return positionIdToLockIdKey, lockIdToPositionIdKey
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
	positionIDBytes := []byte(fmt.Sprintf("%d", positionId))

	keyLen := len(PositionIdPrefix) + len(KeySeparator) + len(positionIDBytes)
	key := make([]byte, 0, keyLen)

	key = append(key, PositionIdPrefix...)
	key = append(key, KeySeparator...)
	key = append(key, positionIDBytes...)

	return key
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
	poolIDBytes := []byte(fmt.Sprintf("%d", poolId))

	keyLen := len(PoolPrefix) + len(poolIDBytes)
	key := make([]byte, 0, keyLen)

	key = append(key, PoolPrefix...)
	key = append(key, poolIDBytes...)

	return key
}

// Incentive Prefix Keys

func KeyIncentiveRecord(poolId uint64, minUptimeIndex uint64, denom string, addr sdk.AccAddress) []byte {
	poolIDBytes := []byte(strconv.FormatUint(poolId, uintBase))
	minUptimeIndexBytes := []byte(strconv.FormatUint(minUptimeIndex, uintBase))

	keyLen := len(IncentivePrefix) + len(KeySeparator) + len(poolIDBytes) + len(KeySeparator) +
		len(minUptimeIndexBytes) + len(KeySeparator) + len(denom) + len(KeySeparator) + len(addr.Bytes())

	key := make([]byte, 0, keyLen)

	key = append(key, IncentivePrefix...)
	key = append(key, KeySeparator...)
	key = append(key, poolIDBytes...)
	key = append(key, KeySeparator...)
	key = append(key, minUptimeIndexBytes...)
	key = append(key, KeySeparator...)
	key = append(key, denom...)
	key = append(key, KeySeparator...)
	// Note that the address below must be bech32 encoded
	// This is because we split the key on KeySeparator
	// Therefore, the address must not contain any KeySeparator which may happen with raw bytes
	// This is not an issue with bech32 encoded addresses
	key = append(key, []byte(addr.String())...)

	return key
}

func KeyUptimeIncentiveRecords(poolId uint64, minUptimeIndex uint64) []byte {
	poolIDBytes := []byte(strconv.FormatUint(poolId, uintBase))
	minUptimeIndexBytes := []byte(strconv.FormatUint(minUptimeIndex, uintBase))

	keyLen := len(IncentivePrefix) + len(KeySeparator) + len(poolIDBytes) + len(KeySeparator) + len(minUptimeIndexBytes)
	key := make([]byte, 0, keyLen)

	key = append(key, IncentivePrefix...)
	key = append(key, KeySeparator...)
	key = append(key, poolIDBytes...)
	key = append(key, KeySeparator...)
	key = append(key, minUptimeIndexBytes...)

	return key
}

func KeyPoolIncentiveRecords(poolId uint64) []byte {
	poolIDBytes := strconv.FormatUint(poolId, uintBase)

	keyLen := len(IncentivePrefix) + len(KeySeparator) + len(poolIDBytes)
	key := make([]byte, 0, keyLen)

	key = append(key, IncentivePrefix...)
	key = append(key, KeySeparator...)
	key = append(key, poolIDBytes...)

	return key
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
	clPoolIDBytes := []byte(strconv.FormatUint(clPoolId, uintBase))
	balancerPoolIDBytes := []byte(strconv.FormatUint(balancerPoolId, uintBase))
	uptimeIndexBytes := []byte(strconv.FormatUint(uptimeIndex, uintBase))

	keyLen := len(BalancerFullRangePrefix) + len(KeySeparator) + len(clPoolIDBytes) + len(KeySeparator) +
		len(balancerPoolIDBytes) + len(KeySeparator) + len(uptimeIndexBytes)
	key := make([]byte, 0, keyLen)

	key = append(key, BalancerFullRangePrefix...)
	key = append(key, KeySeparator...)
	key = append(key, clPoolIDBytes...)
	key = append(key, KeySeparator...)
	key = append(key, balancerPoolIDBytes...)
	key = append(key, KeySeparator...)
	key = append(key, uptimeIndexBytes...)

	return key
}

// Helper Functions
func GetPoolIdFromShareDenom(denom string) (uint64, error) {
	if !strings.HasPrefix(denom, ConcentratedLiquidityTokenPrefix) {
		return 0, fmt.Errorf("denom does not start with the cl token prefix")
	}
	parts := strings.Split(denom, "/")
	if len(parts) != 3 {
		return 0, fmt.Errorf("cl token denom does not have the correct number of parts")
	}
	poolIdStr := parts[2]
	poolId, err := strconv.Atoi(poolIdStr)
	if err != nil {
		return 0, fmt.Errorf("failed to convert poolIdStr to integer: %v", err)
	}
	return uint64(poolId), nil
}

func MustGetPoolIdFromShareDenom(denom string) uint64 {
	poolId, err := GetPoolIdFromShareDenom(denom)
	if err != nil {
		panic(err)
	}
	return poolId
}
