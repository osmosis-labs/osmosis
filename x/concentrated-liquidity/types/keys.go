package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "concentratedliquidity"

	StoreKey  = ModuleName
	RouterKey = ModuleName

	QuerierRoute = ModuleName
	// Contract: Coin denoms cannot contain this character
	KeySeparator = "|"
)

var (
	TickPrefix     = "tick_prefix" + KeySeparator
	PositionPrefix = "position_prefix" + KeySeparator
)

// KeyTick uses pool Id and tick index
func KeyTick(poolId uint64, tickIndex sdk.Int) []byte {
	return []byte(fmt.Sprintf("%s%d%s", TickPrefix, poolId, tickIndex.String()))
}

// KeyPosition uses pool Id, owner, lower tick and upper tick for keys
func KeyPosition(poolId uint64, address string, lowerTick, upperTick sdk.Int) []byte {
	return []byte(fmt.Sprintf("%s%d%s%s%s", PositionPrefix, poolId, address, lowerTick.String(), upperTick.String()))
}
