package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"

	epochtypes "github.com/osmosis-labs/osmosis/v12/x/epochs/types"
)

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/protorev keeper.
type AccountKeeper interface {
	NewAccount(sdk.Context, authtypes.AccountI) authtypes.AccountI

	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	SetAccount(ctx sdk.Context, acc authtypes.AccountI)

	GetModuleAddressAndPermissions(moduleName string) (addr sdk.AccAddress, permissions []string)
}

// BankKeeper defines the banking contract that must be fulfilled when
// creating a x/protorev keeper.
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error

	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error

	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error

	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// GAMMKeeper
type GAMMKeeper interface {
	GetPoolAndPoke(ctx sdk.Context, poolId uint64) (gammtypes.PoolI, error)
	GetPoolsAndPoke(ctx sdk.Context) (res []gammtypes.PoolI, err error)
	GetPoolDenoms(ctx sdk.Context, poolId uint64) ([]string, error)
	SwapExactAmountIn(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, tokenIn sdk.Coin, tokenOutDenom string, tokenOutMinAmount sdk.Int) (sdk.Int, error)
	MultihopSwapExactAmountIn(ctx sdk.Context, sender sdk.AccAddress, routes []types.SwapAmountInRoute, tokenIn sdk.Coin, tokenOutMinAmount sdk.Int) (tokenOutAmount sdk.Int, err error)
}

type EpochKeeper interface {
	GetEpochInfo(ctx sdk.Context, identifier string) epochtypes.EpochInfo
}
