package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/osmosis-labs/osmosis/v12/app/params"
	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

// TODO: godoc
type SwapI interface {
	GetPool(ctx sdk.Context, poolId uint64) (gammtypes.PoolI, error)

	SwapExactAmountIn(
		ctx sdk.Context,
		sender sdk.AccAddress,
		poolId gammtypes.PoolI,
		tokenIn sdk.Coin,
		tokenOutDenom string,
		tokenOutMinAmount sdk.Int,
		swapFee sdk.Dec,
	) (sdk.Int, error)

	SwapExactAmountOut(
		ctx sdk.Context,
		sender sdk.AccAddress,
		poolId gammtypes.PoolI,
		tokenInDenom string,
		tokenInMaxAmount sdk.Int,
		tokenOut sdk.Coin,
		swapFee sdk.Dec,
	) (tokenInAmount sdk.Int, err error)
}

// SimulationExtension defines the swap simulation extension.
// TODO: refactor simulator setup logic to avoid having to define these
// extra methods just for the simulation.
type SimulationExtension interface {
	SwapI

	GetPoolAndPoke(ctx sdk.Context, poolId uint64) (gammtypes.TraditionalAmmInterface, error)

	GetNextPoolId(ctx sdk.Context) uint64
}

type SwapAmountInRoutes []SwapAmountInRoute

func (routes SwapAmountInRoutes) IsOsmoRoutedMultihop() bool {
	return len(routes) == 2 && (routes[0].TokenOutDenom == appparams.BaseCoinUnit)
}

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

type SwapAmountOutRoutes []SwapAmountOutRoute

func (routes SwapAmountOutRoutes) IsOsmoRoutedMultihop() bool {
	return len(routes) == 2 && (routes[1].TokenInDenom == appparams.BaseCoinUnit)
}

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
