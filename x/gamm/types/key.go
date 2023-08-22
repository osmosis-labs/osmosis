package types

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cltypes "github.com/osmosis-labs/osmosis/v17/x/concentrated-liquidity/types"
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

// GetPoolIdFromShareDenomForCLandGAMM performs the same action as GetPoolIdFromShareDenom but instead of only checking,
// "gamm/pool/{id}" denom it also checks "cl/pool/{id}" and retrieves the {id} from the denom. This is later used in
// superfluid GetOrCreateIntermediaryAccount where we create the gauge based on a give lock denoms poolId.
func GetPoolIdFromShareDenomForCLandGAMM(denom string) (uint64, error) {
	if strings.HasPrefix(denom, "gamm/") {
		return GetPoolIdFromShareDenom(denom)
	} else if strings.HasPrefix(denom, "cl/") {
		return cltypes.GetPoolIdFromShareDenom(denom)
	} else {
		return 0, fmt.Errorf("Input denom (%s) did not match any valid prefixes", denom)
	}
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
