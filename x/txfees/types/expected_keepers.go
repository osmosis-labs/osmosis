package types

import (
	context "context"
	markettypes "github.com/osmosis-labs/osmosis/v27/x/market/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

// SpotPriceCalculator defines the contract that must be fulfilled by a spot price calculator
// The x/gamm keeper is expected to satisfy this interface.
type SpotPriceCalculator interface {
	CalculateSpotPrice(ctx sdk.Context, poolId uint64, quoteDenom, baseDenom string) (osmomath.BigDec, error)
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
		tokenOutMinAmount osmomath.Int,
	) (osmomath.Int, sdk.Coin, error)

	SwapExactAmountInNoTakerFee(
		ctx sdk.Context,
		sender sdk.AccAddress,
		poolId uint64,
		tokenIn sdk.Coin,
		tokenOutDenom string,
		tokenOutMinAmount osmomath.Int,
	) (osmomath.Int, error)

	GetParams(ctx sdk.Context) (params poolmanagertypes.Params)

	RouteCalculateSpotPrice(
		ctx sdk.Context,
		poolId uint64,
		quoteAssetDenom string,
		baseAssetDenom string,
	) (price osmomath.BigDec, err error)
	UpdateTakerFeeTrackerForCommunityPoolByDenom(ctx sdk.Context, denom string, increasedAmt osmomath.Int) error
	UpdateTakerFeeTrackerForStakersByDenom(ctx sdk.Context, denom string, increasedAmt osmomath.Int) error
	GetAllTakerFeeShareAccumulators(ctx sdk.Context) ([]poolmanagertypes.TakerFeeSkimAccumulator, error)
	GetTakerFeeShareAgreementFromDenomNoCache(ctx sdk.Context, takerFeeShareDenom string) (poolmanagertypes.TakerFeeShareAgreement, bool)
	DeleteAllTakerFeeShareAccumulatorsForTakerFeeShareDenom(ctx sdk.Context, takerFeeShareDenom string)
}

// AccountKeeper defines the contract needed for AccountKeeper related APIs.
// Interface provides support to use non-sdk AccountKeeper for AnteHandler's decorators.
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
}

// FeegrantKeeper defines the expected feegrant keeper.
type FeegrantKeeper interface {
	UseGrantedFees(ctx sdk.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error
}

// BankKeeper defines the contract needed for supply related APIs (noalias)
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	IsSendEnabledCoins(ctx context.Context, coins ...sdk.Coin) error
	SendCoins(ctx context.Context, from, to sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

// TxFeesKeeper defines the expected transaction fee keeper
type TxFeesKeeper interface {
	ConvertToBaseToken(ctx sdk.Context, inputFee sdk.Coin) (sdk.Coin, error)
	GetBaseDenom(ctx sdk.Context) (denom string, err error)
	IsFeeToken(ctx sdk.Context, denom string) bool
}

type ProtorevKeeper interface {
	GetPoolForDenomPairNoOrder(ctx sdk.Context, baseDenom, denomToMatch string) (uint64, error)
}

type DistributionKeeper interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

type ConsensusKeeper interface {
	Params(ctx context.Context, _ *consensustypes.QueryParamsRequest) (*consensustypes.QueryParamsResponse, error)
}

type MarketKeeper interface {
	Swap(
		ctx sdk.Context,
		trader sdk.AccAddress,
		receiver sdk.AccAddress,
		offerCoin sdk.Coin,
		askDenom string,
	) (*markettypes.MsgSwapResponse, error)
}

type OracleKeeper interface {
	GetMelodyExchangeRate(ctx sdk.Context, denom string) (osmomath.Dec, error)
	IterateNoteExchangeRates(ctx sdk.Context, handler func(denom string, exchangeRate osmomath.Dec) (stop bool))
}
