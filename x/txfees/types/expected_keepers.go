package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

<<<<<<< HEAD
	poolmanagertypes "github.com/osmosis-labs/osmosis/v18/x/poolmanager/types"
=======
	"github.com/osmosis-labs/osmosis/osmomath"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
>>>>>>> ca75f4c3 (refactor(deps): switch to cosmossdk.io/math from fork math (#6238))
)

// SpotPriceCalculator defines the contract that must be fulfilled by a spot price calculator
// The x/gamm keeper is expected to satisfy this interface.
type SpotPriceCalculator interface {
	CalculateSpotPrice(ctx sdk.Context, poolId uint64, quoteDenom, baseDenom string) (osmomath.Dec, error)
}

// PoolManager defines the contract needed for swap related APIs.
type PoolManager interface {
	RouteExactAmountIn(
		ctx sdk.Context,
		sender sdk.AccAddress,
		routes []poolmanagertypes.SwapAmountInRoute,
		tokenIn sdk.Coin,
		tokenOutMinAmount osmomath.Int) (tokenOutAmount osmomath.Int, err error)

	SwapExactAmountIn(
		ctx sdk.Context,
		sender sdk.AccAddress,
		poolId uint64,
		tokenIn sdk.Coin,
		tokenOutDenom string,
<<<<<<< HEAD
		tokenOutMinAmount sdk.Int,
	) (sdk.Int, error)
=======
		tokenOutMinAmount osmomath.Int,
	) (osmomath.Int, error)

	SwapExactAmountInNoTakerFee(
		ctx sdk.Context,
		sender sdk.AccAddress,
		poolId uint64,
		tokenIn sdk.Coin,
		tokenOutDenom string,
		tokenOutMinAmount osmomath.Int,
	) (osmomath.Int, error)

	GetParams(ctx sdk.Context) (params poolmanagertypes.Params)
>>>>>>> ca75f4c3 (refactor(deps): switch to cosmossdk.io/math from fork math (#6238))
}

// AccountKeeper defines the contract needed for AccountKeeper related APIs.
// Interface provides support to use non-sdk AccountKeeper for AnteHandler's decorators.
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
}

// FeegrantKeeper defines the expected feegrant keeper.
type FeegrantKeeper interface {
	UseGrantedFees(ctx sdk.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error
}

// BankKeeper defines the contract needed for supply related APIs (noalias)
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
}

// TxFeesKeeper defines the expected transaction fee keeper
type TxFeesKeeper interface {
	ConvertToBaseToken(ctx sdk.Context, inputFee sdk.Coin) (sdk.Coin, error)
	GetBaseDenom(ctx sdk.Context) (denom string, err error)
	GetFeeToken(ctx sdk.Context, denom string) (FeeToken, error)
}
