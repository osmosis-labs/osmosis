package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CreatePoolMsg defines an interface that every CreatePool transaction should implement.
// The gamm logic will use this to create a pool.
type CreatePoolMsg interface {
	// GetPoolType returns the type of the pool to create.
	GetPoolType() PoolType
	// The creator of the pool, who pays the PoolCreationFee, provides initial liquidity,
	// and gets the initial LP shares.
	PoolCreator() sdk.AccAddress
	// A stateful validation function.
	Validate(ctx sdk.Context) error
	// Initial Liquidity for the pool that the sender is required to send to the pool account
	InitialLiquidity() sdk.Coins
	// CreatePool creates a pool implementing PoolI, using data from the message.
	CreatePool(ctx sdk.Context, poolID uint64) (PoolI, error)
}
