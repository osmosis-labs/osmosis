package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// AccountI defines the account contract that must be fulfilled when
// creating a x/gamm keeper.
type AccountI interface {
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI
	NewAccount(sdk.Context, authtypes.AccountI) authtypes.AccountI
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	SetAccount(ctx sdk.Context, acc authtypes.AccountI)
}

// BankI defines the banking contract that must be fulfilled when
// creating a x/gamm keeper.
type BankI interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata)
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

// CommunityPoolI defines the contract needed to be fulfilled for distribution keeper.
type CommunityPoolI interface {
	FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error
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
	) (price sdk.Dec, err error)

	SwapExactAmountIn(
		ctx sdk.Context,
		sender sdk.AccAddress,
		pool PoolI,
		tokenIn sdk.Coin,
		tokenOutDenom string,
		tokenOutMinAmount sdk.Int,
		spreadFactor sdk.Dec,
	) (sdk.Int, error)
	// CalcOutAmtGivenIn calculates the amount of tokenOut given tokenIn and the pool's current state.
	// Returns error if the given pool is not a CFMM pool. Returns error on internal calculations.
	CalcOutAmtGivenIn(
		ctx sdk.Context,
		poolI PoolI,
		tokenIn sdk.Coin,
		tokenOutDenom string,
		spreadFactor sdk.Dec,
	) (tokenOut sdk.Coin, err error)

	SwapExactAmountOut(
		ctx sdk.Context,
		sender sdk.AccAddress,
		pool PoolI,
		tokenInDenom string,
		tokenInMaxAmount sdk.Int,
		tokenOut sdk.Coin,
		spreadFactor sdk.Dec,
	) (tokenInAmount sdk.Int, err error)
	// CalcInAmtGivenOut calculates the amount of tokenIn given tokenOut and the pool's current state.
	// Returns error if the given pool is not a CFMM pool. Returns error on internal calculations.
	CalcInAmtGivenOut(
		ctx sdk.Context,
		poolI PoolI,
		tokenOut sdk.Coin,
		tokenInDenom string,
		spreadFactor sdk.Dec,
	) (tokenIn sdk.Coin, err error)

	// GetTotalPoolLiquidity returns the coins in the pool owned by all LPs
	GetTotalPoolLiquidity(ctx sdk.Context, poolId uint64) (sdk.Coins, error)

	// ValidatePermissionlessPoolCreationEnabled returns nil if permissionless pool creation in the module is enabled.
	// Otherwise, returns an error.
	ValidatePermissionlessPoolCreationEnabled(ctx sdk.Context) error

	// GetTotalLiquidity returns the total liquidity of all the pools in the module.
	GetTotalLiquidity(ctx sdk.Context) (sdk.Coins, error)
}

type PoolIncentivesKeeperI interface {
	IsPoolIncentivized(ctx sdk.Context, poolId uint64) bool
}

type MultihopRoute interface {
	Length() int
	PoolIds() []uint64
	IntermediateDenoms() []string
}

type StakingKeeper interface {
	BondDenom(ctx sdk.Context) string
}

type ProtorevKeeper interface {
	GetPoolForDenomPair(ctx sdk.Context, baseDenom, denomToMatch string) (uint64, error)
}
