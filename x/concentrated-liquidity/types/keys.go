package types

import "fmt"

const (
	ModuleName = "twap"

	StoreKey          = ModuleName
	TransientStoreKey = "transient_" + ModuleName // this is silly we have to do this
	RouterKey         = ModuleName

	QuerierRoute = ModuleName
	// Contract: Coin denoms cannot contain this character
	KeySeparator = "|"
)

var (
	TickPrefix = "tick_prefix" + KeySeparator
)

func KeyTickByPool(poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%d", TickPrefix, poolId))
}
