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
	// KeyLastLockID defines key to store lock ID used by last
	KeyGlobalPoolNumber = []byte{0x01}
	// KeyPrefixPools defines prefix to store pools
	KeyPrefixPools = []byte{0x02}

	// // Used for querying to paginate the registered pool numbers.
	// KeyPrefixPaginationPoolNumbers = []byte{0x03}
)

func GetPoolShareDenom(poolId uint64) string {
	return fmt.Sprintf("gamm/pool/%d", poolId)
}

func GetKeyPrefixPools(poolId uint64) []byte {
	return append(KeyPrefixPools, sdk.Uint64ToBigEndian(poolId)...)
}

// func GetKeyPaginationPoolNumbers(poolId uint64) []byte {
// 	return append(KeyPrefixPaginationPoolNumbers, sdk.Uint64ToBigEndian(poolId)...)
// }
