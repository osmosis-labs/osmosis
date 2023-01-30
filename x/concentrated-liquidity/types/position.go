package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FullPositionByOwnerResult represents the result of a full position
// entry when querying it by owner.
// Normally, the full position is a collection of owner, pool id,
// lower tick, upper tick, frozen until, and liquidity.
// Here, we omit owner since it is already known when queried.
type FullPositionByOwnerResult struct {
	PoolId      uint64
	LowerTick   int64
	UpperTick   int64
	FrozenUntil time.Time
	Liquidity   sdk.Dec
}
