package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

// BankKeeper defines the banking contract that must be fulfilled when
// creating a x/concentrated-liquidity keeper.
type BankKeeper interface {
	GetDenomMetaData(ctx sdk.Context, denom string) (banktypes.Metadata, bool)
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
}

// PoolManagerKeeper defines the interface needed to be fulfilled for
// the poolmanager keeper.
type PoolManagerKeeper interface {
	CreatePool(ctx sdk.Context, msg poolmanagertypes.CreatePoolMsg) (uint64, error)
	GetNextPoolId(ctx sdk.Context) uint64
}
