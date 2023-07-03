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
	base10         = 10

	ConcentratedLiquidityTokenPrefix = "cl/pool"
)

// Key prefixes
var (
	TickPrefix      = []byte{0x01}
	PositionPrefix  = []byte{0x02}
	PoolPrefix      = []byte{0x03}
	IncentivePrefix = []byte{0x04}

	// n.b. negative prefix must be less than the positive prefix for proper iteration
	TickNegativePrefix = []byte{0x05}
	TickPositivePrefix = []byte{0x06}

	KeyNextGlobalPositionId = []byte{0x07}

	PositionIdPrefix                      = []byte{0x08}
	PoolPositionPrefix                    = []byte{0x09}
	SpreadRewardPositionAccumulatorPrefix = []byte{0x0A}
	KeySpreadRewardPoolAccumulatorPrefix  = []byte{0x0B}
	UptimeAccumulatorPrefix               = []byte{0x0C}
	PositionToLockPrefix                  = []byte{0x0D}
	FullRangeLiquidityPrefix              = []byte{0x0E}
	BalancerFullRangePrefix               = []byte{0x0F}
	LockToPositionPrefix                  = []byte{0x10}
	ConcentratedLockPrefix                = []byte{0x11}

	KeyNextGlobalIncentiveRecordId = []byte{0x12}

	KeyTotalLiquidity = []byte{0x13}

	// TickPrefix + pool id
	KeyTickPrefixByPoolIdLengthBytes = len(TickPrefix) + uint64ByteSize
	// TickPrefix + pool id + sign byte(negative / positive prefix) + tick index: 18bytes in total
	KeyTickLengthBytes = KeyTickPrefixByPoolIdLengthBytes + 1 + uint64ByteSize
)

// TickIndexToBytes converts a tick index to a byte slice. The encoding is:
// - Negative tick indexes are prefixed with a byte `b`
// - Positive tick indexes are prefixed with a byte `b + 1`.
// - Then we encode sign || BigEndian(uint64(tickIndex))
//
// This leading sign byte is to ensure we can iterate over the tick indexes in order.
// 2's complement guarantees that negative integers are in order when iterating.
// However they are not in order relative to positive integers (as 2's complement flips the leading bit)
// Hence we use the leading sign byte to ensure that negative tick indexes
// are in order relative to positive tick indexes.
// TODO: Test key iteration property
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

	i := int64(sdk.BigEndianToUint64(bz[1:]))
	// ensure sign byte is correct, these errors should never occur.
	if bz[0] == TickNegativePrefix[0] && i >= 0 {
		return 0, InvalidTickIndexEncodingError{Length: len(bz)}
	} else if bz[0] == TickPositivePrefix[0] && i < 0 {
		return 0, InvalidTickIndexEncodingError{Length: len(bz)}
	}
	return i, nil
}

// KeyTick generates a tick key for a given pool and tick index by concatenating
// the tick prefix key (generated using keyTickPrefixByPoolIdPrealloc) with the sign prefix(TickNegativePrefix / TickPositivePrefix)
// and the tick index bytes. This function is used to create unique keys for ticks
// and store specified tick info for each pool.
//
// Parameters:
// - poolId (uint64): The pool id for which the tick key is to be generated.
// - tickIndex (int64): The tick index for which the tick key is to be generated.
//
// Returns:
// - []byte: A byte slice representing the generated tick key.
func KeyTick(poolId uint64, tickIndex int64) []byte {
	// 8 bytes for unsigned pool id and 8 bytes for signed tick index.
	key := keyTickPrefixByPoolIdPrealloc(poolId, KeyTickLengthBytes)
	key = append(key, TickIndexToBytes(tickIndex)...)
	return key
}

// KeyTickPrefixByPoolId generates a tick prefix key for a given pool by calling
// the keyTickPrefixByPoolIdPrealloc function with the appropriate pre-allocated memory size.
// This key indicates the first prefix bytes of the KeyTick and can be used to iterate
// over ticks for the given pool id.
// The resulting tick prefix key is used as a base for generating unique tick keys
// within a pool.
//
// Parameters:
// - poolId (uint64): The pool id for which the tick prefix key is to be generated.
//
// Returns:
// - []byte: A byte slice representing the generated tick prefix key.
func KeyTickPrefixByPoolId(poolId uint64) []byte {
	return keyTickPrefixByPoolIdPrealloc(poolId, KeyTickPrefixByPoolIdLengthBytes)
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
	positionIdToLockIdKey = KeyPositionIdForLock(positionId)
	lockIdToPositionIdKey = KeyLockIdForPositionId(lockId)
	return positionIdToLockIdKey, lockIdToPositionIdKey
}

// KeyPositionIdForLock returns the key consisted of (PositionToLockPrefix | position Id)
func KeyPositionIdForLock(positionId uint64) []byte {
	return []byte(fmt.Sprintf("%s%d", PositionToLockPrefix, positionId))
}

// KeyLockIdForPositionId returns the key consisted of (KeyLockIdForPositionId | lockId)
func KeyLockIdForPositionId(lockId uint64) []byte {
	return []byte(fmt.Sprintf("%s%d", LockToPositionPrefix, lockId))
}

// KeyFullRangeLiquidityPrefix returns the prefix used to keep track of full range liquidity for each pool.
func KeyFullRangeLiquidityPrefix(poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%d", FullRangeLiquidityPrefix, poolId))
}

// KeyPositionId returns the prefix the key consisted of (PositionIdPrefix | position Id) and is used to store position info.
func KeyPositionId(positionId uint64) []byte {
	return []byte(fmt.Sprintf("%s%d", PositionIdPrefix, positionId))
}

// Position Prefix Keys

// KeyAddressPoolIdPositionId returns the full key needed to store the position id for given addr + pool id + position id combination.
func KeyAddressPoolIdPositionId(addr sdk.AccAddress, poolId uint64, positionId uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%x%s%d%s%d", PositionPrefix, KeySeparator, addr.Bytes(), KeySeparator, poolId, KeySeparator, positionId))
}

// KeyAddressAndPoolId returns the prefix key used to create KeyAddressPoolIdPositionId, which only includes addr + pool id.
// This key can be used to iterate over users positions for a specific pool.
func KeyAddressAndPoolId(addr sdk.AccAddress, poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%x%s%d%s", PositionPrefix, KeySeparator, addr.Bytes(), KeySeparator, poolId, KeySeparator))
}

// KeyUserPositions returns the prefix key used to create KeyAddressPoolIdPositionId, which only includes the addr.
// This key can be used to iterate over all positions that a specific address has.
func KeyUserPositions(addr sdk.AccAddress) []byte {
	return []byte(fmt.Sprintf("%s%s%x%s", PositionPrefix, KeySeparator, addr.Bytes(), KeySeparator))
}

// Pool Position Prefix Keys
// Used to map a pool id to a position id

func KeyPoolPositionPositionId(poolId uint64, positionId uint64) []byte {
	poolIdBz := sdk.Uint64ToBigEndian(poolId)
	positionIdBz := sdk.Uint64ToBigEndian(positionId)
	key := make([]byte, 0, len(PoolPositionPrefix)+uint64ByteSize+len(KeySeparator)+uint64ByteSize)
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
// KeyPool is used to map a pool id to a pool struct
func KeyPool(poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%d", PoolPrefix, poolId))
}

// Incentive Prefix Keys
// KeyIncentiveRecord is the key used to store incentive records using the combination of
// pool id + min uptime index + incentive record id.
func KeyIncentiveRecord(poolId uint64, minUptimeIndex int, id uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%d%s%d%s%d%s", IncentivePrefix, KeySeparator, poolId, KeySeparator, minUptimeIndex, KeySeparator, id, KeySeparator))
}

// KeyUptimeIncentiveRecords returns the prefix key for incentives records using the combination of pool id + min uptime index.
// This can be used to iterate over incentive records for the pool id + min upttime index combination.
func KeyUptimeIncentiveRecords(poolId uint64, minUptimeIndex int) []byte {
	return []byte(fmt.Sprintf("%s%s%d%s%d%s", IncentivePrefix, KeySeparator, poolId, KeySeparator, minUptimeIndex, KeySeparator))
}

// KeyPoolIncentiveRecords returns the prefix key for incentives records using given pool id.
// This can be used to iterate over all incentive records for the pool.
func KeyPoolIncentiveRecords(poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%d%s", IncentivePrefix, KeySeparator, poolId, KeySeparator))
}

// Spread Reward Accumulator Prefix Keys

func KeySpreadRewardPositionAccumulator(positionId uint64) string {
	return strings.Join([]string{string(SpreadRewardPositionAccumulatorPrefix), strconv.FormatUint(positionId, 10)}, KeySeparator)
}

// This is guaranteed to not contain "||" so it can be used as an accumulator name.
func KeySpreadRewardPoolAccumulator(poolId uint64) string {
	poolIdStr := strconv.FormatUint(poolId, base10)
	return strings.Join([]string{string(KeySpreadRewardPoolAccumulatorPrefix), poolIdStr}, "/")
}

// Uptme Accumulator Prefix Keys
// This is guaranteed to not contain "||" so it can be used as an accumulator name.
func KeyUptimeAccumulator(poolId uint64, uptimeIndex uint64) string {
	poolIdStr := strconv.FormatUint(poolId, base10)
	uptimeIndexStr := strconv.FormatUint(uptimeIndex, base10)
	return strings.Join([]string{string(UptimeAccumulatorPrefix), poolIdStr, uptimeIndexStr}, "/")
}

// Balancer Full Range Prefix Keys

func KeyBalancerFullRange(clPoolId, balancerPoolId, uptimeIndex uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%d%s%d%s%d", BalancerFullRangePrefix, KeySeparator, clPoolId, KeySeparator, balancerPoolId, KeySeparator, uptimeIndex))
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

func GetDenomPrefix(denom string) []byte {
	return append(KeyTotalLiquidity, []byte(denom)...)
}
