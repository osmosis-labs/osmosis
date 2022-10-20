package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// TODO: godoc
type SwapI interface {
	MultihopSwapExactAmountIn(
		ctx sdk.Context,
		sender sdk.AccAddress,
		routes SwapAmountInRoute,
		tokenIn sdk.Coin,
		tokenOutMinAmount sdk.Int,
	) (tokenOutAmount sdk.Int, err error)

	MultihopSwapExactAmountOut(
		ctx sdk.Context,
		sender sdk.AccAddress,
		routes SwapAmountOutRoute,
		tokenInMaxAmount sdk.Int,
		tokenOut sdk.Coin,
	) (tokenInAmount sdk.Int, err error)
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
