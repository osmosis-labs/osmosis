package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// CreatePoolMsg defines an interface that every CreatePool transaction should implement.
// The gamm logic will use this to create a pool.
type CreatePoolMsg interface {
	PoolCreator() sdk.AccAddress
	Validate(ctx sdk.Context) error
	// Initial Liquidity for the pool that the sender is required to send to the pool account
	InitialLiquidity() sdk.Coins
	CreatePool(ctx sdk.Context, poolID uint64) (PoolI, error)
}
