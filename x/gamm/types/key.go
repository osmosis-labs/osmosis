package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "gamm"

	StoreKey = ModuleName

	RouterKey = ModuleName

	QuerierRoute = ModuleName
)

var (
	// KeyNextGlobalPoolNumber defines key to store the next Pool ID to be used
	KeyNextGlobalPoolNumber = []byte{0x01}
	// KeyPrefixPools defines prefix to store pools
	KeyPrefixPools = []byte{0x02}
)

func GetPoolShareDenom(poolId uint64) string {
	return fmt.Sprintf("gamm/pool/%d", poolId)
}

func GetKeyPrefixPools(poolId uint64) []byte {
	return append(KeyPrefixPools, sdk.Uint64ToBigEndian(poolId)...)
}
