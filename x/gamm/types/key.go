package types

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "gamm"

	StoreKey = ModuleName

	RouterKey = ModuleName

	QuerierRoute = ModuleName

	GAMMTokenPrefix = "gamm/pool/"
)

var (
	// KeyNextGlobalPoolId defines key to store the next Pool ID to be used.
	KeyNextGlobalPoolId = []byte{0x01}
	// KeyPrefixPools defines prefix to store pools.
	KeyPrefixPools = []byte{0x02}
	// KeyTotalLiquidity defines key to store total liquidity.
	KeyTotalLiquidity = []byte{0x03}

	KeyPrefixMigrationInfoBalancerPool = []byte{0x04}
	KeyPrefixMigrationInfoCLPool       = []byte{0x05}
)

func MustGetPoolIdFromShareDenom(denom string) uint64 {
	number, err := GetPoolIdFromShareDenom(denom)
	if err != nil {
		panic(err)
	}
	return number
}

func GetPoolIdFromShareDenom(denom string) (uint64, error) {
	numberStr := strings.TrimLeft(denom, GAMMTokenPrefix)
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		return 0, err
	}
	return uint64(number), nil
}

func GetDenomPrefix(denom string) []byte {
	return append(KeyTotalLiquidity, []byte(denom)...)
}

func GetPoolShareDenom(poolId uint64) string {
	return fmt.Sprintf("%s%d", GAMMTokenPrefix, poolId)
}

func GetKeyPrefixPools(poolId uint64) []byte {
	return append(KeyPrefixPools, sdk.Uint64ToBigEndian(poolId)...)
}

func GetKeyPrefixMigrationInfoBalancerPool(balancerPoolId uint64) []byte {
	return append(KeyPrefixMigrationInfoBalancerPool, sdk.Uint64ToBigEndian(balancerPoolId)...)
}

func GetKeyPrefixMigrationInfoPoolCLPool(concentratedPoolId uint64) []byte {
	return append(KeyPrefixMigrationInfoCLPool, sdk.Uint64ToBigEndian(concentratedPoolId)...)
}
