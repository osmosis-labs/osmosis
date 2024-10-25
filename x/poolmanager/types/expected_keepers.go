package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// AccountI defines the account contract that must be fulfilled when
// creating a x/gamm keeper.
type AccountI interface {
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
	NewAccount(context.Context, sdk.AccountI) sdk.AccountI
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)
}

// BankI defines the banking contract that must be fulfilled when
// creating a x/gamm keeper.
type BankI interface {
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata)
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

// CommunityPoolI defines the contract needed to be fulfilled for distribution keeper.
type CommunityPoolI interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

// PoolModuleI is the interface that must be fulfillled by the module
// storing and containing the pools.
type PoolModuleI interface {
	InitializePool(ctx sdk.Context, pool PoolI, creatorAddress sdk.AccAddress) error

	GetPool(ctx sdk.Context, poolId uint64) (PoolI, error)

	GetPools(ctx sdk.Context) ([]PoolI, error)

	GetPoolDenoms(ctx sdk.Context, poolId uint64) (denoms []string, err error)

	CalculateSpotPrice(
		ctx sdk.Context,
		poolId uint64,
		quoteAssetDenom string,
		baseAssetDenom string,
	) (price osmomath.BigDec, err error)

	SwapExactAmountIn(
		ctx sdk.Context,
		sender sdk.AccAddress,
		pool PoolI,
		tokenIn sdk.Coin,
		tokenOutDenom string,
		tokenOutMinAmount osmomath.Int,
		spreadFactor osmomath.Dec,
	) (osmomath.Int, error)
	// CalcOutAmtGivenIn calculates the amount of tokenOut given tokenIn and the pool's current state.
	// Returns error if the given pool is not a CFMM pool. Returns error on internal calculations.
	CalcOutAmtGivenIn(
		ctx sdk.Context,
		poolI PoolI,
		tokenIn sdk.Coin,
		tokenOutDenom string,
		spreadFactor osmomath.Dec,
	) (tokenOut sdk.Coin, err error)

	SwapExactAmountOut(
		ctx sdk.Context,
		sender sdk.AccAddress,
		pool PoolI,
		tokenInDenom string,
		tokenInMaxAmount osmomath.Int,
		tokenOut sdk.Coin,
		spreadFactor osmomath.Dec,
	) (tokenInAmount osmomath.Int, err error)
	// CalcInAmtGivenOut calculates the amount of tokenIn given tokenOut and the pool's current state.
	// Returns error if the given pool is not a CFMM pool. Returns error on internal calculations.
	CalcInAmtGivenOut(
		ctx sdk.Context,
		poolI PoolI,
		tokenOut sdk.Coin,
		tokenInDenom string,
		spreadFactor osmomath.Dec,
	) (tokenIn sdk.Coin, err error)

	// GetTotalPoolLiquidity returns the coins in the pool owned by all LPs
	GetTotalPoolLiquidity(ctx sdk.Context, poolId uint64) (sdk.Coins, error)

	// GetTotalLiquidity returns the total liquidity of all the pools in the module.
	GetTotalLiquidity(ctx sdk.Context) (sdk.Coins, error)
}

type ConcentratedI interface {
	PoolModuleI
	GetWhitelistedAddresses(ctx sdk.Context) []string
}

type PoolIncentivesKeeperI interface {
	IsPoolIncentivized(ctx sdk.Context, poolId uint64) (bool, error)
}

type MultihopRoute interface {
	Length() int
	PoolIds() []uint64
	IntermediateDenoms() []string
}

type StakingKeeper interface {
	BondDenom(ctx context.Context) (string, error)
}

type ProtorevKeeper interface {
	GetPoolForDenomPair(ctx sdk.Context, baseDenom, denomToMatch string) (uint64, error)
}

type WasmKeeper interface {
	QuerySmart(ctx context.Context, contractAddress sdk.AccAddress, queryMsg []byte) ([]byte, error)
}

type AffiliateKeeper interface {
	IsAffiliate(ctx sdk.Context, address sdk.AccAddress) (bool, error)
	GetAffiliates(ctx sdk.Context, address sdk.AccAddress) ([]sdk.AccAddress, error)
}
