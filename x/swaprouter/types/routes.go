package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/osmosis-labs/osmosis/v12/app/params"
	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

// TODO: godoc
type SwapI interface {
	MultihopSwapExactAmountIn(
		ctx sdk.Context,
		sender sdk.AccAddress,
		routes []SwapAmountInRoute,
		tokenIn sdk.Coin,
		tokenOutMinAmount sdk.Int,
	) (tokenOutAmount sdk.Int, err error)

	MultihopSwapExactAmountOut(
		ctx sdk.Context,
		sender sdk.AccAddress,
		routes []SwapAmountOutRoute,
		tokenInMaxAmount sdk.Int,
		tokenOut sdk.Coin,
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
