package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"

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
		swapFee sdk.Dec,
	) (sdk.Int, error)
	// CalcOutAmtGivenIn calculates the amount of tokenOut given tokenIn and the pool's current state.
	// Returns error if the given pool is not a CFMM pool. Returns error on internal calculations.
	CalcOutAmtGivenIn(
		ctx sdk.Context,
		poolI PoolI,
		tokenIn sdk.Coin,
		tokenOutDenom string,
		swapFee sdk.Dec,
	) (tokenOut sdk.Coin, err error)

	SwapExactAmountOut(
		ctx sdk.Context,
		sender sdk.AccAddress,
		pool PoolI,
		tokenInDenom string,
		tokenInMaxAmount sdk.Int,
		tokenOut sdk.Coin,
		swapFee sdk.Dec,
	) (tokenInAmount sdk.Int, err error)
	// CalcInAmtGivenOut calculates the amount of tokenIn given tokenOut and the pool's current state.
	// Returns error if the given pool is not a CFMM pool. Returns error on internal calculations.
	CalcInAmtGivenOut(
		ctx sdk.Context,
		poolI PoolI,
		tokenOut sdk.Coin,
		tokenInDenom string,
		swapFee sdk.Dec,
	) (tokenIn sdk.Coin, err error)

	// GetTotalPoolLiquidity returns the coins in the pool owned by all LPs
	GetTotalPoolLiquidity(ctx sdk.Context, poolId uint64) (sdk.Coins, error)

	// ValidatePermissionlessPoolCreationEnabled returns nil if permissionless pool creation in the module is enabled.
	// Otherwise, returns an error.
	ValidatePermissionlessPoolCreationEnabled(ctx sdk.Context) error
}

type PoolIncentivesKeeperI interface {
	IsPoolIncentivized(ctx sdk.Context, poolId uint64) bool
}

type MultihopRoute interface {
	Length() int
	PoolIds() []uint64
	IntermediateDenoms() []string
}

type SwapAmountInRoutes []SwapAmountInRoute

func (routes SwapAmountInRoutes) Validate() error {
	if len(routes) == 0 {
		return ErrEmptyRoutes
	}

	for _, route := range routes {
		err := sdk.ValidateDenom(route.TokenOutDenom)
		if err != nil {
			return err
		}
	}

	return nil
}

func (routes SwapAmountInRoutes) IntermediateDenoms() []string {
	if len(routes) < 2 {
		return nil
	}
	intermediateDenoms := make([]string, 0, len(routes)-1)
	for _, route := range routes[:len(routes)-1] {
		intermediateDenoms = append(intermediateDenoms, route.TokenOutDenom)
	}

	return intermediateDenoms
}

func (routes SwapAmountInRoutes) PoolIds() []uint64 {
	poolIds := make([]uint64, 0, len(routes))
	for _, route := range routes {
		poolIds = append(poolIds, route.PoolId)
	}
	return poolIds
}

func (routes SwapAmountInRoutes) Length() int {
	return len(routes)
}

type SwapAmountOutRoutes []SwapAmountOutRoute

func (routes SwapAmountOutRoutes) Validate() error {
	if len(routes) == 0 {
		return ErrEmptyRoutes
	}

	for _, route := range routes {
		err := sdk.ValidateDenom(route.TokenInDenom)
		if err != nil {
			return err
		}
	}

	return nil
}

func (routes SwapAmountOutRoutes) IntermediateDenoms() []string {
	if len(routes) < 2 {
		return nil
	}
	intermediateDenoms := make([]string, 0, len(routes)-1)
	for _, route := range routes[1:] {
		intermediateDenoms = append(intermediateDenoms, route.TokenInDenom)
	}

	return intermediateDenoms
}

func (routes SwapAmountOutRoutes) PoolIds() []uint64 {
	poolIds := make([]uint64, 0, len(routes))
	for _, route := range routes {
		poolIds = append(poolIds, route.PoolId)
	}
	return poolIds
}

func (routes SwapAmountOutRoutes) Length() int {
	return len(routes)
}

// ValidateSwapAmountInSplitRoute validates a slice of SwapAmountInSplitRoute and returns an error if any of the following are true:
// - the slice is empty
// - any SwapAmountInRoute in the slice is invalid
// - the last TokenOutDenom of any SwapAmountInRoute in the slice does not match the TokenOutDenom of the previous SwapAmountInRoute in the slice
// - there are duplicate SwapAmountInRoutes in the slice
func ValidateSwapAmountInSplitRoute(splitRoutes []SwapAmountInSplitRoute) error {
	if len(splitRoutes) == 0 {
		return ErrEmptyRoutes
	}

	// validate every multihop path
	previousLastDenomOut := ""
	multihopRoutes := make([]SwapAmountInRoutes, 0, len(splitRoutes))
	for _, splitRoute := range splitRoutes {
		multihopRoute := splitRoute.Pools

		err := SwapAmountInRoutes(multihopRoute).Validate()
		if err != nil {
			return err
		}

		lastDenomOut := multihopRoute[len(multihopRoute)-1].TokenOutDenom

		if previousLastDenomOut != "" && lastDenomOut != previousLastDenomOut {
			return InvalidFinalTokenOutError{TokenOutGivenA: previousLastDenomOut, TokenOutGivenB: lastDenomOut}
		}

		previousLastDenomOut = lastDenomOut

		multihopRoutes = append(multihopRoutes, multihopRoute)
	}

	if osmoutils.ContainsDuplicateDeepEqual(multihopRoutes) {
		return ErrDuplicateRoutesNotAllowed
	}

	return nil
}

// ValidateSwapAmountOutSplitRoute validates a slice of SwapAmountOutSplitRoute and returns an error if any of the following are true:
// - the slice is empty
// - any SwapAmountOutRoute in the slice is invalid
// - the first TokenInDenom of any SwapAmountOutRoute in the slice does not match the TokenInDenom of the previous SwapAmountOutRoute in the slice
// - there are duplicate SwapAmountOutRoutes in the slice
func ValidateSwapAmountOutSplitRoute(splitRoutes []SwapAmountOutSplitRoute) error {
	if len(splitRoutes) == 0 {
		return ErrEmptyRoutes
	}

	// validate every multihop path
	previousFirstDenomIn := ""
	multihopRoutes := make([]SwapAmountOutRoutes, 0, len(splitRoutes))
	for _, splitRoute := range splitRoutes {
		multihopRoute := splitRoute.Pools

		err := SwapAmountOutRoutes(multihopRoute).Validate()
		if err != nil {
			return err
		}

		firstDenomIn := multihopRoute[0].TokenInDenom

		if previousFirstDenomIn != "" && firstDenomIn != previousFirstDenomIn {
			return InvalidFinalTokenOutError{TokenOutGivenA: previousFirstDenomIn, TokenOutGivenB: firstDenomIn}
		}

		previousFirstDenomIn = firstDenomIn

		multihopRoutes = append(multihopRoutes, multihopRoute)
	}

	if osmoutils.ContainsDuplicateDeepEqual(multihopRoutes) {
		return ErrDuplicateRoutesNotAllowed
	}

	return nil
}
