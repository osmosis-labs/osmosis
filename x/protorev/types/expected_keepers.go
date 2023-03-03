package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochtypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/protorev keeper.
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}

// BankKeeper defines the banking contract that must be fulfilled when
// creating a x/protorev keeper.
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) error
}

// GAMMKeeper defines the Gamm contract that must be fulfilled when
// creating a x/protorev keeper.
type GAMMKeeper interface {
	GetPoolAndPoke(ctx sdk.Context, poolId uint64) (gammtypes.CFMMPoolI, error)
	GetPoolsAndPoke(ctx sdk.Context) (res []gammtypes.CFMMPoolI, err error)
	GetPoolDenoms(ctx sdk.Context, poolId uint64) ([]string, error)
	GetPoolType(ctx sdk.Context, poolId uint64) (poolmanagertypes.PoolType, error)
}

// PoolManagerKeeper defines the PoolManager contract that must be fulfilled when
// creating a x/protorev keeper.
type PoolManagerKeeper interface {
	RouteExactAmountIn(
		ctx sdk.Context,
		sender sdk.AccAddress,
		routes []poolmanagertypes.SwapAmountInRoute,
		tokenIn sdk.Coin,
		tokenOutMinAmount sdk.Int) (tokenOutAmount sdk.Int, err error)

	MultihopEstimateOutGivenExactAmountIn(
		ctx sdk.Context,
		routes []poolmanagertypes.SwapAmountInRoute,
		tokenIn sdk.Coin,
	) (tokenOutAmount sdk.Int, err error)
}

// EpochKeeper defines the Epoch contract that must be fulfilled when
// creating a x/protorev keeper.
type EpochKeeper interface {
	GetEpochInfo(ctx sdk.Context, identifier string) epochtypes.EpochInfo
}
