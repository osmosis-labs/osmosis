package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/gamm keeper.
type AccountKeeper interface {
	NewAccount(sdk.Context, authtypes.AccountI) authtypes.AccountI

	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	SetAccount(ctx sdk.Context, acc authtypes.AccountI)

	GetModuleAddressAndPermissions(moduleName string) (addr sdk.AccAddress, permissions []string)
}

// BankKeeper defines the banking contract that must be fulfilled when
// creating a x/gamm keeper.
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error

	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error

	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error

	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) error

	SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata)

	// Only needed for simulation interface matching
	// TODO: Look into golang syntax to make this "Everything in stakingtypes.bankkeeper + extra funcs"
	// I think it has to do with listing another interface as the first line here?
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// CommunityPoolKeeper defines the contract needed to be fulfilled for distribution keeper.
type CommunityPoolKeeper interface {
	FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

// PoolManager defines the interface needed to be fulfilled for
// the pool manger.
type PoolManager interface {
	CreatePool(ctx sdk.Context, msg swaproutertypes.CreatePoolMsg) (uint64, error)

	GetNextPoolId(ctx sdk.Context) uint64

	RouteExactAmountIn(
		ctx sdk.Context,
		sender sdk.AccAddress,
		routes []swaproutertypes.SwapAmountInRoute,
		tokenIn sdk.Coin,
		tokenOutMinAmount sdk.Int) (tokenOutAmount sdk.Int, err error)

	RouteExactAmountOut(ctx sdk.Context,
		sender sdk.AccAddress,
		routes []swaproutertypes.SwapAmountOutRoute,
		tokenInMaxAmount sdk.Int,
		tokenOut sdk.Coin,
	) (tokenInAmount sdk.Int, err error)

	MultihopEstimateOutGivenExactAmountIn(
		ctx sdk.Context,
		routes []swaproutertypes.SwapAmountInRoute,
		tokenIn sdk.Coin,
	) (tokenOutAmount sdk.Int, err error)

	MultihopEstimateInGivenExactAmountOut(
		ctx sdk.Context,
		routes []swaproutertypes.SwapAmountOutRoute,
		tokenOut sdk.Coin) (tokenInAmount sdk.Int, err error)
}
