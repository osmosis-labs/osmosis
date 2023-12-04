package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	epochtypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

// SpotPriceCalculator defines the contract that must be fulfilled by a spot price calculator
// The x/gamm keeper is expected to satisfy this interface.
type SpotPriceCalculator interface {
	CalculateSpotPrice(ctx sdk.Context, poolId uint64, quoteDenom, baseDenom string) (sdk.Dec, error)
	GetPoolDenoms(ctx sdk.Context, poolId uint64) ([]string, error)
}

// PoolManager defines the contract needed for swap related APIs.
type PoolManager interface {
	RouteExactAmountIn(
		ctx sdk.Context,
		sender sdk.AccAddress,
		routes []poolmanagertypes.SwapAmountInRoute,
		tokenIn sdk.Coin,
		tokenOutMinAmount sdk.Int) (tokenOutAmount sdk.Int, err error)
}

type EpochKeeper interface {
	GetEpochInfo(ctx sdk.Context, identifier string) epochtypes.EpochInfo
}

// AccountKeeper defines the contract needed for AccountKeeper related APIs.
// Interface provides support to use non-sdk AccountKeeper for AnteHandler's decorators.
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI
}

// FeegrantKeeper defines the expected feegrant keeper.
type FeegrantKeeper interface {
	UseGrantedFees(ctx sdk.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error
}

// BankKeeper defines the contract needed for supply related APIs (noalias)
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) error
}

// TxFeesKeeper defines the expected transaction fee keeper
type TxFeesKeeper interface {
	ConvertToBaseToken(ctx sdk.Context, inputFee sdk.Coin) (sdk.Coin, error)
	GetBaseDenom(ctx sdk.Context) (denom string, err error)
	GetFeeToken(ctx sdk.Context, denom string) (FeeToken, error)
}
