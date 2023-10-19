package common

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

// PoolKeeper is an interface for getting pools from a keeper.
type PoolKeeper interface {
	GetPools(ctx sdk.Context) ([]poolmanagertypes.PoolI, error)
}

// CosmWasmPoolKeeper is an interface for getting CosmWasm pools from a keeper.
type CosmWasmPoolKeeper interface {
	GetPoolsWithWasmKeeper(ctx sdk.Context) ([]poolmanagertypes.PoolI, error)
}

// BankKeeper is an interface for getting bank balances.
type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}
